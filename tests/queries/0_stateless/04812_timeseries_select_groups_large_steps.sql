-- A legitimate timeSeriesSelect*Groups state can have more than 1,000,000 time steps:
-- add/serialize do not cap the grid size, and deserialize must round-trip the same states
-- (including through a MergeTree table and the query-time merge path).

SET allow_experimental_time_series_aggregate_functions = 1;

SELECT '-- timeSeriesSelect*Groups: more than 1,000,000 time steps round-trip';

-- add() itself has no million-step cap.
SELECT length(res), length(res[1].2)
FROM
(
    SELECT finalizeAggregation(initializeAggregation(
        'timeSeriesSelectTopKGroupsState',
        1::UInt64,
        arrayResize([1.], 1000001),
        1::UInt8)) AS res
);

-- serialize then deserialize via the AggregateFunction text/binary state encoding.
WITH initializeAggregation(
        'timeSeriesSelectTopKGroupsState',
        1::UInt64,
        arrayResize([1.], 1000001),
        1::UInt8) AS st
SELECT length(finalizeAggregation(CAST(unhex(hex(st)), 'AggregateFunction(timeSeriesSelectTopKGroups, UInt64, Array(Float64), UInt8)'))[1].2);

DROP TABLE IF EXISTS topk_large_steps;
CREATE TABLE topk_large_steps
(
    part UInt8,
    st AggregateFunction(timeSeriesSelectTopKGroups, UInt64, Array(Float64), UInt8)
)
ENGINE = MergeTree
ORDER BY part;

INSERT INTO topk_large_steps SELECT 0, timeSeriesSelectTopKGroupsState(1::UInt64, arrayResize([1.], 1000001), 1::UInt8);
INSERT INTO topk_large_steps SELECT 1, timeSeriesSelectTopKGroupsState(2::UInt64, arrayResize([2.], 1000001), 1::UInt8);

-- Query-time merge deserializes each part's state. With k = 1 the greater series wins at every step.
SELECT length(res), length(res[1].2), res[1].1
FROM
(
    SELECT timeSeriesSelectTopKGroupsMerge(st) AS res
    FROM topk_large_steps
);

DROP TABLE topk_large_steps;
