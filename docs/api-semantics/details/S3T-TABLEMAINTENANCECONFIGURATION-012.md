<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEMAINTENANCECONFIGURATION-012: Maintenance defaults CAN be restored: Put auto/512 and 1/120 read back fresh-identical; status-only re-Put is a no-op but bumps modifiedAt
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablemaintenanceconfiguration-012"></a>**S3T-TABLEMAINTENANCECONFIGURATION-012** `server-default` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **Maintenance defaults CAN be restored: Put auto/512 and 1/120 read back fresh-identical; status-only re-Put is a no-op but bumps modifiedAt**
  Fresh table: compaction {"settings": {"icebergCompaction": {"strategy": "auto", "targetFileSizeMB": 512}},
  "status": "enabled"}, snapshots {"settings": {"icebergSnapshotManagement": {"maxSnapshotAgeHours": 120,
  "minSnapshotsToKeep": 1}}, "status": "enabled"}. Put {binpack,128} -> HTTP 204, reads {"settings":
  {"icebergCompaction": {"strategy": "binpack", "targetFileSizeMB": 128}}, "status": "enabled"}. Put {strategy
  auto} only -> HTTP 204, reads {"settings": {"icebergCompaction": {"strategy": "auto", "targetFileSizeMB":
  128}}, "status": "enabled"}. Put {binpack,128} again then Put {auto, 512} (the fresh defaults) -> HTTP 204,
  reads {"settings": {"icebergCompaction": {"strategy": "auto", "targetFileSizeMB": 512}}, "status":
  "enabled"} (identical to fresh: True). Three Put {status enabled} only on that state -> HTTP 204; compaction
  changed: False, GetTable.modifiedAt changed: True. Put value {} -> HTTP 204, reads {"settings":
  {"icebergCompaction": {"strategy": "auto", "targetFileSizeMB": 512}}, "status": "enabled"}. Snapshots Put
  {5,48} -> {"settings": {"icebergSnapshotManagement": {"maxSnapshotAgeHours": 48, "minSnapshotsToKeep": 5}},
  "status": "enabled"} then Put {1,120} -> reads {"settings": {"icebergSnapshotManagement":
  {"maxSnapshotAgeHours": 120, "minSnapshotsToKeep": 1}}, "status": "enabled"} (identical to fresh: True). Put
  compaction {status disabled} -> reads {"settings": {"icebergCompaction": {"strategy": "auto",
  "targetFileSizeMB": 512}}, "status": "disabled"}; then {status enabled} -> reads {"settings":
  {"icebergCompaction": {"strategy": "auto", "targetFileSizeMB": 512}}, "status": "enabled"} (settings survive
  the flip: True).
  - ACK: late_initialize, compare.nil_equals_zero_value, custom_update · ops:
    PutTableMaintenanceConfiguration, GetTableMaintenanceConfiguration, GetTable · fields:
    icebergCompaction.strategy, icebergCompaction.targetFileSizeMB, icebergSnapshotManagement
  - repro: fresh table -> Put binpack/128 -> Put strategy auto -> Put binpack/128 -> Put auto/512 -> 3x Put
    status-only -> Put {} -> snapshots 5/48 -> 1/120 -> disabled -> enabled; Get after each
  - measurements: n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-205](../table-maintenance.md#s3t-table-205), [S3T-TABLE-203](../table-maintenance.md#s3t-table-203), [S3T-TABLE-202](../table-maintenance.md#s3t-table-202), [S3T-TABLE-206](../table-maintenance.md#s3t-table-206), [S3T-TABLE-227](../table-maintenance.md#s3t-table-227), [S3T-TABLE-228](../table-maintenance.md#s3t-table-228),
    [S3T-TABLE-230](../table-maintenance.md#s3t-table-230), [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001), [S3T-TABLE-232](../table-maintenance.md#s3t-table-232), [S3T-TABLE-319](../table-encryption-storage.md#s3t-table-319), [S3T-TABLE-265](../table-maintenance.md#s3t-table-265) ·
    evidence: tablemaintenanceconfiguration/creative/reset-to-default

## Notes

qualifies [S3T-TABLE-205](../table-maintenance.md#s3t-table-205) ('no reset'): the explicit enum value 'auto' is accepted and applied (alone it keeps
the stored 128 MB; with 512 MB the read-back equals the fresh table byte-for-byte), so a controller CAN
converge a removed spec field by sending the fresh-table defaults; there is still no delete/unset call.
Extends [S3T-TABLE-206](../table-maintenance.md#s3t-table-206): a status-only Put that changes nothing still advances GetTable.modifiedAt (3/3), so a
controller that re-Puts status on every reconcile looks like a writer to anyone watching modifiedAt - gate the
Put on an actual diff. Settings survive a disabled/enabled flip. Single run.

Full original title: Maintenance defaults CAN be restored: Put strategy=auto/512 and snapshots 1/120 read back
fresh-identical; status-only re-Put is a no-op but bumps modifiedAt
