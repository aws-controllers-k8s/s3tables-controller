<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEMAINTENANCECONFIGURATION-005: t_prop: unreferenced-file removal Not_Yet_Run -> Successful (lastRunTimestamp set) 6.8 min after a metadata-only UpdateTableMetadataLocation
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablemaintenanceconfiguration-005"></a>**S3T-TABLEMAINTENANCECONFIGURATION-005** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **t_prop: unreferenced-file removal Not_Yet_Run -> Successful (lastRunTimestamp set) 6.8 min after a metadata-only UpdateTableMetadataLocation**
  Cell t_prop (CreateTable with schema, then ONE metadata-only commit: copy of metadata.json + properties
  history.expire.max-snapshot-age-ms=1, no snapshot, UpdateTableMetadataLocation -> 200).
  GetTableMaintenanceJobStatus sampled every 60 s: icebergUnreferencedFileRemoval changed Not_Yet_Run ->
  Successful at soak minute 6.8, i.e. 6.8 min after the commit (table age 6.8 min); lastRunTimestamp
  2026-10-13T23:33:02.148Z (the run itself happened ~2026-10-13T23:33:02.148Z). The two untouched tables
  polled in the same loop stayed Not_Yet_Run. Raw body: {"status": {"icebergCompaction": {"lastRunTimestamp":
  "2026-10-13T23:29:52.066Z", "status": "Successful"}, "icebergUnreferencedFileRemoval": {"lastRunTimestamp":
  "2026-10-13T23:33:02.148Z", "status": "Successful"}, "icebergSnapshotManagement": {"lastRunTimestamp":
  "2026-10-13T23:31:07.881Z", "status": "Successful"}}, "tableARN":
  "arn:aws:s3tables:us-west-2:<account>:bucket/ackq-cf508b-mjob/table/e707e087-85b8-4393-aa1e-fd9550ae04f2"}.
  - ACK: synced.when, requeue · ops: GetTableMaintenanceJobStatus · fields: status.*.status,
    status.*.lastRunTimestamp, status.*.failureMessage
  - repro: CreateTable -> (cell setup) -> GetTableMaintenanceJobStatus every 60 s
  - measurements: commit_to_successful_seen_min=[6.8], transition_at_table_age_s=[411], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEMAINTENANCECONFIGURATION-003](../table-maintenance.md#s3t-tablemaintenanceconfiguration-003), [S3T-TABLEMAINTENANCECONFIGURATION-004](../table-maintenance.md#s3t-tablemaintenanceconfiguration-004),
    [S3T-TABLEMAINTENANCECONFIGURATION-007](../table-maintenance.md#s3t-tablemaintenanceconfiguration-007), [S3T-TABLEMAINTENANCECONFIGURATION-006](../table-maintenance.md#s3t-tablemaintenanceconfiguration-006), [S3T-TABLE-236](../table-maintenance.md#s3t-table-236),
    [S3T-TABLE-237](../table-maintenance.md#s3t-table-237), [S3T-TABLE-238](../table-maintenance.md#s3t-table-238) · hypotheses: H-U-059, H-U-060, H-U-057, H-U-062 · evidence:
    tablemaintenanceconfiguration/state-machine/job-status-soak

## Notes

Transition recorded immediately during the soak. H-U-060 (snapshot management Failed within 48 h because of
the history.expire property): the job ran within minutes and reported Successful, not Failed. H-U-059: jobs on
an empty table run when a commit happens, not on a 24-48 h clock. H-U-062 Successful cell: members status +
lastRunTimestamp, no failureMessage.

Full original title: t_prop: unreferenced-file removal Not_Yet_Run -> Successful (lastRunTimestamp set) seen
6.8 min after a metadata-only UpdateTableMetadataLocation
