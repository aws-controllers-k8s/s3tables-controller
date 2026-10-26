<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-189: IAM explicit Deny on the record-expiration actions -> AccessDeniedException/403 for Put/Get/GetJobStatus, ahead of the table's 405 gate
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-189"></a>**S3T-TABLE-189** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **IAM explicit Deny on the record-expiration actions -> AccessDeniedException/403 for Put/Get/GetJobStatus, ahead of the table's 405 gate**
  Assumed role with Allow s3tables:* plus an explicit Deny on s3tables:PutTableRecordExpirationConfiguration,
  s3tables:GetTableRecordExpirationConfiguration and s3tables:GetTableRecordExpirationJobStatus. On an
  existing customer table, where own (admin) credentials get MethodNotAllowedException 405 at the same moment:
  GetTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus and
  PutTableRecordExpirationConfiguration(enabled, days=1) each -> AccessDeniedException (HTTP 403) with message
  'User: arn:aws:sts::<ACCOUNT>:assumed-role/<role>/<session> is not authorized to perform: s3tables:<Action>
  on resource: <table ARN>' (the action name and the full table ARN are in the message). Control calls through
  the same role: GetTable -> HTTP 200, GetTableMaintenanceConfiguration -> HTTP 200. The Deny was effective on
  the first call made after AssumeRole succeeded (AssumeRole itself needed 4 attempts over 9.1 s for the new
  role to propagate).
  - ACK: terminal_codes, requeue · ops: PutTableRecordExpirationConfiguration,
    GetTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus
  - repro: CreateRole(trust: account) + inline policy Allow s3tables:* / Deny the three record-expiration
    actions -> AssumeRole -> call the three ops on an existing table
  - measurements: assume_role_attempts=4, assume_role_wait_s=9.1, deny_effective_on_first_call=true, n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-190](../table-record-expiration.md#s3t-table-190), [S3T-TABLE-170](../table-record-expiration.md#s3t-table-170), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-002) · hypotheses: H-U-107 ·
    evidence: table/error-taxonomy/expiration-iam-deny

## Notes

H-U-107 confirmed: AccessDeniedException (HTTP 403), not the declared ForbiddenException (the model declares
only ForbiddenException for these three operations; AccessDeniedException is undeclared for them).
Authorization is evaluated before the table-eligibility gate: the same table, same moment, own credentials -> 405.
Single role, single table, n=1 per operation.
