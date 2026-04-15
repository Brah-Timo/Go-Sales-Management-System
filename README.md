# Gestion Commerciale Pro — نظام تسيير تجاري جزائري

Système de gestion commerciale complet pour les **PME/TPE algériennes**.  
Développé en **Go** avec interface graphique **Fyne v2** et base de données **SQLite** locale, embarqué et sans installation de serveur.

---

## Fonctionnalités implémentées (Lots 1–10) — COMPLET ✅

| Module | Fonctionnalités | Statut |
|--------|-----------------|--------|
| **Authentification** | Connexion multi-société, sélection année fiscale, verrouillage 3 tentatives | ✅ |
| **Dashboard** | 7 KPIs colorés (ventes, achats, caisse, bénéfice, créances, dettes, ruptures), raccourcis | ✅ |
| **Articles & Stock** | CRUD articles (3 onglets), codes-barres, catégories, marques, unités, dépôts | ✅ |
| **Inventaire** | Création inventaire, saisie ligne par ligne, scan codes-barres, confirmation stock | ✅ |
| **Mouvements Stock** | Historique filtrable par type/dépôt, ajustement manuel, CMUP automatique | ✅ |
| **Listes de Prix** | CRUD listes de prix personnalisées par article et client | ✅ |
| **Factures Vente (FA)** | Saisie complète, calcul TVA/remise/timbre en temps réel, confirmation stock | ✅ |
| **Devis (DV)** | Création devis avec conversion automatique en FA/PF/BL | ✅ |
| **Proforma (PF)** | Facture proforma convertible en FA ou BL | ✅ |
| **Bons de Livraison (BL)** | Édition BL avec affectation chauffeur | ✅ |
| **Commandes Clients (CC)** | Commandes clients avec conversion | ✅ |
| **Avoirs Clients (AV)** | Avoirs sur ventes avec retour stock | ✅ |
| **Factures Achat (FAC)** | Saisie FAC fournisseurs, N° facture fournisseur, mise à jour stock/CMUP | ✅ |
| **Bons de Réception (BR)** | Réception marchandises fournisseurs | ✅ |
| **Commandes Fournisseurs (BCF)** | Commandes aux fournisseurs | ✅ |
| **Retours Fournisseurs (BCC)** | Retours avec ajustement stock | ✅ |
| **POS** | Point de Vente rapide, scan code-barres, panier, rendu monnaie, facture auto | ✅ |
| **Clients** | CRUD complet (NIF/NIS/RC/AI), 58 wilayas, blocage, relevé de compte filtré | ✅ |
| **Fournisseurs** | CRUD complet + notation livraison/qualité/prix (étoiles), NIF/NIS/RC/AI | ✅ |
| **Chauffeurs** | Gestion livreurs, immatriculation, compteur livraisons | ✅ |
| **Caisse** | Journal de caisse, entrées/sorties, solde courant, filtrage période | ✅ |
| **Banque** | Comptes bancaires multiples, relevé mouvements, rapprochement | ✅ |
| **Chèques** | Chèques reçus/émis, suivi échéances, alertes 7 jours, changement statut | ✅ |
| **Encaissements** | Paiements clients avec ventilation FIFO automatique sur factures | ✅ |
| **Décaissements** | Paiements fournisseurs avec mise à jour solde | ✅ |
| **Balance Âgée** | Créances clients et dettes fournisseurs par tranche (0-30j, 31-60j, 61-90j, >90j) | ✅ |
| **Dépenses** | Dépenses diverses par catégorie, suivi mensuel | ✅ |
| **Rapport Ventes** | Liste factures filtrée (période/client/statut/mode), totaux HT/TVA/TTC/timbre | ✅ |
| **Rapport Achats** | Liste FAC filtrée (période/fournisseur/statut), totaux | ✅ |
| **Rapport Stock** | Valorisation stock, articles bas/rupture, top produits mois | ✅ |
| **Rapport Trésorerie** | Flux de caisse filtrés, totaux entrées/sorties | ✅ |
| **Rapport Rentabilité** | CA, CMUP, bénéfice brut/net, marges, top clients | ✅ |
| **Rapport Créances/Dettes** | Soldes clients/fournisseurs non soldés | ✅ |
| **Indicateurs** | 6 KPIs annuels: rotation stock, DSO, marge brute/nette, valeur facture | ✅ |
| **Déclaration G50** | G50 mensuel calculé: TVA collectée/déductible, TAP, timbre + historique | ✅ |
| **Registre TVA Ventes** | Registre légal mensuel avec NIF client | ✅ |
| **Registre TVA Achats** | Registre légal mensuel avec NIF fournisseur | ✅ |
| **Déclaration Annuelle** | Synthèse annuelle G50-bis: CA, TVA nette, TAP, bénéfice | ✅ |
| **Services PDF** | Génération factures PDF avec en-tête société (gopdf) | ✅ |
| **Export Excel** | Export données en XLSX (excelize) | ✅ |
| **Backup / Restore** | Sauvegarde et restauration des bases de données SQLite | ✅ |
| **Informations Société** | NIF/NIS/RC/AI, coordonnées, RIB, logo, pied de page | ✅ |
| **Années Fiscales** | CRUD exercices, activation, compteurs FA | ✅ |
| **Numérotation** | Préfixes + compteurs par type document, RAZ annuel, aperçu | ✅ |
| **Paramètres Fiscaux** | TVA 0/9/19%, TAP, timbre fiscal, régime d'imposition | ✅ |
| **Paramètres Impression** | Format papier, marges, logo, filigrane, répertoire PDF | ✅ |
| **Gestion Utilisateurs** | CRUD utilisateurs, rôles, permissions granulaires 15 clés | ✅ |
| **Devises** | Gestion multi-devises avec taux de change DZD | ✅ |
| **Code-barres** | Type EAN/QR, préfixe GS1, mode scanner, génération auto | ✅ |
| **Sauvegarde** | Backup ZIP (DB + assets), planification quotidienne | ✅ |
| **Restauration** | Restauration depuis ZIP avec confirmation de sécurité | ✅ |
| **Calculatrice** | Calculatrice avancée, historique, raccourcis TVA/timbre | ✅ |
| **Mise à Jour Prix** | Hausse/baisse % ou montant fixe par catégorie, arrondi | ✅ |
| **Calendrier & Rappels** | Rappels commerciaux, alertes chèques à encaisser 7j | ✅ |
| **Journal d'Audit** | Historique complet des actions utilisateurs, export CSV | ✅ |
| **Centre d'Impression** | Réimpression documents filtrés (type/date/tiers) | ✅ |
| **À Propos** | Version, stack technique, session courante | ✅ |

