<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-104: Re-sending UpdateTableMetadataLocation with the current token and the SAME location (x3) -> 200, token unchanged, but modifiedAt advances
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-104"></a>**S3T-TABLE-104** `idempotency` · impact high · handled · verified 2026-10-13
  **Re-sending UpdateTableMetadataLocation with the current token and the SAME location (x3) -> 200, token unchanged, but modifiedAt advances**
  Three consecutive UpdateTableMetadataLocation calls, each with the then-current token and the location
  already in effect: outcomes ['HTTP 200', 'HTTP 200', 'HTTP 200']; token changed per call: [False, False,
  False] (0 distinct tokens over 3 calls); GetTable.modifiedAt 2026-10-13 22:33:19.776587+00:00 -> 2026-10-13
  22:33:24.985951+00:00.
  - ACK: compare.is_ignored+delta_pre_compare, hook:sdk_update_pre_build_request · ops:
    UpdateTableMetadataLocation, GetTable · fields: metadataLocation, versionToken, modifiedAt
  - repro: UpdateTableMetadataLocation(T, loc) -> UpdateTableMetadataLocation(T', loc) x3 at 1 s intervals ->
    GetTable
  - measurements: update_latency_ms=[13, 14, 17], n=3
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-101](../table-metadata-location.md#s3t-table-101), [S3T-TABLE-102](../table-metadata-location.md#s3t-table-102), [S3T-TABLE-123](../table-metadata-location.md#s3t-table-123), [S3T-TABLE-015](../table-schema.md#s3t-table-015), [S3T-TABLE-134](../table-metadata-location.md#s3t-table-134), [S3T-TABLE-108](../table-metadata-location.md#s3t-table-108),
    [S3T-TABLE-302](../table-metadata-location.md#s3t-table-302), [S3T-TABLE-098](../table-metadata-location.md#s3t-table-098), [S3T-TABLE-103](../table-metadata-location.md#s3t-table-103), [S3T-TABLE-192](../table-metadata-location.md#s3t-table-192), [S3T-TABLE-007](../table-metadata-location.md#s3t-table-007), [S3T-TABLE-124](../table-metadata-location.md#s3t-table-124), [S3T-TABLE-316](../table-metadata-location.md#s3t-table-316),
    [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320), [S3T-TABLE-019](../table-metadata-location.md#s3t-table-019), [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193) · hypotheses: H-T-135, H-T-136 · evidence:
    table/mutation-matrix/metadata-location-update

## Notes

H-T-135 QUALIFIED: the no-op re-send is accepted (200) but does NOT mint a new token - the response echoes the
same versionToken (1 distinct token over 3 calls); GetTable.modifiedAt nevertheless advanced from the
pre-resend value to the time of the last re-send. H-T-136 REFUTED: no ConflictException/BadRequestException
for an equal location. A controller re-applying Spec.metadataLocation every reconcile would not loop on tokens
but would bump modifiedAt on every loop. n=3.

Full original title: Re-sending UpdateTableMetadataLocation with the current token and the SAME location (x3)
-> HTTP 200, versionToken UNCHANGED (same token echoed), but modifiedAt advances
