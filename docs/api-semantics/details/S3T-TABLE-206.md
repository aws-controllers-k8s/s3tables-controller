<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-206: PutTableMaintenanceConfiguration does NOT change GetTable versionToken but DOES advance modifiedAt (modifiedBy stays absent)
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-206"></a>**S3T-TABLE-206** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableMaintenanceConfiguration does NOT change GetTable versionToken but DOES advance modifiedAt (modifiedBy stays absent)**
  GetTable before the first maintenance Put and after 21 Puts (both types, enabled/disabled, partial values):
  versionToken identical (2c50...), modifiedAt advanced from the create time to the time of the last
  PutTableMaintenanceConfiguration (evidence: 23:34:57 -> 23:38:48 UTC), modifiedBy absent in both.
  PutTablePolicy/DeleteTablePolicy in the policy-lifecycle probe changed neither.
  - ACK: none · ops: PutTableMaintenanceConfiguration, GetTable · fields: versionToken, modifiedAt
  - repro: GetTable; many PutTableMaintenanceConfiguration; GetTable; diff
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-110](../table-maintenance.md#s3t-table-110), [S3T-TABLE-155](../table-maintenance.md#s3t-table-155), [S3T-TABLEMAINTENANCECONFIGURATION-012](../table-maintenance.md#s3t-tablemaintenanceconfiguration-012),
    [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001), [S3T-TABLE-232](../table-maintenance.md#s3t-table-232), [S3T-TABLE-205](../table-maintenance.md#s3t-table-205) · hypotheses: H-U-047 · evidence:
    table/mutation-matrix/maintenance-config

## Notes

H-U-047 refuted for the versionToken (no ConflictException risk); qualified: modifiedAt moves, so a controller
comparing modifiedAt would see maintenance writes as table modifications.
