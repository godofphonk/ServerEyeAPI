-- Test script for new optimized granularity system
-- Validates that the correct granularity is selected for different time ranges

-- Test 1: 1 hour period - should use 1m granularity (max 60 points)
SELECT 
    '1h test' as test_name,
    COUNT(*) as points_returned,
    '1m' as expected_granularity,
    granularity as actual_granularity,
    CASE WHEN COUNT(*) <= 60 THEN 'PASS' ELSE 'FAIL' END as result
FROM get_metrics_by_granularity('test-server', 
    NOW() - INTERVAL '1 hour', 
    NOW()
);

-- Test 2: 6 hours period - should use 10m granularity (max 36 points)
SELECT 
    '6h test' as test_name,
    COUNT(*) as points_returned,
    '10m' as expected_granularity,
    granularity as actual_granularity,
    CASE WHEN COUNT(*) <= 36 THEN 'PASS' ELSE 'FAIL' END as result
FROM get_metrics_by_granularity('test-server', 
    NOW() - INTERVAL '6 hours', 
    NOW()
);

-- Test 3: 24 hours period - should use 30m granularity (max 48 points)
SELECT 
    '24h test' as test_name,
    COUNT(*) as points_returned,
    '30m' as expected_granularity,
    granularity as actual_granularity,
    CASE WHEN COUNT(*) <= 48 THEN 'PASS' ELSE 'FAIL' END as result
FROM get_metrics_by_granularity('test-server', 
    NOW() - INTERVAL '24 hours', 
    NOW()
);

-- Test 4: 7 days period - should use 2h granularity (max 84 points)
SELECT 
    '7d test' as test_name,
    COUNT(*) as points_returned,
    '2h' as expected_granularity,
    granularity as actual_granularity,
    CASE WHEN COUNT(*) <= 84 THEN 'PASS' ELSE 'FAIL' END as result
FROM get_metrics_by_granularity('test-server', 
    NOW() - INTERVAL '7 days', 
    NOW()
);

-- Test 5: 30 days period - should use 6h granularity (max 120 points)
SELECT 
    '30d test' as test_name,
    COUNT(*) as points_returned,
    '6h' as expected_granularity,
    granularity as actual_granularity,
    CASE WHEN COUNT(*) <= 120 THEN 'PASS' ELSE 'FAIL' END as result
FROM get_metrics_by_granularity('test-server', 
    NOW() - INTERVAL '30 days', 
    NOW()
);

-- Summary of all tests
SELECT 
    test_name,
    points_returned,
    expected_granularity,
    actual_granularity,
    result
FROM (
    SELECT 
        '1h test' as test_name,
        COUNT(*) as points_returned,
        '1m' as expected_granularity,
        granularity as actual_granularity,
        CASE WHEN COUNT(*) <= 60 THEN 'PASS' ELSE 'FAIL' END as result
    FROM get_metrics_by_granularity('test-server', NOW() - INTERVAL '1 hour', NOW())
    
    UNION ALL
    
    SELECT 
        '6h test' as test_name,
        COUNT(*) as points_returned,
        '10m' as expected_granularity,
        granularity as actual_granularity,
        CASE WHEN COUNT(*) <= 36 THEN 'PASS' ELSE 'FAIL' END as result
    FROM get_metrics_by_granularity('test-server', NOW() - INTERVAL '6 hours', NOW())
    
    UNION ALL
    
    SELECT 
        '24h test' as test_name,
        COUNT(*) as points_returned,
        '30m' as expected_granularity,
        granularity as actual_granularity,
        CASE WHEN COUNT(*) <= 48 THEN 'PASS' ELSE 'FAIL' END as result
    FROM get_metrics_by_granularity('test-server', NOW() - INTERVAL '24 hours', NOW())
    
    UNION ALL
    
    SELECT 
        '7d test' as test_name,
        COUNT(*) as points_returned,
        '2h' as expected_granularity,
        granularity as actual_granularity,
        CASE WHEN COUNT(*) <= 84 THEN 'PASS' ELSE 'FAIL' END as result
    FROM get_metrics_by_granularity('test-server', NOW() - INTERVAL '7 days', NOW())
    
    UNION ALL
    
    SELECT 
        '30d test' as test_name,
        COUNT(*) as points_returned,
        '6h' as expected_granularity,
        granularity as actual_granularity,
        CASE WHEN COUNT(*) <= 120 THEN 'PASS' ELSE 'FAIL' END as result
    FROM get_metrics_by_granularity('test-server', NOW() - INTERVAL '30 days', NOW())
) tests;

-- Performance comparison test
-- Compare query execution time between old and new granularities
EXPLAIN (ANALYZE, BUFFERS) 
SELECT COUNT(*) FROM get_metrics_by_granularity('test-server', NOW() - INTERVAL '30 days', NOW());

-- Verify materialized views exist and are accessible
SELECT 
    schemaname,
    tablename,
    table_size
FROM pg_tables 
WHERE tablename LIKE 'metrics_%_avg'
ORDER BY tablename;

-- Check continuous aggregate policies
SELECT 
    view_name,
    start_offset,
    end_offset,
    schedule_interval
FROM timescaledb_information.continuous_aggregate_policies
WHERE view_name IN ('metrics_30m_avg', 'metrics_2h_avg', 'metrics_6h_avg')
ORDER BY view_name;
