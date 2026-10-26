<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEMAINTENANCECONFIGURATION-013: PutTableMaintenanceConfiguration sort/z-order: no metadata location -> 400 for any caller...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablemaintenanceconfiguration-013"></a>**S3T-TABLEMAINTENANCECONFIGURATION-013** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **PutTableMaintenanceConfiguration sort/z-order: no metadata location -> 400 for any caller...**
  Assumed role allowed only s3tables:GetTable, GetTableMaintenanceConfiguration,
  PutTableMaintenanceConfiguration (Resource '*'), 3 runs. On a table created WITHOUT metadata: Put
  icebergCompaction {binpack,256} -> HTTP 204, {targetFileSizeMB 128} -> HTTP 204, {sort,128} and
  {z-order,128} -> BadRequestException (HTTP 400) 'The table metadata location is not specified.' for the role
  AND for the admin caller (6/6 each). On a table created WITH metadata.iceberg.schema (2 runs): role
  {sort}/{z-order} -> AccessDeniedException (HTTP 403) 'Insufficient access to update the table maintenance
  configuration' (a service message, not the IAM 'User: ... is not authorized to perform' form; 4/4); admin ->
  BadRequestException (HTTP 400) 'The specified compaction strategy requires sort order in the table metadata.
  Please refer to the S3 Tables documentation for steps on configuring sort order.' (4/4). Same three actions +
  s3tables:GetTableMetadataLocation -> still AccessDeniedException 403 (2/2); + s3tables:GetTableData -> the
  admin's 400 'requires sort order' (2/2, authorization passed); + both -> the same 400 (2/2). The denied Puts
  left the configuration unchanged (admin read: no-metadata table binpack/128, schema table auto/512).
  - ACK: terminal_codes, requeue, docs-only · ops: PutTableMaintenanceConfiguration,
    GetTableMaintenanceConfiguration, GetTablePolicy, DeleteTable · fields:
    value.settings.icebergCompaction.strategy, metadataLocation
  - repro: IAM roles (3 actions; +GetTableMetadataLocation; +GetTableData; +both) -> AssumeRole ->
    PutTableMaintenanceConfiguration strategy sort / z-order on a table without metadata and on one created
    with metadata.iceberg.schema; same Puts as admin; admin Get afterwards
  - measurements: runs=3, schema_table_runs=2, n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-233](../table-maintenance.md#s3t-table-233) · evidence: table/creative/reverify-set-c

## Notes

Correction of the sort/z-order clause of [S3T-TABLE-233](../table-maintenance.md#s3t-table-233) (all its other clauses reproduced 3/3: GetTable 200,
Get/Put/DeleteTablePolicy, GetTableMaintenanceJobStatus and DeleteTable -> AccessDeniedException 403 with the
IAM message naming the action and the FULL table ARN <bucketARN>/table/<uuid> as the resource). The 403 for
sort/z-order is state-dependent (only with a metadata location) and is a different message form; the missing
action is s3tables:GetTableData (the service reads the metadata.json to check for a sort order). A controller
policy for maintenance configuration therefore needs GetTableData as well when the spec uses sort or z-order.

Full original title: PutTableMaintenanceConfiguration sort/z-order: no metadata location -> 400 for any
caller; with one, role needs s3tables:GetTableData else 403
