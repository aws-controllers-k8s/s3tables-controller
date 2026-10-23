<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-205: No reset for maintenance config: Put of the default values leaves a previously sent strategy (binpack) in place; fresh table says auto
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-205"></a>**S3T-TABLE-205** `unsettable-field` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **No reset for maintenance config: Put of the default values leaves a previously sent strategy (binpack) in place; fresh table says auto**
  There is no Delete or reset input for the maintenance configuration (value {} / settings {} outcomes in the
  request-validation finding). After many mutations, Put compaction {enabled, {512}} and Put snapshot
  {enabled, {1,120}} -> HTTP 204 / HTTP 204; the resulting map {"icebergCompaction": {"status": "enabled",
  "settings": {"icebergCompaction": {"targetFileSizeMB": 512, "strategy": "binpack"}}},
  "icebergSnapshotManagement": {"status": "enabled", "settings": {"icebergSnapshotManagement":
  {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 120}}}} versus the fresh table F: differs at
  icebergCompaction.settings.icebergCompaction.strategy: reset="binpack" fresh="auto".
  - ACK: custom_delete, late_initialize · ops: PutTableMaintenanceConfiguration,
    GetTableMaintenanceConfiguration · fields: configuration
  - repro: mutate T many times; Put defaults for both types; diff Get(T) against Get(fresh F)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-294](../table-maintenance.md#s3t-table-294), [S3T-TABLE-295](../table-maintenance.md#s3t-table-295), [S3T-TABLE-296](../table-maintenance.md#s3t-table-296), [S3T-TABLE-298](../table-maintenance.md#s3t-table-298), [S3T-TABLE-197](../table-maintenance.md#s3t-table-197), [S3T-TABLE-201](../table-maintenance.md#s3t-table-201),
    [S3T-TABLE-202](../table-maintenance.md#s3t-table-202), [S3T-TABLE-199](../table-maintenance.md#s3t-table-199), [S3T-TABLE-212](../table-maintenance.md#s3t-table-212), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183), [S3T-TABLE-200](../table-maintenance.md#s3t-table-200), [S3T-TABLE-203](../table-maintenance.md#s3t-table-203),
    [S3T-TABLEMAINTENANCECONFIGURATION-012](../table-maintenance.md#s3t-tablemaintenanceconfiguration-012), [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001), [S3T-TABLE-232](../table-maintenance.md#s3t-table-232), [S3T-TABLE-206](../table-maintenance.md#s3t-table-206)
    · hypotheses: H-U-049 · evidence: table/mutation-matrix/maintenance-config

## Notes

H-U-049 qualified: Put {512} does not reset strategy because of merge semantics; equality with a fresh table
requires explicitly sending strategy=auto (which is also what a fresh table reports). Snapshot defaults 1/120
enabled round-trip identically.
