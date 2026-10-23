<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-227: Snapshot minSnapshotsToKeep/maxSnapshotAgeHours: 0 and -1 -> 400 'must be a positive integer'; no upper bound (2^31-1 ok, 2^31 clamped)
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-227"></a>**S3T-TABLE-227** `request-validation` · impact high · handled · verified 2026-10-13
  **Snapshot minSnapshotsToKeep/maxSnapshotAgeHours: 0 and -1 -> 400 'must be a positive integer'; no upper bound (2^31-1 ok, 2^31 clamped)**
  PutTableMaintenanceConfiguration icebergSnapshotManagement (minSnapshotsToKeep, maxSnapshotAgeHours) with
  SDK validation off: accepted ['1_1', '10_2500', '1000000_1000000', '2147483647_2147483647', '2147483648_1',
  '1_24', '100_1', 'min_only_3', 'max_only_48', 'empty_member']; rejected {"0_120": "BadRequestException/400
  'The specified minSnapshotsToKeep is invalid. Value must be a positive integer.'", "1_0":
  "BadRequestException/400 'The specified maxSnapshotAgeHours is invalid. Value must be a positive integer.'",
  "-1_120": "BadRequestException/400 'The specified minSnapshotsToKeep is invalid. Value must be a positive
  integer.'", "0_0": "BadRequestException/400 'The specified minSnapshotsToKeep is invalid. Value must be a
  positive integer.'"}. Read-backs: {"1_1": {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 1}, "10_2500":
  {"minSnapshotsToKeep": 10, "maxSnapshotAgeHours": 2500}, "1000000_1000000": {"minSnapshotsToKeep": 1000000,
  "maxSnapshotAgeHours": 1000000}, "2147483647_2147483647": {"minSnapshotsToKeep": 2147483647,
  "maxSnapshotAgeHours": 2147483647}, "2147483648_1": {"minSnapshotsToKeep": 2147483647,
  "maxSnapshotAgeHours": 1}, "1_24": {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 24}, "100_1":
  {"minSnapshotsToKeep": 100, "maxSnapshotAgeHours": 1}, "min_only_3": {"minSnapshotsToKeep": 3,
  "maxSnapshotAgeHours": 1}, "max_only_48": {"minSnapshotsToKeep": 3, "maxSnapshotAgeHours": 48},
  "empty_member": {"minSnapshotsToKeep": 3, "maxSnapshotAgeHours": 48}}.
  - ACK: terminal_codes, compare.nil_equals_zero_value · ops: PutTableMaintenanceConfiguration · fields:
    value.settings.icebergSnapshotManagement.minSnapshotsToKeep,
    value.settings.icebergSnapshotManagement.maxSnapshotAgeHours
  - repro: Put snapshot management with each (min,max) pair with parameter_validation=False; Get after each
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-226](../table-maintenance.md#s3t-table-226), [S3T-TABLE-228](../table-maintenance.md#s3t-table-228), [S3T-TABLE-229](../table-maintenance.md#s3t-table-229), [S3T-TABLE-230](../table-maintenance.md#s3t-table-230), [S3T-TABLEMAINTENANCECONFIGURATION-012](../table-maintenance.md#s3t-tablemaintenanceconfiguration-012)
    · hypotheses: H-U-042 · evidence: table/weird-inputs/maintenance-bounds

## Notes

H-U-042 confirmed: 0 or negative -> BadRequestException 400 'The specified minSnapshotsToKeep is invalid.
Value must be a positive integer.' (resp. maxSnapshotAgeHours); 1/1, 10/2500, 1e6/1e6 and 2147483647 accepted
and read back unchanged. Surprises: 2147483648 is accepted and stored as 2147483647 (silent clamp); min-only /
max-only / empty member Puts are accepted and keep the other field (merge semantics).
