<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-183: DeleteTable with a policy and custom maintenance -> HTTP 204; sub-resource Gets at 129 ms -> NotFoundException...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-183"></a>**S3T-TABLE-183** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTable with a policy and custom maintenance -> HTTP 204; sub-resource Gets at 129 ms -> NotFoundException...**
  Table with policy Sid NS1, compaction 128 MB and snapshot management disabled (3/24):
  DeleteTable(versionToken) -> HTTP 204 with no prerequisite to remove the policy. Reads after delete:
  [{"at_ms": 129, "GetTablePolicy": {"code": "NotFoundException", "http": 404, "message": "The specified table
  does not exist."}, "GetTableMaintenanceConfiguration": {"code": "NotFoundException", "http": 404, "message":
  "The specified table does not exist."}, "GetTableMaintenanceJobStatus": {"code": "NotFoundException",
  "http": 404, "message": "The specified table does not exist."}}, {"at_ms": 154, "GetTablePolicy": {"code":
  "NotFoundException", "http": 404, "message": "The specified table does not exist."},
  "GetTableMaintenanceConfiguration": {"code": "NotFoundException", "http": 404, "message": "The specified
  table does not exist."}, "GetTableMaintenanceJobStatus": {"code": "NotFoundException", "http": 404,
  "message": "The specified table does not exist."}}, {"at_ms": 500, "GetTablePolicy": {"code":
  "NotFoundException", "http": 404, "message": "The specified table does not exist."},
  "GetTableMaintenanceConfiguration": {"code": "NotFoundException", "http": 404, "message": "The specified
  table does not exist."}, "GetTableMaintenanceJobStatus": {"code": "NotFoundException", "http": 404,
  "message": "The specified table does not exist."}}]. CreateTable with the same bucket/namespace/name -> new
  tableARN differs from the old: True. On the recreated table: GetTablePolicy -> NotFoundException/404 'The
  specified table policy does not exist.'; GetTableMaintenanceConfiguration compaction targetFileSizeMB=512,
  snapshot entry={"status": "enabled", "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 1,
  "maxSnapshotAgeHours": 120}}}; GetTableMaintenanceJobStatus statuses={"icebergCompaction": "Not_Yet_Run",
  "icebergUnreferencedFileRemoval": "Not_Yet_Run", "icebergSnapshotManagement": "Not_Yet_Run"}.
  - ACK: pre-delete-cleanup, exceptions.404, is_arn_primary_key · ops: DeleteTable, CreateTable,
    GetTablePolicy, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus
  - repro: PutTablePolicy + 2x PutTableMaintenanceConfiguration -> DeleteTable -> Gets at 0/100/500 ms ->
    CreateTable same triple -> Gets
  - measurements: post_delete_read_offsets_ms=[129, 154, 500], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-160](../table-maintenance.md#s3t-table-160), [S3T-TABLE-197](../table-maintenance.md#s3t-table-197), [S3T-TABLE-198](../table-maintenance.md#s3t-table-198), [S3T-TABLE-204](../table-maintenance.md#s3t-table-204), [S3T-TABLE-211](../table-maintenance.md#s3t-table-211), [S3T-TABLE-199](../table-maintenance.md#s3t-table-199),
    [S3T-TABLE-205](../table-maintenance.md#s3t-table-205), [S3T-TABLE-212](../table-maintenance.md#s3t-table-212), [S3T-TABLE-178](../table-maintenance.md#s3t-table-178), [S3T-TABLE-179](../table-maintenance.md#s3t-table-179), [S3T-TABLE-180](../table-maintenance.md#s3t-table-180), [S3T-TABLE-182](../table-maintenance.md#s3t-table-182), [S3T-TABLE-213](../table-maintenance.md#s3t-table-213),
    [S3T-TABLE-214](../table-policy.md#s3t-table-214), [S3T-TABLE-215](../table-maintenance.md#s3t-table-215), [S3T-TABLE-216](../table-policy.md#s3t-table-216) · hypotheses: H-U-020, H-U-030, H-U-083 · evidence:
    table/identity/subresource-identifiers

## Notes

H-U-030 confirmed (delete allowed; post-delete reads NotFoundException). H-U-020 refuted (sub-resources do not
leak on the recreated triple). H-U-083 confirmed (job status after delete NotFoundException; recreated
statuses {"icebergCompaction": "Not_Yet_Run", "icebergUnreferencedFileRemoval": "Not_Yet_Run",
"icebergSnapshotManagement": "Not_Yet_Run"}).

Full original title: DeleteTable with a policy and custom maintenance -> HTTP 204; sub-resource Gets at 129 ms
-> NotFoundException; recreated same-name table starts clean (policy 404, compaction 512 MB)
