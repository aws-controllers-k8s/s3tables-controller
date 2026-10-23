<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-155: PutTablePolicy/DeleteTablePolicy do NOT change GetTable versionToken (modifiedAt unchanged) across 4 snapshots
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-155"></a>**S3T-TABLE-155** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTablePolicy/DeleteTablePolicy do NOT change GetTable versionToken (modifiedAt unchanged) across 4 snapshots**
  GetTable snapshots before Put, after 2 Puts, after 4 Puts, after Delete: versionToken distinct values=1,
  modifiedAt changed=False, modifiedBy changed=False. Policy writes do not bump the table's version token.
  - ACK: none · ops: PutTablePolicy, DeleteTablePolicy, GetTable · fields: versionToken, modifiedAt,
    modifiedBy
  - repro: GetTable; PutTablePolicy x4; GetTable; DeleteTablePolicy; GetTable; diff
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-110](../table-maintenance.md#s3t-table-110), [S3T-TABLE-206](../table-maintenance.md#s3t-table-206) · hypotheses: H-U-015 · evidence:
    table/sub-resources/policy-lifecycle

## Notes

H-U-015 refuted (versionToken stable).
