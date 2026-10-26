<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-182: After RenameTable the policy and maintenance config follow the table (new name returns them); old name -> NotFoundException/404...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-182"></a>**S3T-TABLE-182** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **After RenameTable the policy and maintenance config follow the table (new name returns them); old name -> NotFoundException/404...**
  With policy Sid NS1 and compaction targetFileSizeMB=128 stored: RenameTable(same ns, new name) -> HTTP 204.
  Under the OLD name: GetTablePolicy NotFoundException/404, GetTableMaintenanceConfiguration
  NotFoundException/404, GetTableMaintenanceJobStatus NotFoundException/404. Under the NEW name: policy Sid
  NS1, compaction 128 MB, job status HTTP 200. RenameTable into ns2 with a third name -> HTTP 204: policy Sid
  NS1, compaction 128. Rename back to the original ns1/name -> HTTP 204: policy Sid NS1, compaction 128.
  tableARN values across the 4 GetTable snapshots: 1 distinct;
  GetTable/GetTableMaintenanceConfiguration/GetTableMaintenanceJobStatus echo the same tableARN.
  - ACK: is_arn_primary_key, custom_find · ops: RenameTable, GetTablePolicy, GetTableMaintenanceConfiguration,
    GetTableMaintenanceJobStatus, GetTable · fields: namespace, name, tableARN
  - repro: PutTablePolicy + PutTableMaintenanceConfiguration -> RenameTable (same ns) -> Gets under old/new ->
    RenameTable (other ns) -> rename back
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-178](../table-maintenance.md#s3t-table-178), [S3T-TABLE-179](../table-maintenance.md#s3t-table-179), [S3T-TABLE-180](../table-maintenance.md#s3t-table-180), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183) · hypotheses: H-U-021, H-U-022,
    H-U-082 · evidence: table/identity/subresource-identifiers

## Notes

H-U-021 confirmed, H-U-022 refuted; H-U-082 confirmed (tableARN identical across the three Gets and unchanged
by renames and maintenance Put).

Full original title: After RenameTable the policy and maintenance config follow the table (new name returns
them); old name -> NotFoundException/404; tableARN unchanged across renames
