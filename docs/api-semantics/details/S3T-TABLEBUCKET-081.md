<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-081: 30 concurrent GetTableBucket/ListTableBuckets/ListTagsForResource/GetTableBucketPolicy (2 runs each): 0 TooManyRequests, 0 other errors
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-081"></a>**S3T-TABLEBUCKET-081** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **30 concurrent GetTableBucket/ListTableBuckets/ListTagsForResource/GetTableBucketPolicy (2 runs each): 0 TooManyRequests, 0 other errors**
  Bursts of 30 concurrent calls (20 threads) on one table bucket, two runs each 20 s apart: GetTableBucket ->
  run1 {"200": 30} (wall 0.16s), run2 {"200": 30} (wall 0.08s); ListTableBuckets -> run1 {"200": 30} (wall
  0.06s), run2 {"200": 30} (wall 0.05s); ListTagsForResource -> run1 {"200": 30} (wall 0.10s), run2 {"200":
  30} (wall 0.12s); GetTableBucketPolicy -> run1 {"200": 30} (wall 0.15s), run2 {"200": 30} (wall 0.15s). 30
  sequential GetTableBucket at 10 TPS -> {"200": 30}. No TooManyRequestsException was returned by any of the
  270 read calls.
  - ACK: requeue, e2e-timing · ops: GetTableBucket, ListTableBuckets, ListTagsForResource,
    GetTableBucketPolicy
  - repro: CreateTableBucket; 20 threads x 30 GetTableBucket (then ListTableBuckets, ListTagsForResource,
    GetTableBucketPolicy); repeat after 20 s
  - measurements: burst_wall_s.GetTableBucket=[0.16, 0.08], burst_wall_s.ListTableBuckets=[0.06, 0.05],
    burst_wall_s.ListTagsForResource=[0.1, 0.12], burst_wall_s.GetTableBucketPolicy=[0.15, 0.15],
    latency_ms_min_max.GetTableBucket=[[5,146], [7,65]], latency_ms_min_max.ListTableBuckets=[[7,44], [7,39]],
    latency_ms_min_max.ListTagsForResource=[[8,80], [10,100]],
    latency_ms_min_max.GetTableBucketPolicy=[[35,130], [29,131]], n=240
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-064](../tablebucket-policy.md#s3t-tablebucket-064), [S3T-TABLEBUCKET-108](../tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLEBUCKET-165](../tablebucket.md#s3t-tablebucket-165), [S3T-TABLEBUCKET-175](../tablebucket-maintenance-metrics.md#s3t-tablebucket-175),
    [S3T-TABLEBUCKET-018](../tablebucket.md#s3t-tablebucket-018), [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033), [S3T-TABLE-218](../table.md#s3t-table-218), [S3T-TABLEBUCKET-094](../tablebucket-maintenance-metrics.md#s3t-tablebucket-094) · hypotheses: H-B-060, H-S-073
    · evidence: tablebucket/limits/read-write-bursts-policy-size

## Notes

H-B-060 and H-S-073 refuted at 30 concurrent reads (no 429): the hypotheses posit 429s at 200 (resp. 100)
concurrent reads; the shard caps bursts at 30 calls / 5 s, so only the 30-concurrent cell was tried (twice,
with [1, 2] other probes running in this checkout). The 100/200-call cells are untested.
