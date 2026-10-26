<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-123: Empty table: first UpdateTableMetadataLocation with the CreateTable token -> 200, GetTable gains metadataLocation; reusing that token -> 409
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-123"></a>**S3T-TABLE-123** `prerequisite` · impact medium · handled · verified 2026-10-13
  **Empty table: first UpdateTableMetadataLocation with the CreateTable token -> 200, GetTable gains metadataLocation; reusing that token -> 409**
  On a table created without metadata (GetTable has no metadataLocation key), S3 PutObject of a metadata.json
  copy under <warehouseLocation>/metadata/ -> HTTP 200; UpdateTableMetadataLocation(versionToken=<CreateTable
  response token>, that path) -> HTTP 200; afterwards GetTable has metadataLocation: True and
  GetTableMetadataLocation returns the path byte-exact: True; new token minted: True. A second
  UpdateTableMetadataLocation with the same CreateTable token -> ConflictException (HTTP 409) 'Provided
  version token does not match the table version token.'.
  - ACK: hook:sdk_create_post_set_output, custom_update · ops: CreateTable, UpdateTableMetadataLocation,
    GetTable, GetTableMetadataLocation · fields: versionToken, metadataLocation
  - repro: CreateTable(no metadata) -> PutObject <wh>/metadata/00000-<uuid>.metadata.json ->
    UpdateTableMetadataLocation(create token, path) -> GetTable -> repeat with the same token
  - measurements: update_latency_ms=[12], n=1
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491; generator.yaml:76-83; generator.yaml:105-110`
  - related: [S3T-TABLE-292](../table-metadata-location.md#s3t-table-292), [S3T-TABLE-277](../table-metadata-location.md#s3t-table-277), [S3T-TABLE-271](../table.md#s3t-table-271), [S3T-TABLE-275](../table-metadata-location.md#s3t-table-275), [S3T-TABLE-282](../table-metadata-location.md#s3t-table-282), [S3T-TABLE-291](../table.md#s3t-table-291),
    [S3T-TABLE-015](../table-schema.md#s3t-table-015), [S3T-TABLE-034](../table-metadata-location.md#s3t-table-034), [S3T-TABLE-019](../table-metadata-location.md#s3t-table-019), [S3T-TABLE-311](../table-schema.md#s3t-table-311), [S3T-TABLE-316](../table-metadata-location.md#s3t-table-316), [S3T-TABLE-101](../table-metadata-location.md#s3t-table-101), [S3T-TABLE-102](../table-metadata-location.md#s3t-table-102),
    [S3T-TABLE-104](../table-metadata-location.md#s3t-table-104), [S3T-TABLE-065](../table-metadata-location.md#s3t-table-065), [S3T-TABLE-103](../table-metadata-location.md#s3t-table-103), [S3T-TABLE-051](../table-metadata-location.md#s3t-table-051), [S3T-TABLE-098](../table-metadata-location.md#s3t-table-098), [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320) · hypotheses:
    H-T-140 · evidence: table/idempotency/metadata-location-after-delete

## Notes

H-T-140 confirmed: the CreateTable-response token is a valid CAS token for an empty table; after the first
commit GetTable gains metadataLocation and a new token is minted; reusing the create token ->
ConflictException 409 'Provided version token does not match the table version token.'. One PutObject (a copy
of the deleted table's metadata.json with 'location' rewritten) was made; PutObject via the S3 API into the
warehouse -> 200 for the admin principal. n=1.

Full original title: Empty table: UpdateTableMetadataLocation with the CreateTable-response token -> HTTP 200
(first commit); metadataLocation appears in GetTable (False -> True); reusing that token -> ConflictException
409
