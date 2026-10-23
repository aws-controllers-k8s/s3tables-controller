<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-319: RenameTable across namespaces (HTTP 204): tableARN unchanged, old triple -> NotFoundException 404, encryption/storage class follow
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-319"></a>**S3T-TABLE-319** `identity` · impact high · handled · verified 2026-10-14
  **RenameTable across namespaces (HTTP 204): tableARN unchanged, old triple -> NotFoundException 404, encryption/storage class follow**
  Table created with encryptionConfiguration={sseAlgorithm:aws:kms, kmsKeyArn:<key>},
  storageClass=INTELLIGENT_TIERING and a 2-field schema; RenameTable(namespace=ns1, name=t,
  newNamespaceName=ns2, newName=t2, no versionToken) -> HTTP 204 (empty body). Then GetTable(ns1,t) ->
  NotFoundException (HTTP 404) (poll codes right after the call: ['ERR:NotFoundException'], 404 after 0.01 s);
  GetTable(ns2,t2) -> HTTP 200 (poll codes ['OK'], 200 after 0.01 s); GetTable(tableArn=<pre-rename ARN>) ->
  HTTP 200 with namespace=['<ns2>'] name=<t2>; tableARN identical to the Create ARN in both reads: True.
  ListTables(ns1) -> [], ListTables(ns2) -> ['<t2>']. Under the new triple GetTableEncryption -> HTTP 200
  {"kmsKeyArn_eq_sent_key": true, "sseAlgorithm": "aws:kms"} (same as before rename: True),
  GetTableStorageClass -> HTTP 200 {"storageClass": "INTELLIGENT_TIERING"} (same: True),
  GetTableMetadataLocation -> HTTP 200, GetTableMaintenanceConfiguration -> HTTP 200; under the old triple the
  four sub-resource reads -> ['NotFoundException']. versionToken changed by the rename: False (identical token
  before/after); modifiedAt changed: True; modifiedBy was ABSENT from GetTable on the fresh table and PRESENT
  (== caller account) after the rename; ListTagsForResource(tableArn) after rename -> HTTP 200 with the create
  tag intact: True.
  - ACK: is_arn_primary_key, custom_update, is_immutable · ops: RenameTable, GetTable, ListTables,
    GetTableEncryption, GetTableStorageClass, GetTableMetadataLocation, ListTagsForResource · fields:
    namespace, name, newNamespaceName, newName, tableARN, versionToken, encryptionConfiguration,
    storageClassConfiguration
  - repro: CreateTable(ICEBERG, aws:kms key, INTELLIGENT_TIERING) in ns1 -> RenameTable(newNamespaceName=ns2,
    newName=t2) -> GetTable by old triple, new triple, tableArn; GetTableEncryption/GetTableStorageClass under
    both triples; ListTables(ns1), ListTables(ns2)
  - measurements: old_triple_404_after_s=0.01, new_triple_200_after_s=0.01, n=1
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLEMAINTENANCECONFIGURATION-012](../table-maintenance.md#s3t-tablemaintenanceconfiguration-012), [S3T-TABLE-265](../table-maintenance.md#s3t-table-265), [S3T-TABLE-318](../table-maintenance.md#s3t-table-318), [S3T-TABLE-312](../table-schema.md#s3t-table-312) · evidence:
    table/sub-resources/fu-07-rename-kms-table

## Notes

FU-07. A controller keyed by (bucket, namespace, name) loses the table on an out-of-band rename, while a
controller keyed by tableARN keeps finding it (namespace/name are then the drift). RenameTable has no output
body; the new state must be read back. CreateTable with aws:kms succeeded (key policy: account root +
maintenance.s3tables.amazonaws.com GenerateDataKey/Decrypt). Single run.
