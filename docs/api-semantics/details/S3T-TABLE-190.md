<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-190: Under an explicit Deny: ghost-UUID ARN and days=0 -> AccessDeniedException 403 (authz before lookup/validation); bucket ARN -> 400 first
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-190"></a>**S3T-TABLE-190** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Under an explicit Deny: ghost-UUID ARN and days=0 -> AccessDeniedException 403 (authz before lookup/validation); bucket ARN -> 400 first**
  Same denied role: Put(enabled, days=1) on <bucketARN>/table/<random uuid> -> AccessDeniedException (HTTP
  403); Get on that ghost ARN -> AccessDeniedException 403 (own credentials: NotFoundException 404 'The
  specified table does not exist.'). Put with value.settings.days=0 on the existing table ->
  AccessDeniedException 403 (own credentials: ValidationException 400 about the minimum). Get with the table
  BUCKET ARN as tableArn -> BadRequestException (HTTP 400) 'The specified table ARN is not valid.' under the
  denied role too (own credentials: the same 400), so the ARN-shape check runs before authorization while the
  existence lookup and the days range validation run after it.
  - ACK: terminal_codes · ops: PutTableRecordExpirationConfiguration, GetTableRecordExpirationConfiguration ·
    fields: tableArn, value.settings.days
  - repro: Denied role -> Put/Get with a ghost UUID ARN, with days=0 (parameter_validation=False client), with
    the bucket ARN
  - measurements: n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-189](../table-record-expiration.md#s3t-table-189), [S3T-TABLE-170](../table-record-expiration.md#s3t-table-170), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-002), [S3T-TABLE-243](../table-record-expiration.md#s3t-table-243),
    [S3T-NAMESPACE-030](../namespace.md#s3t-namespace-030), [S3T-TABLE-179](../table-maintenance.md#s3t-table-179) · hypotheses: H-U-107, H-U-109, H-U-112 · evidence:
    table/error-taxonomy/expiration-iam-deny

## Notes

Ordering observed: ARN-shape validation (400) -> IAM authorization (403) -> value validation
(ValidationException 400) / table lookup (404) -> table eligibility gate (405). Compare with the
own-credentials codes in table/mutation-matrix/expiration-status-days-variants and
table/identity/expiration-arn-forms-rename-delete. n=1 per cell.
