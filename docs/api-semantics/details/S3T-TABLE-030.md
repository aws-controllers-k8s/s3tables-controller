<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-030: Model: GetTable input 'tableArn' vs output 'tableARN'; DeleteTable requires bucket ARN+namespace+name; no UpdateTable operation
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-030"></a>**S3T-TABLE-030** `codegen-artifact` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Model: GetTable input 'tableArn' vs output 'tableARN'; DeleteTable requires bucket ARN+namespace+name; no UpdateTable operation**
  botocore s3tables model (api version 2018-05-10): GetTable input members ['tableBucketARN', 'namespace',
  'name', 'tableArn'], output ARN member ['tableARN']; DeleteTable input members ['tableBucketARN',
  'namespace', 'name', 'versionToken'] with required ['tableBucketARN', 'namespace', 'name'] (no tableArn
  form); 'UpdateTable' in operation names: False; table-level mutation operations present:
  ['PutTableMaintenanceConfiguration', 'PutTablePolicy', 'PutTableRecordExpirationConfiguration',
  'PutTableReplication', 'RenameTable', 'UpdateTableMetadataLocation']. Verified live: GetTable(tableArn=...)
  -> HTTP 200.
  - ACK: renames, is_immutable, custom_delete · ops: GetTable, DeleteTable, CreateTable · fields: tableArn,
    tableARN, format, metadata, encryptionConfiguration, storageClassConfiguration
  - repro: Inspect botocore service model for s3tables; GetTable(tableArn=<arn>)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-001](../table-metadata-location.md#s3t-table-001), [S3T-TABLE-173](../table-replication.md#s3t-table-173), [S3T-TABLE-175](../table-replication.md#s3t-table-175) · hypotheses: H-T-018, H-T-048 · evidence:
    table/identity/get-table-identifier-forms

## Notes

H-T-018 confirmed from the model (TableArn vs TableARN will be distinct generated Go fields; ReadOne must copy
status ARN into tableArn or use the composite key). H-T-048 confirmed: DeleteTable has no tableArn member, so
bucket ARN, namespace and name must stay resolvable at delete time (GetTable(tableArn) provides them);
format/metadata/encryptionConfiguration/storageClassConfiguration have no update operation (RenameTable
changes only name/namespace).
