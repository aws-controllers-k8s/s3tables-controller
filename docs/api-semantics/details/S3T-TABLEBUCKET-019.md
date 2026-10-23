<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-019: DeleteTableBucket on a bucket holding a table -> BadRequestException (HTTP 400); with only an empty namespace -> same code
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-019"></a>**S3T-TABLEBUCKET-019** `delete-semantics` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **DeleteTableBucket on a bucket holding a table -> BadRequestException (HTTP 400); with only an empty namespace -> same code**
  Bucket with one namespace and one ICEBERG table: DeleteTableBucket -> BadRequestException HTTP 400 'The
  bucket that you tried to delete is not empty.' (x-amzn-ErrorType
  'BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/', raw body {"message":"The
  bucket that you tried to delete is not empty."}); GetTableBucket afterwards -> HTTP 200 and GetTable -> HTTP
  200 (unchanged); three retries 2 s apart -> ["BadRequestException HTTP 400 'The bucket that you tried to
  delete is not empty.'", "BadRequestException HTTP 400 'The bucket that you tried to delete is not empty.'",
  "BadRequestException HTTP 400 'The bucket that you tried to delete is not empty.'"]. After DeleteTable (HTTP 204)
  with the empty namespace still present: DeleteTableBucket -> BadRequestException HTTP 400 'The bucket that
  you tried to delete is not empty.'. After DeleteNamespace (204) DeleteTableBucket -> HTTP 204 issued 0.549 s
  later.
  - ACK: terminal_codes, requeue, pre-delete-cleanup · ops: DeleteTableBucket, DeleteTable, DeleteNamespace,
    GetTableBucket
  - repro: CreateTableBucket -> CreateNamespace -> CreateTable(ICEBERG) -> DeleteTableBucket (x4) ->
    DeleteTable -> DeleteTableBucket -> DeleteNamespace -> DeleteTableBucket
  - measurements: retries=3, namespace_delete_to_bucket_delete_s=0.549, n=1
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-044](../tablebucket.md#s3t-tablebucket-044), [S3T-TABLEBUCKET-132](../tablebucket.md#s3t-tablebucket-132), [S3T-TABLEBUCKET-070](../tablebucket.md#s3t-tablebucket-070), [S3T-NAMESPACE-024](../namespace.md#s3t-namespace-024),
    [S3T-NAMESPACE-037](../namespace.md#s3t-namespace-037), [S3T-NAMESPACE-039](../namespace.md#s3t-namespace-039), [S3T-NAMESPACE-027](../namespace.md#s3t-namespace-027), [S3T-NAMESPACE-028](../namespace.md#s3t-namespace-028), [S3T-NAMESPACE-021](../namespace.md#s3t-namespace-021) ·
    hypotheses: H-B-022, H-B-023, H-B-024, H-B-025 · evidence:
    tablebucket/dependencies/delete-with-children-and-configs

## Notes

H-B-022 refuted (code BadRequestException); H-B-023 confirmed; H-B-024 confirmed; H-B-025 refuted. Compare the
message with the 400 for an uppercase name: BadRequestException HTTP 400 "The specified bucket name isn't
valid. Specify a different bucket name, and then try again.".

Suspected controller bug confirmed by evidence: DeleteTableBucket/DeleteNamespace on a non-empty parent return
BadRequestException 400 ('The bucket|namespace that you tried to delete is not empty.'), never
ConflictException (-019, -044, -132), so with BadRequestException in terminal_codes a TableBucket or Namespace
CR deleted before its children takes the Terminal branch instead of being retried, although a retry succeeds
0.5 s after the last child is gone (-019); treat this 400 (match the message) as retryable or hold the
parent's finalizer until the child CRs are gone.
