<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-235: Maintenance grant revoked: UpdateTableMetadataLocation + warehouse GetObject/PutObject still 200; grant checked only by CreateTable and jobs
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-235"></a>**S3T-TABLE-235** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Maintenance grant revoked: UpdateTableMetadataLocation + warehouse GetObject/PutObject still 200; grant checked only by CreateTable and jobs**
  After PutKeyPolicy(root statement only; the maintenance principal grant removed) the caller (under the root
  statement) read the current metadata.json: GetObject -> HTTP 200 (ServerSideEncryption aws:kms, SSEKMSKeyId
  is the bucket key: True); PutObject of a copy -> HTTP 200; UpdateTableMetadataLocation(versionToken, new
  location) -> HTTP 200. Job status right after:
  'icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run'.
  - ACK: terminal_codes, references · ops: UpdateTableMetadataLocation, PutKeyPolicy, GetTable · fields:
    metadataLocation, versionToken, encryptionConfiguration.kmsKeyArn
  - repro: SSE-KMS table -> PutKeyPolicy(root only) -> GetObject/PutObject metadata.json ->
    UpdateTableMetadataLocation
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-253](../table-maintenance.md#s3t-table-253), [S3T-TABLE-301](../table-maintenance.md#s3t-table-301), [S3T-TABLE-236](../table-maintenance.md#s3t-table-236), [S3T-TABLE-237](../table-maintenance.md#s3t-table-237), [S3T-TABLE-238](../table-maintenance.md#s3t-table-238),
    [S3T-TABLEMAINTENANCECONFIGURATION-011](../table-maintenance.md#s3t-tablemaintenanceconfiguration-011), [S3T-TABLE-242](../table-encryption-storage.md#s3t-table-242) · hypotheses: H-U-126 · evidence:
    table/dependencies/kms-key-revoked-jobs

## Notes

H-U-126 context: the control plane does not re-check the maintenance principal grant at
UpdateTableMetadataLocation (200), only CreateTable rejects (400, see the phase finding). The commit gives the
maintenance jobs something to do; they then fail asynchronously (sibling findings).

Full original title: Maintenance grant revoked from the key: UpdateTableMetadataLocation (and the caller's
warehouse GetObject/PutObject) still -> 200; the grant is checked only at CreateTable and by the jobs
