package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	appstate "gestion-commerciale/internal/app"
)

// BuildMainLayout construit le layout principal de l'application
func BuildMainLayout(w fyne.Window, a fyne.App) fyne.CanvasObject {
	// Zone de contenu centrale
	contentArea := container.NewStack()

	// Navigation vers une route
	navigateTo := func(route string, params ...interface{}) {
		var content fyne.CanvasObject

		switch route {
		case appstate.RouteDashboard:
			content = BuildDashboardScreen()
		case appstate.RouteArticles:
			content = BuildArticlesListScreen()
		case appstate.RouteCategories:
			content = BuildCategoriesScreen()
		case appstate.RouteBrands:
			content = BuildBrandsScreen()
		case appstate.RouteUnits:
			content = BuildUnitsScreen()
		case appstate.RouteInventory:
			content = BuildInventoryScreen()
		case appstate.RouteStockMovements:
			content = BuildStockMovementsScreen()
		case appstate.RouteWarehouses:
			content = BuildWarehousesScreen()
		case appstate.RoutePriceLists:
			content = BuildPriceListsScreen()
		case appstate.RouteSaleInvoices:
			content = BuildSaleInvoicesListScreen()
		case appstate.RouteSaleInvoice:
			var docID int
			if len(params) > 0 {
				if id, ok := params[0].(int); ok {
					docID = id
				}
			}
			content = BuildSaleInvoiceScreen(docID)
		case appstate.RouteQuotations:
			content = BuildQuotationsScreen()
		case appstate.RouteProforma:
			content = BuildProformaScreen()
		case appstate.RouteDeliveryNotes:
			content = BuildDeliveryNotesScreen()
		case appstate.RouteClientOrders:
			content = BuildClientOrdersScreen()
		case appstate.RouteCreditNotes:
			content = BuildCreditNotesScreen()
		case appstate.RoutePurchaseInvoices:
			content = BuildPurchaseInvoicesListScreen()
		case appstate.RoutePurchaseInvoice:
			content = BuildPurchaseInvoiceScreen(0)
		case appstate.RouteReceptionNotes:
			content = BuildReceptionNotesScreen()
		case appstate.RouteSupplierOrders:
			content = BuildSupplierOrdersScreen()
		case appstate.RoutePurchaseReturns:
			content = BuildPurchaseReturnsScreen()
		case appstate.RouteClients:
			content = BuildClientsScreen()
		case appstate.RouteSuppliers:
			content = BuildSuppliersScreen()
		case appstate.RouteDrivers:
			content = BuildDriversScreen()
		case appstate.RouteCash:
			content = BuildCashScreen()
		case appstate.RouteBank:
			content = BuildBankScreen()
		case appstate.RouteCheques:
			content = BuildChequesScreen()
		case appstate.RouteCollections:
			content = BuildCollectionsScreen()
		case appstate.RouteDisbursements:
			content = BuildDisbursementsScreen()
		case appstate.RouteAging:
			content = BuildAgingScreen()
		case appstate.RouteExpenses:
			content = BuildExpensesScreen()
		case appstate.RouteG50:
			content = BuildG50Screen()
		case appstate.RouteTVASalesRegister:
			content = BuildTVASalesRegisterScreen()
		case appstate.RouteTVAPurchaseRegister:
			content = BuildTVAPurchaseRegisterScreen()
		case appstate.RouteAnnualDeclaration:
			content = BuildAnnualDeclarationScreen()
		case appstate.RouteReports, appstate.RouteSalesReports:
			content = BuildSalesReportsScreen()
		case appstate.RoutePurchaseReports:
			content = BuildPurchaseReportsScreen()
		case appstate.RouteStockReports:
			content = BuildStockReportsScreen()
		case appstate.RouteTreasuryReports:
			content = BuildTreasuryReportsScreen()
		case appstate.RouteProfitReports:
			content = BuildProfitReportsScreen()
		case appstate.RouteDebtReports:
			content = BuildDebtReportsScreen()
		case appstate.RoutePOS:
			content = BuildPOSScreen(w)
		case appstate.RoutePrintCenter:
			content = BuildPrintCenterScreen()
		case appstate.RouteCompanyInfo:
			content = BuildCompanyInfoScreen()
		case appstate.RouteFiscalYears:
			content = BuildFiscalYearsScreen()
		case appstate.RouteNumbering:
			content = BuildNumberingScreen()
		case appstate.RouteTaxSettings:
			content = BuildTaxSettingsScreen()
		case appstate.RoutePrintSettings:
			content = BuildPrintSettingsScreen()
		case appstate.RouteUsers:
			content = BuildUsersScreen()
		case appstate.RouteCurrencies:
			content = BuildCurrenciesScreen()
		case appstate.RouteBarcodeConf:
			content = BuildBarcodeSettingsScreen()
		case appstate.RouteBackup:
			content = BuildBackupScreen()
		case appstate.RouteRestore:
			content = BuildRestoreScreen()
		case appstate.RouteCalculator:
			content = BuildCalculatorScreen()
		case appstate.RoutePriceUpdate:
			content = BuildPriceUpdateScreen()
		case appstate.RouteIndicators:
			content = BuildIndicatorsScreen()
		case appstate.RouteCalendar:
			content = BuildCalendarScreen()
		case appstate.RouteAuditLog:
			content = BuildAuditLogScreen()
		case appstate.RouteAbout:
			content = BuildAboutScreen()
		default:
			content = BuildDashboardScreen()
		}

		contentArea.Objects = []fyne.CanvasObject{content}
		contentArea.Refresh()
	}

	// Enregistrer la fonction de navigation globale
	appstate.GlobalNavigate = navigateTo

	// Construire la sidebar
	sidebar := BuildSidebar(navigateTo)

	// Construire la toolbar
	toolbar := BuildTopToolbar(w, navigateTo)

	// Construire la status bar
	statusBar := BuildStatusBar()

	// Diviser sidebar et contenu
	split := container.NewHSplit(sidebar, contentArea)
	split.SetOffset(0.20) // Sidebar 20% de la largeur

	// Layout principal vertical
	mainLayout := container.NewBorder(toolbar, statusBar, nil, nil, split)

	// Afficher le dashboard par défaut
	navigateTo(appstate.RouteDashboard)

	return mainLayout
}

// PlaceholderScreen retourne un écran temporaire avec un message
func PlaceholderScreen(title string) fyne.CanvasObject {
	label := widget.NewLabel("🚧 " + title)
	label.Alignment = fyne.TextAlignCenter
	label.TextStyle = fyne.TextStyle{Bold: true}
	msg := widget.NewLabel("Écran en cours de développement")
	msg.Alignment = fyne.TextAlignCenter
	return container.NewCenter(container.NewVBox(label, msg))
}

// Stubs restants (Lot 10 — Paramètres, Outils, Aide)
// NOTE: Articles (Lot 5)         → articles_screen.go
// NOTE: Ventes/Achats/POS (Lot 6)→ sales_screen.go
// NOTE: Clients/Frs/Chauf (Lot 7)→ tiers_screen.go
// NOTE: Trésorerie (Lot 8)       → treasury_screen.go
// NOTE: Rapports/Fiscalité (Lot 9)→ reports_screen.go
// NOTE: Paramètres/Outils/About (Lot 10) → settings_screen.go
// BuildPOSScreen        → sales_screen.go (Lot 6)
// BuildIndicatorsScreen → reports_screen.go (Lot 9)
// Toutes les fonctions Build* ci-dessous sont implémentées dans settings_screen.go
