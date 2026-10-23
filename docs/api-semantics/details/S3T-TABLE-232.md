<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-232: Admin caller: Put strategy sort / z-order on a table WITH metadata but no sort order -> 400 'requires sort order in the table metadata'
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-232"></a>**S3T-TABLE-232** `prerequisite` · impact medium · handled · verified 2026-10-13
  **Admin caller: Put strategy sort / z-order on a table WITH metadata but no sort order -> 400 'requires sort order in the table metadata'**
  Table created with metadata.iceberg.schema (2 fields, no sort order; GetTable.metadataLocation
  present=True): Put icebergCompaction {enabled, strategy sort} -> BadRequestException (HTTP 400) 'The
  specified compaction strategy requires sort order in the table metadata. Please refer to the S3 Tables
  documentation for steps on configuring sort order.'; Put {enabled, strategy z-order} -> BadRequestException
  (HTTP 400) 'The specified compaction strategy requires sort order in the table metadata. Please refer to the
  S3 Tables documentation for steps on configuring sort order.'; GetTableMaintenanceConfiguration compaction
  settings afterwards {"targetFileSizeMB": 512, "strategy": "auto"}; job status {"icebergCompaction":
  "Not_Yet_Run", "icebergUnreferencedFileRemoval": "Not_Yet_Run", "icebergSnapshotManagement": "Not_Yet_Run"}.
  Compare: on a table without metadata both Puts fail with BadRequestException 'The table metadata location is
  not specified.' (maintenance-config probe).
  - ACK: terminal_codes · ops: PutTableMaintenanceConfiguration, CreateTable · fields:
    value.settings.icebergCompaction.strategy, metadata
  - repro: CreateTable(format ICEBERG, metadata.iceberg.schema) -> Put compaction strategy sort / z-order ->
    Get
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-297](../table-maintenance.md#s3t-table-297), [S3T-TABLE-294](../table-maintenance.md#s3t-table-294), [S3T-TABLE-204](../table-maintenance.md#s3t-table-204), [S3T-TABLE-230](../table-maintenance.md#s3t-table-230), [S3T-TABLE-234](../table-maintenance.md#s3t-table-234), [S3T-TABLE-233](../table-maintenance.md#s3t-table-233),
    [S3T-TABLEMAINTENANCECONFIGURATION-012](../table-maintenance.md#s3t-tablemaintenanceconfiguration-012), [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001), [S3T-TABLE-205](../table-maintenance.md#s3t-table-205), [S3T-TABLE-206](../table-maintenance.md#s3t-table-206)
    · hypotheses: H-U-051, H-U-052 · evidence: table/error-taxonomy/iam-deny-subresource-codes

## Notes

H-U-052 confirmed for a table with metadata (synchronous BadRequestException 400 'The specified compaction
strategy requires sort order in the table metadata. Please refer to the S3 Tables documentation for steps on
configuring sort order.'); H-U-051 refuted. On a metadata-less table the message is 'The table metadata
location is not specified.' instead. Stored strategy stays what it was (auto).

Contradiction with [S3T-TABLE-234](../table-maintenance.md#s3t-table-234): Same Put icebergCompaction {strategy sort|z-order} on a table with Iceberg
metadata (schema only, no sort order): 232 reports BadRequestException 400 'requires sort order in the table
metadata', 234 reports AccessDeniedException 403 'Insufficient access to update the table maintenance
configuration'. 234's own notes reconcile: the caller differs - the restricted role
(Get/PutTableMaintenanceConfiguration + GetTable, no GetTableData) hits the permission check first, the admin
caller on the same table gets the sort-order 400. Consequential for ACK: 400 is terminal ([GT-S3T-039](../service.md#gt-s3t-039) (controller hooks catalog entry)) while
403 is requeued, so the same user mistake is terminal or an endless requeue depending on the controller's IAM
role Resolution: keep both; [S3T-TABLE-234](../table-maintenance.md#s3t-table-234) is canonical for a restricted role, [S3T-TABLE-232](../table-maintenance.md#s3t-table-232) for a
full-permission caller; 232 over-generalizes (omits the caller), see title_fixes
