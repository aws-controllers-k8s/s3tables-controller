<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-089: DeleteTable (no versionToken, empty table) -> 204 empty body in 20/20; GetTable composite and by ARN -> 404 on the next call in 20/20
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-089"></a>**S3T-TABLE-089** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTable (no versionToken, empty table) -> 204 empty body in 20/20; GetTable composite and by ARN -> 404 on the next call in 20/20**
  20 trials: DeleteTable without versionToken on a freshly created empty table -> HTTP 204 (parsed body empty:
  True; call latency 105-168 ms). GetTable composite issued 7-46 ms later -> NotFoundException in 20/20;
  GetTable(tableArn) issued 10-58 ms later -> NotFoundException in 20/20. Outcomes seen after delete:
  ['NotFoundException (HTTP 404)'].
  - ACK: exceptions.404, custom_delete · ops: DeleteTable, GetTable · fields: versionToken
  - repro: Loop 20x: CreateTable; DeleteTable(composite, no token); GetTable composite and by tableArn at once
  - measurements: after_delete_composite_first_call_ms=[9, 27, 9, 24, 13, 26, 8, 20, 8, 18, 7, 22, 8, 18, 8,
    18, 46, 19, 8, 22], after_delete_arn_first_call_ms=[20, 13, 20, 16, 23, 17, 18, 11, 17, 11, 17, 11, 17,
    11, 19, 10, 58, 11, 19, 13], delete_latency_ms=[130, 129, 119, 128, 111, 111, 140, 168, 111, 120, 118,
    119, 116, 105, 121, 110, 132, 135, 115, 121], n=20, still_visible_windows=[]
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-066](../table-metadata-location.md#s3t-table-066), [S3T-TABLE-099](../table-metadata-location.md#s3t-table-099), [S3T-TABLE-122](../table-metadata-location.md#s3t-table-122), [S3T-TABLE-129](../table-metadata-location.md#s3t-table-129), [S3T-TABLE-121](../table-metadata-location.md#s3t-table-121), [S3T-TABLE-067](../table-metadata-location.md#s3t-table-067),
    [S3T-TABLE-064](../table-metadata-location.md#s3t-table-064), [S3T-TABLE-028](../table-metadata-location.md#s3t-table-028), [S3T-TABLE-313](../table-schema.md#s3t-table-313), [S3T-TABLE-231](../table.md#s3t-table-231) · hypotheses: H-T-036 · evidence:
    table/consistency-windows/create-delete-visibility

## Notes

H-T-036 confirmed: delete of an empty table is synchronous and the token is optional; the controller may
finalize on 204 without polling (empty tables; non-empty tables are covered by
table/mutation-matrix/delete-version-token).

Full original title: DeleteTable (no versionToken, empty table) -> HTTP 204 empty body in 20/20; GetTable ->
404 on the next call by composite 20/20 and by ARN 20/20
