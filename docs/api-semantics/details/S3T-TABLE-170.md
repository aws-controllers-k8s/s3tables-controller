<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-170: PutTableRecordExpirationConfiguration on a never-existing table UUID -> NotFoundException 404 for well-formed/{} values; days=0 -> 400 first
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-170"></a>**S3T-TABLE-170** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableRecordExpirationConfiguration on a never-existing table UUID -> NotFoundException 404 for well-formed/{} values; days=0 -> 400 first**
  tableArn = <real bucket ARN>/table/<random uuid> (never existed), region us-west-2: enabled-days-1 ->
  NotFoundException/404 'The specified table does not exist.'; enabled-days-0 -> ValidationException/400 '1
  validation error detected: Value at 'value.settings.days' failed to satisfy constraint: Member must have
  value greater than or equal to 1'; status-Disabled-pascal -> ValidationException/400 '1 validation error
  detected: Value at 'value.status' failed to satisfy constraint: Member must satisfy enum value set:
  [enabled, disabled]'; enabled-no-settings -> NotFoundException/404 'The specified table does not exist.';
  value-empty -> NotFoundException/404 'The specified table does not exist.'; status-missing ->
  NotFoundException/404 'The specified table does not exist.'.
  - ACK: exceptions.404, terminal_codes · ops: PutTableRecordExpirationConfiguration · fields: tableArn, value
  - repro: Put(value) with tableArn=<bucket ARN>/table/<uuid4> for a well-formed and several malformed values
  - measurements: n=6
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-169](../table-record-expiration.md#s3t-table-169), [S3T-TABLE-172](../table-record-expiration.md#s3t-table-172), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-002), [S3T-TABLE-189](../table-record-expiration.md#s3t-table-189),
    [S3T-TABLE-190](../table-record-expiration.md#s3t-table-190) · hypotheses: H-U-109 · evidence: table/mutation-matrix/expiration-status-days-variants

## Notes

H-U-109 (Put side) confirmed: NotFoundException (HTTP 404) 'The specified table does not exist.' for a
well-formed value, also for {} and for a value without status; days=0 and status='Disabled' get
ValidationException 400 instead, so value validation precedes the existence lookup.
GetTableRecordExpirationJobStatus on a ghost UUID is in table/identity/expiration-arn-forms-rename-delete. n=1
per variant.
