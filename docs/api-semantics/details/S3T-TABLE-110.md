<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-110: versionToken is NOT changed by TagResource, PutTableMaintenanceConfiguration or RenameTable (newName / newNamespaceName)...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-110"></a>**S3T-TABLE-110** `idempotency` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **versionToken is NOT changed by TagResource, PutTableMaintenanceConfiguration or RenameTable (newName / newNamespaceName)...** (hypothesis refuted; behavior confirmed)
  Table with metadata. TagResource -> HTTP 200: token changed False, modifiedAt changed True,
  UpdateTableMetadataLocation with the pre-tag token -> HTTP 200.
  PutTableMaintenanceConfiguration(icebergSnapshotManagement enabled) -> HTTP 204: token changed False,
  modifiedAt changed True, CAS with the pre-call token -> HTTP 200. RenameTable(newName) -> HTTP 204: token
  changed False, modifiedAt changed True, CAS with the pre-rename token -> HTTP 200.
  RenameTable(newNamespaceName) -> HTTP 204: token changed False, modifiedAt changed True, CAS with the
  pre-rename token -> HTTP 200.
  - ACK: hook:sdk_update_pre_build_request, requeue · ops: TagResource, PutTableMaintenanceConfiguration,
    RenameTable, UpdateTableMetadataLocation, GetTable · fields: versionToken, modifiedAt
  - repro: CreateTable with schema -> GetTable token -> TagResource / PutTableMaintenanceConfiguration /
    RenameTable, each followed by GetTable and UpdateTableMetadataLocation with the pre-mutation token
  - measurements: rename_latency_ms=[130, 141], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-155](../table-maintenance.md#s3t-table-155), [S3T-TABLE-206](../table-maintenance.md#s3t-table-206) · hypotheses: H-T-139 · evidence:
    table/identity/version-token-rename

## Notes

H-T-139 REFUTED: none of the four mutations changed GetTable.versionToken (each did advance modifiedAt), and
UpdateTableMetadataLocation with the token read before each mutation -> HTTP 200 every time. Together with the
metadata-location-update probe (same-location re-send echoes the same token; rollback returns the create-time
token) the token behaves like a function of the current metadataLocation, not a per-mutation counter. Each
mutation tried once.

Full original title: versionToken is NOT changed by TagResource, PutTableMaintenanceConfiguration or
RenameTable (newName / newNamespaceName); the pre-mutation token still works as the
UpdateTableMetadataLocation CAS token
