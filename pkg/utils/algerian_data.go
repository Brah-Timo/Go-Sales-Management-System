package utils

import "fmt"

// Wilaya représente une wilaya algérienne
type Wilaya struct {
	Code int
	Name string
}

// Wilayas58 liste les 58 wilayas algériennes
var Wilayas58 = []Wilaya{
	{1, "Adrar"}, {2, "Chlef"}, {3, "Laghouat"}, {4, "Oum El Bouaghi"},
	{5, "Batna"}, {6, "Béjaïa"}, {7, "Biskra"}, {8, "Béchar"},
	{9, "Blida"}, {10, "Bouira"}, {11, "Tamanrasset"}, {12, "Tébessa"},
	{13, "Tlemcen"}, {14, "Tiaret"}, {15, "Tizi Ouzou"}, {16, "Alger"},
	{17, "Djelfa"}, {18, "Jijel"}, {19, "Sétif"}, {20, "Saïda"},
	{21, "Skikda"}, {22, "Sidi Bel Abbès"}, {23, "Annaba"}, {24, "Guelma"},
	{25, "Constantine"}, {26, "Médéa"}, {27, "Mostaganem"}, {28, "M'Sila"},
	{29, "Mascara"}, {30, "Ouargla"}, {31, "Oran"}, {32, "El Bayadh"},
	{33, "Illizi"}, {34, "Bordj Bou Arréridj"}, {35, "Boumerdès"}, {36, "El Tarf"},
	{37, "Tindouf"}, {38, "Tissemsilt"}, {39, "El Oued"}, {40, "Khenchela"},
	{41, "Souk Ahras"}, {42, "Tipaza"}, {43, "Mila"}, {44, "Aïn Defla"},
	{45, "Naâma"}, {46, "Aïn Témouchent"}, {47, "Ghardaïa"}, {48, "Relizane"},
	{49, "El M'Ghair"}, {50, "El Meniaa"}, {51, "Ouled Djellal"}, {52, "Bordj Baji Mokhtar"},
	{53, "Béni Abbès"}, {54, "Timimoun"}, {55, "Touggourt"}, {56, "Djanet"},
	{57, "In Salah"}, {58, "In Guezzam"},
}

// WilayaNames retourne uniquement les noms des wilayas pour les listes déroulantes
func WilayaNames() []string {
	names := make([]string, len(Wilayas58))
	for i, w := range Wilayas58 {
		names[i] = w.Name
	}
	return names
}

// WilayaNamesWithCode retourne les noms avec code (ex: "16 - Alger")
func WilayaNamesWithCode() []string {
	names := make([]string, len(Wilayas58))
	for i, w := range Wilayas58 {
		names[i] = fmt.Sprintf("%02d - %s", w.Code, w.Name)
	}
	return names
}

// ActivityTypes liste les types d'activités commerciales
var ActivityTypes = []string{
	"Commerce Général",
	"Commerce de Gros",
	"Commerce de Détail",
	"Services",
	"Production / Fabrication",
	"Import / Export",
	"Restauration",
	"Transport",
	"BTP / Construction",
	"Artisanat",
	"Autre",
}

// TaxRegimes liste les régimes fiscaux algériens
var TaxRegimes = []string{
	"Régime Réel",
	"Régime Forfaitaire (IFU)",
	"Impôt Forfaitaire Unique (IFU)",
}

// TVARatesStr liste des taux de TVA algériens (sous forme de chaîne)
var TVARatesStr = []string{"0", "9", "19"}

// PaymentTermsOptions liste les conditions de paiement
var PaymentTermsOptions = []string{
	"Immédiat",
	"30 jours",
	"45 jours",
	"60 jours",
	"90 jours",
}

// ChequeStatusLabels libellés des statuts de chèques
var ChequeStatusLabels = map[string]string{
	"pending":   "En attente",
	"deposited": "Déposé",
	"collected": "Encaissé",
	"rejected":  "Rejeté",
	"returned":  "Retourné",
	"replaced":  "Remplacé",
}

// CashCategories catégories de mouvements de caisse
var CashCategories = []string{
	"Encaissement client",
	"Paiement fournisseur",
	"Dépense diverse",
	"Versement banque",
	"Retrait banque",
	"Autre",
}

// BankMovementTypes types de mouvements bancaires
var BankMovementTypes = []string{
	"Dépôt",
	"Retrait",
	"Virement reçu",
	"Virement émis",
	"Chèque émis",
	"Chèque reçu",
	"Frais bancaires",
	"Autre",
}

// RoleLabels libellés français des rôles
var RoleLabels = map[string]string{
	"admin":     "Administrateur",
	"seller":    "Vendeur",
	"cashier":   "Caissier",
	"assistant": "Assistant",
}

// PermissionLabels libellés des permissions
var PermissionLabels = map[string]string{
	"create_sale_invoice":      "Créer facture de vente",
	"create_purchase_invoice":  "Créer facture d'achat",
	"edit_confirmed_invoice":   "Modifier facture confirmée",
	"delete_invoice":           "Supprimer facture",
	"edit_prices":              "Modifier les prix",
	"view_purchase_prices":     "Voir les prix d'achat",
	"view_profit_margin":       "Voir la marge bénéficiaire",
	"manage_stock":             "Gérer le stock",
	"manage_clients_suppliers": "Gérer clients/fournisseurs",
	"access_financial_reports": "Accéder aux rapports financiers",
	"collect_payments":         "Encaisser/Payer",
	"manage_settings":          "Gérer les paramètres",
	"backup_restore":           "Sauvegarde/Restauration",
	"apply_discount_above_10":  "Remise > 10%",
	"inventory":                "Inventaire",
}
