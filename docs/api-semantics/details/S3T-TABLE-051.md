<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-051: RenameTable with a stale or garbage versionToken -> ConflictException 409 'Provided version token does not match'; current token -> 204
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-051"></a>**S3T-TABLE-051** `error-code` · impact high · handled · verified 2026-10-13
  **RenameTable with a stale or garbage versionToken -> ConflictException 409 'Provided version token does not match'; current token -> 204**
  UpdateTableMetadataLocation(v0, <warehouse>/metadata/<new file>) -> HTTP 200 (versionToken changed: True).
  RenameTable(newName=u, versionToken=<pre-update token>) -> ConflictException (HTTP 409) 'Provided version
  token does not match the table version token.'; RenameTable(newName=u, versionToken='xyz') ->
  ConflictException (HTTP 409) 'Provided version token does not match the table version token.'; GetTable then
  shows the name unchanged: True. RenameTable(newName=u, versionToken=<current token>) -> HTTP 204.
  - ACK: requeue, terminal_codes, custom_update · ops: RenameTable, UpdateTableMetadataLocation, GetTable ·
    fields: versionToken, metadataLocation
  - repro: CreateTable(schema) -> GetTable v0 -> UpdateTableMetadataLocation(v0, new path) -> v1 ->
    RenameTable(versionToken=v0), RenameTable(versionToken='xyz'), RenameTable(versionToken=v1)
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-112](../table-metadata-location.md#s3t-table-112), [S3T-TABLE-100](../table-metadata-location.md#s3t-table-100), [S3T-TABLE-192](../table-metadata-location.md#s3t-table-192), [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320), [S3T-TABLE-108](../table-metadata-location.md#s3t-table-108), [S3T-TABLE-065](../table-metadata-location.md#s3t-table-065),
    [S3T-TABLE-103](../table-metadata-location.md#s3t-table-103), [S3T-TABLE-098](../table-metadata-location.md#s3t-table-098), [S3T-TABLE-123](../table-metadata-location.md#s3t-table-123) · hypotheses: H-T-061 · evidence:
    table/error-taxonomy/rename-errors

## Notes

H-T-061 confirmed: stale token (from before UpdateTableMetadataLocation) and the garbage token 'xyz' both ->
ConflictException 409 with the same message 'Provided version token does not match the table version token.';
the table was not renamed; the current token -> 204. Side observation (outside this shard):
UpdateTableMetadataLocation(v0, <warehouse>/metadata/<file that does not exist>) was accepted with HTTP 200
and a new versionToken - no object-existence check.
