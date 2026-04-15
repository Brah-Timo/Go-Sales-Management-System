package database

// SchemaSQL contient le schéma complet de la base de données SQLite
const SchemaSQL = `
-- ═══════════════════════════════════════════════════════════════
-- Entreprise / Commerce
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS companies (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    name_ar        TEXT NOT NULL DEFAULT '',
    name_fr        TEXT NOT NULL DEFAULT '',
    activity       TEXT DEFAULT 'Commerce Général',
    address        TEXT DEFAULT '',
    wilaya         TEXT DEFAULT '',
    commune        TEXT DEFAULT '',
    postal_code    TEXT DEFAULT '',
    phone          TEXT DEFAULT '',
    mobile         TEXT DEFAULT '',
    fax            TEXT DEFAULT '',
    email          TEXT DEFAULT '',
    website        TEXT DEFAULT '',
    nif            TEXT DEFAULT '',
    nis            TEXT DEFAULT '',
    rc             TEXT DEFAULT '',
    ai             TEXT DEFAULT '',
    rib            TEXT DEFAULT '',
    bank_name      TEXT DEFAULT '',
    capital        REAL DEFAULT 0,
    logo_path      TEXT DEFAULT '',
    stamp_path     TEXT DEFAULT '',
    signature_path TEXT DEFAULT '',
    footer_text    TEXT DEFAULT 'Merci de votre confiance',
    created_at     TEXT DEFAULT (datetime('now','localtime')),
    updated_at     TEXT DEFAULT (datetime('now','localtime'))
);

-- ═══════════════════════════════════════════════════════════════
-- Années fiscales
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS fiscal_years (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    year       INTEGER NOT NULL UNIQUE,
    start_date TEXT NOT NULL,
    end_date   TEXT NOT NULL,
    status     TEXT DEFAULT 'open' CHECK(status IN ('open','closed')),
    created_at TEXT DEFAULT (datetime('now','localtime'))
);

-- ═══════════════════════════════════════════════════════════════
-- Utilisateurs
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS users (
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    username             TEXT NOT NULL UNIQUE,
    full_name            TEXT NOT NULL,
    password_hash        TEXT NOT NULL,
    role                 TEXT NOT NULL CHECK(role IN ('admin','seller','cashier','assistant')),
    permissions_json     TEXT DEFAULT '{}',
    is_active            INTEGER DEFAULT 1,
    security_question    TEXT DEFAULT '',
    security_answer_hash TEXT DEFAULT '',
    last_login           TEXT,
    created_at           TEXT DEFAULT (datetime('now','localtime'))
);

-- ═══════════════════════════════════════════════════════════════
-- Familles / Catégories de produits (hiérarchique)
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS categories (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name_fr     TEXT NOT NULL,
    name_ar     TEXT DEFAULT '',
    parent_id   INTEGER DEFAULT NULL,
    description TEXT DEFAULT '',
    FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL
);

-- ═══════════════════════════════════════════════════════════════
-- Marques
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS brands (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    name    TEXT NOT NULL,
    country TEXT DEFAULT ''
);

-- ═══════════════════════════════════════════════════════════════
-- Unités de mesure
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS units (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    name_fr TEXT NOT NULL,
    name_ar TEXT DEFAULT '',
    symbol  TEXT NOT NULL
);

-- ═══════════════════════════════════════════════════════════════
-- Conversions d'unités
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS unit_conversions (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    article_id   INTEGER,
    from_unit_id INTEGER NOT NULL,
    to_unit_id   INTEGER NOT NULL,
    factor       REAL NOT NULL,
    FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
    FOREIGN KEY (from_unit_id) REFERENCES units(id),
    FOREIGN KEY (to_unit_id) REFERENCES units(id)
);

-- ═══════════════════════════════════════════════════════════════
-- Articles / Produits (table centrale)
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS articles (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    reference           TEXT NOT NULL UNIQUE,
    barcode             TEXT DEFAULT '',
    name_ar             TEXT DEFAULT '',
    name_fr             TEXT NOT NULL,
    description         TEXT DEFAULT '',
    category_id         INTEGER,
    brand_id            INTEGER,
    unit_id             INTEGER,
    purchase_price      REAL DEFAULT 0,
    cmup                REAL DEFAULT 0,
    sale_price_ht       REAL DEFAULT 0,
    sale_price_ttc      REAL DEFAULT 0,
    wholesale_price     REAL DEFAULT 0,
    semi_wholesale_price REAL DEFAULT 0,
    margin_percent      REAL DEFAULT 0,
    tva_rate            REAL DEFAULT 19,
    stock_qty           REAL DEFAULT 0,
    stock_min           REAL DEFAULT 0,
    stock_max           REAL DEFAULT 0,
    valuation_method    TEXT DEFAULT 'CMUP' CHECK(valuation_method IN ('CMUP','FIFO')),
    warehouse_location  TEXT DEFAULT '',
    lot_tracking        INTEGER DEFAULT 0,
    expiry_tracking     INTEGER DEFAULT 0,
    image_path          TEXT DEFAULT '',
    is_active           INTEGER DEFAULT 1,
    created_at          TEXT DEFAULT (datetime('now','localtime')),
    updated_at          TEXT DEFAULT (datetime('now','localtime')),
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
    FOREIGN KEY (brand_id)    REFERENCES brands(id)     ON DELETE SET NULL,
    FOREIGN KEY (unit_id)     REFERENCES units(id)      ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_articles_barcode   ON articles(barcode);
CREATE INDEX IF NOT EXISTS idx_articles_reference ON articles(reference);
CREATE INDEX IF NOT EXISTS idx_articles_name_fr   ON articles(name_fr);
CREATE INDEX IF NOT EXISTS idx_articles_category  ON articles(category_id);
CREATE INDEX IF NOT EXISTS idx_articles_stock     ON articles(stock_qty, stock_min);

-- ═══════════════════════════════════════════════════════════════
-- Listes de prix
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS price_lists (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    description TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS article_prices (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    article_id   INTEGER NOT NULL,
    price_list_id INTEGER NOT NULL,
    price        REAL NOT NULL,
    UNIQUE(article_id, price_list_id),
    FOREIGN KEY (article_id)    REFERENCES articles(id)     ON DELETE CASCADE,
    FOREIGN KEY (price_list_id) REFERENCES price_lists(id)  ON DELETE CASCADE
);

-- ═══════════════════════════════════════════════════════════════
-- Lots / Batches de produits
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS article_lots (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    article_id      INTEGER NOT NULL,
    lot_number      TEXT NOT NULL,
    production_date TEXT,
    expiry_date     TEXT,
    quantity        REAL DEFAULT 0,
    FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
);

-- ═══════════════════════════════════════════════════════════════
-- Dépôts / Magasins
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS warehouses (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    name    TEXT NOT NULL,
    address TEXT DEFAULT '',
    manager TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS warehouse_stock (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    article_id   INTEGER NOT NULL,
    warehouse_id INTEGER NOT NULL,
    quantity     REAL DEFAULT 0,
    UNIQUE(article_id, warehouse_id),
    FOREIGN KEY (article_id)   REFERENCES articles(id)   ON DELETE CASCADE,
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE
);

-- ═══════════════════════════════════════════════════════════════
-- Clients
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS clients (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    code          TEXT NOT NULL UNIQUE,
    name_ar       TEXT DEFAULT '',
    name_fr       TEXT NOT NULL,
    type          TEXT DEFAULT 'person' CHECK(type IN ('person','company')),
    address       TEXT DEFAULT '',
    wilaya        TEXT DEFAULT '',
    commune       TEXT DEFAULT '',
    phone         TEXT DEFAULT '',
    mobile        TEXT DEFAULT '',
    fax           TEXT DEFAULT '',
    email         TEXT DEFAULT '',
    nif           TEXT DEFAULT '',
    nis           TEXT DEFAULT '',
    rc            TEXT DEFAULT '',
    ai            TEXT DEFAULT '',
    price_list_id INTEGER,
    credit_limit  REAL DEFAULT 0,
    payment_terms TEXT DEFAULT 'immediate',
    discount_rate REAL DEFAULT 0,
    balance       REAL DEFAULT 0,
    is_blocked    INTEGER DEFAULT 0,
    notes         TEXT DEFAULT '',
    created_at    TEXT DEFAULT (datetime('now','localtime')),
    FOREIGN KEY (price_list_id) REFERENCES price_lists(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_clients_name    ON clients(name_fr);
CREATE INDEX IF NOT EXISTS idx_clients_balance ON clients(balance);

-- ═══════════════════════════════════════════════════════════════
-- Fournisseurs
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS suppliers (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    code            TEXT NOT NULL UNIQUE,
    name_ar         TEXT DEFAULT '',
    name_fr         TEXT NOT NULL,
    address         TEXT DEFAULT '',
    wilaya          TEXT DEFAULT '',
    phone           TEXT DEFAULT '',
    mobile          TEXT DEFAULT '',
    fax             TEXT DEFAULT '',
    email           TEXT DEFAULT '',
    nif             TEXT DEFAULT '',
    nis             TEXT DEFAULT '',
    rc              TEXT DEFAULT '',
    ai              TEXT DEFAULT '',
    payment_terms   TEXT DEFAULT 'immediate',
    balance         REAL DEFAULT 0,
    rating_delivery INTEGER DEFAULT 3,
    rating_quality  INTEGER DEFAULT 3,
    rating_pricing  INTEGER DEFAULT 3,
    notes           TEXT DEFAULT '',
    created_at      TEXT DEFAULT (datetime('now','localtime'))
);

-- ═══════════════════════════════════════════════════════════════
-- Chauffeurs / Livreurs
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS drivers (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,
    phone         TEXT DEFAULT '',
    vehicle_plate TEXT DEFAULT ''
);

-- ═══════════════════════════════════════════════════════════════
-- Documents commerciaux (table pivot centrale)
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS documents (
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    doc_type                TEXT NOT NULL CHECK(doc_type IN ('FA','FAC','BL','BR','DV','PF','BCC','BCF','AV','BP','BRE')),
    doc_number              TEXT NOT NULL UNIQUE,
    date                    TEXT NOT NULL,
    fiscal_year_id          INTEGER,
    client_id               INTEGER,
    supplier_id             INTEGER,
    warehouse_id            INTEGER,
    payment_method          TEXT DEFAULT 'cash' CHECK(payment_method IN ('cash','cheque','transfer','credit','mixed')),
    payment_terms           TEXT DEFAULT 'immediate',
    price_list_id           INTEGER,
    total_ht                REAL DEFAULT 0,
    total_discount          REAL DEFAULT 0,
    global_discount_pct     REAL DEFAULT 0,
    net_ht                  REAL DEFAULT 0,
    total_tva               REAL DEFAULT 0,
    total_ttc               REAL DEFAULT 0,
    timbre                  REAL DEFAULT 0,
    net_amount              REAL DEFAULT 0,
    amount_paid             REAL DEFAULT 0,
    amount_remaining        REAL DEFAULT 0,
    status                  TEXT DEFAULT 'draft' CHECK(status IN ('draft','confirmed','paid','partial','cancelled')),
    notes                   TEXT DEFAULT '',
    driver_id               INTEGER,
    delivery_address        TEXT DEFAULT '',
    supplier_invoice_number TEXT DEFAULT '',
    validity_days           INTEGER DEFAULT 0,
    source_doc_id           INTEGER,
    created_by              INTEGER,
    created_at              TEXT DEFAULT (datetime('now','localtime')),
    updated_at              TEXT DEFAULT (datetime('now','localtime')),
    FOREIGN KEY (fiscal_year_id) REFERENCES fiscal_years(id),
    FOREIGN KEY (client_id)      REFERENCES clients(id),
    FOREIGN KEY (supplier_id)    REFERENCES suppliers(id),
    FOREIGN KEY (warehouse_id)   REFERENCES warehouses(id),
    FOREIGN KEY (price_list_id)  REFERENCES price_lists(id),
    FOREIGN KEY (driver_id)      REFERENCES drivers(id),
    FOREIGN KEY (source_doc_id)  REFERENCES documents(id),
    FOREIGN KEY (created_by)     REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_documents_type        ON documents(doc_type);
CREATE INDEX IF NOT EXISTS idx_documents_date        ON documents(date);
CREATE INDEX IF NOT EXISTS idx_documents_client      ON documents(client_id);
CREATE INDEX IF NOT EXISTS idx_documents_supplier    ON documents(supplier_id);
CREATE INDEX IF NOT EXISTS idx_documents_status      ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_fiscal_year ON documents(fiscal_year_id);
CREATE INDEX IF NOT EXISTS idx_documents_number      ON documents(doc_number);

-- ═══════════════════════════════════════════════════════════════
-- Lignes de documents
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS document_lines (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    document_id      INTEGER NOT NULL,
    line_number      INTEGER NOT NULL,
    article_id       INTEGER,
    designation      TEXT NOT NULL,
    quantity         REAL NOT NULL DEFAULT 1,
    unit             TEXT DEFAULT '',
    unit_price_ht    REAL NOT NULL DEFAULT 0,
    discount_percent REAL DEFAULT 0,
    discount_amount  REAL DEFAULT 0,
    amount_ht        REAL DEFAULT 0,
    tva_rate         REAL DEFAULT 19,
    tva_amount       REAL DEFAULT 0,
    amount_ttc       REAL DEFAULT 0,
    lot_number       TEXT DEFAULT '',
    expiry_date      TEXT,
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (article_id)  REFERENCES articles(id)  ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_doclines_document ON document_lines(document_id);
CREATE INDEX IF NOT EXISTS idx_doclines_article  ON document_lines(article_id);

-- ═══════════════════════════════════════════════════════════════
-- Paiements (encaissements et décaissements)
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS payments (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    type           TEXT NOT NULL CHECK(type IN ('collection','disbursement')),
    date           TEXT NOT NULL,
    client_id      INTEGER,
    supplier_id    INTEGER,
    amount         REAL NOT NULL,
    payment_method TEXT NOT NULL CHECK(payment_method IN ('cash','cheque','transfer')),
    cheque_number  TEXT DEFAULT '',
    bank_name      TEXT DEFAULT '',
    reference      TEXT DEFAULT '',
    notes          TEXT DEFAULT '',
    created_by     INTEGER,
    created_at     TEXT DEFAULT (datetime('now','localtime')),
    FOREIGN KEY (client_id)   REFERENCES clients(id),
    FOREIGN KEY (supplier_id) REFERENCES suppliers(id),
    FOREIGN KEY (created_by)  REFERENCES users(id)
);

-- ═══════════════════════════════════════════════════════════════
-- Ventilation paiements / factures
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS payment_allocations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    payment_id  INTEGER NOT NULL,
    document_id INTEGER NOT NULL,
    amount      REAL NOT NULL,
    FOREIGN KEY (payment_id)  REFERENCES payments(id)  ON DELETE CASCADE,
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
);

-- ═══════════════════════════════════════════════════════════════
-- Chèques
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS cheques (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    type               TEXT NOT NULL CHECK(type IN ('issued','received')),
    cheque_number      TEXT NOT NULL,
    date               TEXT NOT NULL,
    due_date           TEXT NOT NULL,
    amount             REAL NOT NULL,
    payer_payee        TEXT NOT NULL,
    bank_name          TEXT DEFAULT '',
    status             TEXT DEFAULT 'pending' CHECK(status IN ('pending','deposited','collected','rejected','returned','replaced')),
    reject_reason      TEXT DEFAULT '',
    related_payment_id INTEGER,
    notes              TEXT DEFAULT '',
    created_at         TEXT DEFAULT (datetime('now','localtime')),
    FOREIGN KEY (related_payment_id) REFERENCES payments(id)
);

CREATE INDEX IF NOT EXISTS idx_cheques_due_date ON cheques(due_date);
CREATE INDEX IF NOT EXISTS idx_cheques_status   ON cheques(status);

-- ═══════════════════════════════════════════════════════════════
-- Mouvements de caisse
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS cash_movements (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    date        TEXT NOT NULL DEFAULT (datetime('now','localtime')),
    type        TEXT NOT NULL CHECK(type IN ('in','out')),
    category    TEXT DEFAULT '',
    description TEXT DEFAULT '',
    reference   TEXT DEFAULT '',
    party_name  TEXT DEFAULT '',
    amount      REAL NOT NULL,
    created_by  INTEGER,
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_cash_date ON cash_movements(date);

-- ═══════════════════════════════════════════════════════════════
-- Comptes bancaires
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS bank_accounts (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    bank_name      TEXT NOT NULL,
    branch         TEXT DEFAULT '',
    account_number TEXT DEFAULT '',
    rib            TEXT DEFAULT '',
    balance        REAL DEFAULT 0
);

-- ═══════════════════════════════════════════════════════════════
-- Mouvements bancaires
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS bank_movements (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    bank_account_id INTEGER NOT NULL,
    date            TEXT NOT NULL,
    type            TEXT DEFAULT '',
    description     TEXT DEFAULT '',
    reference       TEXT DEFAULT '',
    debit           REAL DEFAULT 0,
    credit          REAL DEFAULT 0,
    is_reconciled   INTEGER DEFAULT 0,
    created_at      TEXT DEFAULT (datetime('now','localtime')),
    FOREIGN KEY (bank_account_id) REFERENCES bank_accounts(id)
);

-- ═══════════════════════════════════════════════════════════════
-- Mouvements de stock
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS stock_movements (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    date             TEXT NOT NULL DEFAULT (datetime('now','localtime')),
    type             TEXT NOT NULL CHECK(type IN ('purchase_in','sale_out','transfer_in','transfer_out','return_in','return_out','adjustment_in','adjustment_out','damage')),
    article_id       INTEGER NOT NULL,
    warehouse_id     INTEGER,
    quantity         REAL NOT NULL,
    unit_price       REAL DEFAULT 0,
    reference_doc_id INTEGER,
    notes            TEXT DEFAULT '',
    created_by       INTEGER,
    FOREIGN KEY (article_id)       REFERENCES articles(id),
    FOREIGN KEY (warehouse_id)     REFERENCES warehouses(id),
    FOREIGN KEY (reference_doc_id) REFERENCES documents(id),
    FOREIGN KEY (created_by)       REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_article ON stock_movements(article_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_date    ON stock_movements(date);
CREATE INDEX IF NOT EXISTS idx_stock_movements_type    ON stock_movements(type);

-- ═══════════════════════════════════════════════════════════════
-- Inventaires
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS inventories (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    date       TEXT NOT NULL,
    type       TEXT DEFAULT 'full' CHECK(type IN ('full','partial','single')),
    status     TEXT DEFAULT 'draft' CHECK(status IN ('draft','confirmed')),
    notes      TEXT DEFAULT '',
    created_by INTEGER,
    created_at TEXT DEFAULT (datetime('now','localtime')),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS inventory_lines (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    inventory_id    INTEGER NOT NULL,
    article_id      INTEGER NOT NULL,
    theoretical_qty REAL DEFAULT 0,
    physical_qty    REAL DEFAULT 0,
    difference      REAL DEFAULT 0,
    value           REAL DEFAULT 0,
    note            TEXT DEFAULT '',
    FOREIGN KEY (inventory_id) REFERENCES inventories(id) ON DELETE CASCADE,
    FOREIGN KEY (article_id)   REFERENCES articles(id)
);

-- ═══════════════════════════════════════════════════════════════
-- Dépenses diverses
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS expense_categories (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS expenses (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    date           TEXT NOT NULL,
    category_id    INTEGER,
    description    TEXT DEFAULT '',
    amount         REAL NOT NULL,
    payment_method TEXT DEFAULT 'cash' CHECK(payment_method IN ('cash','bank')),
    created_by     INTEGER,
    created_at     TEXT DEFAULT (datetime('now','localtime')),
    FOREIGN KEY (category_id) REFERENCES expense_categories(id),
    FOREIGN KEY (created_by)  REFERENCES users(id)
);

-- ═══════════════════════════════════════════════════════════════
-- Déclarations fiscales
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS tax_declarations (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    type       TEXT NOT NULL CHECK(type IN ('G50','annual')),
    year       INTEGER NOT NULL,
    month      INTEGER,
    data_json  TEXT DEFAULT '{}',
    status     TEXT DEFAULT 'draft' CHECK(status IN ('draft','submitted')),
    created_at TEXT DEFAULT (datetime('now','localtime'))
);

-- ═══════════════════════════════════════════════════════════════
-- Devises
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS currencies (
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    name   TEXT NOT NULL,
    code   TEXT NOT NULL UNIQUE,
    symbol TEXT DEFAULT '',
    rate   REAL DEFAULT 1.0
);

-- ═══════════════════════════════════════════════════════════════
-- Configuration de numérotation
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS numbering_config (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    doc_type       TEXT NOT NULL UNIQUE,
    prefix         TEXT NOT NULL,
    current_number INTEGER DEFAULT 0,
    reset_yearly   INTEGER DEFAULT 1
);

-- ═══════════════════════════════════════════════════════════════
-- Rappels / Alertes
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS reminders (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    date        TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT DEFAULT '',
    type        TEXT DEFAULT 'custom' CHECK(type IN ('custom','tax','cheque','delivery')),
    related_id  INTEGER,
    is_done     INTEGER DEFAULT 0,
    created_by  INTEGER,
    FOREIGN KEY (created_by) REFERENCES users(id)
);

-- ═══════════════════════════════════════════════════════════════
-- Journal d'audit
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS audit_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp   TEXT DEFAULT (datetime('now','localtime')),
    user_id     INTEGER,
    action_type TEXT NOT NULL,
    module      TEXT NOT NULL,
    description TEXT DEFAULT '',
    old_data    TEXT DEFAULT '',
    new_data    TEXT DEFAULT '',
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_user      ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_module    ON audit_log(module);

-- ═══════════════════════════════════════════════════════════════
-- Paramètres généraux (clé-valeur)
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS settings (
    id    INTEGER PRIMARY KEY AUTOINCREMENT,
    key   TEXT NOT NULL UNIQUE,
    value TEXT DEFAULT ''
);

-- ═══════════════════════════════════════════════════════════════
-- Triggers
-- ═══════════════════════════════════════════════════════════════

CREATE TRIGGER IF NOT EXISTS update_article_timestamp
AFTER UPDATE ON articles
BEGIN
    UPDATE articles SET updated_at = datetime('now','localtime') WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_document_timestamp
AFTER UPDATE ON documents
BEGIN
    UPDATE documents SET updated_at = datetime('now','localtime') WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_document_remaining
AFTER UPDATE OF amount_paid ON documents
BEGIN
    UPDATE documents 
    SET amount_remaining = net_amount - NEW.amount_paid 
    WHERE id = NEW.id;
END;
`
