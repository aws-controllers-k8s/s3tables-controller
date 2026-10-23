<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-204: Put compaction strategy sort/z-order on a table without metadata -> BadRequestException 400 'The table metadata location is not specified.'
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-204"></a>**S3T-TABLE-204** `prerequisite` · impact medium · handled · verified 2026-10-13
  **Put compaction strategy sort/z-order on a table without metadata -> BadRequestException 400 'The table metadata location is not specified.'**
  Table created with format=ICEBERG and no metadata (no schema, no sort order): Put {enabled, {strategy sort}}
  -> BadRequestException (HTTP 400) 'The table metadata location is not specified.', Get settings
  {"icebergCompaction": {"targetFileSizeMB": 128, "strategy": "binpack"}}. Put {enabled, {strategy z-order}}
  -> BadRequestException (HTTP 400) 'The table metadata location is not specified.', Get settings
  {"icebergCompaction": {"targetFileSizeMB": 128, "strategy": "binpack"}}. GetTableMaintenanceJobStatus right
  after: {"icebergCompaction": "Not_Yet_Run", "icebergUnreferencedFileRemoval": "Disabled",
  "icebergSnapshotManagement": "Disabled"}.
  - ACK: terminal_codes · ops: PutTableMaintenanceConfiguration, GetTableMaintenanceJobStatus · fields:
    value.settings.icebergCompaction.strategy
  - repro: fresh ICEBERG table without metadata; Put compaction strategy sort; Put strategy z-order; Get;
    GetTableMaintenanceJobStatus
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-297](../table-maintenance.md#s3t-table-297), [S3T-TABLE-294](../table-maintenance.md#s3t-table-294), [S3T-TABLE-232](../table-maintenance.md#s3t-table-232), [S3T-TABLE-230](../table-maintenance.md#s3t-table-230), [S3T-TABLE-234](../table-maintenance.md#s3t-table-234), [S3T-TABLE-160](../table-maintenance.md#s3t-table-160),
    [S3T-TABLE-197](../table-maintenance.md#s3t-table-197), [S3T-TABLE-198](../table-maintenance.md#s3t-table-198), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183), [S3T-TABLE-211](../table-maintenance.md#s3t-table-211), [S3T-TABLE-199](../table-maintenance.md#s3t-table-199), [S3T-TABLE-203](../table-maintenance.md#s3t-table-203), [S3T-TABLE-016](../table-schema.md#s3t-table-016) ·
    hypotheses: H-U-051, H-U-052 · evidence: table/mutation-matrix/maintenance-config

## Notes

H-U-052 confirmed (synchronous 400) but the message names the missing metadata location, not a missing sort
order; H-U-051 refuted for a metadata-less table. A table WITH metadata but no sort order was not tried
(untested cell).

Full original title: Put compaction strategy=sort / z-order on a table without metadata -> BadRequestException
400 'The table metadata location is not specified.'
