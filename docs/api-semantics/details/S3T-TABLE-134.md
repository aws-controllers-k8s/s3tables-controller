<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-134: versionToken is a deterministic function of the full metadataLocation: A->B->A->B gives A0==A1==A2, B1==B2; another table's same key differs
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-134"></a>**S3T-TABLE-134** `idempotency` · impact medium · handled · verified 2026-10-13
  **versionToken is a deterministic function of the full metadataLocation: A->B->A->B gives A0==A1==A2, B1==B2; another table's same key differs**
  Table with location A (create token A0). UpdateTableMetadataLocation A->B, B->A, A->B, B->A (all HTTP 200):
  token after each return to A equals A0: True/True; the two B tokens are equal: True; A and B tokens differ:
  True. A second table pointing at the same relative key under ITS warehouse gets the same token as B: False
  (so the token depends on the full location, including the warehouse). Token is 20 chars, hex: True.
  - ACK: hook:sdk_update_pre_build_request, docs-only · ops: UpdateTableMetadataLocation, GetTable · fields:
    versionToken, metadataLocation
  - repro: CreateTable with schema -> UpdateTableMetadataLocation to B, back to A, to B, back to A -> compare
    the five tokens
  - measurements: updates=4, n=4
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-108](../table-metadata-location.md#s3t-table-108), [S3T-TABLE-104](../table-metadata-location.md#s3t-table-104), [S3T-TABLE-302](../table-metadata-location.md#s3t-table-302), [S3T-TABLE-098](../table-metadata-location.md#s3t-table-098), [S3T-TABLE-103](../table-metadata-location.md#s3t-table-103), [S3T-TABLE-192](../table-metadata-location.md#s3t-table-192) ·
    hypotheses: H-T-135, H-T-139 · evidence: table/creative/token-and-suffix-followup

## Notes

Creative follow-up: explains why same-location re-sends (H-T-135), renames/tags (H-T-139) and the rollback
(H-T-147) did not change the token - it is derived from the location, not a write counter. Consequence: a
token cached before an external commit becomes valid again if the location is restored; ConflictException
means 'the location changed since you read', nothing more.

Full original title: versionToken is a deterministic function of the table's metadataLocation: A->B->A->B
yields A0==A1==A2 True, B1==B2 True; another table at the same relative key gets a DIFFERENT token
