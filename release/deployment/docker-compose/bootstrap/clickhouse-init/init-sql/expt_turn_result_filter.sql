CREATE TABLE IF NOT EXISTS `expt_turn_result_filter` (
    `space_id` String,
    `expt_id` String,
    `item_id` String,
    `item_idx` Int32,
    `turn_id` String,
    `status` Int32,
    `eval_target_data` Map(String, String),
    `evaluator_score` Map(String, Float64),
    `annotation_float` Map(String, Float64),
    `annotation_bool` Map(String, Int8),
    `annotation_string` Map(String, String),
    `evaluator_score_corrected` Int32,
    `eval_set_version_id` String,
    `created_date` Date,
    `updated_at` DateTime,
    `created_at` DateTime
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (space_id, expt_id, item_id, turn_id, created_date)
PARTITION BY toYYYYMM(created_date);
