<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-318: Table format, metadata, encryptionConfiguration and storageClassConfiguration are CreateTable-only inputs: no per-table Put/Update exists
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-318"></a>**S3T-TABLE-318** `immutable-field` · impact high · handled · verified 2026-10-14
  **Table format, metadata, encryptionConfiguration and storageClassConfiguration are CreateTable-only inputs: no per-table Put/Update exists**
  botocore s3tables model (apiVersion 2018-05-10, 49 operations): the input members
  ['encryptionConfiguration', 'storageClassConfiguration', 'format', 'metadata'] appear only in
  {"CreateTable": ["format", "metadata", "encryptionConfiguration", "storageClassConfiguration"],
  "CreateTableBucket": ["encryptionConfiguration", "storageClassConfiguration"], "PutTableBucketEncryption":
  ["encryptionConfiguration"], "PutTableBucketStorageClass": ["storageClassConfiguration"]}. The table-scoped
  mutating operations are ['DeleteTable', 'DeleteTablePolicy', 'PutTableMaintenanceConfiguration',
  'PutTablePolicy', 'RenameTable', 'UpdateTableMetadataLocation']; none of them accepts encryption, storage
  class, format or metadata, so a table's encryption/storage class can be changed only by delete+recreate (the
  bucket-level Put*Encryption/Put*StorageClass operations apply to the bucket default, not to an existing
  table).
  - ACK: is_immutable, custom_update · ops: CreateTable, CreateTableBucket, PutTableBucketEncryption,
    PutTableBucketStorageClass, DeleteTable, DeleteTablePolicy, PutTableMaintenanceConfiguration,
    PutTablePolicy, RenameTable, UpdateTableMetadataLocation · fields: format, metadata,
    encryptionConfiguration, storageClassConfiguration
  - repro: walk every operation's input shape in the botocore s3tables model for the four member names
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLE-319](../table-encryption-storage.md#s3t-table-319), [S3T-TABLE-312](../table-schema.md#s3t-table-312) · evidence: table/sub-resources/fu-07-rename-kms-table

## Notes

FU-07 (static half). Confirms the sketch expectation: CreateTable is the only operation taking
format/metadata, and the only table-level carrier of encryptionConfiguration/storageClassConfiguration
(CreateTableBucket, PutTableBucketEncryption and PutTableBucketStorageClass carry them for the bucket).
Derived from a programmatic walk of the botocore model inside the probe (recorded in result.yaml
model_input_carriers), no AWS call needed.
