<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-250: IAM explicit Deny on RenameTable/TagResource/UntagResource/ListTables/GetTable -> AccessDeniedException 403 on all 8 calls, never Forbidden
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-250"></a>**S3T-TABLE-250** `error-code` · impact high · handled · verified 2026-10-14
  **IAM explicit Deny on RenameTable/TagResource/UntagResource/ListTables/GetTable -> AccessDeniedException 403 on all 8 calls, never Forbidden**
  Role = AdministratorAccess + inline Deny on s3tables:RenameTable, TagResource, UntagResource, ListTables,
  GetTable (assumable 9.2 s after CreateRole; deny effective on the first call). As that role, all 8 calls
  returned AccessDeniedException HTTP 403: RenameTable, TagResource(table ARN), TagResource(bucket ARN),
  UntagResource(table), ListTables(bucket), ListTables(bucket+namespace), GetTable(bucketARN+namespace+name),
  GetTable(tableArn). Message: 'User: <assumed-role ARN> is not authorized to perform: s3tables:<Action> on
  resource: <table or bucket ARN> with an explicit deny in an identity-based policy'. The wire body is only
  {"Message": ...} with no __type; the code comes from the x-amzn-ErrorType header
  'AccessDeniedException:http://internal.amazon.com/coral/com.amazon.coral.service/' (framework namespace),
  whereas service-side errors such as BadRequestException carry
  'http://internal.amazon.com/coral/com.amazon.s3tablebuckets/'. Controls as the same role: GetTableBucket,
  ListTagsForResource(table), ListNamespaces -> HTTP 200. The denied RenameTable/TagResource left the table
  name and tags unchanged (admin read).
  - ACK: terminal_codes · ops: RenameTable, TagResource, UntagResource, ListTables, GetTable
  - repro: IAM role: AdministratorAccess + Deny on the five actions; AssumeRole; RenameTable / TagResource /
    UntagResource / ListTables / GetTable on an existing table
  - measurements: create_role_to_assumable_s=9.2, deny_effective_after_first_assume_s=0.06, n=1
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLE-109](../table-policy.md#s3t-table-109), [S3T-TABLEBUCKET-038](../tablebucket.md#s3t-tablebucket-038), [S3T-TABLE-026](../table-policy.md#s3t-table-026),
    [S3T-TABLEBUCKET-181](../tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-131](../table-policy.md#s3t-table-131), [S3T-TABLE-132](../table-policy.md#s3t-table-132), [S3T-TABLE-140](../table-policy.md#s3t-table-140), [S3T-TABLE-142](../table-policy.md#s3t-table-142), [S3T-TABLEBUCKET-109](../tablebucket-encryption-storage.md#s3t-tablebucket-109),
    [S3T-TABLEBUCKET-067](../tablebucket-encryption-storage.md#s3t-tablebucket-067), [S3T-TABLE-083](../table-policy.md#s3t-table-083), [S3T-TABLE-120](../table.md#s3t-table-120) · hypotheses: H-T-066 · evidence:
    table/error-taxonomy/iam-deny-rename-tag

## Notes

H-T-066 (IAM half) confirmed: RenameTable/TagResource/UntagResource/ListTables - which model only
ForbiddenException - return the unmodeled AccessDeniedException; GetTable (models both) also returns
AccessDeniedException. ForbiddenException was never observed in this probe. Single role, single run; deny
effective on the first call after AssumeRole.

Full original title: IAM explicit Deny on RenameTable/TagResource/UntagResource/ListTables/GetTable ->
AccessDeniedException 403 on all 8 calls, never ForbiddenException
