<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-038: GetTableBucket/Delete/ListTags: missing bucket -> NotFoundException/404...
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-038"></a>**S3T-TABLEBUCKET-038** `identity` · impact high · handled · verified 2026-10-13
  **GetTableBucket/Delete/ListTags: missing bucket -> NotFoundException/404...**
  Per identifier form (code/HTTP): valid-missing -> Get NotFoundException/404 | ListTags NotFoundException/404
  | Delete NotFoundException/404; plain-name -> Get BadRequestException/400 | ListTags BadRequestException/400
  | Delete BadRequestException/400; wrong-service-s3 -> Get BadRequestException/400 | ListTags
  BadRequestException/400 | Delete BadRequestException/400; table-arn-under-bucket -> Get
  BadRequestException/400 | ListTags NotFoundException/404 | Delete BadRequestException/400; uppercase-name ->
  Get BadRequestException/400 | ListTags BadRequestException/400 | Delete BadRequestException/400;
  foreign-account -> Get ForbiddenException/403 | ListTags ForbiddenException/403 | Delete
  ForbiddenException/403; trailing-slash -> Get BadRequestException/400 | ListTags BadRequestException/400 |
  Delete BadRequestException/400; trailing-space -> Get BadRequestException/400 | ListTags
  BadRequestException/400 | Delete BadRequestException/400; leading-space -> Get BadRequestException/400 |
  ListTags BadRequestException/400 | Delete BadRequestException/400; uppercase-arn-prefix -> Get
  BadRequestException/400 | ListTags BadRequestException/400 | Delete BadRequestException/400; other-region ->
  Get BadRequestException/400 | ListTags BadRequestException/400 | Delete BadRequestException/400;
  other-partition -> Get BadRequestException/400 | ListTags BadRequestException/400 | Delete
  BadRequestException/400; empty-region -> Get BadRequestException/400 | ListTags BadRequestException/400 |
  Delete BadRequestException/400; empty-account -> Get BadRequestException/400 | ListTags
  BadRequestException/400 | Delete BadRequestException/400; underscore-name -> Get NotFoundException/404 |
  ListTags NotFoundException/404 | Delete NotFoundException/404; aws-s3-pseudo-bucket -> Get
  NotFoundException/404 | ListTags NotFoundException/404 | Delete NotFoundException/404. The real bucket was
  untouched (GetTableBucket -> HTTP 200). Messages seen: ['The provided Partition does not match the Partition
  of the resource.', 'The provided Region does not match the Region of the resource.', 'The specified bucket
  does not exist.', "The specified bucket name isn't valid. Specify a different bucket name, and then try
  again.", 'The specified resourceArn is not valid.', 'The specified table bucket ARN is not valid.', 'The
  specified table does not exist.', 'Unauthorized'].
  - ACK: exceptions.404, terminal_codes, is_arn_primary_key · ops: GetTableBucket, DeleteTableBucket,
    ListTagsForResource · fields: tableBucketARN, resourceArn
  - repro: GetTableBucket/ListTagsForResource/DeleteTableBucket with each identifier form
    (parameter_validation=False)
  - handling: handled via `generator.yaml:191-192; pkg/resource/table_bucket/sdk.go:133-155; generator.yaml:99-104; generator.yaml:170-175; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304`
  - related: [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLE-109](../table-policy.md#s3t-table-109), [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLE-026](../table-policy.md#s3t-table-026),
    [S3T-TABLEBUCKET-181](../tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-131](../table-policy.md#s3t-table-131), [S3T-TABLEBUCKET-061](../tablebucket-policy.md#s3t-tablebucket-061), [S3T-TABLEBUCKET-066](../tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLEBUCKET-178](../tablebucket-policy.md#s3t-tablebucket-178),
    [S3T-TABLEBUCKET-065](../tablebucket-policy.md#s3t-tablebucket-065), [S3T-TABLEBUCKET-109](../tablebucket-encryption-storage.md#s3t-tablebucket-109), [S3T-TABLE-081](../table.md#s3t-table-081), [S3T-TABLEBUCKET-162](../tablebucket.md#s3t-tablebucket-162), [S3T-TABLEBUCKET-163](../tablebucket-policy.md#s3t-tablebucket-163),
    [S3T-TABLEBUCKET-164](../tablebucket.md#s3t-tablebucket-164), [S3T-TABLEBUCKET-127](../tablebucket.md#s3t-tablebucket-127), [S3T-NAMESPACE-029](../namespace.md#s3t-namespace-029) · hypotheses: H-B-034, H-B-035, H-B-067 ·
    evidence: tablebucket/error-taxonomy/arn-forms-get-delete-tags

## Notes

H-B-034: confirmed. H-B-035 (foreign account): ForbiddenException/403. H-B-067: trailing slash
BadRequestException/400, trailing space BadRequestException/400, leading space BadRequestException/400,
uppercase 'ARN:AWS:S3TABLES' BadRequestException/400, uppercase name BadRequestException/400. Other region in
the ARN BadRequestException/400; other partition BadRequestException/400; empty region
BadRequestException/400; empty account BadRequestException/400; underscore name NotFoundException/404.
Distinct 400 messages: 'The specified table bucket ARN is not valid.' (Get/Delete) vs 'The specified
resourceArn is not valid.' (ListTags) for malformed ARNs; names that fail the naming rules inside a
well-formed ARN yield the Create-style message "The specified bucket name isn't valid..."; a table ARN passed
to ListTagsForResource is accepted (404 'table does not exist') but to Get/Delete is a 400. Foreign account ->
ForbiddenException 403 'Unauthorized' on all three ops.

Full original title: GetTableBucket/Delete/ListTags: missing bucket -> NotFoundException/404; non-ARN, wrong
service, table ARN, uppercase, trailing slash ->
BadRequestException/400/BadRequestException/400/NotFoundException/404; foreign account ->
ForbiddenException/403
