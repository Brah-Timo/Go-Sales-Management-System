package app

import "fyne.io/fyne/v2"

// NavigationFunc est le type de la fonction de navigation principale
type NavigationFunc func(route string, params ...interface{})

// GlobalNavigate est la fonction de navigation enregistrée par le layout principal
var GlobalNavigate NavigationFunc

// Navigate navigue vers un écran
func Navigate(route string, params ...interface{}) {
	if GlobalNavigate != nil {
		GlobalNavigate(route, params...)
	}
}

// Routes définit les routes de l'application
const (
	RouteLogin         = "login"
	RouteDashboard     = "dashboard"

	// Paramètres
	RouteCompanyInfo   = "settings/company"
	RouteFiscalYears   = "settings/fiscal_years"
	RouteNumbering     = "settings/numbering"
	RouteTaxSettings   = "settings/taxes"
	RoutePrintSettings = "settings/print"
	RouteUsers         = "settings/users"
	RouteCurrencies    = "settings/currencies"
	RouteBarcodeConf   = "settings/barcode"

	// Articles / Stock
	RouteArticles       = "articles"
	RouteArticleNew     = "articles/new"
	RouteArticleEdit    = "articles/edit"
	RouteCategories     = "articles/categories"
	RouteBrands         = "articles/brands"
	RouteUnits          = "articles/units"
	RouteInventory      = "articles/inventory"
	RouteStockMovements = "articles/stock_movements"
	RouteWarehouses     = "articles/warehouses"
	RoutePriceLists     = "articles/price_lists"

	// Ventes
	RouteSaleInvoice    = "sales/invoice"
	RouteSaleInvoices   = "sales/invoices"
	RouteQuotations     = "sales/quotations"
	RouteProforma       = "sales/proforma"
	RouteDeliveryNotes  = "sales/delivery_notes"
	RouteClientOrders   = "sales/client_orders"
	RouteCreditNotes    = "sales/credit_notes"

	// Achats
	RoutePurchaseInvoice  = "purchases/invoice"
	RoutePurchaseInvoices = "purchases/invoices"
	RouteReceptionNotes   = "purchases/reception_notes"
	RouteSupplierOrders   = "purchases/supplier_orders"
	RoutePurchaseReturns  = "purchases/returns"

	// Tiers
	RouteClients    = "tiers/clients"
	RouteSuppliers  = "tiers/suppliers"
	RouteDrivers    = "tiers/drivers"

	// Trésorerie
	RouteCash          = "treasury/cash"
	RouteBank          = "treasury/bank"
	RouteCheques       = "treasury/cheques"
	RouteCollections   = "treasury/collections"
	RouteDisbursements = "treasury/disbursements"
	RouteAging         = "treasury/aging"
	RouteExpenses      = "treasury/expenses"

	// Déclarations fiscales
	RouteG50                 = "tax/g50"
	RouteTVASalesRegister    = "tax/tva_sales"
	RouteTVAPurchaseRegister = "tax/tva_purchases"
	RouteAnnualDeclaration   = "tax/annual"

	// Rapports
	RouteReports          = "reports"
	RouteSalesReports     = "reports/sales"
	RoutePurchaseReports  = "reports/purchases"
	RouteStockReports     = "reports/stock"
	RouteTreasuryReports  = "reports/treasury"
	RouteProfitReports    = "reports/profits"
	RouteDebtReports      = "reports/debts"

	// POS
	RoutePOS = "pos"

	// Centre d'impression
	RoutePrintCenter = "print_center"

	// Outils
	RouteBackup       = "tools/backup"
	RouteRestore      = "tools/restore"
	RouteImport       = "tools/import"
	RouteExport       = "tools/export"
	RouteCalculator   = "tools/calculator"
	RoutePriceUpdate  = "tools/price_update"
	RouteIndicators   = "tools/indicators"
	RouteCalendar     = "tools/calendar"
	RouteAuditLog     = "tools/audit_log"

	// Aide
	RouteHelp    = "help"
	RouteAbout   = "help/about"
)

// Window est la fenêtre principale de l'application
var MainWindow fyne.Window
