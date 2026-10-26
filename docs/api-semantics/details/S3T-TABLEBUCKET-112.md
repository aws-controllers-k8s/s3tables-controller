<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-112: Sub-resource Puts sent <1 ms after CreateTableBucket returned: 48/48 first attempts succeeded on 12 fresh buckets (metrics/policy/maint/SC)
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-112"></a>**S3T-TABLEBUCKET-112** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Sub-resource Puts sent <1 ms after CreateTableBucket returned: 48/48 first attempts succeeded on 12 fresh buckets (metrics/policy/maint/SC)** (hypothesis refuted; behavior confirmed)
  12 x CreateTableBucket (latency 1435-1940 ms) each immediately followed by
  PutTableBucketMetricsConfiguration, PutTableBucketPolicy, PutTableBucketMaintenanceConfiguration and
  PutTableBucketStorageClass (first Put sent 0-0 ms after the create response): first-attempt codes
  {'metrics': ['HTTP 204'], 'policy': ['HTTP 200'], 'maintenance': ['HTTP 204'], 'storage_class': ['HTTP
  200']}. The Get of each noun right after the Puts reflected the written value in all 12 buckets.
  - ACK: none · ops: CreateTableBucket, PutTableBucketMetricsConfiguration, PutTableBucketPolicy,
    PutTableBucketMaintenanceConfiguration, PutTableBucketStorageClass · fields: tableBucketARN
  - repro: loop: CreateTableBucket -> four sub-resource Puts with no delay -> four Gets -> DeleteTableBucket
  - measurements: first_put_offset_ms_min_max=[0, 0], put_success_at_ms_from_create.metrics=[70, 128],
    put_success_at_ms_from_create.policy=[175, 248], put_success_at_ms_from_create.maintenance=[550, 642],
    put_success_at_ms_from_create.storage_class=[564, 655], create_latency_ms_min_max=[1435, 1940], n=12
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-056](../tablebucket-maintenance-metrics.md#s3t-tablebucket-056), [S3T-TABLEBUCKET-059](../tablebucket-encryption-storage.md#s3t-tablebucket-059), [S3T-TABLEBUCKET-058](../tablebucket-encryption-storage.md#s3t-tablebucket-058) · hypotheses: H-S-094 · evidence:
    tablebucket/consistency-windows/create-then-subresource-put

## Notes

H-S-094 (contrarian: some fraction of sub-resource Puts within ~100 ms of create fail with 404) refuted in 12
buckets x 4 nouns (0 failures, 0 retries); the Get of each noun immediately after the Puts reflected the
written value on every bucket. Sequential buckets, one at a time, single client, us-west-2.

Full original title: Sub-resource Puts sent < 1 ms after CreateTableBucket returned: 48 of 48 first attempts
succeeded on 12 fresh buckets (metrics, policy, maintenance, storage class)
