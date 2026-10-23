<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-049: DeleteTableBucketEncryption is NOT idempotent - NotFoundException 404 when never set and on a second Delete; 204 only when a config exists
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-049"></a>**S3T-TABLEBUCKET-049** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableBucketEncryption is NOT idempotent - NotFoundException 404 when never set and on a second Delete; 204 only when a config exists** (hypothesis refuted; behavior confirmed)
  On one plain bucket: DeleteTableBucketEncryption never set -> NotFoundException (HTTP 404) 'The specified
  encryption configuration does not exist.'; Get right after -> '200
  {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}'. PutTableBucketEncryption(AES256) -> HTTP 200; Get ->
  '200 {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}'. Delete -> HTTP 204; the first Get right after
  the Delete returned '200 {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}' (equal to the
  never-configured body); 4 more Gets at 1 s intervals all identical. Second Delete -> NotFoundException (HTTP 404)
  'The specified encryption configuration does not exist.'.
  - ACK: exceptions.404, custom_delete · ops: DeleteTableBucketEncryption, PutTableBucketEncryption,
    GetTableBucketEncryption · fields: encryptionConfiguration
  - repro: Fresh bucket: Delete; Put(AES256); Get; Delete; Get immediately + 4 polls at 1 s; Delete again
  - measurements: get_converged_on_first_read=true, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-005](../tablebucket-encryption-storage.md#s3t-tablebucket-005), [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLEBUCKET-051](../tablebucket-encryption-storage.md#s3t-tablebucket-051), [S3T-TABLEBUCKET-052](../tablebucket-encryption-storage.md#s3t-tablebucket-052),
    [S3T-TABLEBUCKET-072](../tablebucket-encryption-storage.md#s3t-tablebucket-072) · hypotheses: H-S-003 · evidence: tablebucket/sub-resources/encryption-get-put-delete

## Notes

H-S-003 REFUTED on the never-set/idempotency part - DeleteTableBucketEncryption returns NotFoundException 404
'The specified encryption configuration does not exist.' when nothing was ever set and again on a repeated
Delete (Delete is NOT idempotent); the convergence part held - the first Get after a successful Delete already
returned the never-configured body (200 AES256). Contrast - DeleteTableBucketPolicy IS idempotent (204 when
absent).
