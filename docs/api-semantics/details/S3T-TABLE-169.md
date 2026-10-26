<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-169: PutTableRecordExpirationConfiguration: days<1 or status not in [enabled, disabled] -> undeclared ValidationException 400 before table lookup
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-169"></a>**S3T-TABLE-169** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableRecordExpirationConfiguration: days<1 or status not in [enabled, disabled] -> undeclared ValidationException 400 before table lookup**
  Through a client without botocore parameter validation: value.settings.days=0 and days=-1 ->
  ValidationException (HTTP 400) '1 validation error detected: Value at 'value.settings.days' failed to
  satisfy constraint: Member must have value greater than or equal to 1'; value.status='Disabled', 'ENABLED'
  and '' -> ValidationException (HTTP 400) '1 validation error detected: Value at 'value.status' failed to
  satisfy constraint: Member must satisfy enum value set: [enabled, disabled]'. The same values sent with a
  never-existing table UUID get the same ValidationException (not 404), and on an existing customer table they
  get it instead of the table's MethodNotAllowedException 405, so this validation runs before the table is
  resolved. days as a JSON string "7" -> BadRequestException (HTTP 400) 'Invalid input. Consult the S3 Tables
  and/or Iceberg REST Catalog API reference for constraints on request parameters.' The other 12 of 18
  variants reached the 405 gate.
  - ACK: terminal_codes, custom_update · ops: PutTableRecordExpirationConfiguration · fields:
    value.settings.days, value.status
  - repro: CreateTable (customer) -> Put each value variant through a client with parameter_validation=False
    -> record code/HTTP/message
  - measurements: variants_rejected_400=6, variants_total=18, n=18
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-006](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-006), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-005),
    [S3T-TABLE-172](../table-record-expiration.md#s3t-table-172), [S3T-TABLE-170](../table-record-expiration.md#s3t-table-170), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-002) · hypotheses: H-U-112, H-U-128 ·
    evidence: table/mutation-matrix/expiration-status-days-variants

## Notes

H-U-112 qualified: days=0 is rejected with HTTP 400 and a message stating the minimum (1), but the wire code
is ValidationException, which the model does not declare (it declares BadRequestException); the 'stored
configuration unchanged' half is unobservable on customer tables (Get -> 405). H-U-128 confirmed with the same
qualification: status 'Disabled' (job-status casing) and 'ENABLED' -> ValidationException 400, not
BadRequestException. botocore's own validator rejects days<1 and days>2147483647 client-side
(ParamValidationError) but does NOT check enum values, so a controller passing a mis-cased status reaches the
service. n=1 call per variant.
