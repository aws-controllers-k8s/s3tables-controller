<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-213: Table policy Deny s3tables:DeleteTable for the caller: own DeleteTable -> AccessDeniedException/403 after 0.03s...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-213"></a>**S3T-TABLE-213** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Table policy Deny s3tables:DeleteTable for the caller: own DeleteTable -> AccessDeniedException/403 after 0.03s...**
  PutTablePolicy (Deny DeleteTable + GetTableMaintenanceConfiguration, Principal = caller's role ARN,
  DateLessThan guard) -> HTTP 200. The read canary GetTableMaintenanceConfiguration polled at 1/s went
  ['AccessDeniedException'] and was first denied after 0.03 s (1 polls). DeleteTable(versionToken) while
  denied -> AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not
  authorized to perform: s3tables:DeleteTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-7b3e8f-deny/tab'; GetTable afterwards -> HTTP 200.
  DeleteTablePolicy by the same (denied-for-delete) caller -> HTTP 204. Canary restored to 200 after 0.01 s
  (codes ['OK']). DeleteTable then -> HTTP 204 (after 1 attempt(s)).
  - ACK: terminal_codes, pre-delete-cleanup, requeue · ops: PutTablePolicy, DeleteTable, DeleteTablePolicy,
    GetTableMaintenanceConfiguration · fields: resourcePolicy
  - repro: PutTablePolicy Deny DeleteTable for own role -> poll a denied read at 1/s -> DeleteTable ->
    DeleteTablePolicy -> poll -> DeleteTable
  - measurements: deny_enforced_after_s=0.03, deny_lifted_after_s=0.01, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-214](../table-policy.md#s3t-table-214), [S3T-TABLE-215](../table-maintenance.md#s3t-table-215), [S3T-TABLE-216](../table-policy.md#s3t-table-216), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183) · hypotheses: H-U-016 · evidence:
    table/dependencies/policy-deny-enforcement

## Notes

H-U-016 confirmed: denial code AccessDeniedException (HTTP 403), enforced within 0.03 s, DeleteTablePolicy
still allowed=True, delete after removal=True. Exact message: 'User:
arn:aws:sts::<account>:assumed-role/<role>/<session> is not authorized to perform: s3tables:DeleteTable on
resource: <table ARN> with an explicit deny in a resource-based policy'. A role ARN principal matches the
role's assumed-role sessions.

Full original title: Table policy Deny s3tables:DeleteTable for the caller: own DeleteTable ->
AccessDeniedException/403 after 0.03s; DeleteTablePolicy HTTP 204; delete then HTTP 204
