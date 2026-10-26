<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-009: After an out-of-band rename DeleteTable(old triple) -> 404; GetTable(tableArn) recovers the current triple; pre-rename token still deletes
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-009"></a>**S3T-TABLE-009** `identity` · impact high · handled · verified 2026-10-15, re-verified
  **After an out-of-band rename DeleteTable(old triple) -> 404; GetTable(tableArn) recovers the current triple; pre-rename token still deletes**
  Table created as ns1/t1, renamed 3 times (now ns1/t3). DeleteTable(ns1,t1) -> NotFoundException (HTTP 404)
  'The specified table does not exist.'. GetTable(tableArn=<stored ARN>) -> HTTP 200 returning
  namespace=[<current ns>], name=<current name>. DeleteTable(recovered triple with versionToken=<CreateTable
  token>) -> HTTP 204. GetTable(tableArn) afterwards -> NotFoundException (HTTP 404) 'The specified table does
  not exist.'.
  - ACK: custom_delete, is_arn_primary_key, exceptions.404 · ops: DeleteTable, GetTable, RenameTable · fields:
    tableArn, namespace, name, versionToken
  - repro: CreateTable ns1/t1 -> RenameTable (x3) -> DeleteTable(ns1,t1) -> GetTable(tableArn) ->
    DeleteTable(current namespace/name) -> GetTable(tableArn)
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122; templates/hooks/table/sdk_read_one_post_build_request.go.tpl:1-13; pkg/resource/table/sdk.go:79-91; generator.yaml:10-13; generator.yaml:31-39`
  - related: [S3T-TABLE-003](../table-metadata-location.md#s3t-table-003), [S3T-TABLE-111](../table-metadata-location.md#s3t-table-111), [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320), [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193), [S3T-TABLE-007](../table-metadata-location.md#s3t-table-007), [S3T-TABLE-005](../table-metadata-location.md#s3t-table-005),
    [S3T-TABLE-112](../table-metadata-location.md#s3t-table-112), [S3T-TABLE-316](../table-metadata-location.md#s3t-table-316) · hypotheses: H-T-099, H-T-054 · evidence: table/identity/rename-identity

## Notes

H-T-099 confirmed. H-T-054: DeleteTable guarded with the CreateTable-time versionToken after three renames ->
HTTP 204 (token still valid: True).

Full original title: After an out-of-band rename DeleteTable(old triple) -> NotFoundException (HTTP 404) 'The
specified table does not exist.'; GetTable(tableArn) recovers the triple; DeleteTable(current triple,
pre-rename versionToken) -> HTTP 204
