<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEPOLICY-010: Deleted principal role: AROA and ARN documents both re-Put 200 for ~20-40 s after DeleteRole (IAM lag), then 400 at bucket and table level
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablepolicy-010"></a>**S3T-TABLEPOLICY-010** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Deleted principal role: AROA and ARN documents both re-Put 200 for ~20-40 s after DeleteRole (IAM lag), then 400 at bucket and table level**
  Allow policies naming a fresh role's ARN stored on the bucket (accepted after [7.69, 7.59] s) and on the
  table (accepted after [1.23, 0.14] s); read back as [{'bucket': 'ARN', 'table': 'ARN'}, {'bucket': 'ARN',
  'table': 'ARN'}]. iam DeleteRole, then at each offset: PutTableBucketPolicy(AROA doc) /
  PutTableBucketPolicy(ARN doc) / PutTablePolicy(AROA doc) / PutTablePolicy(ARN doc). Bucket level: +5s AROA
  HTTP 200|HTTP 200 / ARN HTTP 200|HTTP 200; +10s AROA HTTP 200|HTTP 200 / ARN HTTP 200|HTTP 200; +20s AROA
  HTTP 200|HTTP 200 / ARN HTTP 200|HTTP 200; +40s AROA HTTP 200|BadRequestException (HTTP 400) / ARN
  BadRequestException (HTTP 400)|BadRequestException (HTTP 400); +90s AROA BadRequestException (HTTP
  400)|BadRequestException (HTTP 400) / ARN BadRequestException (HTTP 400)|BadRequestException (HTTP 400).
  Table level: +5s AROA HTTP 200|HTTP 200 / ARN HTTP 200|HTTP 200; +10s AROA HTTP 200|HTTP 200 / ARN HTTP
  200|HTTP 200; +20s AROA HTTP 200|HTTP 200 / ARN HTTP 200|HTTP 200; +40s AROA BadRequestException (HTTP
  400)|BadRequestException (HTTP 400) / ARN BadRequestException (HTTP 400)|BadRequestException (HTTP 400);
  +90s AROA BadRequestException (HTTP 400)|BadRequestException (HTTP 400) / ARN BadRequestException (HTTP
  400)|BadRequestException (HTTP 400). (per offset: results of role 1 | role 2.)
  GetTableBucketPolicy/GetTablePolicy rendering per offset, role 1: [(5, 'ARN', 'ARN'), (10, 'ARN', 'ARN'),
  (20, 'ARN', 'ARN'), (40, 'AROA', 'AROA'), (90, 'AROA', 'AROA')]; role 2: [(5, 'ARN', 'ARN'), (10, 'ARN',
  'ARN'), (20, 'ARN', 'ARN'), (40, 'AROA', 'AROA'), (90, 'AROA', 'AROA')].
  - ACK: is_iam_policy, requeue, references · ops: PutTableBucketPolicy, PutTablePolicy, GetTableBucketPolicy,
    GetTablePolicy · fields: resourcePolicy, Principal
  - repro: CreateRole -> PutTableBucketPolicy+PutTablePolicy(Principal=role ARN) until 200 -> DeleteRole -> at
    +5/10/20/40/90 s Put the AROA (RoleId) document and the ARN document at both levels
  - measurements: offsets_s=[5, 10, 20, 40, 90], roles=2, aroa_accepted_offsets_bucket=[5, 10, 20, 40],
    aroa_accepted_offsets_table=[5, 10, 20], arn_accepted_offsets_both_levels=[5, 10, 20],
    all_rejected_from_s=90, arn_doc_accepted_after_create_s=[7.69, 7.59, 1.23, 0.14], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLE-165](../table-policy.md#s3t-table-165), [S3T-TABLEPOLICY-004](../table-policy.md#s3t-tablepolicy-004) · evidence:
    table/identity/fu2-06-missing-ns-messages-subread-iam-aroa

## Notes

FU2-02: both readings are right under different conditions - time since DeleteRole (IAM propagation of the
deletion), not the policy level, decides. At +5/+10/+20 s both the AROA document and the ARN document were
accepted (HTTP 200) at bucket AND table level while GetTableBucketPolicy/GetTablePolicy still rendered the ARN
(2/2 roles, 12/12 Puts per level). At +40 s the reads rendered the AROA and 7 of 8 Puts were
BadRequestException 400 'Invalid principal in policy' (the bucket-level AROA Put of role 1 was still 200: the
boundary, not a level rule). At +90 s all 8 Puts were 400. 013's 200 at +9 s fell inside the window; 165 (+30
s) and 004 were outside. A controller re-Putting its spec right after an IAM Role is deleted can therefore get
a transient 200 that the next reconcile cannot reproduce. The AROA document was built from CreateRole's RoleId
(same string the read renders) so it could be sent before the read flipped.

Contradiction with [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLE-165](../table-policy.md#s3t-table-165), [S3T-TABLEPOLICY-004](../table-policy.md#s3t-tablepolicy-004): FU2-02 settled by follow-up
probe table/identity/fu2-06-missing-ns-messages-subread-iam-aroa: both readings are right under different
conditions - time since DeleteRole (IAM propagation of the deletion), not the policy level, decides. At
+5/+10/+20 s both the AROA document and the ARN document were accepted (HTTP 200) at bucket AND table level
while GetTableBucketPolicy/GetTablePolicy still rendered the ARN (2/2 roles, 12/12 Puts per level). At +40 s
the reads rendered the AROA and Resolution: keep all; [S3T-TABLEPOLICY-010](../table-policy.md#s3t-tablepolicy-010) is canonical for the settled rule
