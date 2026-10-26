<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-192: UpdateTableMetadataLocation same-token race: exactly one 200 + one 409 in 10/10 rounds; also 200 for a missing .metadata.json object
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-192"></a>**S3T-TABLE-192** `idempotency` · impact high · handled · verified 2026-10-13
  **UpdateTableMetadataLocation same-token race: exactly one 200 + one 409 in 10/10 rounds; also 200 for a missing .metadata.json object**
  UpdateTableMetadataLocation(versionToken=current,
  metadataLocation=<warehouseLocation>/metadata/00000-ackq-probe.metadata.json) on a fresh table where that
  object does NOT exist -> HTTP 200, versionToken changed in the response. Then 10 rounds of two concurrent
  updates carrying the same (fresh) versionToken and different locations: code pairs {'ConflictException (HTTP
  409)+HTTP 200': 10}; messages {"B": "Provided version token does not match the table version token.", "A":
  "Provided version token does not match the table version token."}; in 10/10 rounds GetTableMetadataLocation
  afterwards showed the winner's location and in 10/10 rounds the versionToken had changed.
  - ACK: requeue, custom_update · ops: UpdateTableMetadataLocation, GetTableMetadataLocation · fields:
    versionToken, metadataLocation
  - repro: GetTableMetadataLocation -> token; 2 threads x UpdateTableMetadataLocation(token, locA|locB);
    GetTableMetadataLocation; repeat 10x
  - measurements: cas_round_wall_s=[0.02, 0.01, 0.01, 0.02, 0.02, 0.03, 0.02, 0.01, 0.01, 0.01], n=10
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491; generator.yaml:84-88; generator.yaml:103-104`
  - related: [S3T-TABLE-100](../table-metadata-location.md#s3t-table-100), [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320), [S3T-TABLE-051](../table-metadata-location.md#s3t-table-051), [S3T-TABLE-108](../table-metadata-location.md#s3t-table-108), [S3T-TABLE-134](../table-metadata-location.md#s3t-table-134), [S3T-TABLE-104](../table-metadata-location.md#s3t-table-104),
    [S3T-TABLE-302](../table-metadata-location.md#s3t-table-302), [S3T-TABLE-098](../table-metadata-location.md#s3t-table-098), [S3T-TABLE-103](../table-metadata-location.md#s3t-table-103), [S3T-TABLE-064](../table-metadata-location.md#s3t-table-064), [S3T-TABLE-065](../table-metadata-location.md#s3t-table-065), [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193) · hypotheses:
    H-T-145 · evidence: table/limits/policy-size-write-races

## Notes

H-T-145 confirmed: in 10/10 rounds two concurrent UpdateTableMetadataLocation calls with the same fresh
versionToken returned exactly one 200 and one ConflictException 409 'Provided version token does not match the
table version token.'; GetTableMetadataLocation then showed the winner's location and a changed token every
time (wall 0.01-0.02 s, no serialization delay). Surprise: the very first update pointed at
<warehouseLocation>/metadata/00000-ackq-probe.metadata.json, an object that does not exist, and was accepted
with 200 (token changed) - the control plane validates only the URI prefix/suffix, not the object, so a
controller can set a dangling metadata location.

Full original title: UpdateTableMetadataLocation does not check that the object exists (200 for a missing
.metadata.json); same-token races -> exactly one 200 + one 409 in 10/10 rounds
