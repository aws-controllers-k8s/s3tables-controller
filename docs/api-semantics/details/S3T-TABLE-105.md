<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-105: metadataLocation prefix check is a bare string-prefix match: sibling/other-bucket/foreign paths -> 400, but s3://<warehouse-host>x/.. -> 200
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-105"></a>**S3T-TABLE-105** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **metadataLocation prefix check is a bare string-prefix match: sibling/other-bucket/foreign paths -> 400, but s3://<warehouse-host>x/.. -> 200**
  With the current token: another table's warehouse in the same bucket -> BadRequestException (HTTP 400) 'The
  specified metadata location is not valid.'; a table's warehouse in another table bucket ->
  BadRequestException (HTTP 400) 'The specified metadata location is not valid.'; s3://<non-existent
  general-purpose bucket>/... -> BadRequestException (HTTP 400) 'The specified metadata location is not
  valid.'; a bucket whose name is the warehouse host plus 'x' (string-prefix match but a different bucket) ->
  HTTP 200. GetTableMetadataLocation unchanged after each rejection: {'sibling_table_warehouse': True,
  'other_table_bucket_warehouse': True, 'foreign_gp_bucket': True, 'scheme_upper_S3': True, 'scheme_s3a':
  True, 'https_virtual_hosted': True}.
  - ACK: terminal_codes, hook:sdk_update_pre_build_request · ops: UpdateTableMetadataLocation · fields:
    metadataLocation, warehouseLocation
  - repro: UpdateTableMetadataLocation(A, token, <B.warehouseLocation>/metadata/x.metadata.json), then
    s3://<other>/..., then s3://<A-host>x/...
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-278](../table-metadata-location.md#s3t-table-278), [S3T-TABLE-279](../table-metadata-location.md#s3t-table-279), [S3T-TABLE-280](../table-metadata-location.md#s3t-table-280), [S3T-TABLE-281](../table-metadata-location.md#s3t-table-281), [S3T-TABLE-106](../table-metadata-location.md#s3t-table-106), [S3T-TABLE-107](../table-metadata-location.md#s3t-table-107),
    [S3T-TABLE-135](../table-metadata-location.md#s3t-table-135) · hypotheses: H-T-131 · evidence: table/mutation-matrix/metadata-location-update

## Notes

H-T-131 confirmed for sibling-table, other-table-bucket and foreign-bucket paths (BadRequestException 400 'The
specified metadata location is not valid.'; location unchanged after each). SURPRISE: a location whose bucket
name is the warehouse host plus 'x' (string-prefix match, a different bucket) is ACCEPTED (200) and read back
as such - the controller should compare against warehouseLocation + '/' not a bare prefix. One attempt each.

Full original title: metadataLocation prefix check is a plain string-prefix match: other table/bucket/foreign
paths -> 400, but s3://<warehouse-host>x/... (a different bucket) -> 200
