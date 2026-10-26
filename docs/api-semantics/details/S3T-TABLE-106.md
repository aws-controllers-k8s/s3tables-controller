<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-106: metadataLocation: 'S3://', 's3a://', https:// -> 400; '//' path, '%20' and a literal space -> 200, read back byte-exact (no normalization)
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-106"></a>**S3T-TABLE-106** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **metadataLocation: 'S3://', 's3a://', https:// -> 400; '//' path, '%20' and a literal space -> 200, read back byte-exact (no normalization)**
  scheme_upper_S3: sent
  'S3://1ae9703d-2fb4-4723-ewxhmr5r89mkgmuaf63a1ddwsxxxeusw2b--table-s3/metadata/x.metadata.json' ->
  BadRequestException (HTTP 400) 'The specified metadata location is not valid.'; scheme_s3a: sent
  's3a://1ae9703d-2fb4-4723-ewxhmr5r89mkgmuaf63a1ddwsxxxeusw2b--table-s3/metadata/x.metadata.json' ->
  BadRequestException (HTTP 400) 'The specified metadata location is not valid.'; https_virtual_hosted: sent
  'https://1ae9703d-2fb4-4723-ewxhmr5r89mkgmuaf63a1ddwsxxxeusw2b--table-s3.s3.us-west-2.amazonaws.com/metadata/x.metadata.json'
  -> BadRequestException (HTTP 400) 'The specified metadata location is not valid.'; double_slashes: sent
  '<WH>//metadata//x.metadata.json' -> HTTP 200, GetTableMetadataLocation returns
  '<WH>//metadata//x.metadata.json' (byte-identical: True); percent20: sent
  '<WH>/metadata/a%20b.metadata.json' -> HTTP 200, GetTableMetadataLocation returns
  '<WH>/metadata/a%20b.metadata.json' (byte-identical: True); space: sent '<WH>/metadata/a b.metadata.json' ->
  HTTP 200, GetTableMetadataLocation returns '<WH>/metadata/a b.metadata.json' (byte-identical: True).
  - ACK: docs-only, compare.is_ignored+delta_pre_compare · ops: UpdateTableMetadataLocation,
    GetTableMetadataLocation · fields: metadataLocation
  - repro: UpdateTableMetadataLocation with each rewritten location string -> GetTableMetadataLocation
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-278](../table-metadata-location.md#s3t-table-278), [S3T-TABLE-279](../table-metadata-location.md#s3t-table-279), [S3T-TABLE-280](../table-metadata-location.md#s3t-table-280), [S3T-TABLE-281](../table-metadata-location.md#s3t-table-281), [S3T-TABLE-105](../table-metadata-location.md#s3t-table-105), [S3T-TABLE-107](../table-metadata-location.md#s3t-table-107),
    [S3T-TABLE-135](../table-metadata-location.md#s3t-table-135) · hypotheses: H-T-142 · evidence: table/mutation-matrix/metadata-location-update

## Notes

H-T-142 confirmed. One attempt each.

Full original title: metadataLocation scheme/case/encoding: 'S3://' -> BadRequestException (HTTP 400);
's3a://' -> BadRequestException (HTTP 400); https:// -> BadRequestException (HTTP 400); '//' path -> HTTP 200
(read back exact=True); '%20' -> HTTP 200 (read back exact=True); literal space -> HTTP 200 (read back
exact=True)
