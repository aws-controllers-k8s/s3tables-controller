<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-212: Maintenance config always exists with server defaults and policy may exist out-of-band: a nil spec must mean 'unmanaged', not 'reset/delete'
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-212"></a>**S3T-TABLE-212** `first-sync-destructive` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Maintenance config always exists with server defaults and policy may exist out-of-band: a nil spec must mean 'unmanaged', not 'reset/delete'**
  A never-configured table already returns a full maintenance map (compaction enabled/512, snapshots
  enabled/1/120, strategy absent) and no API distinguishes 'never set' from 'user set the defaults'; an
  explicit Put of strategy makes 'auto' appear permanently (auto). A table policy is a separate sub-resource
  that returns 404 when absent. A reconcile that treats an empty spec as 'write defaults' or 'delete policy'
  alters state the user never declared.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare, custom_delete · ops:
    GetTableMaintenanceConfiguration, PutTableMaintenanceConfiguration, GetTablePolicy, DeleteTablePolicy
  - repro: derived from the fresh-table defaults and strategy findings in this probe
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-197](../table-maintenance.md#s3t-table-197), [S3T-TABLE-199](../table-maintenance.md#s3t-table-199), [S3T-TABLE-205](../table-maintenance.md#s3t-table-205), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183) · hypotheses: H-U-079 · evidence:
    table/mutation-matrix/maintenance-config

## Notes

H-U-079 confirmed as a design consequence of H-U-031/H-U-033 observations in this probe (not a separate
experiment).

Contradiction with [S3T-TABLE-197](../table-maintenance.md#s3t-table-197): 212's behavior says a never-configured table returns compaction with
'strategy absent' and that an explicit Put 'makes auto appear permanently'; 197 (direct
GetTableMaintenanceConfiguration 619 ms after CreateTable, full map) reads strategy 'auto' on the fresh table,
as do 199 ('auto is present on a never-configured table') and 205 (fresh='auto'). 212 is a derived finding
(repro: 'derived from the fresh-table defaults and strategy findings', confidence medium) that restated the
refuted hypothesis H-U-033 instead of the observation Resolution: keep both; [S3T-TABLE-197](../table-maintenance.md#s3t-table-197) is canonical for
the fresh-table strategy value ('auto' present); 212 remains valid for its design claim (no reset input, nil
spec = unmanaged)