---

## Technologies

| Composant | Technologie | Version |
|-----------|-------------|---------|
| Langage | Go | 1.19+ |
| Interface graphique | Fyne | v2.4 |
| Base de données | SQLite 3 | embarqué (mattn/go-sqlite3) |
| PDF | gopdf | v0.22 |
| Excel | excelize | v2.8 |
| Chiffrement | bcrypt | — |

---

## Structure du projet

```
gestion-commerciale/
├── cmd/app/                    → Point d'entrée (main.go)
├── internal/
│   ├── app/                    → Session, routes, état global, navigation
│   ├── database/
│   │   ├── schema.go           → Schéma SQLite (CREATE TABLE + seed data)
│   │   └── queries/            → Requêtes SQL (articles, documents, clients, rapports…)
│   ├── models/                 → Structures de données (Article, Document, Client,
│   │                             Payment, CashMovement, BankAccount, Cheque,
│   │                             Expense, G50Data, TaxDeclaration…)
│   └── services/               → Logique métier
│       ├── article_service.go  → Gestion articles et stock
│       ├── auth_service.go     → Authentification + bcrypt
│       ├── backup_service.go   → Sauvegarde/restauration DB
│       ├── cash_service.go     → Caisse, banque, chèques, dépenses, balance âgée
│       ├── client_service.go   → CRUD clients, fournisseurs, chauffeurs
│       ├── document_service.go → Documents commerciaux + conversions
│       ├── inventory_service.go→ Inventaires physiques
│       ├── numbering_service.go→ Auto-numérotation documents
│       ├── payment_service.go  → Encaissements/décaissements + ventilation FIFO
│       ├── report_service.go   → Rapports: ventes, stock, rentabilité, indicateurs
│       ├── stock_service.go    → Mouvements stock + CMUP
│       └── tax_service.go      → G50, TVA, TAP, registres fiscaux
├── ui/
│   ├── layout/                 → Sidebar (structure de navigation)
│   └── screens/
│       ├── login_screen.go          → Connexion multi-société + verrouillage
│       ├── dashboard_screen.go      → Tableau de bord KPIs + raccourcis
│       ├── main_layout.go           → Layout principal + routage 50+ écrans
│       ├── sidebar_screen.go        → Barre de navigation latérale
│       ├── toolbar_statusbar.go     → Barre outils session + horloge
│       ├── articles_screen.go       → Articles, stock, inventaire (Lot 5) ~1700 lignes
│       ├── sales_screen.go          → Ventes/Achats/POS (Lot 6) ~1600 lignes
│       ├── tiers_screen.go          → Clients/Fournisseurs/Chauffeurs (Lot 7) ~880 lignes
│       ├── treasury_screen.go       → Trésorerie (Lot 8) ~1361 lignes
│       ├── reports_screen.go        → Rapports & Fiscalité (Lot 9) ~1240 lignes
│       └── settings_screen.go      → Paramètres/Utilisateurs/Outils (Lot 10) ~2060 lignes
├── pkg/utils/                  → Utilitaires (formatage, calculs TVA, wilayas 58, montant-en-lettres)
├── assets/                     → Ressources (polices, icônes)
├── data/                       → Bases de données SQLite (.db, un fichier par société/année)
└── scripts/                    → Scripts de compilation Linux/Windows
```

