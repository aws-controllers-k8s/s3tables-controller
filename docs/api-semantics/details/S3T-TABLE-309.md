<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-309: CreateTable with iceberg.schema: metadataLocation present at once; GetTableMaintenanceConfiguration 200 with compaction + snapshot defaults
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-309"></a>**S3T-TABLE-309** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **CreateTable with iceberg.schema: metadataLocation present at once; GetTableMaintenanceConfiguration 200 with compaction + snapshot defaults**
  Right after CreateTable with a schema (HTTP 200, versionToken present: True) GetTable showed
  metadataLocation=present, versionToken equal to the Create response: True, type=customer, format=ICEBERG,
  managedByService present: False. GetTableMetadataLocation -> HTTP 200 with keys ['metadataLocation',
  'versionToken', 'warehouseLocation'] and the same versionToken: True. GetTableMaintenanceConfiguration ->
  HTTP 200 {"icebergCompaction": {"settings": {"icebergCompaction": {"strategy": "auto", "targetFileSizeMB":
  512}}, "status": "enabled"}, "icebergSnapshotManagement": {"settings": {"icebergSnapshotManagement":
  {"maxSnapshotAgeHours": 120, "minSnapshotsToKeep": 1}}, "status": "enabled"}}.
  - ACK: hook:sdk_create_post_set_output, late_initialize · ops: CreateTable, GetTable,
    GetTableMetadataLocation, GetTableMaintenanceConfiguration · fields: metadataLocation, versionToken,
    metadata.iceberg.schema
  - repro: CreateTable(format=ICEBERG, metadata.iceberg.schema.fields=[2 fields], aws:kms,
    INTELLIGENT_TIERING, tags) -> GetTable, GetTableMetadataLocation
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-014](../table-metadata-location.md#s3t-table-014), [S3T-TABLE-031](../table-metadata-location.md#s3t-table-031), [S3T-TABLE-033](../table-schema.md#s3t-table-033), [S3T-TABLE-315](../table-metadata-location.md#s3t-table-315) · evidence:
    table/round-trip/fu-06-full-create-read-paths

## Notes

FU-06, n=1. Contrast with the smoke probe (table created WITHOUT metadata): there GetTableMetadataLocation had
no metadataLocation. With a schema the service writes the first metadata.json at create time, so
metadataLocation is populated without any data-plane write.

Full original title: CreateTable with metadata.iceberg.schema: GetTable/GetTableMetadataLocation return
metadataLocation immediately; GetTableMaintenanceConfiguration returns HTTP 200 with ['icebergCompaction',
'icebergSnapshotManagement']
