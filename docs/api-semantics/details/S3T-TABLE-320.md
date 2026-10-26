<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-320: Renamed table accepts UpdateTableMetadataLocation (200) and TagResource (200); stale versionToken -> ConflictException 409
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-320"></a>**S3T-TABLE-320** `sub-resource-api` · impact medium · handled · verified 2026-10-14
  **Renamed table accepts UpdateTableMetadataLocation (200) and TagResource (200); stale versionToken -> ConflictException 409**
  On the renamed table (new triple) UpdateTableMetadataLocation(versionToken=<current from GetTable>,
  metadataLocation=<warehouseLocation>/metadata/00001-<uuid>.metadata.json (no object was written there)) ->
  HTTP 200; the same call addressed to the OLD triple -> NotFoundException (HTTP 404) 'The specified table
  does not exist.'; with the pre-rename/stale versionToken -> ConflictException (HTTP 409) 'Provided version
  token does not match the table version token.'. GetTable after the update: metadataLocation changed to the
  sent URI, versionToken changed: True, modifiedAt changed: True. TagResource(resourceArn=<tableARN>) -> HTTP
  200, ListTagsForResource -> HTTP 200 tags=['ackq-run', 'after-rename']. RenameTable with the stale
  versionToken -> ConflictException (HTTP 409) 'Provided version token does not match the table version
  token.'; RenameTable into a namespace that does not exist -> NotFoundException (HTTP 404) 'The specified
  destination namespace does not exist.'. Across create/rename/update: createdBy==caller account True,
  modifiedBy present True and ==caller account True, ownerAccountId==caller True.
  - ACK: custom_update, exceptions.404, terminal_codes · ops: UpdateTableMetadataLocation, TagResource,
    ListTagsForResource, RenameTable, GetTable · fields: versionToken, metadataLocation, modifiedAt,
    modifiedBy, createdBy, ownerAccountId, tags
  - repro: RenameTable -> GetTable(new triple) for versionToken -> UpdateTableMetadataLocation(new triple,
    token, <warehouse>/metadata/00001-<uuid>.metadata.json); same on old triple; with stale token;
    TagResource(tableARN); RenameTable(stale token); RenameTable(newNamespaceName=<missing>)
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLE-100](../table-metadata-location.md#s3t-table-100), [S3T-TABLE-192](../table-metadata-location.md#s3t-table-192), [S3T-TABLE-051](../table-metadata-location.md#s3t-table-051), [S3T-TABLE-108](../table-metadata-location.md#s3t-table-108), [S3T-TABLE-065](../table-metadata-location.md#s3t-table-065), [S3T-TABLE-103](../table-metadata-location.md#s3t-table-103),
    [S3T-TABLE-098](../table-metadata-location.md#s3t-table-098), [S3T-TABLE-123](../table-metadata-location.md#s3t-table-123), [S3T-TABLE-003](../table-metadata-location.md#s3t-table-003), [S3T-TABLE-009](../table-metadata-location.md#s3t-table-009), [S3T-TABLE-111](../table-metadata-location.md#s3t-table-111), [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193), [S3T-TABLE-007](../table-metadata-location.md#s3t-table-007),
    [S3T-TABLE-005](../table-metadata-location.md#s3t-table-005), [S3T-TABLE-112](../table-metadata-location.md#s3t-table-112), [S3T-TABLE-316](../table-metadata-location.md#s3t-table-316), [S3T-TABLE-101](../table-metadata-location.md#s3t-table-101), [S3T-TABLE-104](../table-metadata-location.md#s3t-table-104), [S3T-TABLE-124](../table-metadata-location.md#s3t-table-124), [S3T-TABLE-019](../table-metadata-location.md#s3t-table-019),
    [S3T-TABLE-121](../table-metadata-location.md#s3t-table-121), [S3T-TABLE-245](../table-metadata-location.md#s3t-table-245) · evidence: table/sub-resources/fu-07-rename-kms-table

## Notes

FU-07 (mutations after rename) and FU-11 (modifiedBy/createdBy/ownerAccountId are all the caller's 12-digit
account id in a single-principal lab, so a change of modifiedBy to another principal is not observable: only
presence and equality were recorded; modifiedAt/versionToken changes are recorded above).
UpdateTableMetadataLocation accepted a metadataLocation whose object does not exist (no S3 PutObject was
made): the service validates the URI, not the object. UpdateTableMetadataLocation changed versionToken and
modifiedAt; RenameTable changed modifiedAt only.
