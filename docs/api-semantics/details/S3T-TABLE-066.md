<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-066: DeleteTable -> 204 empty body; GetTable (composite and ARN) -> 404 at once; 2nd DeleteTable -> 404, same message as a never-created name
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-066"></a>**S3T-TABLE-066** `idempotency` · impact high · handled · verified 2026-10-13
  **DeleteTable -> 204 empty body; GetTable (composite and ARN) -> 404 at once; 2nd DeleteTable -> 404, same message as a never-created name**
  DeleteTable(correct versionToken) -> HTTP 204 (parsed body empty: True). GetTable composite ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; GetTable(tableArn) -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'. Second DeleteTable without token -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'; with the old token -> NotFoundException (HTTP 404) 'The
  specified table does not exist.'; DeleteTable of a never-created name -> NotFoundException (HTTP 404) 'The
  specified table does not exist.' (identical message to the second delete: True). DeleteTable of t2 WITHOUT
  versionToken -> HTTP 204.
  - ACK: exceptions.404, custom_delete · ops: DeleteTable, GetTable · fields: versionToken
  - repro: Create t1; DeleteTable(t1, token); GetTable both forms; DeleteTable(t1) twice more;
    DeleteTable(never_existed)
  - handling: handled via `templates/hooks/table/sdk_read_one_post_build_request.go.tpl:1-13; pkg/resource/table/sdk.go:79-91; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-292](../table-metadata-location.md#s3t-table-292), [S3T-TABLE-283](../table-schema.md#s3t-table-283), [S3T-TABLE-291](../table.md#s3t-table-291), [S3T-TABLE-015](../table-schema.md#s3t-table-015), [S3T-TABLE-088](../table.md#s3t-table-088), [S3T-TABLE-089](../table-metadata-location.md#s3t-table-089),
    [S3T-TABLE-099](../table-metadata-location.md#s3t-table-099), [S3T-TABLE-122](../table-metadata-location.md#s3t-table-122), [S3T-TABLE-129](../table-metadata-location.md#s3t-table-129), [S3T-TABLE-121](../table-metadata-location.md#s3t-table-121) · hypotheses: H-T-039, H-T-036 · evidence:
    table/idempotency/duplicate-create-double-delete

## Notes

H-T-039 confirmed: no distinct already-deleted signal. H-T-036 partially covered here (single trial; the
20-trial window measurement is in table/consistency-windows/create-delete-visibility).

Full original title: DeleteTable -> HTTP 204 empty body; GetTable right after -> NotFoundException (composite
and ARN); 2nd DeleteTable -> NotFoundException, same message as never-existed: True
