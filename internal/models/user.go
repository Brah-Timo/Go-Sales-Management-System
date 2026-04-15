package models

import "time"

// User représente un utilisateur du système
type User struct {
	ID                  int       `db:"id"`
	Username            string    `db:"username"`
	FullName            string    `db:"full_name"`
	PasswordHash        string    `db:"password_hash"`
	Role                string    `db:"role"`
	PermissionsJSON     string    `db:"permissions_json"`
	IsActive            bool      `db:"is_active"`
	SecurityQuestion    string    `db:"security_question"`
	SecurityAnswerHash  string    `db:"security_answer_hash"`
	LastLogin           *time.Time `db:"last_login"`
	CreatedAt           time.Time `db:"created_at"`

	// Relations chargées dynamiquement
	Permissions map[string]bool `db:"-"`
}

// Permissions disponibles dans le système
const (
	PermCreateSaleInvoice      = "create_sale_invoice"
	PermCreatePurchaseInvoice  = "create_purchase_invoice"
	PermEditConfirmedInvoice   = "edit_confirmed_invoice"
	PermDeleteInvoice          = "delete_invoice"
	PermEditPrices             = "edit_prices"
	PermViewPurchasePrices     = "view_purchase_prices"
	PermViewProfitMargin       = "view_profit_margin"
	PermManageStock            = "manage_stock"
	PermManageClientSupplier   = "manage_clients_suppliers"
	PermAccessFinancialReports = "access_financial_reports"
	PermCollectPayments        = "collect_payments"
	PermManageSettings         = "manage_settings"
	PermBackupRestore          = "backup_restore"
	PermApplyDiscountAbove10   = "apply_discount_above_10"
	PermInventory              = "inventory"
)

// DefaultPermissionsByRole retourne les permissions par défaut selon le rôle
func DefaultPermissionsByRole(role string) map[string]bool {
	switch role {
	case RoleAdmin:
		return map[string]bool{
			PermCreateSaleInvoice:      true,
			PermCreatePurchaseInvoice:  true,
			PermEditConfirmedInvoice:   true,
			PermDeleteInvoice:          true,
			PermEditPrices:             true,
			PermViewPurchasePrices:     true,
			PermViewProfitMargin:       true,
			PermManageStock:            true,
			PermManageClientSupplier:   true,
			PermAccessFinancialReports: true,
			PermCollectPayments:        true,
			PermManageSettings:         true,
			PermBackupRestore:          true,
			PermApplyDiscountAbove10:   true,
			PermInventory:              true,
		}
	case RoleSeller:
		return map[string]bool{
			PermCreateSaleInvoice:    true,
			PermManageStock:          true,
			PermManageClientSupplier: true,
			PermCollectPayments:      true,
			PermInventory:            true,
		}
	case RoleCashier:
		return map[string]bool{
			PermCreateSaleInvoice: true,
			PermCollectPayments:   true,
		}
	case RoleAssistant:
		return map[string]bool{
			PermInventory: true,
		}
	}
	return map[string]bool{}
}

// Session représente la session utilisateur active
type Session struct {
	UserID     int
	Username   string
	FullName   string
	Role       string
	FiscalYear int
	CompanyID  int
	CompanyName string
	Permissions map[string]bool
	DBPath     string
}

// HasPermission vérifie si la session a une permission
func (s *Session) HasPermission(perm string) bool {
	if s == nil || s.Permissions == nil {
		return false
	}
	return s.Permissions[perm]
}

// AuditLog enregistre les actions des utilisateurs
type AuditLog struct {
	ID          int    `db:"id"`
	Timestamp   string `db:"timestamp"`
	UserID      int    `db:"user_id"`
	UserName    string `db:"user_name"`
	ActionType  string `db:"action_type"`
	Module      string `db:"module"`
	Description string `db:"description"`
	OldData     string `db:"old_data"`
	NewData     string `db:"new_data"`
}
