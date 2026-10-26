<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-099: DeleteTable of a table with committed metadata + a data object -> HTTP 204 in 3/3; GetTable -> NotFoundException (HTTP 404) on the next call
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-099"></a>**S3T-TABLE-099** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTable of a table with committed metadata + a data object -> HTTP 204 in 3/3; GetTable -> NotFoundException (HTTP 404) on the next call**
  3 trials: table with schema, one UpdateTableMetadataLocation commit and one 2 KB object under data/
  (['None'] objects listed in the warehouse via S3 ListObjectsV2 before delete). DeleteTable(current token) ->
  HTTP 204 (latency [140, 192, 118] ms). First GetTable composite [9, 11, 9] ms later -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'; first GetTable(tableArn) -> NotFoundException (HTTP 404)
  'The specified table does not exist.'. After the delete, S3 ListObjectsV2 on the warehouse ->
  MethodNotAllowed (HTTP 405) 'The specified method is not allowed against this resource.'; HeadObject on the
  data file -> HTTP 200. Data-plane access to the warehouse location through the S3 API: CopyObject ->
  MethodNotAllowed (HTTP 405) 'The specified method is not allowed against this resource.', ListObjectsV2 ->
  MethodNotAllowed (HTTP 405) (before and after the delete); GetObject, PutObject and HeadObject -> HTTP 200;
  HeadObject on the data object still returned 200 immediately AFTER the table was deleted (3/3), so the
  storage is drained asynchronously even though the table itself is gone from GetTable at once.
  - ACK: exceptions.404, custom_delete · ops: DeleteTable, GetTable, UpdateTableMetadataLocation · fields:
    versionToken, warehouseLocation
  - repro: CreateTable with schema; CopyObject metadata; UpdateTableMetadataLocation; PutObject data/...;
    DeleteTable(token); GetTable both forms; poll at 0.5 s if still visible
  - measurements: delete_latency_ms=[140, 192, 118], first_get_after_delete_ms=[9, 11, 9],
    visible_windows_s=[], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-066](../table-metadata-location.md#s3t-table-066), [S3T-TABLE-089](../table-metadata-location.md#s3t-table-089), [S3T-TABLE-122](../table-metadata-location.md#s3t-table-122), [S3T-TABLE-129](../table-metadata-location.md#s3t-table-129), [S3T-TABLE-121](../table-metadata-location.md#s3t-table-121) · hypotheses: H-T-040 ·
    evidence: table/mutation-matrix/delete-version-token

## Notes

H-T-040 (delete becomes asynchronous once the table is non-empty) refuted: GetTable returned 404 on the very
next call in every trial. Only 3 trials with a tiny data object; a large table may behave differently.
Surprise for follow-up: the warehouse objects remain readable via HeadObject right after DeleteTable 204 (n=3,
single read each); a controller must not use the S3 data plane to confirm deletion.
