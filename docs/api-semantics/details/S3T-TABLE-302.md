<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-302: Stale versionToken revived by another writer's rollback (A->B->A): DeleteTable(A0) -> 204, UpdateTableMetadataLocation(C, A0) -> 200
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-302"></a>**S3T-TABLE-302** `idempotency` · impact medium · handled · verified 2026-10-14
  **Stale versionToken revived by another writer's rollback (A->B->A): DeleteTable(A0) -> 204, UpdateTableMetadataLocation(C, A0) -> 200**
  Table at location A with the controller's cached token A0. External UpdateTableMetadataLocation to B -> HTTP
  200; the controller's A0 while at B -> ConflictException (HTTP 409) 'Provided version token does not match
  the table version token.'; external rollback to A -> HTTP 200 (token A2 == A0: True); the external writer's
  B token after the rollback -> ConflictException (HTTP 409) 'Provided version token does not match the table
  version token.'. The controller's DeleteTable with the revived A0 -> HTTP 204 (table gone). On a second
  table with the same A->B->A history, UpdateTableMetadataLocation(C, A0) -> HTTP 200 and GetTable then points
  at C: True. Three fresh tables without metadataLocation had 3 distinct tokens.
  - ACK: annotation-shadow-state, custom_delete · ops: UpdateTableMetadataLocation, DeleteTable, GetTable ·
    fields: versionToken, metadataLocation
  - repro: CreateTable -> Update(A) token A0 -> Update(B, A0) -> Update(B, A0) 409 -> Update(A, B1) ->
    DeleteTable(versionToken=A0) / Update(C, A0)
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-134](../table-metadata-location.md#s3t-table-134), [S3T-TABLE-098](../table-metadata-location.md#s3t-table-098), [S3T-TABLE-108](../table-metadata-location.md#s3t-table-108), [S3T-TABLE-104](../table-metadata-location.md#s3t-table-104), [S3T-TABLE-103](../table-metadata-location.md#s3t-table-103), [S3T-TABLE-192](../table-metadata-location.md#s3t-table-192) ·
    evidence: table/creative/token-revival

## Notes

Extends [S3T-TABLE-134](../table-metadata-location.md#s3t-table-134) (token = f(location)) with the consequence for optimistic locking: ConflictException
means 'the location differs from the one your token encodes', NOT 'the table changed since you read it' - two
intervening commits that end at the cached location are invisible, so a controller must not treat a token
match as proof that nothing happened (use modifiedAt / re-read before destructive calls). Single run, 2
tables.

Full original title: Stale versionToken revived by another writer's rollback: DeleteTable(A0) -> HTTP 204,
UpdateTableMetadataLocation(A0) -> HTTP 200 after A->B->A (clobber)

Contradiction with [S3T-TABLE-098](../table-metadata-location.md#s3t-table-098): same input (DeleteTable with the create-time token A0 after external
UpdateTableMetadataLocation commits): 098 -> ConflictException 409 after one commit that moved the location,
302 -> HTTP 204 after two commits A->B->A that restored it; 134 reconciles: versionToken =
f(metadataLocation), so the 409 means 'location differs from the token's', not 'table changed since read'
Resolution: keep both; [S3T-TABLE-134](../table-metadata-location.md#s3t-table-134) is canonical for the rule, [S3T-TABLE-302](../table-metadata-location.md#s3t-table-302) for the revival exception; 098
holds for any history that ends at a different location
