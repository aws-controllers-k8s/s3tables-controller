<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-077: Maintenance type variants (IcebergUnreferencedFileRemoval, icebergCompaction, icebergSnapshotManagement, ALL-CAPS, foo) -> Validation 400
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-077"></a>**S3T-TABLEBUCKET-077** `request-validation` · impact medium · handled · verified 2026-10-13
  **Maintenance type variants (IcebergUnreferencedFileRemoval, icebergCompaction, icebergSnapshotManagement, ALL-CAPS, foo) -> Validation 400**
  type 'IcebergUnreferencedFileRemoval' -> ValidationException (HTTP 400) '1 validation error detected: Value
  at 'type' failed to satisfy constraint: Member must satisfy enum value set:
  [icebergUnreferencedFileRemoval]'; type 'icebergCompaction' -> ValidationException (HTTP 400) '1 validation
  error detected: Value at 'type' failed to satisfy constraint: Member must satisfy enum value set:
  [icebergUnreferencedFileRemoval]'; type 'icebergSnapshotManagement' -> ValidationException (HTTP 400) '1
  validation error detected: Value at 'type' failed to satisfy constraint: Member must satisfy enum value set:
  [icebergUnreferencedFileRemoval]'; type 'ICEBERGUNREFERENCEDFILEREMOVAL' -> ValidationException (HTTP 400)
  '1 validation error detected: Value at 'type' failed to satisfy constraint: Member must satisfy enum value
  set: [icebergUnreferencedFileRemoval]'; type 'foo' -> ValidationException (HTTP 400) '1 validation error
  detected: Value at 'type' failed to satisfy constraint: Member must satisfy enum value set:
  [icebergUnreferencedFileRemoval]'. Get before and after -> enabled/8/8 / enabled/8/8 (stored configuration
  unchanged).
  - ACK: terminal_codes, none · ops: PutTableBucketMaintenanceConfiguration · fields: type
  - repro: Put value {status: enabled} with each type string above -> Get
  - handling: handled via `generator.yaml:239-262; generator.yaml:269-275`
  - related: [S3T-TABLEBUCKET-074](../tablebucket-maintenance-metrics.md#s3t-tablebucket-074), [S3T-TABLEBUCKET-075](../tablebucket-maintenance-metrics.md#s3t-tablebucket-075), [S3T-TABLEBUCKET-076](../tablebucket-maintenance-metrics.md#s3t-tablebucket-076), [S3T-TABLEBUCKET-080](../tablebucket-maintenance-metrics.md#s3t-tablebucket-080) · hypotheses:
    H-S-032 · evidence: tablebucket/weird-inputs/maintenance-validation

## Notes

H-S-032 qualified: all four non-bucket type strings (plus the all-caps variant) are rejected with HTTP 400 and
the enum message 'Value at 'type' failed to satisfy constraint: Member must satisfy enum value set:
[icebergUnreferencedFileRemoval]', not 404; the wire code is ValidationException rather than the modeled
BadRequestException.

Full original title: Maintenance type path variants (IcebergUnreferencedFileRemoval, icebergCompaction,
icebergSnapshotManagement, foo) -> ValidationException 400, config unchanged
