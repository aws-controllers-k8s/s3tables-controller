<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-031: Explicit deny (bucket or IAM session policy) -> AccessDeniedException 403 on all 6 namespace ops; foreign-account ARN -> ForbiddenException
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-031"></a>**S3T-NAMESPACE-031** `error-code` · impact high · handled · verified 2026-10-13
  **Explicit deny (bucket or IAM session policy) -> AccessDeniedException 403 on all 6 namespace ops; foreign-account ARN -> ForbiddenException**
  With a table bucket policy Deny (Principal '*') on the six actions, and separately with an IAM session
  policy Deny while assuming a role that allows s3tables:*, GetNamespace, ListNamespaces, CreateNamespace,
  DeleteNamespace, GetTableBucket and ListTables ALL returned AccessDeniedException (HTTP 403,
  x-amzn-ErrorType 'AccessDeniedException') with the IAM-style message 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<role>/<session> is not authorized to perform: s3tables:<Action> on
  resource: <bucket ARN> with an explicit deny in a resource-based policy' (IAM path: '... with an explicit
  deny in an identity-based policy'). CreateNamespace, DeleteNamespace and ListTables do not declare
  AccessDeniedException in the model. ForbiddenException (HTTP 403, message 'Unauthorized') was returned only
  for a well-formed ARN of a bucket in another account (see the parent-ARN finding). A non-denied op on the
  same session (GetTable of a missing table) returned NotFoundException, proving the credentials worked.
  - ACK: terminal_codes, requeue · ops: GetNamespace, ListNamespaces, CreateNamespace, DeleteNamespace,
    GetTableBucket, ListTables
  - repro: PutTableBucketPolicy(Deny the six actions for '*') on an otherwise empty bucket; and AssumeRole
    with a session policy denying them; call each op
  - measurements: assume_role_success_after_create_role_s=9.1, n=1
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-017](../namespace.md#s3t-namespace-017), [S3T-NAMESPACE-029](../namespace.md#s3t-namespace-029), [S3T-NAMESPACE-030](../namespace.md#s3t-namespace-030), [S3T-NAMESPACE-003](../namespace.md#s3t-namespace-003) · hypotheses: H-N-041,
    H-N-042 · evidence: namespace/error-taxonomy/parent-arn-iam-deny-cross-region

## Notes

H-N-041 refuted (ForbiddenException was never returned for an IAM/resource-policy deny; AccessDeniedException
was returned on every op, including the three that do not model it). H-N-042 refuted (reads and writes return
the same code). New role: AssumeRole succeeded 9.1 s after CreateRole; the first s3tables call with the fresh
session token returned 200 in this run, but in a previous run (same probe, run 3de968) the first call returned
UnrecognizedClientException 'The security token included in the request is invalid.' for a token minted ~9 s
after CreateRole.

Full original title: Explicit deny (bucket policy or IAM session policy) -> AccessDeniedException 403 on all 6
namespace ops; ForbiddenException only for a foreign-account ARN
