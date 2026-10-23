<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-165: Deleting the IAM role named in a stored table policy: GetTablePolicy principal rewritten to the AROA unique id...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-165"></a>**S3T-TABLE-165** `normalization` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Deleting the IAM role named in a stored table policy: GetTablePolicy principal rewritten to the AROA unique id...**
  CreateRole then PutTablePolicy(Principal=<new role ARN>): accepted after 10.45 s (6 attempts; first
  rejection ('BadRequestException', 'Invalid principal in policy')). Principal read back before role deletion:
  {"AWS": "arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL>"}. After DeleteRole, GetTablePolicy at 0/5/15/30 s returned
  principals [{"AWS": "arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL>"}, {"AWS":
  "arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL>"}, {"AWS": "<AROA>"}, {"AWS": "<AROA>"}]. Re-Put of the read-back
  document -> BadRequestException/400 'Invalid principal in policy'; re-Put of the original ARN document (role
  gone) -> BadRequestException/400 'Invalid principal in policy'.
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy, GetTablePolicy · fields:
    resourcePolicy
  - repro: iam CreateRole R -> PutTablePolicy(Principal R ARN) -> iam DeleteRole R -> GetTablePolicy (x4 over
    30 s) -> PutTablePolicy(read-back)
  - measurements: new_role_accepted_after_s=10.45, put_attempts=6, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLEPOLICY-004](../table-policy.md#s3t-tablepolicy-004), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLEPOLICY-002](../table-policy.md#s3t-tablepolicy-002),
    [S3T-TABLE-164](../table-policy.md#s3t-table-164), [S3T-TABLE-157](../table-policy.md#s3t-table-157), [S3T-TABLEPOLICY-010](../table-policy.md#s3t-tablepolicy-010) · hypotheses: H-U-010 · evidence:
    table/error-taxonomy/policy-validation

## Notes

H-U-010 second half (AROA rewrite) confirmed within the 30 s observed. Single run.

Full original title: Deleting the IAM role named in a stored table policy: GetTablePolicy principal rewritten
to the AROA unique id (new role accepted after 10.45s)

Contradiction with [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLEPOLICY-004](../table-policy.md#s3t-tablepolicy-004): 013: re-Put of the AROA read-back document
while the role is gone -> HTTP 200 (bucket policy, +9 s after DeleteRole); 165 and 004: re-Put of the AROA
read-back -> BadRequestException 400 'Invalid principal in policy' (table policy, +30 s and later); level and
time-since-delete are confounded, 013 itself suggests a timing window before IAM forgets the unique id
Resolution: re-probe: creative - DeleteRole then re-Put the AROA read-back via PutTableBucketPolicy AND
PutTablePolicy at +5/+10/+20/+40/+90 s, n>=2 roles, to separate level from time-since-delete

Contradiction with [S3T-TABLEPOLICY-010](../table-policy.md#s3t-tablepolicy-010), [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLEPOLICY-004](../table-policy.md#s3t-tablepolicy-004): FU2-02 settled by
follow-up probe table/identity/fu2-06-missing-ns-messages-subread-iam-aroa: both readings are right under
different conditions - time since DeleteRole (IAM propagation of the deletion), not the policy level, decides.
At +5/+10/+20 s both the AROA document and the ARN document were accepted (HTTP 200) at bucket AND table level
while GetTableBucketPolicy/GetTablePolicy still rendered the ARN (2/2 roles, 12/12 Puts per level). At +40 s
the reads rendered the AROA and Resolution: keep all; [S3T-TABLEPOLICY-010](../table-policy.md#s3t-tablepolicy-010) is canonical for the settled rule