---

## Installation et compilation

### Prérequis système

```bash
# Ubuntu / Debian
sudo apt-get install -y golang-go gcc pkg-config \
    libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev \
    libgl1-mesa-dev xorg-dev libxxf86vm-dev

# Fedora / RHEL
sudo dnf install golang gcc pkg-config libX11-devel libXcursor-devel \
    libXrandr-devel libXinerama-devel mesa-libGL-devel

# Windows (via MSYS2/MinGW)
# Installer Go 1.19+ depuis https://golang.org/dl/
# Installer MinGW64 et gcc
```

### Compilation

```bash
# Cloner le projet
git clone https://github.com/GPT4-AI/go-gestion-commerciale.git
cd go-gestion-commerciale

# Télécharger les dépendances
go mod download

# Compilation Linux
go build -o gestion_commerciale ./cmd/app/
./gestion_commerciale

# Compilation Windows (cross depuis Linux)
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags="-H windowsgui" -o gestion_commerciale.exe ./cmd/app/

# Ou utiliser le script fourni
chmod +x scripts/build_linux.sh && ./scripts/build_linux.sh
```

---

## Connexion par défaut

| Champ | Valeur |
|-------|--------|
| Utilisateur | `admin` |
| Mot de passe | `admin123` |
| Base de données | Fichier `.db` dans le dossier `data/` |

> **Note** : Au premier lancement, aucune base de données n'existe.  
> Cliquer **"Nouvelle Société"** sur l'écran de connexion pour créer la première société.

---

## Guide d'utilisation

### 1. Créer une société

1. Lancer l'application (`./gestion_commerciale`)
2. Écran de connexion → cliquer **"Nouvelle Société"**
3. Renseigner : nom de la société, année fiscale
4. Se connecter avec `admin` / `admin123`

---

### 2. Configurer les articles

1. **Articles & Stock → Articles** → bouton **Nouvel Article**
2. Onglet **Général** : Nom, Référence (auto-générée), Code-barres, Catégorie
3. Onglet **Prix & TVA** : Prix vente HT/TTC, TVA (0/9/19 %), marge auto-calculée
4. Onglet **Stock** : Stock initial, stock min (alerte rupture), dépôt
5. **Catégories / Marques / Unités** : Gérer depuis les sous-menus dédiés

---

### 3. Gérer les clients et fournisseurs

**Clients** (`Tiers → Clients`) :
- **Nouveau Client** → 3 onglets : Général (coordonnées), Fiscal (NIF/NIS/RC/AI), Conditions (crédit/remise/délai)
- **Bloquer/Débloquer** : empêche la création de nouvelles factures pour ce client
- **Relevé de Compte** : sélectionner une période → tableau débit/crédit/solde

