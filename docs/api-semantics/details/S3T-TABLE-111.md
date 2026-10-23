<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-111: RenameTable (newName, then newNamespaceName): tableARN, warehouseLocation, metadataLocation stable; old identifiers -> 404 within 0.2 s
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-111"></a>**S3T-TABLE-111** `identity` · impact high · handled · verified 2026-10-13
  **RenameTable (newName, then newNamespaceName): tableARN, warehouseLocation, metadataLocation stable; old identifiers -> 404 within 0.2 s**
  RenameTable(newName) -> HTTP 204; under the new name GetTable shows the same tableARN True,
  warehouseLocation True, metadataLocation True, createdAt True; GetTable/GetTableMetadataLocation under the
  OLD name (0.17 s after the rename response) -> NotFoundException (HTTP 404) 'The specified table does not
  exist.' / NotFoundException (HTTP 404) 'The specified table does not exist.'. RenameTable(newNamespaceName)
  -> HTTP 204; tableARN same True, warehouseLocation same True, metadataLocation same True, namespaceId
  changed True; old namespace+name reads -> NotFoundException (HTTP 404) 'The specified table does not exist.'
  / NotFoundException (HTTP 404) 'The specified table does not exist.'. GetTable(tableArn) after both renames
  -> HTTP 200 and reports the new namespace/name.
  - ACK: is_arn_primary_key, custom_update · ops: RenameTable, GetTable, GetTableMetadataLocation · fields:
    tableARN, warehouseLocation, metadataLocation, namespaceId, name, namespace
  - repro: CreateTable with schema -> RenameTable(newName) -> GetTable/GetTableMetadataLocation under old and
    new names -> RenameTable(newNamespaceName) -> same reads -> GetTable(tableArn)
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122; generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLE-003](../table-metadata-location.md#s3t-table-003), [S3T-TABLE-009](../table-metadata-location.md#s3t-table-009), [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320), [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193), [S3T-TABLE-007](../table-metadata-location.md#s3t-table-007) · hypotheses: H-T-144 ·
    evidence: table/identity/version-token-rename

## Notes

H-T-144 confirmed: tableARN, warehouseLocation, metadataLocation and createdAt identical under the new name
and the new namespace (namespaceId changes with the namespace move); old name and old namespace+name ->
NotFoundException 404 'The specified table does not exist.' on both GetTable and GetTableMetadataLocation at
0.17-0.19 s after the rename response; GetTable(tableArn) resolves the renamed table and reports the new
namespace/name. n=1 per rename.

Full original title: RenameTable (newName, then newNamespaceName): tableARN stable, warehouseLocation stable,
metadataLocation stable; old identifiers -> NotFoundException (HTTP 404) within 0.2 s
