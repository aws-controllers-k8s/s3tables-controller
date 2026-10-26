<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-053: GetTableBucketMaintenanceConfiguration on a fresh bucket -> 200 enabled/3/10; Put {status: enabled} w/o settings -> 204, still enabled/3/10
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-053"></a>**S3T-TABLEBUCKET-053** `server-default` · impact high · handled · verified 2026-10-15, re-verified
  **GetTableBucketMaintenanceConfiguration on a fresh bucket -> 200 enabled/3/10; Put {status: enabled} w/o settings -> 204, still enabled/3/10**
  A table bucket created with name+tags only, never Put: GetTableBucketMaintenanceConfiguration returned HTTP
  200 at +0 s and +5 s (bodies identical); raw JSON top-level keys ['configuration', 'tableBucketARN'],
  configuration map keys ['icebergUnreferencedFileRemoval'], value enabled/3/10; tableBucketARN echoed
  byte-identical: True. PutTableBucketMaintenanceConfiguration(type=icebergUnreferencedFileRemoval,
  value={status: enabled}) (no settings) -> HTTP 204; Get afterwards -> enabled/3/10.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare · ops: GetTableBucketMaintenanceConfiguration,
    PutTableBucketMaintenanceConfiguration · fields: configuration, configuration{}.status,
    configuration{}.settings.icebergUnreferencedFileRemoval.unreferencedDays,
    configuration{}.settings.icebergUnreferencedFileRemoval.nonCurrentDays
  - repro: CreateTableBucket(name, tags) -> GetTableBucketMaintenanceConfiguration at +0 s and +5 s -> Put
    {status: enabled} -> Get
  - measurements: get_latency_ms=[49, 42], n=2
  - handling: handled via `generator.yaml:239-262; generator.yaml:269-275; generator.yaml:243-249; pkg/resource/table_bucket/hooks.go:268-278; pkg/resource/table_bucket/hooks.go:345-363; generator.yaml:249; generator.yaml:250-257; generator.yaml:278-283`
  - related: [S3T-TABLEBUCKET-001](../tablebucket-encryption-storage.md#s3t-tablebucket-001), [S3T-TABLEBUCKET-057](../tablebucket-maintenance-metrics.md#s3t-tablebucket-057), [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093), [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100),
    [S3T-TABLEBUCKET-158](../tablebucket.md#s3t-tablebucket-158), [S3T-TABLEBUCKET-054](../tablebucket-maintenance-metrics.md#s3t-tablebucket-054), [S3T-TABLEBUCKET-055](../tablebucket-maintenance-metrics.md#s3t-tablebucket-055), [S3T-TABLEBUCKET-076](../tablebucket-maintenance-metrics.md#s3t-tablebucket-076), [S3T-TABLEBUCKET-056](../tablebucket-maintenance-metrics.md#s3t-tablebucket-056),
    [S3T-TABLEBUCKET-170](../tablebucket-maintenance-metrics.md#s3t-tablebucket-170), [S3T-TABLEBUCKET-171](../tablebucket-maintenance-metrics.md#s3t-tablebucket-171), [S3T-TABLEBUCKET-172](../tablebucket-maintenance-metrics.md#s3t-tablebucket-172), [S3T-TABLEBUCKET-075](../tablebucket-maintenance-metrics.md#s3t-tablebucket-075), [S3T-TABLEBUCKET-173](../tablebucket-maintenance-metrics.md#s3t-tablebucket-173),
    [S3T-TABLEBUCKET-059](../tablebucket-encryption-storage.md#s3t-tablebucket-059), [S3T-TABLEBUCKET-176](../tablebucket-maintenance-metrics.md#s3t-tablebucket-176), [S3T-TABLEBUCKET-097](../tablebucket-replication.md#s3t-tablebucket-097) · hypotheses: H-S-025, H-S-026 · evidence:
    tablebucket/sub-resources/maintenance-defaults-fill

## Notes

H-S-025 confirmed (fresh read is enabled/3/10); H-S-026 refuted (the contrarian 'empty map until first Put'
did not hold). Single bucket, two reads.

Full original title: GetTableBucketMaintenanceConfiguration on a fresh bucket -> HTTP 200 enabled/3/10; Put
{status:enabled} w/o settings -> HTTP 204, then enabled/3/10
