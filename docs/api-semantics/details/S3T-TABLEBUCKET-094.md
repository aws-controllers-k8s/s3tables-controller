<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-094: 10x metrics Put,Put,Get,Delete,Get without sleep: no ConflictException, no stale read; Gets throttled 429 from the 20th Get in 1.3 s
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-094"></a>**S3T-TABLEBUCKET-094** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **10x metrics Put,Put,Get,Delete,Get without sleep: no ConflictException, no stale read; Gets throttled 429 from the 20th Get in 1.3 s**
  Ten iterations of PutTableBucketMetricsConfiguration x2, Get, Delete, Get back-to-back on one bucket (50
  calls in 1.31 s): iterations 1-9 returned {put 204, put 204, get 200, delete 204, get
  NotFoundException/404}; in iteration 10 both Gets returned TooManyRequestsException (HTTP 429, 3 ms) while
  the Puts and the Delete in the same iteration still returned 204. The first 429 came on the 20th
  GetTableBucketMetricsConfiguration within 1.29 s (37 calls on the bucket in the preceding second: 15 Put, 7
  Delete, 15 Get); the following 3 Gets within 300 ms were also 429, the Get 0.6 s later succeeded. No
  ConflictException, no stale read.
  - ACK: requeue · ops: PutTableBucketMetricsConfiguration, GetTableBucketMetricsConfiguration,
    DeleteTableBucketMetricsConfiguration · fields: tableBucketARN
  - repro: loop 10x without sleep: Put, Put, Get, Delete, Get
  - measurements: call_latency_ms_min_max=[3, 50], n=50
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093), [S3T-TABLEBUCKET-095](../tablebucket-maintenance-metrics.md#s3t-tablebucket-095), [S3T-TABLEBUCKET-099](../tablebucket-maintenance-metrics.md#s3t-tablebucket-099), [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100),
    [S3T-TABLEBUCKET-113](../tablebucket-maintenance-metrics.md#s3t-tablebucket-113), [S3T-TABLEBUCKET-165](../tablebucket.md#s3t-tablebucket-165), [S3T-TABLEBUCKET-175](../tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-081](../tablebucket-policy.md#s3t-tablebucket-081), [S3T-TABLEBUCKET-018](../tablebucket.md#s3t-tablebucket-018),
    [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033), [S3T-TABLE-218](../table.md#s3t-table-218), [S3T-TABLEBUCKET-174](../tablebucket-maintenance-metrics.md#s3t-tablebucket-174), [S3T-TABLEBUCKET-176](../tablebucket-maintenance-metrics.md#s3t-tablebucket-176), [S3T-TABLEBUCKET-092](../tablebucket-replication.md#s3t-tablebucket-092),
    [S3T-TABLEBUCKET-097](../tablebucket-replication.md#s3t-tablebucket-097) · hypotheses: H-S-045 · evidence: tablebucket/sub-resources/metrics-lifecycle

## Notes

H-S-045 confirmed for the toggle semantics (double Put ok, Delete then immediate Get 404 on the first read,
Put right after Delete ok). Surprise: only the Get was throttled (the shard brief's documented 10/s
sub-resource Get limit is consistent with ~15 Gets/s); the throttle message was 'Too Many Requests'. Single
burst, n=50 calls.

Full original title: 10x metrics Put,Put,Get,Delete,Get without sleep: no ConflictException, reads reflect the
last mutation; Gets throttled 429 from the 20th Get in 1.3 s
