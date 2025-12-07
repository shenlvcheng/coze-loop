CREATE TABLE IF NOT EXISTS `dataset_item_snapshot` (
    `version_id` String,
    `item_id` String,
    `sync_ck_date` Date,
    `float_map` Map(String, Float64),
    `int_map` Map(String, Int64),
    `bool_map` Map(String, Int8),
    `string_map` Map(String, String),
    `created_at` DateTime,
    `updated_at` DateTime
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (version_id, item_id, sync_ck_date)
PARTITION BY toYYYYMM(sync_ck_date);
