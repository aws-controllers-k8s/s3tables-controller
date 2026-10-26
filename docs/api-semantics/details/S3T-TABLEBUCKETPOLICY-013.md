<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETPOLICY-013: Principal role deleted then re-created (same name): read-back stays the old AROA id...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketpolicy-013"></a>**S3T-TABLEBUCKETPOLICY-013** `normalization` · impact medium · handled · verified 2026-10-14
  **Principal role deleted then re-created (same name): read-back stays the old AROA id...**
  Role v1 (path /ackq/) created; PutTableBucketPolicy naming its ARN every 1 s: 8 x BadRequestException 400
  'Invalid principal in policy' (0-7.5 s), HTTP 200 at +8.6 s; GetTableBucketPolicy echoed the ARN with path
  verbatim. DeleteRole -> 200; the read-back principal stayed the ARN for 8 s and flipped to the AROA unique
  id (== v1 RoleId) at +8.4 s. While the role was gone (+9 s): re-Put of the read-back AROA document -> HTTP
  200 (contrast [S3T-TABLE-165](../table-policy.md#s3t-table-165): 400 for a table policy at +30 s); re-Put of the ARN document -> 400 'Invalid
  principal in policy'; read-back still the AROA. CreateRole with the SAME name (same ARN, new RoleId): the
  read-back stayed the OLD AROA for the 20 s watched (no re-binding by name). Re-Put of the ARN document at
  +20.7 s -> HTTP 200 and the read-back shows the ARN (now bound to the new role).
  - ACK: is_iam_policy, requeue, references · ops: PutTableBucketPolicy, GetTableBucketPolicy · fields:
    resourcePolicy
  - repro: CreateRole -> PutTableBucketPolicy(principal ARN) -> DeleteRole -> Get (AROA) -> CreateRole same
    name -> Get x10 -> PutTableBucketPolicy(ARN) -> Get
  - measurements: accept_after_createrole_s=[8.6, 20.7], aroa_flip_after_delete_s=8.4, n=2
  - handling: handled via `generator.yaml:234-238; pkg/resource/table_bucket/delta.go:69-75`
  - related: [S3T-TABLE-165](../table-policy.md#s3t-table-165), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKET-063](../tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](../tablebucket-policy.md#s3t-tablebucket-111),
    [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLEBUCKETPOLICY-004](../tablebucket-policy.md#s3t-tablebucketpolicy-004), [S3T-TABLEBUCKETPOLICY-017](../tablebucket-policy.md#s3t-tablebucketpolicy-017), [S3T-TABLEPOLICY-004](../table-policy.md#s3t-tablepolicy-004),
    [S3T-TABLEPOLICY-002](../table-policy.md#s3t-tablepolicy-002), [S3T-TABLE-164](../table-policy.md#s3t-table-164), [S3T-TABLE-157](../table-policy.md#s3t-table-157), [S3T-TABLEPOLICY-010](../table-policy.md#s3t-tablepolicy-010) · evidence:
    tablebucketpolicy/creative/recreated-role-principal

## Notes

Extends [S3T-TABLE-165](../table-policy.md#s3t-table-165) (deleted role -> AROA in the read-back) with the re-creation case a controller meets
when an IAM Role CR is recreated: the stored policy is bound to the OLD unique id and does not follow the
name; the diff spec(ARN) != read-back(AROA) is therefore a correct signal and the controller's re-Put of the
spec document is the fix once the new role has propagated (the first Put after CreateRole fails 400 for ~8-10
s, [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001); here n=2: 8.6 s and <=20.7 s). Qualifies 165's 'read-back cannot be re-Put': for
a bucket policy ~9 s after the delete the AROA document WAS accepted (200) - possibly a timing window before
IAM forgets the id. Single run.

Full original title: Principal role deleted then re-created (same name): read-back stays the old AROA id;
re-Put of the ARN document -> HTTP 200 and read-back shows the ARN again

Contradiction with [S3T-TABLE-165](../table-policy.md#s3t-table-165), [S3T-TABLEPOLICY-004](../table-policy.md#s3t-tablepolicy-004): 013: re-Put of the AROA read-back document while the
role is gone -> HTTP 200 (bucket policy, +9 s after DeleteRole); 165 and 004: re-Put of the AROA read-back ->
BadRequestException 400 'Invalid principal in policy' (table policy, +30 s and later); level and
time-since-delete are confounded, 013 itself suggests a timing window before IAM forgets the unique id
Resolution: re-probe: creative - DeleteRole then re-Put the AROA read-back via PutTableBucketPolicy AND
PutTablePolicy at +5/+10/+20/+40/+90 s, n>=2 roles, to separate level from time-since-delete

Contradiction with [S3T-TABLEPOLICY-010](../table-policy.md#s3t-tablepolicy-010), [S3T-TABLE-165](../table-policy.md#s3t-table-165), [S3T-TABLEPOLICY-004](../table-policy.md#s3t-tablepolicy-004): FU2-02 settled by follow-up probe
table/identity/fu2-06-missing-ns-messages-subread-iam-aroa: both readings are right under different conditions -
time since DeleteRole (IAM propagation of the deletion), not the policy level, decides. At +5/+10/+20 s both
the AROA document and the ARN document were accepted (HTTP 200) at bucket AND table level while
GetTableBucketPolicy/GetTablePolicy still rendered the ARN (2/2 roles, 12/12 Puts per level). At +40 s the
reads rendered the AROA and Resolution: keep all; [S3T-TABLEPOLICY-010](../table-policy.md#s3t-tablepolicy-010) is canonical for the settled rule
