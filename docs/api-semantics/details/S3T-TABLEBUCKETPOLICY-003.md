<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETPOLICY-003: GetTableBucketPolicy after Put v1->v2 and after Delete, polled at 200 ms for 4 s x3: 0 stale reads in 101 polls (first read 64-68 ms)
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketpolicy-003"></a>**S3T-TABLEBUCKETPOLICY-003** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableBucketPolicy after Put v1->v2 and after Delete, polled at 200 ms for 4 s x3: 0 stale reads in 101 polls (first read 64-68 ms)** (hypothesis refuted; behavior confirmed)
  Three repetitions on one bucket: PutTableBucketPolicy v1, 2 s, PutTableBucketPolicy v2 (HTTP 200) then
  GetTableBucketPolicy every 200 ms for 4 s: all 50 reads (first at +0.064-0.068 s) carried the v2 statement
  Sid, none returned v1. DeleteTableBucketPolicy (HTTP 204) then the same poll: all 51 reads returned
  NotFoundException 404 'The specified bucket policy does not exist.', none returned the deleted policy. (Byte
  comparison of the document is not possible because the server normalizes it, see the normalization finding.)
  - ACK: none · ops: PutTableBucketPolicy, GetTableBucketPolicy, DeleteTableBucketPolicy · fields:
    resourcePolicy
  - repro: Put v1 -> 2 s -> Put v2 -> Get every 200 ms x 20 -> Delete -> Get every 200 ms x 20; x3
  - measurements: first_poll_offset_s=[0.068, 0.064, 0.064], n=101
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-108](../tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLEBUCKET-065](../tablebucket-policy.md#s3t-tablebucket-065), [S3T-TABLE-158](../table-policy.md#s3t-table-158), [S3T-TABLEBUCKETPOLICY-005](../tablebucket-policy.md#s3t-tablebucketpolicy-005),
    [S3T-TABLEBUCKETPOLICY-009](../tablebucket-policy.md#s3t-tablebucketpolicy-009), [S3T-TABLEBUCKET-064](../tablebucket-policy.md#s3t-tablebucket-064), [S3T-TABLEPOLICY-006](../table-policy.md#s3t-tablepolicy-006), [S3T-TABLE-153](../table-policy.md#s3t-table-153), [S3T-TABLE-151](../table-policy.md#s3t-table-151) ·
    hypotheses: H-S-021 · evidence: tablebucket/consistency-windows/create-then-subresource-put

## Notes

H-S-021 (contrarian: policy Get returns the old value for ~2-5 s) refuted for the policy noun in 101 polls at
200 ms; maintenance/metrics/storage-class read-after-Put showed no stale reads either
(tablebucket/sub-resources/maintenance-defaults-fill, metrics-lifecycle).

Full original title: GetTableBucketPolicy after Put v1->v2 and after Delete, polled at 200 ms for 4 s x3: 0
stale reads in 101 polls (first read 64-68 ms after the Put)
