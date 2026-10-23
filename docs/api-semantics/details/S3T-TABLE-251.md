<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-251: Table-bucket-policy Deny (Principal=role, Resource=<bucket>/table/*) on GetTableEncryption -> AccessDeniedException 403 like IAM deny
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-251"></a>**S3T-TABLE-251** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Table-bucket-policy Deny (Principal=role, Resource=<bucket>/table/*) on GetTableEncryption -> AccessDeniedException 403 like IAM deny**
  PutTableBucketPolicy with Deny {Principal: {AWS: role A ARN}, Action: [s3tables:GetTableEncryption,
  s3tables:GetTableStorageClass], Resource: <bucketARN>/table/*} -> HTTP 200 (the /table/* resource form is
  accepted). As role A (IAM allows both actions): GetTableEncryption -> AccessDeniedException 403 and
  GetTableStorageClass -> AccessDeniedException 403, message '... is not authorized to perform:
  s3tables:GetTableEncryption on resource: <table ARN> with an explicit deny in a resource-based policy';
  x-amzn-ErrorType 'AccessDeniedException:http://internal.amazon.com/coral/com.amazon.coral.service/', body
  {"Message": ...} without __type. GetTableMaintenanceConfiguration (not in the policy) -> HTTP 200 as the
  role; GetTableEncryption as the admin caller -> HTTP 200. The deny was effective on the first call 0.03 s
  after PutTableBucketPolicy returned.
  - ACK: terminal_codes · ops: PutTableBucketPolicy, GetTableEncryption, GetTableStorageClass
  - repro: PutTableBucketPolicy(Deny role on GetTableEncryption for <bucket>/table/*) -> GetTableEncryption as
    the role
  - measurements: bucket_policy_deny_effective_s=0.03, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-233](../table-maintenance.md#s3t-table-233), [S3T-TABLE-234](../table-maintenance.md#s3t-table-234) · hypotheses: H-T-066 · evidence:
    table/error-taxonomy/iam-deny-rename-tag

## Notes

Characterizes H-T-066 beyond the IAM case: identity-policy deny, resource-policy deny and implicit deny (see
the role-B finding) all surface as AccessDeniedException 403; only the message tail differs ('with an explicit
deny in an identity-based policy' / '... in a resource-based policy' / 'because no identity-based policy
allows the <action> action'). Single run.

Full original title: Table-bucket-policy Deny (Principal=role, Resource=<bucket>/table/*) on
GetTableEncryption -> AccessDeniedException 403, same code as an IAM deny
