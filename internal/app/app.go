package app

import "gestion-commerciale/internal/models"

// LoginResult contient le résultat d'une connexion réussie
type LoginResult struct {
	Session *models.Session
	DBPath  string
}

// NotificationType type d'une notification
type NotificationType int

const (
	NotifInfo NotificationType = iota
	NotifWarning
	NotifError
	NotifSuccess
)

// Notification représente une notification système
type Notification struct {
	Type    NotificationType
	Title   string
	Message string
}

// NotificationChannel canal global pour les notifications
var NotificationChannel = make(chan Notification, 100)

// SendNotification envoie une notification
func SendNotification(t NotificationType, title, message string) {
	select {
	case NotificationChannel <- Notification{t, title, message}:
	default:
		// Ne pas bloquer si le canal est plein
	}
}

// RefreshChannel canal pour déclencher le rafraîchissement de l'UI
var RefreshChannel = make(chan string, 100)

// TriggerRefresh demande un rafraîchissement d'une vue
func TriggerRefresh(viewName string) {
	select {
	case RefreshChannel <- viewName:
	default:
	}
}
