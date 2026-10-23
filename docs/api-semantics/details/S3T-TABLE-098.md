<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-098: DeleteTable with the create-time versionToken after one UpdateTableMetadataLocation (location moved) -> 409; current token -> 204
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-098"></a>**S3T-TABLE-098** `delete-semantics` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **DeleteTable with the create-time versionToken after one UpdateTableMetadataLocation (location moved) -> 409; current token -> 204**
  Table created with metadata.iceberg.schema (token T0, metadataLocation 00000-<uuid>.metadata.json).
  UpdateTableMetadataLocation(versionToken=T0, metadataLocation=<copied 00001 file>) -> HTTP 200 (token
  changed: True; response keys ['metadataLocation', 'name', 'namespace', 'tableARN', 'versionToken']; GetTable
  token equals the update response token: True; GetTable metadataLocation equals the sent location: True;
  modifiedAt changed: True; modifiedBy now present: False). Then UpdateTableMetadataLocation(T0) ->
  ConflictException (HTTP 409) 'Provided version token does not match the table version token.';
  DeleteTable(versionToken=T0) -> ConflictException (HTTP 409) 'Provided version token does not match the
  table version token.'; GetTable -> table still exists with token T1: True; DeleteTable(versionToken=T1) ->
  HTTP 204; GetTable -> NotFoundException (HTTP 404) 'The specified table does not exist.'.
  - ACK: custom_delete, requeue · ops: DeleteTable, UpdateTableMetadataLocation, GetTable · fields:
    versionToken, metadataLocation, modifiedAt, modifiedBy
  - repro: CreateTable with schema (T0) -> S3 CopyObject of the 00000 metadata file to 00001 ->
    UpdateTableMetadataLocation(T0) -> DeleteTable(T0) -> GetTable -> DeleteTable(T1)
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLE-134](../table-metadata-location.md#s3t-table-134), [S3T-TABLE-108](../table-metadata-location.md#s3t-table-108), [S3T-TABLE-104](../table-metadata-location.md#s3t-table-104), [S3T-TABLE-302](../table-metadata-location.md#s3t-table-302), [S3T-TABLE-103](../table-metadata-location.md#s3t-table-103), [S3T-TABLE-192](../table-metadata-location.md#s3t-table-192),
    [S3T-TABLE-065](../table-metadata-location.md#s3t-table-065), [S3T-TABLE-051](../table-metadata-location.md#s3t-table-051), [S3T-TABLE-123](../table-metadata-location.md#s3t-table-123), [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320) · hypotheses: H-T-037, H-T-030 · evidence:
    table/mutation-matrix/delete-version-token

## Notes

H-T-037 confirmed. A controller caching the create-time token must re-read it (GetTable) before DeleteTable or
omit versionToken entirely (a tokenless delete succeeds, see the idempotency probe). H-T-030 re-confirmed: the
schema-at-create table had metadataLocation and a token usable for the first commit. Note: modifiedBy was
still ABSENT from GetTable after the UpdateTableMetadataLocation commit although modifiedAt changed.

Full original title: DeleteTable with the create-time versionToken after one UpdateTableMetadataLocation ->
ConflictException (HTTP 409); with the current token -> HTTP 204

Contradiction with [S3T-TABLE-302](../table-metadata-location.md#s3t-table-302): same input (DeleteTable with the create-time token A0 after external
UpdateTableMetadataLocation commits): 098 -> ConflictException 409 after one commit that moved the location,
302 -> HTTP 204 after two commits A->B->A that restored it; 134 reconciles: versionToken =
f(metadataLocation), so the 409 means 'location differs from the token's', not 'table changed since read'
Resolution: keep both; [S3T-TABLE-134](../table-metadata-location.md#s3t-table-134) is canonical for the rule, [S3T-TABLE-302](../table-metadata-location.md#s3t-table-302) for the revival exception; 098
holds for any history that ends at a different location

Suspected controller bug confirmed by evidence: not flagged; evidence-driven. The controller passes the cached
Status.VersionToken to DeleteTable and RenameTable although the token is optional there, so any engine commit
between the reconcile's GetTable and the write turns a wanted delete/rename into ConflictException 409
'Provided version token does not match the table version token.' ([S3T-TABLE-098](../table-metadata-location.md#s3t-table-098) for DeleteTable, [S3T-TABLE-051](../table-metadata-location.md#s3t-table-051)
for RenameTable) while a tokenless DeleteTable succeeds ([S3T-TABLE-067](../table-metadata-location.md#s3t-table-067)) and a token match is not proof the
table is unchanged anyway (token = f(location), [S3T-TABLE-134](../table-metadata-location.md#s3t-table-134)/302); omit the token on DeleteTable/RenameTable,
or at minimum map this 409 to requeue-after-re-read, never to Terminal (the same code is permanent for
duplicate CreateTable, [S3T-TABLE-064](../table-metadata-location.md#s3t-table-064)).
