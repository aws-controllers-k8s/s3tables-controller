<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEPOLICY-004: Policy principal role deleted: Get flaps ARN/AROA for ~65 s; re-Put observed or desired -> 400...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablepolicy-004"></a>**S3T-TABLEPOLICY-004** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Policy principal role deleted: Get flaps ARN/AROA for ~65 s; re-Put observed or desired -> 400...**
  With Allow Principal=<role ARN> stored, iam DeleteRole: GetTablePolicy (1/s for 75 s) first shows the AROA
  unique id at +33 s, but the rendering flips between ARN and AROA 9 times before it is stable from +64 s
  (second deletion: first AROA +0 s, 4 flips, stable from +65 s); every read is HTTP 200. Controller moves:
  re-Put of the observed document -> BadRequestException/400 'Invalid principal in policy'; re-Put of the
  desired document (ARN) -> BadRequestException/400 'Invalid principal in policy'; DeleteTablePolicy -> HTTP
  204; Put(desired) with the role still absent -> BadRequestException/400 'Invalid principal in policy'. After
  CreateRole with the SAME name (same ARN, new RoleId): Put(desired) -> HTTP 200 after 7.53 s (7 rejections,
  ['BadRequestException/400']); the first GetTablePolicy after that 200 still rendered the NEW role's unique
  id instead of the ARN (the read renderer lags the Put validator). Deleting the new role makes the read show
  the NEW unique id and never the old one: the stored document is bound to the role's unique id, not its name.
  - ACK: requeue, references, terminal_codes, is_iam_policy · ops: PutTablePolicy, GetTablePolicy,
    DeleteTablePolicy · fields: resourcePolicy, Principal
  - repro: PutTablePolicy(Allow role) -> iam DeleteRole -> GetTablePolicy (poll) -> Put(observed) ->
    Put(desired) -> DeleteTablePolicy -> Put(desired) -> CreateRole(same name) -> Put(desired) poll ->
    DeleteRole -> Get
  - measurements: aroa_first_seen_s=[33, 0], aroa_stable_from_s=[64, 65], flips=[9, 4],
    form_sequences_1_per_s=[AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAAIAAIAAAAAAAAAAAAAAIIAIIIIIIIIIII,
    IAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAIIIIIIIIII],
    put_ok_after_recreate_s=7.53, n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-165](../table-policy.md#s3t-table-165), [S3T-TABLE-164](../table-policy.md#s3t-table-164), [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001),
    [S3T-TABLEPOLICY-002](../table-policy.md#s3t-tablepolicy-002), [S3T-TABLE-157](../table-policy.md#s3t-table-157), [S3T-TABLEPOLICY-010](../table-policy.md#s3t-tablepolicy-010) · evidence:
    table/creative/policy-principal-lifecycle

## Notes

extends [S3T-TABLE-165](../table-policy.md#s3t-table-165) (AROA rewrite) with the propagation flapping and the full recovery sequence a controller
would run. A controller diffing (n=2 deletions, 1 poll/s for 75 s) GetTablePolicy against spec sees the
principal change back and forth for ~20 s with no write of its own. Decision rule: observed-with-AROA !=
desired is a real, externally caused drift the controller cannot fix by re-Putting either document; it must
requeue until the referenced role exists again and must not 'reset' by DeleteTablePolicy (that only removes
protection). Single run.

Full original title: Policy principal role deleted: Get flaps ARN/AROA for ~65 s; re-Put observed or desired
-> 400; only a role with the same name heals it (200 after ~8 s)

Contradiction with [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLE-165](../table-policy.md#s3t-table-165): 013: re-Put of the AROA read-back document while
the role is gone -> HTTP 200 (bucket policy, +9 s after DeleteRole); 165 and 004: re-Put of the AROA read-back
-> BadRequestException 400 'Invalid principal in policy' (table policy, +30 s and later); level and
time-since-delete are confounded, 013 itself suggests a timing window before IAM forgets the unique id
Resolution: re-probe: creative - DeleteRole then re-Put the AROA read-back via PutTableBucketPolicy AND
PutTablePolicy at +5/+10/+20/+40/+90 s, n>=2 roles, to separate level from time-since-delete

Contradiction with [S3T-TABLEPOLICY-010](../table-policy.md#s3t-tablepolicy-010), [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLE-165](../table-policy.md#s3t-table-165): FU2-02 settled by follow-up
probe table/identity/fu2-06-missing-ns-messages-subread-iam-aroa: both readings are right under different
conditions - time since DeleteRole (IAM propagation of the deletion), not the policy level, decides. At
+5/+10/+20 s both the AROA document and the ARN document were accepted (HTTP 200) at bucket AND table level
while GetTableBucketPolicy/GetTablePolicy still rendered the ARN (2/2 roles, 12/12 Puts per level). At +40 s
the reads rendered the AROA and Resolution: keep all; [S3T-TABLEPOLICY-010](../table-policy.md#s3t-tablepolicy-010) is canonical for the settled rule
