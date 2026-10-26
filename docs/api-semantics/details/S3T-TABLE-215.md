<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-215: Table policy Deny PutTableMaintenanceConfiguration: Put -> AccessDeniedException/403 after 0.03 s; GetTableMaintenanceConfiguration -> 200
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-215"></a>**S3T-TABLE-215** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Table policy Deny PutTableMaintenanceConfiguration: Put -> AccessDeniedException/403 after 0.03 s; GetTableMaintenanceConfiguration -> 200**
  Deny s3tables:PutTableMaintenanceConfiguration for the caller's role via PutTablePolicy -> HTTP 200.
  PutTableMaintenanceConfiguration polled at 1/s: codes ['AccessDeniedException'], first denial after 0.03 s
  with AccessDeniedException. While denied: GetTableMaintenanceConfiguration -> HTTP 200,
  GetTableMaintenanceJobStatus -> HTTP 200. DeleteTablePolicy -> HTTP 204; Put allowed again after 0.04 s
  (codes ['OK']). Table policies govern these control-plane sub-resource operations per action.
  - ACK: terminal_codes, requeue · ops: PutTablePolicy, PutTableMaintenanceConfiguration,
    GetTableMaintenanceConfiguration, DeleteTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy Deny PutTableMaintenanceConfiguration for own role -> poll Put at 1/s -> Get ->
    DeleteTablePolicy -> poll Put
  - measurements: deny_enforced_after_s=0.03, deny_lifted_after_s=0.04, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-213](../table-maintenance.md#s3t-table-213), [S3T-TABLE-214](../table-policy.md#s3t-table-214), [S3T-TABLE-216](../table-policy.md#s3t-table-216), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183) · hypotheses: H-U-080 · evidence:
    table/dependencies/policy-deny-enforcement

## Notes

H-U-080 confirmed (denial code AccessDeniedException).

Full original title: Table policy Deny PutTableMaintenanceConfiguration: Put -> AccessDeniedException after
0.03s while GetTableMaintenanceConfiguration -> HTTP 200; lifted 0.04s after delete
