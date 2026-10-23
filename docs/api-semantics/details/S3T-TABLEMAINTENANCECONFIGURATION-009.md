<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEMAINTENANCECONFIGURATION-009: GetTableMaintenanceJobStatus entry members: Not_Yet_Run {status}, Disabled {status}, Successful {status, lastRunTimestamp}; absent = omitted
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablemaintenanceconfiguration-009"></a>**S3T-TABLEMAINTENANCECONFIGURATION-009** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **GetTableMaintenanceJobStatus entry members: Not_Yet_Run {status}, Disabled {status}, Successful {status, lastRunTimestamp}; absent = omitted**
  Raw HTTP bodies captured before parsing. Not_Yet_Run (fresh tables): per-type object ['status']. Disabled
  (Put disabled on a never-run job): ['status']. Successful (after a metadata commit): ['lastRunTimestamp',
  'status'] - lastRunTimestamp is an ISO-8601 string with milliseconds and 'Z' (e.g.
  2026-10-13T23:29:52.066Z). No member was ever sent as null; failureMessage was never present. Failed was not
  reached.
  - ACK: hook:sdk_read_one_post_set_output, late_initialize · ops: GetTableMaintenanceJobStatus · fields:
    status.*.status, status.*.lastRunTimestamp, status.*.failureMessage
  - repro: GetTableMaintenanceJobStatus in each reachable state; inspect the raw JSON body
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001), [S3T-TABLEMAINTENANCECONFIGURATION-003](../table-maintenance.md#s3t-tablemaintenanceconfiguration-003),
    [S3T-TABLEMAINTENANCECONFIGURATION-011](../table-maintenance.md#s3t-tablemaintenanceconfiguration-011), [S3T-TABLE-236](../table-maintenance.md#s3t-table-236), [S3T-TABLE-237](../table-maintenance.md#s3t-table-237), [S3T-TABLE-238](../table-maintenance.md#s3t-table-238) · hypotheses: H-U-062 ·
    evidence: tablemaintenanceconfiguration/state-machine/job-status-soak

## Notes

H-U-062 confirmed for Not_Yet_Run, Disabled (never-run) and Successful. The Failed cell ({status,
lastRunTimestamp, failureMessage}) and 'Disabled after a run' were reached in the sibling probe
table/dependencies/kms-key-revoked-jobs (Disabled after a Failed run drops BOTH lastRunTimestamp and
failureMessage - refuting the 'keeps lastRunTimestamp' part).

Full original title: GetTableMaintenanceJobStatus entry members: Not_Yet_Run {status}, Disabled {status},
Successful {status, lastRunTimestamp}; absent members are omitted, never null
