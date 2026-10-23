<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-010: Fresh IAM role as policy Principal: Put*Policy -> BadRequestException 'Invalid principal in policy' 7.2/7.9/7.8 s after CreateRole, then OK
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-010"></a>**S3T-SERVICE-010** `eventual-consistency` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-14
  **Fresh IAM role as policy Principal: Put*Policy -> BadRequestException 'Invalid principal in policy' 7.2/7.9/7.8 s after CreateRole, then OK**
  CreateRole, then PutTableBucketPolicy (roles 0,2) / PutTablePolicy (role 1) with Principal.AWS=<new role
  ARN> every 0.5 s: first HTTP 200 at [7.22, 7.93, 7.79] s ([13, 14, 14] x BadRequestException before it,
  codes seen ['BadRequestException'], messages ['Invalid principal in policy']). sts:AssumeRole on the same
  new role polled in the same loop: first success at [8.1, 8.75, 8.6] s; before that ['User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: sts:AssumeRole'].
  - ACK: requeue, references · ops: PutTableBucketPolicy, PutTablePolicy · fields: resourcePolicy
  - repro: iam CreateRole(trust=account root) -> immediately loop PutTableBucketPolicy/PutTablePolicy with the
    role as Principal at 0.5 s until 200
  - measurements: put_policy_first_ok_s=[7.22, 7.93, 7.79], assume_role_first_ok_s=[8.1, 8.75, 8.6], n=3
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLE-164](../table-policy.md#s3t-table-164), [S3T-SERVICE-025](../service.md#s3t-service-025), [S3T-SERVICE-012](../service.md#s3t-service-012) · evidence:
    service/creative/iam-authz-variants

## Notes

Creative round, seed error-messages: the BadRequestException 'Invalid principal in policy' class is AMBIGUOUS -
PERMANENT-SPEC for a typo/deleted principal but transient (WAIT-FOR-STATE on IAM propagation) for a principal
created seconds earlier, e.g. an ACK Role and a TableBucketPolicy applied in the same reconcile. Extends
[S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001) (n=1, 7.2 s) with n=3 and the AssumeRole window.

Full original title: Fresh IAM role as policy Principal: Put*Policy -> BadRequestException 'Invalid principal
in policy' for 7.2/7.9/7.8 s after CreateRole, then 200 (n=3)

Suspected controller bug confirmed by evidence: not flagged; evidence-driven. BadRequestException 'Invalid
principal in policy' is transient for 7.2-7.9 s after iam:CreateRole (n=3; 13-14 consecutive 400s then 200)
and only afterwards a permanent spec error, so terminal_codes: [BadRequestException] sets ACK.Terminal on a
TableBucket/Table whose policy names a role created in the same pass (the common Role + policy manifest);
exempt this message with a ~10 s requeue or re-evaluate Terminal on the next reconcile.
