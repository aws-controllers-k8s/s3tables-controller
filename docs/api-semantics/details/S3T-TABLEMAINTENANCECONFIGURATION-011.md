<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEMAINTENANCECONFIGURATION-011: Put(disabled) on a Failed job -> entry collapses to {status: Disabled} (lastRunTimestamp AND failureMessage dropped)...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablemaintenanceconfiguration-011"></a>**S3T-TABLEMAINTENANCECONFIGURATION-011** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Put(disabled) on a Failed job -> entry collapses to {status: Disabled} (lastRunTimestamp AND failureMessage dropped)...** (hypothesis refuted; behavior confirmed)
  Compaction job had run and failed (KMS key disabled): raw entry {"failureMessage": "Insufficient access to
  perform table maintenance. Ensure that the key used to encrypt the table still exists, is active, and has a
  resource policy granting access to the S3 service principal 'maintenance.s3tables.amazonaws.com'.",
  "lastRunTimestamp": "2026-10-13T23:53:03.778Z", "status": "Failed"}.
  PutTableMaintenanceConfiguration(icebergCompaction, {disabled, 512/auto}) -> HTTP 204;
  GetTableMaintenanceJobStatus right after -> entry {"status": "Disabled"}; 5 s later -> {"status":
  "Disabled"}. Put(icebergCompaction, {enabled, 512/auto}) -> HTTP 204; right after -> entry
  {"failureMessage": "Insufficient access to perform table maintenance. Ensure that the key used to encrypt
  the table still exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'.", "lastRunTimestamp": "2026-10-13T23:53:03.778Z", "status": "Failed"}.
  - ACK: hook:sdk_read_one_post_set_output, synced.when · ops: PutTableMaintenanceConfiguration,
    GetTableMaintenanceJobStatus · fields: status.icebergCompaction.status,
    status.icebergCompaction.lastRunTimestamp, status.icebergCompaction.failureMessage
  - repro: job in Failed (with lastRunTimestamp+failureMessage) -> Put disabled ->
    GetTableMaintenanceJobStatus -> Put enabled -> GetTableMaintenanceJobStatus
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-235](../table-maintenance.md#s3t-table-235), [S3T-TABLE-236](../table-maintenance.md#s3t-table-236), [S3T-TABLE-237](../table-maintenance.md#s3t-table-237), [S3T-TABLE-238](../table-maintenance.md#s3t-table-238), [S3T-TABLE-242](../table-encryption-storage.md#s3t-table-242),
    [S3T-TABLEMAINTENANCECONFIGURATION-002](../table-maintenance.md#s3t-tablemaintenanceconfiguration-002), [S3T-TABLEMAINTENANCECONFIGURATION-008](../table-maintenance.md#s3t-tablemaintenanceconfiguration-008),
    [S3T-TABLEMAINTENANCECONFIGURATION-010](../table-maintenance.md#s3t-tablemaintenanceconfiguration-010), [S3T-TABLE-265](../table-maintenance.md#s3t-table-265), [S3T-TABLEMAINTENANCECONFIGURATION-009](../table-maintenance.md#s3t-tablemaintenanceconfiguration-009),
    [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001), [S3T-TABLEMAINTENANCECONFIGURATION-003](../table-maintenance.md#s3t-tablemaintenanceconfiguration-003) · hypotheses: H-U-062,
    H-U-057 · evidence: table/dependencies/kms-key-revoked-jobs

## Notes

H-U-062 Disabled-after-a-run cell REFUTED in part: the hypothesis says Disabled keeps lastRunTimestamp of the
previous run - observed members were ['status'] only (both lastRunTimestamp and failureMessage omitted). The
previous run's result is not lost: re-enabling restored the identical Failed entry (same lastRunTimestamp
2026-10-13T23:53:03.778Z and message) within 1 s, so Disabled is a view over the configuration, not a reset of
run history. H-U-057: the disable/enable flips are immediate (n=1 each).

Full original title: Put(disabled) on a Failed job -> entry collapses to {status: Disabled} (lastRunTimestamp
AND failureMessage dropped); Put(enabled) -> the full Failed entry returns
