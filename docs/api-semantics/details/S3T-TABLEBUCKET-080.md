<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-080: PutTableBucketMaintenanceConfiguration input validation returns wire code ValidationException (HTTP 400)...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-080"></a>**S3T-TABLEBUCKET-080** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketMaintenanceConfiguration input validation returns wire code ValidationException (HTTP 400)...**
  Every server-side input validation failure seen on PutTableBucketMaintenanceConfiguration (status not in
  [enabled, disabled], type not icebergUnreferencedFileRemoval, unreferencedDays/nonCurrentDays < 1, missing
  value) came back as code 'ValidationException', HTTP 400, message '1 validation error detected: Value at
  '<path>' failed to satisfy constraint: ...'. The SDK model lists BadRequestException for this operation and
  does not model ValidationException; BadRequestException (HTTP 400) was seen only for an invalid bucket name
  / region in the ARN and for the table-level Put of the bucket-only type ('Invalid table maintenance type:
  icebergUnreferencedFileRemoval').
  - ACK: terminal_codes · ops: PutTableBucketMaintenanceConfiguration, PutTableMaintenanceConfiguration ·
    fields: value.status, type, value.settings.icebergUnreferencedFileRemoval.unreferencedDays,
    value.settings.icebergUnreferencedFileRemoval.nonCurrentDays
  - repro: Put with status 'Enabled' / type 'foo' / unreferencedDays 0 (parameter_validation=False) and
    inspect the error code
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-074](../tablebucket-maintenance-metrics.md#s3t-tablebucket-074), [S3T-TABLEBUCKET-075](../tablebucket-maintenance-metrics.md#s3t-tablebucket-075), [S3T-TABLEBUCKET-076](../tablebucket-maintenance-metrics.md#s3t-tablebucket-076), [S3T-TABLEBUCKET-077](../tablebucket-maintenance-metrics.md#s3t-tablebucket-077) · hypotheses:
    H-S-106, H-S-029, H-S-032, H-S-030 · evidence: tablebucket/weird-inputs/maintenance-validation,
    tablebucket/sub-resources/maintenance-defaults-fill

## Notes

Qualifies the '400 BadRequestException' expectation in H-S-106, H-S-029, H-S-032 and H-S-030: the HTTP status
matches but the code string does not. A terminal-code list keyed on BadRequestException would miss these user
errors; match on ValidationException too (11 of 11 validation rejections in this probe).

Full original title: PutTableBucketMaintenanceConfiguration input validation returns wire code
ValidationException (HTTP 400), which is not in the modeled error list
