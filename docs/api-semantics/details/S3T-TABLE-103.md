<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-103: UpdateTableMetadataLocation: superseded-location or garbage versionToken -> ConflictException 409; empty token -> ValidationException 400
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-103"></a>**S3T-TABLE-103** `error-code` · impact high · handled · verified 2026-10-13
  **UpdateTableMetadataLocation: superseded-location or garbage versionToken -> ConflictException 409; empty token -> ValidationException 400**
  Stale token (the CreateTable token after 3 later commits) -> ConflictException (HTTP 409) 'Provided version
  token does not match the table version token.'. Garbage token 'bogus-token-ackq' -> ConflictException (HTTP 409)
  'Provided version token does not match the table version token.'. Empty string token (SDK validation
  bypassed) -> ValidationException (HTTP 400) '1 validation error detected: Value at 'versionToken' failed to
  satisfy constraint: Member must have length greater than or equal to 1'. GetTableMetadataLocation after each
  still returned the last accepted token/location: True/True/True.
  - ACK: requeue, terminal_codes · ops: UpdateTableMetadataLocation · fields: versionToken
  - repro: UpdateTableMetadataLocation with the pre-commit token, with 'bogus-token-ackq', and with '' ->
    GetTableMetadataLocation
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-134](../table-metadata-location.md#s3t-table-134), [S3T-TABLE-108](../table-metadata-location.md#s3t-table-108), [S3T-TABLE-104](../table-metadata-location.md#s3t-table-104), [S3T-TABLE-302](../table-metadata-location.md#s3t-table-302), [S3T-TABLE-098](../table-metadata-location.md#s3t-table-098), [S3T-TABLE-192](../table-metadata-location.md#s3t-table-192),
    [S3T-TABLE-065](../table-metadata-location.md#s3t-table-065), [S3T-TABLE-051](../table-metadata-location.md#s3t-table-051), [S3T-TABLE-123](../table-metadata-location.md#s3t-table-123), [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320) · hypotheses: H-T-134 · evidence:
    table/mutation-matrix/metadata-location-update

## Notes

H-T-134 confirmed: stale and garbage tokens -> ConflictException 409 'Provided version token does not match
the table version token.'; the empty-string token -> ValidationException 400 '1 validation error detected:
Value at 'versionToken' failed to satisfy constraint: Member must have length greater than ...' - note
ValidationException is NOT in the operation's modeled error list (BadRequestException is). Location unchanged
after each. One attempt each.

Full original title: UpdateTableMetadataLocation with a stale versionToken -> ConflictException 409; garbage
token -> ConflictException 409; empty token -> ValidationException 400; location unchanged after each
