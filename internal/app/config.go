package app

import (
	"gestion-commerciale/internal/models"
	"gestion-commerciale/internal/database"
	"strconv"
)

// AppConfig contient la configuration globale de l'application
type AppConfig struct {
	AppName    string
	AppVersion string
	DataDir    string
	AssetsDir  string
	DocsDir    string
	TaxConfig  models.TaxConfig
}

// DefaultConfig retourne la configuration par défaut
func DefaultConfig() *AppConfig {
	return &AppConfig{
		AppName:    "Gestion Commerciale Pro",
		AppVersion: "1.0.0",
		DataDir:    "data",
		AssetsDir:  "assets",
		DocsDir:    "docs",
	}
}

// LoadTaxConfig charge la configuration fiscale depuis la base de données
func LoadTaxConfig() models.TaxConfig {
	db := GetDB()
	if db == nil {
		return models.TaxConfig{
			TVANormal: 19, TVAReduced: 9, TimbreRate: 1,
			TimbreMax: 2500, TAPRate: 1, IsTVASubject: true,
			AutoTimbre: true,
		}
	}

	tc := models.TaxConfig{}
	tc.TVANormal, _ = strconv.ParseFloat(database.GetSetting(db, "tva_normal"), 64)
	tc.TVAReduced, _ = strconv.ParseFloat(database.GetSetting(db, "tva_reduced"), 64)
	tc.TimbreRate, _ = strconv.ParseFloat(database.GetSetting(db, "timbre_rate"), 64)
	tc.TimbreMax, _ = strconv.ParseFloat(database.GetSetting(db, "timbre_max"), 64)
	tc.TimbreExemption, _ = strconv.ParseFloat(database.GetSetting(db, "timbre_exemption"), 64)
	tc.TAPRate, _ = strconv.ParseFloat(database.GetSetting(db, "tap_rate"), 64)
	tc.IsTVASubject = database.GetSetting(db, "is_tva_subject") == "1"
	tc.AutoTimbre = database.GetSetting(db, "auto_timbre") == "1"
	tc.AutoUpdateSalePrice = database.GetSetting(db, "auto_update_sale_price") == "1"
	tc.TaxRegime = database.GetSetting(db, "tax_regime")

	// Valeurs par défaut si manquantes
	if tc.TVANormal == 0 { tc.TVANormal = 19 }
	if tc.TimbreRate == 0 { tc.TimbreRate = 1 }
	if tc.TimbreMax == 0 { tc.TimbreMax = 2500 }
	if tc.TAPRate == 0 { tc.TAPRate = 1 }

	return tc
}

// GlobalTaxConfig contient la configuration fiscale active
var GlobalTaxConfig models.TaxConfig

// RefreshTaxConfig recharge la configuration fiscale
func RefreshTaxConfig() {
	GlobalTaxConfig = LoadTaxConfig()
}
