-- Dolt database: alphanet_fred
-- Purpose: FRED macro, rates, commodity, gold, dollar, and fallback volatility observations.
-- Normal users usually clone the public database instead of running this schema manually.

CREATE TABLE IF NOT EXISTS source_metadata (
    source_id VARCHAR(64) PRIMARY KEY,
    provider_name VARCHAR(128) NOT NULL,
    api_docs_url TEXT,
    terms_url TEXT,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS series_catalog (
    series_id VARCHAR(128) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    category VARCHAR(128),
    frequency VARCHAR(64),
    units VARCHAR(128),
    seasonal_adjustment VARCHAR(128),
    source_url TEXT,
    alpha_symbol VARCHAR(64),
    notes TEXT
);

CREATE TABLE IF NOT EXISTS observations (
    series_id VARCHAR(128) NOT NULL,
    date DATE NOT NULL,
    value DECIMAL(30,10),
    realtime_start DATE NOT NULL DEFAULT '1776-07-04',
    realtime_end DATE NOT NULL DEFAULT '9999-12-31',
    source_id VARCHAR(64) NOT NULL DEFAULT 'fred',
    ingestion_id VARCHAR(128),
    PRIMARY KEY (series_id, date, realtime_start, realtime_end),
    KEY idx_observations_series_date (series_id, date)
);

CREATE TABLE IF NOT EXISTS series_aliases (
    alias VARCHAR(128) PRIMARY KEY,
    series_id VARCHAR(128) NOT NULL,
    meaning VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS ingestion_runs (
    ingestion_id VARCHAR(128) PRIMARY KEY,
    started_at DATETIME NOT NULL,
    finished_at DATETIME,
    status VARCHAR(64) NOT NULL,
    request_url TEXT,
    request_params JSON,
    rows_read INT DEFAULT 0,
    rows_written INT DEFAULT 0,
    error_message TEXT,
    metadata JSON
);

CREATE TABLE IF NOT EXISTS data_quality_issues (
    issue_id VARCHAR(128) PRIMARY KEY,
    series_id VARCHAR(128),
    date DATE,
    severity VARCHAR(64) NOT NULL,
    issue_type VARCHAR(128) NOT NULL,
    message TEXT NOT NULL,
    detected_at DATETIME NOT NULL,
    metadata JSON
);
