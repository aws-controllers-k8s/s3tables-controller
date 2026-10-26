<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-067: Immediate CreateTable after DeleteTable (same name) -> 200 in 10/10 (409 in 0); new tableARN every time, old ARN -> 404 in 10/10
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-067"></a>**S3T-TABLE-067** `identity` · impact high · handled · verified 2026-10-13
  **Immediate CreateTable after DeleteTable (same name) -> 200 in 10/10 (409 in 0); new tableARN every time, old ARN -> 404 in 10/10**
  10 trials of CreateTable -> DeleteTable(no token) -> CreateTable(same name) issued 308-403 ms after the
  delete response: immediate HTTP 200 in 10, ConflictException in 0 (retries to 200: []). The re-created table
  has a different tableARN (uuid) in all trials: True, a different versionToken: True, a different
  warehouseLocation (trial 0): True. GetTable(old tableArn) afterwards -> NotFoundException (HTTP 404) 'The
  specified table does not exist.' in 10/10 trials; GetTable(new tableArn) -> HTTP 200.
  - ACK: is_arn_primary_key, custom_find, requeue · ops: CreateTable, DeleteTable, GetTable · fields:
    tableARN, versionToken, warehouseLocation
  - repro: Loop 10x: CreateTable t; DeleteTable t; CreateTable t at once (retry each 1 s on 409); GetTable by
    old and new ARN; DeleteTable t
  - measurements: recreate_gap_ms=[361, 326, 308, 338, 336, 403, 316, 326, 329, 368], immediate_200=10,
    conflict_409=0, n=10
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-122](../table-metadata-location.md#s3t-table-122), [S3T-TABLE-129](../table-metadata-location.md#s3t-table-129), [S3T-TABLE-064](../table-metadata-location.md#s3t-table-064), [S3T-TABLE-089](../table-metadata-location.md#s3t-table-089), [S3T-TABLE-028](../table-metadata-location.md#s3t-table-028) · hypotheses: H-T-041,
    H-T-042 · evidence: table/idempotency/duplicate-create-double-delete

## Notes

H-T-041 confirmed. H-T-042 (409 name reservation for 1-2 s) refuted. A controller holding the old ARN sees 404
by ARN while the composite read returns the NEW table: it must compare ARNs and re-adopt.

Full original title: Immediate CreateTable after DeleteTable (same name) -> HTTP 200 in 10/10 trials, 409 in
0; new tableARN every time: True; old ARN -> 404 in 10/10
