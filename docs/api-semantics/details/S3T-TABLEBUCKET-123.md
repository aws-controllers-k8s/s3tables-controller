<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-123: TagResource at +0 ms and PutTableBucketEncryption at +69-142 ms after CreateTableBucket -> 5/5 first attempts 200, no ConflictException
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-123"></a>**S3T-TABLEBUCKET-123** `eventual-consistency` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **TagResource at +0 ms and PutTableBucketEncryption at +69-142 ms after CreateTableBucket -> 5/5 first attempts 200, no ConflictException** (hypothesis refuted; behavior confirmed)
  5 fresh buckets created without the probe tag (CreateTableBucket latency 1561-1823 ms); TagResource({t:1})
  was sent [0, 0, 0, 0, 0] ms after the create response and PutTableBucketEncryption(AES256) right after it
  ([93, 69, 142, 74, 76] ms): first-attempt outcomes ['HTTP 200'] and ['HTTP 200'] (messages none).
  ListTagsForResource right after showed the tag in 5/5 and GetTableBucketEncryption showed AES256 in 5/5.
  - ACK: none, tags.custom-sync · ops: CreateTableBucket, TagResource, PutTableBucketEncryption,
    ListTagsForResource, GetTableBucketEncryption · fields: tags, encryptionConfiguration.sseAlgorithm
  - repro: CreateTableBucket(no tags) -> immediately TagResource({t:1}) -> immediately
    PutTableBucketEncryption(AES256) -> ListTags + GetEncryption; x5
  - measurements: tag_offset_ms=[0, 0, 0, 0, 0], encryption_offset_ms=[93, 69, 142, 74, 76],
    tag_first_attempt_ok=5, encryption_first_attempt_ok=5, create_latency_ms=[1562, 1600, 1646, 1823, 1561],
    n=5
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-112](../tablebucket-encryption-storage.md#s3t-tablebucket-112), [S3T-TABLEBUCKET-073](../tablebucket-encryption-storage.md#s3t-tablebucket-073), [S3T-TABLEBUCKET-008](../tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLEBUCKET-009](../tablebucket-encryption-storage.md#s3t-tablebucket-009),
    [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLE-061](../table-encryption-storage.md#s3t-table-061) · hypotheses: H-B-018 · evidence:
    tablebucket/consistency-windows/create-read-paths-tags-n5

## Notes

H-B-018 (409 'conflict with a previous write' window under 2 s for writes right after Create) refuted for
TagResource and PutTableBucketEncryption (0 retries needed); extends [S3T-TABLEBUCKET-112](../tablebucket-encryption-storage.md#s3t-tablebucket-112)
(metrics/policy/maintenance/storage class) to the two remaining write paths. ACK's create-then-TagResource
pattern needs no requeue on a fresh bucket.

Full original title: TagResource at +0-0 ms and PutTableBucketEncryption at +69-142 ms after
CreateTableBucket: 5/5 and 5/5 first attempts succeeded, no ConflictException
