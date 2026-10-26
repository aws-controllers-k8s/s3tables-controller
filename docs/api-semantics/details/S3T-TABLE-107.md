<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-107: metadataLocation: x.metadata.json, x.metadata.json.gz, x.metadata.json.gzip, any folder under the warehouse -> 200; .json/.JSON/.avro -> 400
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-107"></a>**S3T-TABLE-107** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **metadataLocation: x.metadata.json, x.metadata.json.gz, x.metadata.json.gzip, any folder under the warehouse -> 200; .json/.JSON/.avro -> 400**
  Under the table's own warehouse, with fresh tokens: '<WH>/metadata/x.json' -> BadRequestException (HTTP 400)
  'The specified metadata location is not valid.'; '<WH>/metadata/x.metadata.JSON' -> BadRequestException
  (HTTP 400) 'The specified metadata location is not valid.'; '<WH>/metadata/x.metadata.json.gzip' -> HTTP
  200; '<WH>/metadata/x.avro' -> BadRequestException (HTTP 400) 'The specified metadata location is not
  valid.'; '<WH>/metadata/x.metadata.json' -> HTTP 200; '<WH>/metadata/x.metadata.json.gz' -> HTTP 200;
  '<WH>/metadata/' -> BadRequestException (HTTP 400) 'The specified metadata location is not valid.'; '<WH>'
  -> BadRequestException (HTTP 400) 'The specified metadata location is not valid.'; '<WH>/metadata' ->
  BadRequestException (HTTP 400) 'The specified metadata location is not valid.';
  '<WH>/metadata/.metadata.json' -> HTTP 200; '<WH>/data/v2.metadata.json' -> HTTP 200;
  '<WH>/v2.metadata.json' -> HTTP 200; '<WH>/a/b/c/d.metadata.json' -> HTTP 200.
  - ACK: terminal_codes, hook:sdk_update_pre_build_request, docs-only · ops: UpdateTableMetadataLocation ·
    fields: metadataLocation
  - repro: UpdateTableMetadataLocation with each suffix/path variant under the warehouse (objects need not
    exist)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-278](../table-metadata-location.md#s3t-table-278), [S3T-TABLE-279](../table-metadata-location.md#s3t-table-279), [S3T-TABLE-280](../table-metadata-location.md#s3t-table-280), [S3T-TABLE-281](../table-metadata-location.md#s3t-table-281), [S3T-TABLE-105](../table-metadata-location.md#s3t-table-105), [S3T-TABLE-106](../table-metadata-location.md#s3t-table-106),
    [S3T-TABLE-135](../table-metadata-location.md#s3t-table-135) · hypotheses: H-T-133, H-T-141 · evidence: table/mutation-matrix/metadata-location-update

## Notes

H-T-133 QUALIFIED: .metadata.json and .metadata.json.gz accepted; .json, .metadata.JSON (case-sensitive),
.avro, trailing '/', bare warehouse and bare '/metadata' rejected with BadRequestException 400 'The specified
metadata location is not valid.'; but '.metadata.json.gzip' was ACCEPTED (contrary to the literal doc rule -
the check may be 'contains .metadata.json' rather than 'ends with'), and '/metadata/.metadata.json' (no
basename) accepted. H-T-141 confirmed: <wh>/data/v2.metadata.json, <wh>/v2.metadata.json and
<wh>/a/b/c/d.metadata.json all accepted. One attempt each.

Full original title: metadataLocation suffix: .metadata.json / .metadata.json.gz / .metadata.json.gzip / bare
'.metadata.json' and any folder under the warehouse accepted; .json, .metadata.JSON, .avro, '/', bare
warehouse, /metadata -> 400
