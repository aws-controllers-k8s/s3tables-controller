<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-065: DeleteTable with a garbage/other table's/random/upper-cased/2048-char versionToken -> ConflictException 409, table untouched; no 400 seen
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-065"></a>**S3T-TABLE-065** `error-code` · impact medium · handled · verified 2026-10-13
  **DeleteTable with a garbage/other table's/random/upper-cased/2048-char versionToken -> ConflictException 409, table untouched; no 400 seen**
  On an existing table: versionToken=garbage_abc -> ConflictException (HTTP 409) 'Provided version token does
  not match the table version token.'; versionToken=other_tables_token -> ConflictException (HTTP 409)
  'Provided version token does not match the table version token.'; versionToken=empty -> ParamValidationError
  (HTTP None) 'Parameter validation failed:
  Invalid length for parameter versionToken, value: 0, valid min length: 1';
  versionToken=same_length_hex_random -> ConflictException (HTTP 409) 'Provided version token does not match
  the table version token.'; versionToken=uppercased_own_token -> ConflictException (HTTP 409) 'Provided
  version token does not match the table version token.'; versionToken=very_long -> ConflictException (HTTP 409)
  'Provided version token does not match the table version token.'. GetTable afterwards -> table exists with
  its original token: True.
  - ACK: custom_delete, requeue · ops: DeleteTable · fields: versionToken
  - repro: Create t1,t2; DeleteTable(t1, versionToken='abc'), (t1, token(t2)), (t1, ''), (t1, random hex),
    (t1, upper-cased own token), (t1, 2048 x 'a')
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-103](../table-metadata-location.md#s3t-table-103), [S3T-TABLE-051](../table-metadata-location.md#s3t-table-051), [S3T-TABLE-098](../table-metadata-location.md#s3t-table-098), [S3T-TABLE-123](../table-metadata-location.md#s3t-table-123), [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320), [S3T-TABLE-064](../table-metadata-location.md#s3t-table-064),
    [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193), [S3T-TABLE-192](../table-metadata-location.md#s3t-table-192) · hypotheses: H-T-038 · evidence:
    table/idempotency/duplicate-create-double-delete

## Notes

H-T-038 confirmed: every server-side mismatch is ConflictException 'Provided version token does not match the
table version token.'; there is no BadRequestException for a malformed token. The empty-string cell never
reached the server: boto3 rejects it client-side (ParamValidationError, min length 1). A 409 on delete
therefore means 'stale/corrupt token' and should trigger a GetTable re-read (or a tokenless delete), not a
terminal error.

Full original title: DeleteTable with a garbage, another table's, random, upper-cased or 2048-char
versionToken -> ConflictException 409 (never 400); table untouched
