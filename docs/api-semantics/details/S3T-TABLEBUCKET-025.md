<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-025: CreateTableBucket with 51 tags / aws: key / 129-char key / 257-char value -> BadRequestException 400, no bucket; 50 tags, 128/256 accepted
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-025"></a>**S3T-TABLEBUCKET-025** `request-validation` · impact high · handled · verified 2026-10-13
  **CreateTableBucket with 51 tags / aws: key / 129-char key / 257-char value -> BadRequestException 400, no bucket; 50 tags, 128/256 accepted**
  51 tags -> BadRequestException/400 'The provided tags are invalid. Number of user tags must not exceed 50';
  {aws:x, ok} -> BadRequestException/400 "Tags beginning with 'aws:' are reserved for system use."; 129-char
  key -> BadRequestException/400 'The provided tags are invalid. Tag key is empty or exceeds max length';
  257-char value -> BadRequestException/400 'The provided tags are invalid. Tag value exceeds max length'. For
  each rejected call GetTableBucket on the constructed ARN -> ["NotFoundException HTTP 404 'The specified
  bucket does not exist.'"] and ListTableBuckets(prefix) -> ['0'] items. Boundary: one 128-char key with a
  256-char value -> HTTP 200 (deleted again); exactly 50 tags -> HTTP 200 (deleted again).
  - ACK: custom_create, terminal_codes · ops: CreateTableBucket, GetTableBucket, ListTableBuckets · fields:
    tags
  - repro: CreateTableBucket(name, tags=<bad set>) -> GetTableBucket(constructed arn) ->
    ListTableBuckets(prefix=name)
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLEBUCKET-030](../tablebucket.md#s3t-tablebucket-030), [S3T-TABLEBUCKET-029](../tablebucket.md#s3t-tablebucket-029), [S3T-TABLEBUCKET-031](../tablebucket.md#s3t-tablebucket-031), [S3T-TABLEBUCKET-148](../tablebucket.md#s3t-tablebucket-148),
    [S3T-TABLEBUCKET-149](../tablebucket.md#s3t-tablebucket-149), [S3T-TABLEBUCKET-150](../tablebucket.md#s3t-tablebucket-150), [S3T-TABLE-185](../table.md#s3t-table-185), [S3T-TABLEBUCKET-152](../tablebucket.md#s3t-tablebucket-152), [S3T-TABLE-118](../table.md#s3t-table-118) · hypotheses:
    H-B-038, H-B-045 · evidence: tablebucket/weird-inputs/create-names-enums-empty-structs

## Notes

H-B-038: confirmed - create is atomic w.r.t. tags. Also covers the create-time half of H-B-045 (128/256
accepted, 129/257 rejected).

Full original title: CreateTableBucket with 51 tags / aws: key / 129-char key / 257-char value ->
BadRequestException 400 and no bucket is created; 50 tags and 128/256 limits accepted
