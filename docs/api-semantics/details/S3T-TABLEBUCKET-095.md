<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-095: After DeleteTableBucketMetricsConfiguration (204) the first Get was 404 (2 of 3 reps; 3rd a 429 carry-over), no 200 in 57 polls over 3x4 s
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-095"></a>**S3T-TABLEBUCKET-095** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTableBucketMetricsConfiguration (204) the first Get was 404 (2 of 3 reps; 3rd a 429 carry-over), no 200 in 57 polls over 3x4 s**
  Put -> Delete -> poll Get every 200 ms for 4 s, repeated 3 times: first poll after the Delete at [0.003,
  0.013, 0.013] s returned ['TooManyRequestsException/429', 'NotFoundException/404', 'NotFoundException/404'];
  codes seen over all polls ['NotFoundException/404', 'TooManyRequestsException/429']; reads returning 200
  after the Delete: [[], [], []].
  - ACK: requeue · ops: DeleteTableBucketMetricsConfiguration, GetTableBucketMetricsConfiguration · fields:
    tableBucketARN
  - repro: Put metrics -> Delete -> Get every 200 ms for 4 s (x3)
  - measurements: first_poll_offset_s=[0.003, 0.013, 0.013], n=57
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093), [S3T-TABLEBUCKET-094](../tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-TABLEBUCKET-099](../tablebucket-maintenance-metrics.md#s3t-tablebucket-099), [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100),
    [S3T-TABLEBUCKET-113](../tablebucket-maintenance-metrics.md#s3t-tablebucket-113), [S3T-TABLEBUCKET-177](../tablebucket-maintenance-metrics.md#s3t-tablebucket-177), [S3T-TABLEBUCKET-096](../tablebucket-replication.md#s3t-tablebucket-096), [S3T-TABLEBUCKET-092](../tablebucket-replication.md#s3t-tablebucket-092) · hypotheses: H-S-104 ·
    evidence: tablebucket/sub-resources/metrics-lifecycle

## Notes

H-S-104 confirmed for the metrics noun: the first Get at +3-13 ms after the Delete's 204 was NotFoundException
(the one exception was a TooManyRequestsException carried over from the preceding burst, not a stale 200) and
none of 57 polls over 3 x 4 s returned the deleted configuration. Replication and encryption nouns not tested
here (replication is out of this shard; AES256 Delete/Get is indistinguishable from the default).

Full original title: After DeleteTableBucketMetricsConfiguration (204) the first Get was 404 in 2 of 3 reps
(3rd: 429 carry-over), no read flipped back to 200 in 57 polls
