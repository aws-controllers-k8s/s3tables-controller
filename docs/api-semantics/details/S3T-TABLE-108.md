<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-108: Rollback: UpdateTableMetadataLocation back to the original 00000-*.metadata.json -> 200; response versionToken == the CreateTable token
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-108"></a>**S3T-TABLE-108** `idempotency` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **Rollback: UpdateTableMetadataLocation back to the original 00000-*.metadata.json -> 200; response versionToken == the CreateTable token**
  After 20 accepted updates away from the create-time file (last location '<WH>/a/b/c/d.metadata.json'),
  UpdateTableMetadataLocation(current token, <loc0 = 00000-<uuid>.metadata.json>) -> HTTP 200;
  GetTableMetadataLocation shows loc0 again: True. HeadObject: loc0 still exists=True; the dangling 00001 path
  exists=False (never created). No earlier metadata file was deleted by any update.
  - ACK: docs-only, compare.is_ignored+delta_pre_compare · ops: UpdateTableMetadataLocation,
    GetTableMetadataLocation · fields: metadataLocation
  - repro: CreateTable with schema (loc0) -> UpdateTableMetadataLocation to
    <wh>/metadata/00001-<uuid>.metadata.json -> UpdateTableMetadataLocation back to loc0 ->
    GetTableMetadataLocation
  - measurements: update_latency_ms_min=11, update_latency_ms_max=17, n=21
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLE-100](../table-metadata-location.md#s3t-table-100), [S3T-TABLE-192](../table-metadata-location.md#s3t-table-192), [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320), [S3T-TABLE-051](../table-metadata-location.md#s3t-table-051), [S3T-TABLE-134](../table-metadata-location.md#s3t-table-134), [S3T-TABLE-104](../table-metadata-location.md#s3t-table-104),
    [S3T-TABLE-302](../table-metadata-location.md#s3t-table-302), [S3T-TABLE-098](../table-metadata-location.md#s3t-table-098), [S3T-TABLE-103](../table-metadata-location.md#s3t-table-103) · hypotheses: H-T-147 · evidence:
    table/mutation-matrix/metadata-location-update

## Notes

H-T-147 confirmed (no monotonicity check; both files untouched; the probe wrote no objects). SURPRISE: the
token returned for the rollback equals the create-time token exactly, and the same-location re-sends also
returned the same token - versionToken appears to be a deterministic function of the current metadataLocation
(not a per-write nonce), so a token saved before an external commit becomes valid again if the location is
restored. 21 accepted updates, latency 11-17 ms.

Full original title: Rollback: UpdateTableMetadataLocation back to the original 00000-*.metadata.json -> HTTP
200 and the response versionToken is byte-identical to the CreateTable token

Suspected controller bug confirmed by evidence: metadataLocation is engine-managed state and must not be a
reconciled spec field: on any table in use the observed value is non-nil from the first read (schema-at-create
writes 00000-<uuid>.metadata.json synchronously, [S3T-TABLE-015](../table-schema.md#s3t-table-015)/033) and moves with every commit
([S3T-TABLE-101](../table-metadata-location.md#s3t-table-101), [S3T-TABLE-123](../table-metadata-location.md#s3t-table-123)), so an unset spec diffs forever, while a pinned spec makes the controller
re-issue UpdateTableMetadataLocation(latest token, old location) after each engine commit - the service
accepts the rollback with no monotonicity check and even returns the old token ([S3T-TABLE-108](../table-metadata-location.md#s3t-table-108), [S3T-TABLE-134](../table-metadata-location.md#s3t-table-134))
and the engine's next commit then fails with ConflictException 409 'Provided version token does not match the
table version token.' ([S3T-TABLE-302](../table-metadata-location.md#s3t-table-302)). Also drop the create-time Synced=False/requeue path: a desired
metadataLocation on a fresh table can only fail with BadRequestException 400 -> Terminal, because the location
must string-prefix the server-assigned warehouseLocation ([S3T-TABLE-105](../table-metadata-location.md#s3t-table-105)) that is unknowable before CreateTable
([S3T-TABLE-028](../table-metadata-location.md#s3t-table-028), [S3T-TABLE-122](../table-metadata-location.md#s3t-table-122)). Fix: surface metadataLocation as read-only status (or late-initialize and
ignore it in compare).
