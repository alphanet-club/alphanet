-- Dolt database: alphanet_imf_portwatch
-- Purpose: IMF PortWatch shipping, port, and chokepoint observations.
-- Normal users usually clone the public database instead of running this schema manually.

CREATE TABLE IF NOT EXISTS source_metadata (
    source_id VARCHAR(64) PRIMARY KEY,
    provider_name VARCHAR(128) NOT NULL,
    source_url TEXT,
    methodology_url TEXT,
    terms_url TEXT,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS chokepoints (
    chokepoint_id VARCHAR(128) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    region VARCHAR(255),
    description TEXT,
    metadata JSON
);

CREATE TABLE IF NOT EXISTS ports (
    port_id VARCHAR(128) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    country VARCHAR(128),
    region VARCHAR(255),
    latitude DECIMAL(18,10),
    longitude DECIMAL(18,10),
    metadata JSON
);

CREATE TABLE IF NOT EXISTS chokepoint_observations (
    chokepoint_id VARCHAR(128) NOT NULL,
    date DATE NOT NULL,
    metric VARCHAR(128) NOT NULL,
    value DECIMAL(30,10),
    unit VARCHAR(64),
    source_id VARCHAR(64) NOT NULL DEFAULT 'imf_portwatch',
    ingestion_id VARCHAR(128),
    PRIMARY KEY (chokepoint_id, metric, date),
    KEY idx_chokepoint_metric_date (chokepoint_id, metric, date)
);

CREATE TABLE IF NOT EXISTS port_observations (
    port_id VARCHAR(128) NOT NULL,
    date DATE NOT NULL,
    metric VARCHAR(128) NOT NULL,
    value DECIMAL(30,10),
    unit VARCHAR(64),
    source_id VARCHAR(64) NOT NULL DEFAULT 'imf_portwatch',
    ingestion_id VARCHAR(128),
    PRIMARY KEY (port_id, metric, date),
    KEY idx_port_metric_date (port_id, metric, date)
);

CREATE TABLE IF NOT EXISTS ingestion_runs (
    ingestion_id VARCHAR(128) PRIMARY KEY,
    started_at DATETIME NOT NULL,
    finished_at DATETIME,
    status VARCHAR(64) NOT NULL,
    request_url TEXT,
    dataset_id VARCHAR(255),
    rows_read INT DEFAULT 0,
    rows_written INT DEFAULT 0,
    error_message TEXT,
    metadata JSON
);

CREATE TABLE IF NOT EXISTS data_quality_issues (
    issue_id VARCHAR(128) PRIMARY KEY,
    entity_id VARCHAR(128),
    date DATE,
    severity VARCHAR(64) NOT NULL,
    issue_type VARCHAR(128) NOT NULL,
    message TEXT NOT NULL,
    detected_at DATETIME NOT NULL,
    metadata JSON
);
