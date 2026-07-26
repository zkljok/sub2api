CREATE TABLE IF NOT EXISTS model_plaza_vendors (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    icon TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT model_plaza_vendors_status_check CHECK (status IN ('active', 'disabled'))
);

CREATE TABLE IF NOT EXISTS model_plaza_models (
    id BIGSERIAL PRIMARY KEY,
    model_name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    icon TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    vendor_id BIGINT REFERENCES model_plaza_vendors(id) ON DELETE SET NULL,
    endpoints JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    name_rule VARCHAR(20) NOT NULL DEFAULT 'exact',
    sort_order INTEGER NOT NULL DEFAULT 0,
    pricing_override JSONB NOT NULL DEFAULT '{}'::jsonb,
    auto_synced BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT model_plaza_models_status_check CHECK (status IN ('active', 'disabled')),
    CONSTRAINT model_plaza_models_name_rule_check CHECK (name_rule IN ('exact', 'prefix', 'suffix', 'contains')),
    CONSTRAINT model_plaza_models_name_rule_unique UNIQUE (model_name, name_rule)
);

CREATE INDEX IF NOT EXISTS idx_model_plaza_models_vendor_id ON model_plaza_models(vendor_id);
CREATE INDEX IF NOT EXISTS idx_model_plaza_models_active_sort ON model_plaza_models(status, sort_order, model_name);
