<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-109: s3tables wire codes: BadRequestException=400, ConflictException=409, NotFoundException=404, ValidationException=400...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-109"></a>**S3T-TABLEBUCKET-109** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **s3tables wire codes: BadRequestException=400, ConflictException=409, NotFoundException=404, ValidationException=400...**
  45 failed calls in this probe produced 4 distinct tuples: code=BadRequestException http=400
  header-prefix=BadRequestException
  header='BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/' body.__type=None
  ops=['get_table_bucket_policy', 'put_table_bucket_policy']; code=ConflictException http=409
  header-prefix=ConflictException
  header='ConflictException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/' body.__type=None
  ops=['create_table_bucket']; code=NotFoundException http=404 header-prefix=NotFoundException
  header='NotFoundException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/' body.__type=None
  ops=['delete_table_bucket_encryption', 'get_table_bucket_policy']; code=ValidationException http=400
  header-prefix=ValidationException
  header='ValidationException:http://internal.amazon.com/coral/com.amazon.coral.validate/' body.__type=None
  ops=['put_table_bucket_storage_class']. The SDK error code is the part of x-amzn-ErrorType before the first
  ':'; the JSON body carries 'message' and no '__type'.
  - ACK: terminal_codes, exceptions.404 · ops: CreateTableBucket, DeleteTableBucketEncryption,
    GetTableBucketPolicy, PutTableBucketPolicy, PutTableBucketStorageClass
  - repro: Force a bad enum (client validation off), a malformed ARN, a missing policy, malformed policy JSON
    and a duplicate CreateTableBucket; read raw headers
  - measurements: n=45
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-067](../tablebucket-encryption-storage.md#s3t-tablebucket-067), [S3T-TABLEBUCKET-061](../tablebucket-policy.md#s3t-tablebucket-061), [S3T-TABLEBUCKET-066](../tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLEBUCKET-038](../tablebucket.md#s3t-tablebucket-038),
    [S3T-TABLEBUCKET-178](../tablebucket-policy.md#s3t-tablebucket-178), [S3T-TABLEBUCKET-065](../tablebucket-policy.md#s3t-tablebucket-065), [S3T-TABLE-083](../table-policy.md#s3t-table-083), [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLE-109](../table-policy.md#s3t-table-109), [S3T-TABLE-120](../table.md#s3t-table-120) ·
    hypotheses: H-S-072 · evidence: tablebucket/consistency-windows/subresource-read-after-write

## Notes

H-S-072 qualified: the status mapping held for the shapes reached (400/404/409 as modeled) and the code is the
bare shape name after stripping the ':http://internal.amazon.com/coral/...' suffix, but ValidationException
(400) is returned for enum/length violations and is not in the model; AccessDeniedException 403 (not
ForbiddenException) was seen for a resource-policy explicit deny in probe
weird-inputs/policy-validation-lockout.

Full original title: s3tables wire codes: BadRequestException=400, ConflictException=409,
NotFoundException=404, ValidationException=400; x-amzn-ErrorType = '<Shape>:<coral namespace URI>', body has
only 'message'; unmodeled: ValidationException
