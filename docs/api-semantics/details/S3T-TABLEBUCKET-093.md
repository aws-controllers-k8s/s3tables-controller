<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-093: GetTableBucketMetricsConfiguration.id is the constant string 'EntireBucket' across repeated Puts, Delete+Put and on a second bucket
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-093"></a>**S3T-TABLEBUCKET-093** `identity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableBucketMetricsConfiguration.id is the constant string 'EntireBucket' across repeated Puts, Delete+Put and on a second bucket**
  PutTableBucketMetricsConfiguration (no body) -> HTTP 204; three Gets returned id ['EntireBucket',
  'EntireBucket', 'EntireBucket'] (format len12:EaaaaaBaaaaa). Second Put -> HTTP 204; Get id -> EntireBucket.
  Delete -> HTTP 204; Get -> NotFoundException/404; Put immediately after Delete -> HTTP 204; Get id ->
  EntireBucket. On a second bucket Put -> HTTP 204, Get id -> EntireBucket. Delete -> HTTP 204, Delete again
  -> HTTP 204. Raw Get JSON keys: ['id', 'tableBucketARN']; tableBucketARN equals the request ARN and
  GetTableBucket.arn: True.
  - ACK: is_read_only, compare.is_ignored+delta_pre_compare · ops: PutTableBucketMetricsConfiguration,
    GetTableBucketMetricsConfiguration, DeleteTableBucketMetricsConfiguration · fields: id, tableBucketARN
  - repro: Put -> Get x3 -> Put -> Get -> Delete -> Get -> Put -> Get; Put+Get on a second bucket; Delete
    twice
  - measurements: put_latency_ms=[45, 44, 48, 70], get_latency_ms=[12, 13, 12, 12, 13, 12], n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-001](../tablebucket-encryption-storage.md#s3t-tablebucket-001), [S3T-TABLEBUCKET-053](../tablebucket-maintenance-metrics.md#s3t-tablebucket-053), [S3T-TABLEBUCKET-057](../tablebucket-maintenance-metrics.md#s3t-tablebucket-057), [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100),
    [S3T-TABLEBUCKET-158](../tablebucket.md#s3t-tablebucket-158), [S3T-TABLEBUCKET-094](../tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-TABLEBUCKET-095](../tablebucket-maintenance-metrics.md#s3t-tablebucket-095), [S3T-TABLEBUCKET-099](../tablebucket-maintenance-metrics.md#s3t-tablebucket-099), [S3T-TABLEBUCKET-113](../tablebucket-maintenance-metrics.md#s3t-tablebucket-113),
    [S3T-TABLEBUCKET-082](../tablebucket-encryption-storage.md#s3t-tablebucket-082), [S3T-TABLEBUCKET-174](../tablebucket-maintenance-metrics.md#s3t-tablebucket-174), [S3T-TABLEBUCKET-175](../tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-176](../tablebucket-maintenance-metrics.md#s3t-tablebucket-176), [S3T-TABLEBUCKET-092](../tablebucket-replication.md#s3t-tablebucket-092),
    [S3T-TABLEBUCKET-097](../tablebucket-replication.md#s3t-tablebucket-097), [S3T-TABLEBUCKET-177](../tablebucket-maintenance-metrics.md#s3t-tablebucket-177), [S3T-TABLEBUCKET-096](../tablebucket-replication.md#s3t-tablebucket-096) · hypotheses: H-S-044, H-S-092, H-S-097,
    H-S-101 · evidence: tablebucket/sub-resources/metrics-lifecycle

## Notes

H-S-044 confirmed (constant, 'EntireBucket'-style literal; Put returns HTTP 204, not 200). H-S-092 refuted:
the id is unchanged after Delete -> Put. H-S-097 qualified: id is present and non-empty after every Put and
cannot be set, but it is NOT unique per bucket (identical 'EntireBucket' on both buckets). H-S-101 confirmed:
raw JSON has exactly ['id', 'tableBucketARN'] and the ARN equals the request ARN and GetTableBucket.arn
byte-for-byte. One bucket pair, 7 Gets.

Contradiction with [S3T-TABLEBUCKET-001](../tablebucket-encryption-storage.md#s3t-tablebucket-001): 001 ends 'No sub-resource read returned NotFoundException for an unset
setting' (6 Get* reads: bucket encryption/maintenance, table encryption/storage class/metadata
location/maintenance); 093 and 100 show GetTableBucketMetricsConfiguration on a bucket with no metrics
configuration -> NotFoundException 404 'No such metric configuration exists', and 158 shows
GetTableBucketPolicy 404 on a fresh re-created bucket - the 'none 404' statement holds only for the six nouns
001 read Resolution: keep both; [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093) is canonical for the metrics noun; 001 over-generalizes in
its closing sentence, see title_fixes
