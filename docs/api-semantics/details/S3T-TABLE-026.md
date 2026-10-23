<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-026: GetTable(tableArn): foreign-account ARN -> ForbiddenException, own-account unknown uuid -> NotFoundException...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-026"></a>**S3T-TABLE-026** `error-code` · impact medium · handled · verified 2026-10-13
  **GetTable(tableArn): foreign-account ARN -> ForbiddenException, own-account unknown uuid -> NotFoundException...**
  GetTable(tableArn=...) on variants of a real table ARN: foreign_account_random_uuid -> ForbiddenException
  (HTTP 403) 'Unauthorized'; foreign_account_real_uuid -> ForbiddenException (HTTP 403) 'Unauthorized';
  own_account_random_uuid -> NotFoundException (HTTP 404) 'The specified table does not exist.';
  real_uuid_other_existing_bucket -> NotFoundException (HTTP 404) 'The specified table does not exist.';
  real_uuid_nonexistent_bucket -> NotFoundException (HTTP 404) 'The specified table does not exist.';
  bucket_arn_as_table_arn -> BadRequestException (HTTP 400) 'The specified table ARN is not valid.';
  not_an_arn -> BadRequestException (HTTP 400) 'The specified table ARN is not valid.'; trailing_slash ->
  BadRequestException (HTTP 400) 'The specified table ARN is not valid.'; uppercase_uuid ->
  BadRequestException (HTTP 400) 'The specified table ARN is not valid.'; uuid_only -> BadRequestException
  (HTTP 400) 'The specified table ARN is not valid.'; s3_bucket_arn -> BadRequestException (HTTP 400) 'The
  specified table ARN is not valid.'; empty_string -> ParamValidationError (HTTP None) 'Parameter validation
  failed:
  Invalid length for parameter tableArn, value: 0, valid min length: 1'.
  - ACK: terminal_codes, exceptions.404, is_arn_primary_key · ops: GetTable · fields: tableArn
  - repro: Take arn(t1); rewrite account / uuid / bucket segment / strip to bucket ARN / garbage string;
    GetTable(tableArn=variant) for each
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLE-109](../table-policy.md#s3t-table-109), [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLEBUCKET-038](../tablebucket.md#s3t-tablebucket-038),
    [S3T-TABLEBUCKET-181](../tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-131](../table-policy.md#s3t-table-131), [S3T-TABLEBUCKET-066](../tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLE-081](../table.md#s3t-table-081), [S3T-TABLEREPLICATION-023](../table-replication.md#s3t-tablereplication-023),
    [S3T-TABLEREPLICATION-024](../table-replication.md#s3t-tablereplication-024), [S3T-TABLEREPLICATION-013](../table-replication.md#s3t-tablereplication-013), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028),
    [S3T-TABLEBUCKETREPLICATION-012](../tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLE-083](../table-policy.md#s3t-table-083) · hypotheses: H-T-008, H-T-009, H-T-010 · evidence:
    table/identity/get-table-identifier-forms

## Notes

H-T-008 confirmed (foreign account -> ForbiddenException, own account random uuid -> NotFoundException).
H-T-009 confirmed (real uuid under another existing bucket -> NotFoundException; under a nonexistent bucket ->
NotFoundException). H-T-010 confirmed (bucket ARN as tableArn -> BadRequestException; 'not-an-arn' ->
BadRequestException; trailing slash -> BadRequestException). Extra cells: uppercase uuid ->
BadRequestException; uuid only -> BadRequestException; s3 bucket ARN -> BadRequestException; empty string ->
ParamValidationError.

Full original title: GetTable(tableArn): foreign-account ARN -> ForbiddenException, own-account unknown uuid
-> NotFoundException, other bucket+real uuid -> NotFoundException, malformed -> BadRequestException
