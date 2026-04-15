package models

// ─────────────────────────────────────────────────────────────
// Types de documents
// ─────────────────────────────────────────────────────────────

const (
	DocTypeFA  = "FA"  // Facture de Vente
	DocTypeFAC = "FAC" // Facture d'Achat
	DocTypeBL  = "BL"  // Bon de Livraison
	DocTypeBR  = "BR"  // Bon de Réception
	DocTypeDV  = "DV"  // Devis
	DocTypePF  = "PF"  // Proforma
	DocTypeBCC = "BCC" // Bon de Commande Client
	DocTypeBCF = "BCF" // Bon de Commande Fournisseur
	DocTypeAV  = "AV"  // Avoir (Note de Crédit)
	DocTypeBP  = "BP"  // Bon de Paiement
	DocTypeBRE = "BRE" // Bon de Réception Espèces
)

// ─────────────────────────────────────────────────────────────
// Statuts des documents
// ─────────────────────────────────────────────────────────────

const (
	StatusDraft     = "draft"
	StatusConfirmed = "confirmed"
	StatusPaid      = "paid"
	StatusPartial   = "partial"
	StatusCancelled = "cancelled"
)

// ─────────────────────────────────────────────────────────────
// Modes de paiement
// ─────────────────────────────────────────────────────────────

const (
	PaymentCash     = "cash"
	PaymentCheque   = "cheque"
	PaymentTransfer = "transfer"
	PaymentCredit   = "credit"
	PaymentMixed    = "mixed"
)

// ─────────────────────────────────────────────────────────────
// Types de mouvements de stock
// ─────────────────────────────────────────────────────────────

const (
	StockMovePurchaseIn   = "purchase_in"
	StockMoveSaleOut      = "sale_out"
	StockMoveTransferIn   = "transfer_in"
	StockMoveTransferOut  = "transfer_out"
	StockMoveReturnIn     = "return_in"
	StockMoveReturnOut    = "return_out"
	StockMoveAdjustIn     = "adjustment_in"
	StockMoveAdjustOut    = "adjustment_out"
	StockMoveDamage       = "damage"
)

// ─────────────────────────────────────────────────────────────
// Statuts des chèques
// ─────────────────────────────────────────────────────────────

const (
	ChequeStatusPending   = "pending"
	ChequeStatusDeposited = "deposited"
	ChequeStatusCollected = "collected"
	ChequeStatusRejected  = "rejected"
	ChequeStatusReturned  = "returned"
	ChequeStatusReplaced  = "replaced"
)

// ─────────────────────────────────────────────────────────────
// Rôles utilisateurs
// ─────────────────────────────────────────────────────────────

const (
	RoleAdmin     = "admin"
	RoleSeller    = "seller"
	RoleCashier   = "cashier"
	RoleAssistant = "assistant"
)

// ─────────────────────────────────────────────────────────────
// Méthodes de valorisation du stock
// ─────────────────────────────────────────────────────────────

const (
	ValuationCMUP = "CMUP"
	ValuationFIFO = "FIFO"
)

// ─────────────────────────────────────────────────────────────
// Types de paiement (encaissement / décaissement)
// ─────────────────────────────────────────────────────────────

const (
	PaymentTypeCollection   = "collection"
	PaymentTypeDisbursement = "disbursement"
)

// ─────────────────────────────────────────────────────────────
// Taux de TVA
// ─────────────────────────────────────────────────────────────

const (
	TVANormal  = 19.0
	TVAReduit  = 9.0
	TVAExempt  = 0.0
)

// ─────────────────────────────────────────────────────────────
// Libellés français des types de documents
// ─────────────────────────────────────────────────────────────

var DocTypeLabels = map[string]string{
	DocTypeFA:  "Facture de Vente",
	DocTypeFAC: "Facture d'Achat",
	DocTypeBL:  "Bon de Livraison",
	DocTypeBR:  "Bon de Réception",
	DocTypeDV:  "Devis",
	DocTypePF:  "Facture Proforma",
	DocTypeBCC: "Bon de Commande Client",
	DocTypeBCF: "Bon de Commande Fournisseur",
	DocTypeAV:  "Avoir",
	DocTypeBP:  "Bon de Paiement",
	DocTypeBRE: "Bon de Réception Espèces",
}

var StatusLabels = map[string]string{
	StatusDraft:     "Brouillon",
	StatusConfirmed: "Confirmé",
	StatusPaid:      "Payé",
	StatusPartial:   "Partiel",
	StatusCancelled: "Annulé",
}

var PaymentLabels = map[string]string{
	PaymentCash:     "Espèces",
	PaymentCheque:   "Chèque",
	PaymentTransfer: "Virement",
	PaymentCredit:   "Crédit",
	PaymentMixed:    "Mixte",
}
