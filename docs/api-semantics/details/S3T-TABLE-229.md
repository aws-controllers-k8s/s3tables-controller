<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-229: type path segment: icebergUnreferencedFileRemoval / IcebergCompaction / foo -> 400 'Invalid table maintenance type: <value>'; empty -> 404
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-229"></a>**S3T-TABLE-229** `request-validation` · impact medium · handled · verified 2026-10-13
  **type path segment: icebergUnreferencedFileRemoval / IcebergCompaction / foo -> 400 'Invalid table maintenance type: <value>'; empty -> 404**
  PutTableMaintenanceConfiguration with an out-of-enum type: icebergUnreferencedFileRemoval ->
  BadRequestException/400 'Invalid table maintenance type: icebergUnreferencedFileRemoval'; IcebergCompaction
  -> BadRequestException/400 'Invalid table maintenance type: IcebergCompaction'; ICEBERGCOMPACTION ->
  BadRequestException/400 'Invalid table maintenance type: ICEBERGCOMPACTION'; foo -> BadRequestException/400
  'Invalid table maintenance type: foo'; iceberg-compaction -> BadRequestException/400 'Invalid table
  maintenance type: iceberg-compaction'; icebergCompaction%20 -> BadRequestException/400 'Invalid table
  maintenance type: icebergCompaction%20'; ufr_with_its_own_member -> ClientSerializationError/None
  'KeyError('icebergUnreferencedFileRemoval')'; empty -> 404/404 '<UnknownOperationException/>
  '.
  - ACK: terminal_codes · ops: PutTableMaintenanceConfiguration · fields: type
  - repro: Put with type in [icebergUnreferencedFileRemoval, IcebergCompaction, ICEBERGCOMPACTION, foo,
    iceberg-compaction, '']
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-226](../table-maintenance.md#s3t-table-226), [S3T-TABLE-227](../table-maintenance.md#s3t-table-227), [S3T-TABLE-228](../table-maintenance.md#s3t-table-228), [S3T-TABLE-230](../table-maintenance.md#s3t-table-230), [S3T-TABLE-265](../table-maintenance.md#s3t-table-265),
    [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001), [S3T-TABLEMAINTENANCECONFIGURATION-008](../table-maintenance.md#s3t-tablemaintenanceconfiguration-008), [S3T-TABLE-160](../table-maintenance.md#s3t-table-160) · hypotheses:
    H-U-045 · evidence: table/weird-inputs/maintenance-bounds

## Notes

H-U-045 confirmed: icebergUnreferencedFileRemoval (bucket-level type), IcebergCompaction, ICEBERGCOMPACTION,
foo, iceberg-compaction -> BadRequestException 400 'Invalid table maintenance type: <value>'; an empty type
segment -> HTTP 404 with a non-JSON body '<UnknownOperationException/>' (router miss, no modeled code).

Full original title: type path segment: icebergUnreferencedFileRemoval -> BadRequestException/400;
IcebergCompaction -> BadRequestException/400; foo -> BadRequestException/400; empty -> 404/404
