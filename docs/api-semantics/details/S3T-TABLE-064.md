<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-064: Duplicate CreateTable (same bucket/namespace/name) -> ConflictException 409 at t+0/5/10 s and with different metadata/tags; table untouched
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-064"></a>**S3T-TABLE-064** `idempotency` · impact high · handled · verified 2026-10-13
  **Duplicate CreateTable (same bucket/namespace/name) -> ConflictException 409 at t+0/5/10 s and with different metadata/tags; table untouched**
  After CreateTable(t1, format=ICEBERG, tags) -> 200: identical CreateTable at t+0 -> ConflictException (HTTP 409)
  'A table with an identical name already exists in the namespace.'; t+5 s -> ConflictException (HTTP 409) 'A
  table with an identical name already exists in the namespace.'; t+10 s -> ConflictException (HTTP 409) 'A
  table with an identical name already exists in the namespace.'; with metadata.iceberg.schema ->
  ConflictException (HTTP 409) 'A table with an identical name already exists in the namespace.'; without tags
  -> ConflictException (HTTP 409) 'A table with an identical name already exists in the namespace.'; with an
  extra tag -> ConflictException (HTTP 409) 'A table with an identical name already exists in the namespace.'.
  GetTable afterwards: tableARN/versionToken/createdAt/modifiedAt/warehouseLocation unchanged = {'tableARN':
  True, 'versionToken': True, 'createdAt': True, 'modifiedAt': True, 'warehouseLocation': True}; ListTables
  shows 1 entry named t1.
  - ACK: custom_create, terminal_codes · ops: CreateTable, GetTable · fields: name, namespace, versionToken
  - repro: CreateTable t1; CreateTable t1 again at t+0, t+5 s, t+10 s with identical params, then with
    metadata, without tags, with extra tags; GetTable; ListTables
  - measurements: retry_spacing_s=[0, 5, 10], n=3
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-NAMESPACE-036](../namespace.md#s3t-namespace-036), [S3T-TABLE-274](../table.md#s3t-table-274), [S3T-TABLE-276](../table.md#s3t-table-276), [S3T-NAMESPACE-021](../namespace.md#s3t-namespace-021), [S3T-TABLE-050](../table.md#s3t-table-050), [S3T-TABLE-065](../table-metadata-location.md#s3t-table-065),
    [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193), [S3T-TABLE-192](../table-metadata-location.md#s3t-table-192), [S3T-TABLE-067](../table-metadata-location.md#s3t-table-067), [S3T-TABLE-122](../table-metadata-location.md#s3t-table-122), [S3T-TABLE-129](../table-metadata-location.md#s3t-table-129), [S3T-TABLE-089](../table-metadata-location.md#s3t-table-089), [S3T-TABLE-028](../table-metadata-location.md#s3t-table-028) ·
    hypotheses: H-T-023, H-T-024, H-T-025 · evidence: table/idempotency/duplicate-create-double-delete

## Notes

H-T-023 confirmed. H-T-024 (PUT-like idempotent 200) refuted. H-T-025 (409 is permanent across 3 spaced
retries) confirmed. Message text: 'A table with an identical name already exists in the namespace.'. The
duplicate-create ConflictException message differs from the versionToken-mismatch ConflictException ('Provided
version token does not match the table version token.'), so the message is the only way to tell 'already
exists' (permanent) from a token conflict (retryable after re-read).

Full original title: Duplicate CreateTable (same bucket/namespace/name) -> ConflictException (HTTP 409) at
t+0/5/10 s and with different metadata/tags; existing table untouched: True
