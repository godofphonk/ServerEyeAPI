-- Migration 012: Optimized Granularity System
-- Copyright (c) 2026 godofphonk
-- 
-- This migration implements enterprise-level optimized granularity for better visualization performance
-- Reduces data points significantly while maintaining essential information

-- Apply optimized granularity views
\i deployments/timescaledb/timescaledb-optimized-granularity.sql

-- Apply optimized granularity function  
\i deployments/timescaledb/timescaledb-optimized-function.sql

-- Verify new aggregates exist
DO $$
BEGIN
    -- Check if all optimized views exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'metrics_30m_avg') THEN
        RAISE EXCEPTION 'metrics_30m_avg view not found after migration';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'metrics_2h_avg') THEN
        RAISE EXCEPTION 'metrics_2h_avg view not found after migration';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'metrics_6h_avg') THEN
        RAISE EXCEPTION 'metrics_6h_avg view not found after migration';
    END IF;
    
    -- Check if optimized function exists
    IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'get_metrics_by_granularity') THEN
        RAISE EXCEPTION 'get_metrics_by_granularity function not found after migration';
    END IF;
    
    RAISE NOTICE 'Optimized granularity system migration completed successfully';
END $$;

-- Add migration record
INSERT INTO schema_migrations (version, description, applied_at) 
VALUES ('012', 'Optimized granularity system for better visualization performance', NOW())
ON CONFLICT (version) DO NOTHING;

-- Refresh all new aggregates to ensure they have data
CALL refresh_continuous_aggregate('metrics_30m_avg', NULL, NULL);
CALL refresh_continuous_aggregate('metrics_2h_avg', NULL, NULL);  
CALL refresh_continuous_aggregate('metrics_6h_avg', NULL, NULL);

RAISE NOTICE 'Migration 012: Optimized granularity system applied successfully';
