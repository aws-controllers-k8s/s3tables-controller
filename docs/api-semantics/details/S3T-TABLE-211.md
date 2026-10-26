<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-211: Maintenance model hazards: enum value "z-order" (hyphen), required path member named "type", settings is a tagged union keyed by type
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-211"></a>**S3T-TABLE-211** `codegen-artifact` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Maintenance model hazards: enum value "z-order" (hyphen), required path member named "type", settings is a tagged union keyed by type**
  botocore model: PutTableMaintenanceConfiguration input members tableBucketARN, namespace, name, type (enum
  TableMaintenanceType: icebergCompaction, icebergSnapshotManagement), value {status: enabled|disabled,
  settings: union{icebergCompaction, icebergSnapshotManagement}}; IcebergCompactionStrategy enum auto,
  binpack, sort, z-order (hyphenated; the GetTableMaintenanceConfiguration doc string says 'zorder'). Integer
  members carry min=1 in the model (botocore rejects 0/-1 client-side before any request). Observed: Put
  strategy=z-order accepted (BadRequestException/400).
  - ACK: renames, custom_update · ops: PutTableMaintenanceConfiguration · fields: type, value.settings,
    strategy
  - repro: inspect botocore service model; Put strategy z-order
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-160](../table-maintenance.md#s3t-table-160), [S3T-TABLE-197](../table-maintenance.md#s3t-table-197), [S3T-TABLE-198](../table-maintenance.md#s3t-table-198), [S3T-TABLE-204](../table-maintenance.md#s3t-table-204), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183), [S3T-TABLE-199](../table-maintenance.md#s3t-table-199),
    [S3T-TABLE-203](../table-maintenance.md#s3t-table-203), [S3T-TABLE-016](../table-schema.md#s3t-table-016) · hypotheses: H-U-072 · evidence: table/mutation-matrix/maintenance-config

## Notes

H-U-072 confirmed from the model plus the z-order Put.
