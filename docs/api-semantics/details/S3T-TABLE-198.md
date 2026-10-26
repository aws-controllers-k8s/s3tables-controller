<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-198: Fresh table GetTableMaintenanceJobStatus: keys ['icebergCompaction', 'icebergSnapshotManagement', 'icebergUnreferencedFileRemoval']...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-198"></a>**S3T-TABLE-198** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Fresh table GetTableMaintenanceJobStatus: keys ['icebergCompaction', 'icebergSnapshotManagement', 'icebergUnreferencedFileRemoval']...**
  GetTableMaintenanceJobStatus 619 ms after the CreateTable request was issued -> HTTP 200 with status map
  {"tableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-f1afd0-maint/table/8e81c95a-2317-46eb-8e1d-e6a7a12ea8e7",
  "status": {"icebergCompaction": {"status": "Not_Yet_Run"}, "icebergUnreferencedFileRemoval": {"status":
  "Disabled"}, "icebergSnapshotManagement": {"status": "Not_Yet_Run"}}}; polls at 0/1/5/30/60 s returned
  statuses [{"icebergCompaction": "Not_Yet_Run", "icebergUnreferencedFileRemoval": "Disabled",
  "icebergSnapshotManagement": "Not_Yet_Run"}, {"icebergCompaction": "Not_Yet_Run",
  "icebergUnreferencedFileRemoval": "Disabled", "icebergSnapshotManagement": "Not_Yet_Run"},
  {"icebergCompaction": "Not_Yet_Run", "icebergUnreferencedFileRemoval": "Disabled",
  "icebergSnapshotManagement": "Not_Yet_Run"}, {"icebergCompaction": "Not_Yet_Run",
  "icebergUnreferencedFileRemoval": "Disabled", "icebergSnapshotManagement": "Not_Yet_Run"},
  {"icebergCompaction": "Not_Yet_Run", "icebergUnreferencedFileRemoval": "Disabled",
  "icebergSnapshotManagement": "Not_Yet_Run"}]. Per-job fields present: {"icebergCompaction": ["status"],
  "icebergUnreferencedFileRemoval": ["status"], "icebergSnapshotManagement": ["status"]}
  (lastRunTimestamp/failureMessage absent). The bucket-level icebergUnreferencedFileRemoval was disabled
  before the table was created; its table-level job status reads Disabled.
  - ACK: custom_field, hook:sdk_read_one_post_set_output · ops: GetTableMaintenanceJobStatus · fields: status,
    status.icebergCompaction.status, status.icebergSnapshotManagement.status,
    status.icebergUnreferencedFileRemoval.status
  - repro: CreateTable -> GetTableMaintenanceJobStatus at 0/1/5/30/60 s
  - measurements: poll_offsets_s=[0, 1, 5, 30, 60], n=5
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-160](../table-maintenance.md#s3t-table-160), [S3T-TABLE-197](../table-maintenance.md#s3t-table-197), [S3T-TABLE-204](../table-maintenance.md#s3t-table-204), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183), [S3T-TABLE-211](../table-maintenance.md#s3t-table-211) · hypotheses: H-U-054,
    H-U-055, H-U-058, H-U-061, H-U-067 · evidence: table/mutation-matrix/maintenance-config

## Notes

H-U-054 confirmed: exactly three keys, each {status} only (no lastRunTimestamp/failureMessage),
compaction+snapshot Not_Yet_Run. H-U-055 refuted: full map at the first read, no 404 window. H-U-058 confirmed
within the first 60 s: the bucket-level disabled icebergUnreferencedFileRemoval shows as 'Disabled' in the
table-level job status while GetTableMaintenanceConfiguration never lists that type. H-U-061: raw strings
'Not_Yet_Run' and 'Disabled' (mixed case, underscores) vs lowercase 'enabled'/'disabled' in the configuration.
H-U-067 job-status half: 200 at the first read.

Full original title: Fresh table GetTableMaintenanceJobStatus: keys ['icebergCompaction',
'icebergSnapshotManagement', 'icebergUnreferencedFileRemoval'], statuses {"icebergCompaction": "Not_Yet_Run",
"icebergUnreferencedFileRemoval": "Disabled", "icebergSnapshotManagement": "Not_Yet_Run"} at 0-60 s (bucket
unreferencedFileRemoval disabled)
