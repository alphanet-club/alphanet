-- Dolt database: alphanet_cboe
-- Purpose: Cboe volatility index data, especially VIX historical data.
-- Normal users usually clone the public database instead of running this schema manually.

CREATE TABLE IF NOT EXISTS source_metadata (
    source_id VARCHAR(64) PRIMARY KEY,
    provider_name VARCHAR(128) NOT NULL,
    source_url TEXT,
    terms_url TEXT,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS volatility_indexes (
    index_symbol VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    source_url TEXT
);

CREATE TABLE IF NOT EXISTS vix_daily (
    index_symbol VARCHAR(64) NOT NULL,
    date DATE NOT NULL,
    open DECIMAL(30,10),
    high DECIMAL(30,10),
    low DECIMAL(30,10),
    close DECIMAL(30,10),
    source_id VARCHAR(64) NOT NULL DEFAULT 'cboe',
    ingestion_id VARCHAR(128),
    PRIMARY KEY (index_symbol, date),
    KEY idx_vix_daily_symbol_date (index_symbol, date)
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
    index_symbol VARCHAR(64),
    date DATE,
    severity VARCHAR(64) NOT NULL,
    issue_type VARCHAR(128) NOT NULL,
    message TEXT NOT NULL,
    detected_at DATETIME NOT NULL,
    metadata JSON
);
