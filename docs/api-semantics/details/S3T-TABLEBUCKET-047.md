<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-047: Deleted bucket name blocked with ConflictException for 604.7 s after DeleteTableBucket (n=2); re-created bucket: empty tags, default config
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-047"></a>**S3T-TABLEBUCKET-047** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Deleted bucket name blocked with ConflictException for 604.7 s after DeleteTableBucket (n=2); re-created bucket: empty tags, default config**
  2 samples: CreateTableBucket(name, tags={ghost: 1}, INTELLIGENT_TIERING) -> DeleteTableBucket (HTTP 204) ->
  CreateTableBucket(name) every 5 s (15 s after 2 min). Attempts before success returned ['ConflictException']
  with message(s) ['The bucket is in a transitional state because of a previous deletion attempt. Try again
  later.']; the first HTTP 200 came [604.7, 604.7] s after the delete. Meanwhile
  GetTableBucket/ListTableBuckets(prefix) reported [('NotFoundException', '0')]. The re-created bucket: same
  ARN [True, True], new tableBucketId [True, True], tags [{}, {}], storage class ['STANDARD', 'STANDARD'],
  encryption [{'sseAlgorithm': 'AES256'}, {'sseAlgorithm': 'AES256'}].
  - ACK: requeue, custom_create, e2e-timing, is_arn_primary_key · ops: CreateTableBucket, DeleteTableBucket,
    GetTableBucket, ListTableBuckets, ListTagsForResource, GetTableBucketStorageClass · fields: name,
    tableBucketId, tags, storageClassConfiguration
  - repro: CreateTableBucket(name, tags, INTELLIGENT_TIERING) -> DeleteTableBucket -> CreateTableBucket(name)
    every 5 s until 200 -> Get/ListTags/GetStorageClass
  - measurements: blocked_window_s=[604.7, 604.7], attempts=[57, 57], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-013](../tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-131](../tablebucket.md#s3t-tablebucket-131) · hypotheses: H-B-027, H-B-028 · evidence:
    tablebucket/identity/recreate-after-delete-window

## Notes

Refutes H-B-027 (there is a post-delete name-reservation window; see blocked_window_s). H-B-028: refuted - the
second life starts with empty tags and default storage class. Follow-up to the 60 s observation in
tablebucket/idempotency/dup-create-delete-recreate.

Full original title: Deleted bucket name is blocked with ConflictException for 604.7 s after
DeleteTableBucket; re-created bucket has fresh tags/config (empty tags)
