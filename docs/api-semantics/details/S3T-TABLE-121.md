<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-121: GetTable/GTML/UTML on a missing table, namespace or table bucket -> NotFoundException 404, one message 'The specified table does not exist.'
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-121"></a>**S3T-TABLE-121** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **GetTable/GTML/UTML on a missing table, namespace or table bucket -> NotFoundException 404, one message 'The specified table does not exist.'**
  Missing table in an existing namespace: GetTable -> NotFoundException (HTTP 404) 'The specified table does
  not exist.'; GetTableMetadataLocation -> NotFoundException (HTTP 404) 'The specified table does not exist.';
  UpdateTableMetadataLocation(bogus token, s3://<nonexistent>/metadata/x.metadata.json) -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'. Missing namespace: NotFoundException (HTTP 404) 'The
  specified table does not exist.' / NotFoundException (HTTP 404) 'The specified table does not exist.' /
  NotFoundException (HTTP 404) 'The specified table does not exist.'. Missing table bucket ARN:
  NotFoundException (HTTP 404) 'The specified table does not exist.' / NotFoundException (HTTP 404) 'The
  specified table does not exist.' / NotFoundException (HTTP 404) 'The specified table does not exist.'.
  - ACK: exceptions.404 · ops: GetTable, GetTableMetadataLocation, UpdateTableMetadataLocation · fields: name,
    namespace, tableBucketARN
  - repro: Call the three operations with a bogus table name, a bogus namespace and a bogus table bucket ARN
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-066](../table-metadata-location.md#s3t-table-066), [S3T-TABLE-089](../table-metadata-location.md#s3t-table-089), [S3T-TABLE-099](../table-metadata-location.md#s3t-table-099), [S3T-TABLE-122](../table-metadata-location.md#s3t-table-122), [S3T-TABLE-129](../table-metadata-location.md#s3t-table-129), [S3T-TABLE-245](../table-metadata-location.md#s3t-table-245),
    [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320), [S3T-TABLE-003](../table-metadata-location.md#s3t-table-003), [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193) · hypotheses: H-T-146 · evidence:
    table/idempotency/metadata-location-after-delete

## Notes

H-T-146 NotFound part QUALIFIED: all nine cells are NotFoundException (HTTP 404), but the message is identical
for a missing table, a missing namespace and a missing table bucket ARN ('The specified table does not
exist.'), so the message does NOT distinguish which level is missing (hypothesis expected it to).
UpdateTableMetadataLocation reports 404 before validating the token/location. The IAM-denied part is in
table/error-taxonomy/iam-putdata-denied. One attempt per cell.

Full original title: GetTable/GetTableMetadataLocation/UpdateTableMetadataLocation on a missing table, missing
namespace or missing table bucket -> NotFoundException 404 with the SAME message 'The specified table does not
exist.'
