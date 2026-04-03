-- Updated get_metrics_by_granularity function with optimized granularities
-- Replaces the old function with new enterprise-level granularity selection

-- Drop existing function
DROP FUNCTION IF EXISTS get_metrics_by_granularity(TEXT, TIMESTAMPTZ, TIMESTAMPTZ);

CREATE OR REPLACE FUNCTION get_metrics_by_granularity(
    p_server_id TEXT,
    p_start_time TIMESTAMPTZ,
    p_end_time TIMESTAMPTZ
)
RETURNS TABLE (
    bucket TIMESTAMPTZ,
    avg_cpu DOUBLE PRECISION,
    max_cpu DOUBLE PRECISION,
    min_cpu DOUBLE PRECISION,
    avg_memory DOUBLE PRECISION,
    max_memory DOUBLE PRECISION,
    min_memory DOUBLE PRECISION,
    avg_disk DOUBLE PRECISION,
    max_disk DOUBLE PRECISION,
    avg_network DOUBLE PRECISION,
    max_network DOUBLE PRECISION,
    sample_count BIGINT,
    granularity TEXT
) AS $$
BEGIN
    -- Use 1-minute data for last hour (max 60 points)
    IF p_end_time - p_start_time <= INTERVAL '1 hour' THEN
        RETURN QUERY
        SELECT 
            m.bucket, m.avg_cpu, m.max_cpu, m.min_cpu,
            m.avg_memory, m.max_memory, m.min_memory,
            m.avg_disk, m.max_disk, m.avg_network, m.max_network,
            m.sample_count, '1m'::TEXT
        FROM metrics_1m_avg m
        WHERE m.server_id = p_server_id 
        AND m.bucket BETWEEN p_start_time AND p_end_time
        ORDER BY m.bucket;
    
    -- Use 10-minute data for 1-6 hours (max 36 points)
    ELSIF p_end_time - p_start_time <= INTERVAL '6 hours' THEN
        RETURN QUERY
        SELECT 
            m.bucket, m.avg_cpu, m.max_cpu, m.min_cpu,
            m.avg_memory, m.max_memory, m.min_memory,
            m.avg_disk, m.max_disk, m.avg_network, m.max_network,
            m.sample_count, '10m'::TEXT
        FROM metrics_10m_avg m
        WHERE m.server_id = p_server_id 
        AND m.bucket BETWEEN p_start_time AND p_end_time
        ORDER BY m.bucket;
    
    -- Use 30-minute data for 6-24 hours (max 36 points)
    ELSIF p_end_time - p_start_time <= INTERVAL '24 hours' THEN
        RETURN QUERY
        SELECT 
            m.bucket, m.avg_cpu, m.max_cpu, m.min_cpu,
            m.avg_memory, m.max_memory, m.min_memory,
            m.avg_disk, m.max_disk, m.avg_network, m.max_network,
            m.sample_count, '30m'::TEXT
        FROM metrics_30m_avg m
        WHERE m.server_id = p_server_id 
        AND m.bucket BETWEEN p_start_time AND p_end_time
        ORDER BY m.bucket;
    
    -- Use 2-hour data for 1-7 days (max 84 points)
    ELSIF p_end_time - p_start_time <= INTERVAL '7 days' THEN
        RETURN QUERY
        SELECT 
            m.bucket, m.avg_cpu, m.max_cpu, m.min_cpu,
            m.avg_memory, m.max_memory, m.min_memory,
            m.avg_disk, m.max_disk, m.avg_network, m.max_network,
            m.sample_count, '2h'::TEXT
        FROM metrics_2h_avg m
        WHERE m.server_id = p_server_id 
        AND m.bucket BETWEEN p_start_time AND p_end_time
        ORDER BY m.bucket;
    
    -- Use 6-hour data for 7+ days (max 120 points for 30 days)
    ELSE
        RETURN QUERY
        SELECT 
            m.bucket, m.avg_cpu, m.max_cpu, m.min_cpu,
            m.avg_memory, m.max_memory, m.min_memory,
            m.avg_disk, m.max_disk, m.avg_network, m.max_network,
            m.sample_count, '6h'::TEXT
        FROM metrics_6h_avg m
        WHERE m.server_id = p_server_id 
        AND m.bucket BETWEEN p_start_time AND p_end_time
        ORDER BY m.bucket;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Grant permissions
GRANT EXECUTE ON FUNCTION get_metrics_by_granularity(TEXT, TIMESTAMPTZ, TIMESTAMPTZ) TO server_eye_read;

-- Add comment
COMMENT ON FUNCTION get_metrics_by_granularity(TEXT, TIMESTAMPTZ, TIMESTAMPTZ) IS 
'Optimized function that returns metrics with appropriate granularity for visualization: 
- 1m for ≤1h (max 60 points)
- 10m for 1-6h (max 36 points) 
- 30m for 6-24h (max 36 points)
- 2h for 1-7d (max 84 points)
- 6h for >7d (max 120 points for 30d)';
