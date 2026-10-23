<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-131: A rejected CreateTableBucket (400) or CreateTable (400/500) does NOT reserve the name: the immediate same-name retry -> 200 (6/6)
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-131"></a>**S3T-TABLEBUCKET-131** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **A rejected CreateTableBucket (400) or CreateTable (400/500) does NOT reserve the name: the immediate same-name retry -> 200 (6/6)**
  CreateTableBucket failing with kms-ghost-key -> BadRequestException (HTTP 400) 'The specified KMS key ARN
  does not exist.'; 51-tags -> BadRequestException (HTTP 400) 'The provided tags are invalid. Number of user
  tags must not exceed 50'; bad-storage-class -> BadRequestException (HTTP 400) 'The specified storage class
  is not valid.'. A plain CreateTableBucket of the SAME name 1581-1610 ms later -> {'kms-ghost-key': 'HTTP
  200', '51-tags': 'HTTP 200', 'bad-storage-class': 'HTTP 200'}. CreateTable failing with schema-empty-500 ->
  InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.'; partition-bad-source-400
  -> BadRequestException (HTTP 400) 'The specified metadata is not valid'; kms-ghost-key-400 ->
  BadRequestException (HTTP 400) 'The specified KMS key ARN does not exist.'. A plain CreateTable of the SAME
  name right after -> {'schema-empty-500': 'HTTP 200', 'partition-bad-source-400': 'HTTP 200',
  'kms-ghost-key-400': 'HTTP 200'}; three more schema={} attempts -> ['InternalServerErrorException/500',
  'InternalServerErrorException/500', 'InternalServerErrorException/500']; no ghost entries in ListTables.
  - ACK: requeue, terminal_codes · ops: CreateTableBucket, CreateTable, GetTableBucket, GetTable · fields:
    name, encryptionConfiguration, tags, metadata
  - repro: CreateTableBucket(name, aws:kms ghost key | 51 tags | GLACIER) -> 400; CreateTableBucket(name) at
    once. CreateTable(name, schema={} | bad partition sourceId | ghost key) -> 500/400; CreateTable(name) at
    once.
  - measurements: bucket_retry_gap_ms=[1581, 1609, 1610], table_retry_gap_ms=[346, 387, 403], n=6
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-115](../tablebucket.md#s3t-tablebucket-115), [S3T-TABLE-133](../table-schema.md#s3t-table-133), [S3T-TABLEBUCKET-013](../tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-047](../tablebucket.md#s3t-tablebucket-047),
    [S3T-TABLEBUCKET-105](../tablebucket-encryption-storage.md#s3t-tablebucket-105), [S3T-TABLEBUCKET-128](../tablebucket-encryption-storage.md#s3t-tablebucket-128), [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107), [S3T-TABLE-085](../table-encryption-storage.md#s3t-table-085) · evidence:
    tablebucket/creative/failed-create-name-reservation

## Notes

Extends [S3T-TABLEBUCKET-115](../tablebucket.md#s3t-tablebucket-115) (10-min 'transitional state' after a DELETE) with the create-failure side: unlike
a deletion, a rejected create leaves no name reservation, so a controller may retry the same name on the next
reconcile without a backoff; extends [S3T-TABLE-133](../table-schema.md#s3t-table-133) (schema={} -> 500) with the same-name retry outcome. Single
run, one attempt per variant.
