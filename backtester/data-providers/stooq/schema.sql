-- Dolt database: alphanet_stooq
-- Purpose: Stooq market price data for ETFs, equities, indexes, and simple commodity benchmark funds.
-- Normal users usually clone the public database instead of running this schema manually.

CREATE TABLE IF NOT EXISTS source_metadata (
    source_id VARCHAR(64) PRIMARY KEY,
    provider_name VARCHAR(128) NOT NULL,
    source_url TEXT,
    terms_url TEXT,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS symbols (
    symbol VARCHAR(64) PRIMARY KEY,
    stooq_symbol VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(255),
    instrument_type VARCHAR(64) NOT NULL,
    exchange VARCHAR(64),
    currency VARCHAR(16) DEFAULT 'USD',
    active BOOLEAN DEFAULT TRUE,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS daily_prices (
    symbol VARCHAR(64) NOT NULL,
    date DATE NOT NULL,
    open DECIMAL(30,10),
    high DECIMAL(30,10),
    low DECIMAL(30,10),
    close DECIMAL(30,10),
    adjusted_close DECIMAL(30,10),
    volume DECIMAL(30,4),
    source_id VARCHAR(64) NOT NULL DEFAULT 'stooq',
    ingestion_id VARCHAR(128),
    PRIMARY KEY (symbol, date),
    KEY idx_daily_prices_symbol_date (symbol, date)
);

CREATE TABLE IF NOT EXISTS ingestion_runs (
    ingestion_id VARCHAR(128) PRIMARY KEY,
    started_at DATETIME NOT NULL,
    finished_at DATETIME,
    status VARCHAR(64) NOT NULL,
    request_url TEXT,
    rows_read INT DEFAULT 0,
    rows_written INT DEFAULT 0,
    error_message TEXT,
    metadata JSON
);

CREATE TABLE IF NOT EXISTS data_quality_issues (
    issue_id VARCHAR(128) PRIMARY KEY,
    symbol VARCHAR(64),
    date DATE,
    severity VARCHAR(64) NOT NULL,
    issue_type VARCHAR(128) NOT NULL,
    message TEXT NOT NULL,
    detected_at DATETIME NOT NULL,
    metadata JSON
);
