<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-228: Union member that does not match `type` is silently ignored (204): only status is applied; both members -> only the typed one stored
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-228"></a>**S3T-TABLE-228** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Union member that does not match `type` is silently ignored (204): only status is applied; both members -> only the typed one stored**
  Baseline before each case: compaction {enabled,128,auto}, snapshots {enabled,2/24}.
  type=icebergSnapshotManagement with value {disabled, settings.icebergCompaction 256} -> HTTP 204; config
  after: {"icebergCompaction": {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB":
  128, "strategy": "auto"}}}, "icebergSnapshotManagement": {"status": "disabled", "settings":
  {"icebergSnapshotManagement": {"minSnapshotsToKeep": 2, "maxSnapshotAgeHours": 24}}}}.
  type=icebergCompaction with {disabled, settings.icebergSnapshotManagement 7/48} -> HTTP 204; config after:
  {"icebergCompaction": {"status": "disabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 128,
  "strategy": "auto"}}}, "icebergSnapshotManagement": {"status": "enabled", "settings":
  {"icebergSnapshotManagement": {"minSnapshotsToKeep": 2, "maxSnapshotAgeHours": 24}}}}. Both members under
  type=icebergCompaction {enabled, compaction 256 + snapshot 7/48} -> HTTP 204; config after:
  {"icebergCompaction": {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 256,
  "strategy": "auto"}}}, "icebergSnapshotManagement": {"status": "enabled", "settings":
  {"icebergSnapshotManagement": {"minSnapshotsToKeep": 2, "maxSnapshotAgeHours": 24}}}}. {disabled, settings
  {}} -> HTTP 204; config after: {"icebergCompaction": {"status": "disabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 128, "strategy": "auto"}}}, "icebergSnapshotManagement":
  {"status": "enabled", "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 2,
  "maxSnapshotAgeHours": 24}}}}. An unknown member (icebergFoo) or an extra key never leaves boto3
  (ClientSerializationError/None).
  - ACK: terminal_codes, custom_update · ops: PutTableMaintenanceConfiguration,
    GetTableMaintenanceConfiguration · fields: type, value.settings
  - repro: Put with each mismatched union payload (parameter_validation=False); Get after
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-226](../table-maintenance.md#s3t-table-226), [S3T-TABLE-227](../table-maintenance.md#s3t-table-227), [S3T-TABLE-229](../table-maintenance.md#s3t-table-229), [S3T-TABLE-230](../table-maintenance.md#s3t-table-230), [S3T-TABLEMAINTENANCECONFIGURATION-012](../table-maintenance.md#s3t-tablemaintenanceconfiguration-012)
    · hypotheses: H-U-043, H-U-044 · evidence: table/weird-inputs/maintenance-bounds

## Notes

H-U-044 confirmed / H-U-043 refuted: no 400 for a mismatched or empty union member; the request's status is
applied to the typed entry and the foreign member is dropped (the other type's entry is untouched). With both
members present only the member matching `type` is stored. A member unknown to the model cannot be sent
through boto3 at all (serializer KeyError), so that cell is untestable via the SDK.
