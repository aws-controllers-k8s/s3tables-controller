<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-033: CreateTable with iceberg.schema: GetTable metadataLocation = <warehouseLocation>/metadata/00000-<uuid>.metadata.json; token == Create's
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-033"></a>**S3T-TABLE-033** `response-fidelity` · impact medium · handled · verified 2026-10-13
  **CreateTable with iceberg.schema: GetTable metadataLocation = <warehouseLocation>/metadata/00000-<uuid>.metadata.json; token == Create's**
  CreateTable(format=ICEBERG, metadata={iceberg:{schema:{fields:[id int required, v string]}}}) -> HTTP 200;
  immediate GetTable -> HTTP 200 with metadataLocation present(str), starting with warehouseLocation + '/':
  True, relative path pattern 'metadata/00000-<uuid>.metadata.json' (matches 00000-<uuid>.metadata.json:
  True); create versionToken == GetTable versionToken: True; GetTableMetadataLocation agrees on
  metadataLocation: True and token: True; type=customer, managedByService="<absent>".
  - ACK: is_read_only, late_initialize · ops: CreateTable, GetTable, GetTableMetadataLocation · fields:
    metadata, metadataLocation, versionToken, warehouseLocation
  - repro: CreateTable with metadata.iceberg.schema; GetTable(tableArn); GetTableMetadataLocation
  - handling: handled via `generator.yaml:76-83; generator.yaml:105-110`
  - related: [S3T-TABLE-014](../table-metadata-location.md#s3t-table-014), [S3T-TABLE-031](../table-metadata-location.md#s3t-table-031), [S3T-TABLE-309](../table-maintenance.md#s3t-table-309), [S3T-TABLE-315](../table-metadata-location.md#s3t-table-315) · hypotheses: H-T-030 · evidence:
    table/round-trip/fresh-table-fields

## Notes

H-T-030 confirmed: schema-at-create yields a committed initial metadata file in one synchronous call; no
follow-up UpdateTableMetadataLocation is needed.

Full original title: CreateTable with metadata.iceberg.schema: GetTable metadataLocation is
<warehouseLocation>/metadata/00000-<uuid>.metadata.json and create-time versionToken equals GetTable's