**Fournisseurs** (`Tiers → Fournisseurs`) :
- Mêmes champs fiscaux algériens (NIF/NIS/RC/AI)
- Onglet **Évaluation** : noter la qualité livraison (1–5 étoiles)

---

### 4. Créer une facture de vente (FA)

1. **Ventes → Factures de Vente** → bouton **Nouveau**
2. Sélectionner le client, le mode de paiement, la remise globale %
3. **Ajouter lignes** : article, quantité, prix, remise ligne, TVA
4. **Totaux** mis à jour automatiquement : HT → TVA → TTC → Timbre → **Net à Payer**
5. **Sauvegarder** (brouillon) ou **Confirmer** (stock mis à jour)
6. **Imprimer PDF** / **Convertir** vers BL, Avoir…

---

### 5. Encaisser un client

1. **Trésorerie → Encaissements Clients**
2. Sélectionner le client (son solde actuel est affiché)
3. Saisir le montant, la date, le mode (espèces/chèque/virement)
4. **Enregistrer** → ventilation FIFO automatique sur les factures impayées
5. Si espèces : mouvement de caisse créé automatiquement
6. Si chèque : chèque reçu créé avec statut "En attente"

---

### 6. Gérer la caisse

1. **Trésorerie → Caisse**
2. Sélectionner la période (boutons Aujourd'hui / Ce mois / Cette année)
3. **Journal** : toutes les entrées/sorties avec solde cumulé
4. **Nouveau Mouvement** : saisir manuellement une entrée ou sortie de caisse

---

### 7. Gestion des chèques

1. **Trésorerie → Chèques**
2. Filtrer par type (Reçus / Émis) et statut (En attente / Déposé / Encaissé / Rejeté)
3. ⚠️ **Alertes** : les chèques à échéance dans 7 jours sont signalés en haut
4. **Changer statut** : déposer en banque (mouvement bancaire créé), encaisser, rejeter
5. **Nouveau Chèque** : saisir manuellement un chèque reçu ou émis

---

### 8. Générer la déclaration G50

1. **Fiscalité → Déclaration G50**
2. Sélectionner l'année et le mois
3. **Calculer G50** → le formulaire se remplit automatiquement :
   - CA par taux TVA (19%, 9%, exonéré)
   - TVA collectée, TVA déductible sur achats
   - TVA nette due (ou crédit reporté)
   - TAP, Timbre fiscal
   - **Total à payer**
4. **Sauvegarder** (brouillon) ou **Valider** (final)
5. Historique des G50 de l'année visible en panneau latéral

---

### 9. Rapports de ventes

1. **Rapports → Ventes**
2. Filtrer par période, client, statut, mode de paiement
3. Tableau complet avec : N° facture, date, client, HT, TVA, TTC, timbre, payé, reste
4. Résumé : nombre de factures + totaux HT/TVA/TTC/timbre
5. **Exporter Excel** / **Imprimer PDF**

---

### 10. Point de Vente (POS)

1. **POS** depuis la barre latérale ou le bouton 🛒 de la toolbar
2. Scanner le code-barres (Entrée) ou **F2 Rechercher** (sélection article + quantité)
3. Modifier quantités dans le panier — supprimer avec ✕
4. **F8 Payer** → Saisir montant remis → Rendu monnaie calculé
5. Facture FA créée automatiquement, stock mis à jour

---

### 11. Paramètres Société (Lot 10)

1. **Paramètres → Informations Société**
2. Renseigner NIF, NIS, RC, AI, RIB, banque, capital
3. Renseigner le texte pied de page des factures
4. **Enregistrer** → sauvegardé dans la table `settings`

---

### 12. Gestion Utilisateurs (Lot 10)

1. **Paramètres → Gestion Utilisateurs** *(admin uniquement)*
2. **Créer** un utilisateur : identifiant, nom, mot de passe, rôle
3. **Modifier** les permissions : cocher/décocher les 15 clés d'accès
4. **Changer le mot de passe** ou **Activer/Désactiver** le compte

**Rôles disponibles :**
| Rôle | Permissions par défaut |
|------|----------------------|
| `admin` | Toutes les permissions |
| `seller` | Ventes, stock, clients, encaissements, inventaire |
| `cashier` | Ventes, encaissements uniquement |
| `assistant` | Inventaire uniquement |

---

### 13. Numérotation automatique (Lot 10)

1. **Paramètres → Numérotation**
2. Pour chaque type de document : modifier le **préfixe** et le **numéro de départ**
3. **RAZ annuel** : cocher pour remettre le compteur à 0 chaque nouvelle année
4. **Aperçu** : le prochain numéro s'affiche en temps réel

---

### 14. Sauvegarde & Restauration (Lot 10)

**Sauvegarde :**
1. **Paramètres → Sauvegarde**
2. Configurer le répertoire de sortie
3. **Lancer la sauvegarde maintenant** → fichier `.zip` créé avec DB + assets

**Restauration :**
1. **Paramètres → Restauration**
2. Saisir le chemin du fichier `.zip`
3. Confirmer → ⚠️ **TOUTES les données actuelles sont remplacées**
4. Redémarrer l'application après restauration

---

### 15. Calculatrice (Lot 10)

- Opérations de base : +, -, ×, ÷, %
- Fonction racine carrée (√)
- Touches rapides : **TVA 19%**, **TVA 9%**, **Timbre 1‰**
- Historique des calculs affiché en panneau droit

---

### 16. Mise à Jour Prix en Masse (Lot 10)

1. **Outils → Mise à Jour Prix**
2. Filtrer par **catégorie** (ou toutes)
3. Choisir le **mode** : Hausse/Baisse % ou montant fixe, ou nouveau prix fixe
4. Saisir la **valeur** et l'**arrondi** souhaité
5. **Aperçu** : voir les 50 premiers articles affectés avant confirmation
6. **Appliquer** → mise à jour en base + entrée journal d'audit

---

### 17. Calendrier & Rappels (Lot 10)

- Panneau **chèques à encaisser dans 7 jours** affiché automatiquement
- Créer des **rappels** : rendez-vous, relances, paiements
- Marquer comme **fait** ✓ ou supprimer ✗

---

### 18. Journal d'Audit (Lot 10)

1. **Outils → Journal d'Audit**
2. Filtrer par **utilisateur**, **module**, **limite** de lignes
3. Colonnes : timestamp, utilisateur, action, module, description
4. **Exporter CSV** : fichier `audit_log_AAAAMMJJ.csv`

---

## Tableau des conversions de documents

| Source | Convertissable en |
|--------|-------------------|
| Devis (DV) | FA, PF, BL |
| Proforma (PF) | FA, BL |
| Facture (FA) | BL, Avoir (AV) |
| Bon Livraison (BL) | FA |
| Commande Client (CC) | FA, BL |

---

## Lots de développement

| Lot | Contenu | Statut | Lignes |
|-----|---------|--------|--------|
| **Lot 1** | Structure projet, modèles, schéma SQLite, seed data | ✅ Terminé | ~800 |
| **Lot 2** | Services métier : documents, stock, CMUP, paiements FIFO | ✅ Terminé | ~1200 |
| **Lot 3** | Génération PDF (gopdf), Export Excel (excelize), Backup | ✅ Terminé | ~600 |
| **Lot 4** | UI Login + verrouillage, Dashboard KPIs, Layout, Sidebar, Toolbar | ✅ Terminé | ~800 |
| **Lot 5** | UI Articles : liste filtrée, CRUD, catégories, inventaire, dépôts, prix | ✅ Terminé | ~1700 |
| **Lot 6** | UI Ventes (FA/DV/PF/BL/CC/AV), Achats (FAC/BR/BCF/BCC), POS complet | ✅ Terminé | ~1600 |
| **Lot 7** | UI Clients (CRUD + relevé compte), Fournisseurs (CRUD + notation), Chauffeurs | ✅ Terminé | ~880 |
| **Lot 8** | UI Trésorerie : Caisse, Banque, Chèques, Encaissements, Décaissements, Balance âgée, Dépenses | ✅ Terminé | ~1361 |
| **Lot 9** | UI Rapports : Ventes, Achats, Stock, Trésorerie, Rentabilité, Créances, Indicateurs, G50, TVA, Déclaration Annuelle | ✅ Terminé | ~1240 |
| **Lot 10** | UI Paramètres (Société, Fiscal, Impression, Code-barres), Utilisateurs & permissions, Devises, Backup/Restore, Calculatrice, Mise à Jour Prix, Calendrier/Rappels, Journal Audit, Centre Impression, À Propos | ✅ Terminé | ~2060 |

---

## Particularités algériennes intégrées

| Feature | Détail |
|---------|--------|
| **58 Wilayas** | Sélecteur complet des wilayas algériennes avec codes |
| **Timbre fiscal** | Calcul automatique : 1‰ du TTC, minimum 5 DA, max 2500 DA |
| **TVA** | Taux 0 %, 9 %, 19 % configurables par article |
| **TAP** | Taxe sur Activité Professionnelle dans le G50 |
| **G50** | Déclaration fiscale mensuelle calculée automatiquement (TVA nette + TAP + Timbre) |
| **G50-bis** | Synthèse annuelle pour déclaration fiscale annuelle |
| **Registres TVA** | Registre légal ventes et achats mensuel avec NIF |
| **NIF / NIS / RC / AI** | Champs fiscaux sur clients et fournisseurs |
| **CMUP** | Coût Moyen Unitaire Pondéré recalculé à chaque achat |
| **Multi-société** | Plusieurs fichiers `.db` — une base par société et par année fiscale |
| **Montant en lettres** | Conversion automatique en français |
| **Balance âgée** | Créances/dettes par tranche 0-30j, 31-60j, 61-90j, >90j |
| **Chèques** | Suivi reçus/émis, alertes échéances, dépôt en banque |
| **FIFO paiements** | Ventilation automatique des paiements sur les factures les plus anciennes |

---

## Architecture base de données (SQLite)

| Table | Description |
|-------|-------------|
| `users` | Utilisateurs, mots de passe (bcrypt), permissions |
| `fiscal_years` | Années fiscales par société |
| `articles` | Articles / Produits avec stock et prix |
| `categories`, `brands`, `units` | Nomenclatures articles |
| `warehouses` | Dépôts de stockage multi-sites |
| `price_lists`, `price_list_items` | Listes de prix clients |
| `documents` | Tous documents commerciaux (FA/DV/BL/FAC…) |
| `document_lines` | Lignes des documents avec TVA/remise |
| `clients` | Clients avec données fiscales algériennes |
| `suppliers` | Fournisseurs avec notation qualité |
| `drivers` | Chauffeurs / Livreurs |
| `payments` | Encaissements et décaissements |
| `payment_allocations` | Ventilation FIFO des paiements sur factures |
| `cash_movements` | Journal de caisse (entrées/sorties) |
| `bank_accounts` | Comptes bancaires |
| `bank_movements` | Mouvements bancaires |
| `cheques` | Gestion chèques reçus/émis avec suivi statut |
| `expenses` | Dépenses diverses par catégorie |
| `expense_categories` | Catégories de dépenses |
| `stock_movements` | Historique complet des mouvements de stock |
| `inventory`, `inventory_lines` | Inventaires physiques |
| `tax_declarations` | Déclarations G50 et annuelles (JSON) |
| `settings` | Paramètres application (TVA, timbre, numérotation…) |
| `audit_log` | Journal de toutes les actions utilisateurs |

---

## Raccourcis clavier

| Écran | Touche | Action |
|-------|--------|--------|
| POS | `Entrée` (code-barres) | Ajouter l'article au panier |
| POS | `F2` | Ouvrir la recherche article |
| POS | `F8` | Ouvrir le dialogue de paiement |

---

## Dépannage

**"DB non connectée"** : Aucune base de données sélectionnée — retourner à l'écran de connexion.

**Article introuvable au POS** : Vérifier que le code-barres est bien saisi sur la fiche article.

**Stock négatif** : Le système bloque la vente si le stock est insuffisant à la confirmation.

**Timbre non calculé** : Le timbre fiscal s'applique uniquement sur les paiements en espèces (`cash`).

**G50 affiche des zéros** : Vérifier que des factures confirmées existent pour la période sélectionnée.

**Chèque en attente non visible** : Filtrer par statut "pending" dans l'écran Chèques.

---

## Licence

Projet propriétaire — Algérie 🇩🇿  
Développé pour les PME/TPE algériennes conformément à la réglementation fiscale en vigueur (TVA, Timbre, TAP, G50, NIF/NIS).
