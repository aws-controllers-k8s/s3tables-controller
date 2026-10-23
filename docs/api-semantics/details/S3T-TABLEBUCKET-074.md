<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-074: MaintenanceStatus reads back as lowercase 'enabled'; Put with Enabled/ENABLED/Disabled/DISABLED/on -> ValidationException (HTTP 400)...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-074"></a>**S3T-TABLEBUCKET-074** `request-validation` · impact medium · handled · verified 2026-10-13
  **MaintenanceStatus reads back as lowercase 'enabled'; Put with Enabled/ENABLED/Disabled/DISABLED/on -> ValidationException (HTTP 400)...**
  After Put {status: enabled, 3/2} (HTTP 204) the raw Get JSON carries status 'enabled'. Put with the same
  settings and status 'Enabled' -> ValidationException (HTTP 400) '1 validation error detected: Value at
  'value.status' failed to satisfy constraint: Member must satisfy enum value set: [enabled, disabled]';
  'ENABLED' -> ValidationException (HTTP 400) '1 validation error detected: Value at 'value.status' failed to
  satisfy constraint: Member must satisfy enum value set: [enabled, disabled]'; 'Disabled' ->
  ValidationException (HTTP 400) '1 validation error detected: Value at 'value.status' failed to satisfy
  constraint: Member must satisfy enum value set: [enabled, disabled]'; 'DISABLED' -> ValidationException
  (HTTP 400) '1 validation error detected: Value at 'value.status' failed to satisfy constraint: Member must
  satisfy enum value set: [enabled, disabled]'; 'on' -> ValidationException (HTTP 400) '1 validation error
  detected: Value at 'value.status' failed to satisfy constraint: Member must satisfy enum value set:
  [enabled, disabled]'. Get afterwards -> enabled/3/2 (unchanged: True).
  - ACK: terminal_codes, none · ops: PutTableBucketMaintenanceConfiguration,
    GetTableBucketMaintenanceConfiguration · fields: value.status, configuration{}.status
  - repro: Put {enabled, 3/2} -> Get -> Put with status Enabled / ENABLED / Disabled / DISABLED / on -> Get
  - handling: handled via `generator.yaml:250-257; generator.yaml:278-283`
  - related: [S3T-TABLEBUCKET-075](../tablebucket-maintenance-metrics.md#s3t-tablebucket-075), [S3T-TABLEBUCKET-076](../tablebucket-maintenance-metrics.md#s3t-tablebucket-076), [S3T-TABLEBUCKET-077](../tablebucket-maintenance-metrics.md#s3t-tablebucket-077), [S3T-TABLEBUCKET-080](../tablebucket-maintenance-metrics.md#s3t-tablebucket-080) · hypotheses:
    H-S-106 · evidence: tablebucket/weird-inputs/maintenance-validation

## Notes

H-S-106 qualified: lowercase literal on read and case variants are rejected with HTTP 400 as hypothesized, but
the wire code is ValidationException ('1 validation error detected: Value at 'value.status' failed to satisfy
constraint: Member must satisfy enum value set: [enabled, disabled]'), not the modeled BadRequestException.
Stored value unchanged after the five rejected Puts. botocore does not validate enum values client-side, so
these reached the server.

Full original title: MaintenanceStatus reads back as lowercase 'enabled'; Put with
Enabled/ENABLED/Disabled/DISABLED/on -> ValidationException (HTTP 400), not case-folded
