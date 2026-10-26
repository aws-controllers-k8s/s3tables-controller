<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEMAINTENANCECONFIGURATION-006: Untouched tables stay Not_Yet_Run (no-metadata 60 min, schema 31 min); a compaction settings Put did not trigger a run in 29 min (n=1 each)
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablemaintenanceconfiguration-006"></a>**S3T-TABLEMAINTENANCECONFIGURATION-006** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Untouched tables stay Not_Yet_Run (no-metadata 60 min, schema 31 min); a compaction settings Put did not trigger a run in 29 min (n=1 each)**
  GetTableMaintenanceJobStatus every 60 s for 60 min. t_nometa (CreateTable without metadata, never touched):
  final
  'icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run',
  transitions 0. t_schema (CreateTable with schema -> initial metadata.json, no snapshots): untouched for 31
  min with 0 transition(s); then PutTableMaintenanceConfiguration(icebergCompaction, {enabled,
  targetFileSizeMB 256, auto}) -> HTTP 204 at soak minute 31; in the remaining 29 min the job status did not
  change. Final t_schema
  'icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run'.
  Soak errors: 0.
  - ACK: synced.when, e2e-timing · ops: GetTableMaintenanceJobStatus, PutTableMaintenanceConfiguration ·
    fields: status.*.status, status.*.lastRunTimestamp
  - repro: CreateTable (with/without schema) -> leave untouched -> GetTableMaintenanceJobStatus every 60 s; at
    minute 30 Put a compaction config on the schema table
  - measurements: soak_window_min=[60.0], untouched_min=[30.8, 60.0], nudge_to_first_run_min=[], n=61
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEMAINTENANCECONFIGURATION-003](../table-maintenance.md#s3t-tablemaintenanceconfiguration-003), [S3T-TABLEMAINTENANCECONFIGURATION-004](../table-maintenance.md#s3t-tablemaintenanceconfiguration-004),
    [S3T-TABLEMAINTENANCECONFIGURATION-005](../table-maintenance.md#s3t-tablemaintenanceconfiguration-005), [S3T-TABLEMAINTENANCECONFIGURATION-007](../table-maintenance.md#s3t-tablemaintenanceconfiguration-007) · hypotheses: H-U-059,
    H-U-057 · evidence: tablemaintenanceconfiguration/state-machine/job-status-soak

## Notes

H-U-059 (Not_Yet_Run -> Successful within 24-48 h on an untouched table) qualified: nothing ran on untouched
tables inside the 60-min cap, while the table that received a metadata-only commit had all three jobs run
within 7 min (sibling finding) - job runs are commit-driven, not clock-driven, in this window. A controller's
own PutTableMaintenanceConfiguration call did not make the status leave Not_Yet_Run within 29 min. The 24-48 h
claim itself is untested. H-U-057 cross-check: see the Disabled findings.

Full original title: Untouched tables stay Not_Yet_Run (no-metadata 60 min, schema 31 min);
PutTableMaintenanceConfiguration(enabled) does NOT trigger a run (29 min)
