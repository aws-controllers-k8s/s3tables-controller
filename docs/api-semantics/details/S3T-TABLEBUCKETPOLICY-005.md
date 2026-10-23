<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETPOLICY-005: Bucket-policy Deny on a role takes effect on the first poll (<= 0.07 s) and stops on the first poll after DeleteTableBucketPolicy...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketpolicy-005"></a>**S3T-TABLEBUCKETPOLICY-005** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket-policy Deny on a role takes effect on the first poll (<= 0.07 s) and stops on the first poll after DeleteTableBucketPolicy...** (hypothesis refuted; behavior confirmed)
  Owner PutTableBucketPolicy {Deny <role ARN> s3tables:GetTableBucketMaintenanceConfiguration on the bucket}
  -> HTTP 200; the role's GetTableBucketMaintenanceConfiguration (HTTP 200 just before) failed on the very
  first poll after the Put (+0.07/+0.01/+0.01 s) with AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<account>:assumed-role/... is not authorized to perform:
  s3tables:GetTableBucketMaintenanceConfiguration on resource: <bucket ARN> with an explicit deny in a
  resource-based policy'. Owner DeleteTableBucketPolicy -> HTTP 204 and owner GetTableBucketPolicy ->
  NotFoundException 404 immediately; the role's Get returned HTTP 200 on its first poll after the delete
  (+0.12/+0.13/+0.12 s), 0 denied polls, in all 3 repetitions. Note the wire code for a resource-policy deny
  is AccessDeniedException, not the ForbiddenException seen for IAM-level/foreign-account denials.
  - ACK: none · ops: PutTableBucketPolicy, DeleteTableBucketPolicy, GetTableBucketPolicy,
    GetTableBucketMaintenanceConfiguration · fields: resourcePolicy
  - repro: role R denied by bucket policy; R polls GetTableBucketMaintenanceConfiguration at 1 s; owner
    deletes the policy; R polls until 200 (x3)
  - measurements: deny_effective_after_s=[0.07, 0.01, 0.01], allowed_again_after_s=[0.12, 0.13, 0.12],
    denied_polls_after_delete=[0, 0, 0], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-129](../tablebucket-policy.md#s3t-tablebucket-129), [S3T-TABLEBUCKET-091](../tablebucket-policy.md#s3t-tablebucket-091),
    [S3T-TABLEBUCKETPOLICY-006](../tablebucket-policy.md#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-007](../tablebucket-policy.md#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKETPOLICY-016](../tablebucket-policy.md#s3t-tablebucketpolicy-016), [S3T-TABLEPOLICY-009](../table-policy.md#s3t-tablepolicy-009),
    [S3T-TABLEBUCKET-108](../tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLEBUCKETPOLICY-003](../tablebucket-policy.md#s3t-tablebucketpolicy-003), [S3T-TABLEBUCKET-065](../tablebucket-policy.md#s3t-tablebucket-065), [S3T-TABLE-158](../table-policy.md#s3t-table-158) · hypotheses: H-S-140 ·
    evidence: tablebucket/dependencies/iam-cloudwatch-deny-metrics

## Notes

H-S-140 (contrarian: the Deny keeps being enforced for up to ~30 s after the Delete's 204) refuted at 1 s
polling resolution in 3 repetitions: policy changes were enforced read-your-writes in both directions.
GetTableBucketMaintenanceConfiguration is listed with ForbiddenException but returned AccessDeniedException
for the resource-policy deny.

Full original title: Bucket-policy Deny on a role takes effect on the first poll (<= 0.07 s) and stops on the
first poll after DeleteTableBucketPolicy (<= 0.13 s), 3 reps

Contradiction with [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046): 005's behavior text calls ForbiddenException the code 'seen for
IAM-level/foreign-account denials'; 046 (seven ops), 045, 091 and 129 show every IAM-level denial
(identity-policy explicit Deny and implicit deny) as AccessDeniedException 403 - ForbiddenException belongs
only to foreign-account ARNs (038) and downstream KMS/PutTableData checks (181, [S3T-TABLE-131](../table-policy.md#s3t-table-131)) Resolution:
keep both; [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046) is canonical for the IAM-level code; 005's remark conflates IAM-level with
foreign-account/KMS ForbiddenException
