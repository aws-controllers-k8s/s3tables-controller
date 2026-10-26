<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-160: PutTableMaintenanceConfiguration(type=icebergUnreferencedFileRemoval) -> BadRequestException (HTTP 400)...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-160"></a>**S3T-TABLE-160** `scope` · impact medium · handled · verified 2026-10-13
  **PutTableMaintenanceConfiguration(type=icebergUnreferencedFileRemoval) -> BadRequestException (HTTP 400)...**
  **Scope verdict: field-on-parent**
  On a fresh ICEBERG table: PutTableMaintenanceConfiguration(type=icebergUnreferencedFileRemoval,
  value={status: enabled}) -> BadRequestException (HTTP 400) 'Invalid table maintenance type:
  icebergUnreferencedFileRemoval'. GetTableMaintenanceConfiguration before -> HTTP 200 with configuration keys
  ['icebergCompaction', 'icebergSnapshotManagement']; after the Put and after
  PutTableBucketMaintenanceConfiguration {disabled, 5/5} on the parent bucket -> keys ['icebergCompaction',
  'icebergSnapshotManagement'] (identical body). The table-level configuration map carries no entry for the
  bucket-level type in either read.
  - ACK: scope:field-on-parent, terminal_codes · ops: PutTableMaintenanceConfiguration,
    GetTableMaintenanceConfiguration, PutTableBucketMaintenanceConfiguration · fields: type, configuration
  - repro: CreateTable -> GetTableMaintenanceConfiguration ->
    PutTableMaintenanceConfiguration(type=icebergUnreferencedFileRemoval) -> Put bucket maintenance disabled
    -> GetTableMaintenanceConfiguration
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-197](../table-maintenance.md#s3t-table-197), [S3T-TABLE-198](../table-maintenance.md#s3t-table-198), [S3T-TABLE-204](../table-maintenance.md#s3t-table-204), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183), [S3T-TABLE-211](../table-maintenance.md#s3t-table-211), [S3T-TABLE-265](../table-maintenance.md#s3t-table-265),
    [S3T-TABLE-229](../table-maintenance.md#s3t-table-229), [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001), [S3T-TABLEMAINTENANCECONFIGURATION-008](../table-maintenance.md#s3t-tablemaintenanceconfiguration-008) · hypotheses:
    H-S-035 · evidence: tablebucket/weird-inputs/maintenance-validation

## Notes

H-S-035 confirmed: the bucket-only type is rejected at table level and the table read does not reflect the
bucket setting, so there is no per-table propagation to wait for (and nothing to show in Table status).
Whether the bucket setting is enforced on this table's files was not tested (needs data-plane snapshots).

Full original title: PutTableMaintenanceConfiguration(type=icebergUnreferencedFileRemoval) ->
BadRequestException (HTTP 400); GetTableMaintenanceConfiguration keys ['icebergCompaction',
'icebergSnapshotManagement'] (bucket setting not visible per table)
