<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-167: PutTablePolicy does not validate Resource or s3tables action names...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-167"></a>**S3T-TABLE-167** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTablePolicy does not validate Resource or s3tables action names...**
  On table t1: Resource = t2's ARN, the bucket ARN, <bucketARN>/table/<name>, <bucketARN>/table/<ns>/<name>,
  "*", <t1 ARN>/*, <bucketARN>/table/*, [t1, t2] and a statement with NO Resource key were all accepted (HTTP 200)
  and read back as sent. Action "s3tables:NoSuchAction" and "s3tables:*" are accepted; Action "s3:GetObject"
  -> BadRequestException 400 'Invalid Action'. Only the service prefix is checked.
  - ACK: none · ops: PutTablePolicy, GetTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy on t1 with each Resource/Action variant; GetTablePolicy after each
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-163](../table-policy.md#s3t-table-163), [S3T-TABLE-157](../table-policy.md#s3t-table-157), [S3T-TABLE-162](../table-policy.md#s3t-table-162), [S3T-TABLE-164](../table-policy.md#s3t-table-164), [S3T-TABLE-166](../table-policy.md#s3t-table-166), [S3T-TABLEPOLICY-001](../table-policy.md#s3t-tablepolicy-001),
    [S3T-TABLEBUCKET-088](../tablebucket-policy.md#s3t-tablebucket-088), [S3T-TABLEPOLICY-009](../table-policy.md#s3t-tablepolicy-009), [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLE-004](../table-policy.md#s3t-table-004) · hypotheses: H-U-009,
    H-U-068, H-U-008 · evidence: table/error-taxonomy/policy-validation

## Notes

Refutes H-U-009's rejection half (no Resource-vs-table check; '*' accepted as hypothesized) and H-U-068's
Resource-suffix half (<arn>/* accepted). A controller templating Resource from the spec name instead of
status.tableARN will NOT get an error; the policy is simply ineffective.

Full original title: PutTablePolicy does not validate Resource or s3tables action names: other table's ARN,
bucket ARN, name-based ARN, "*", <arn>/*, no Resource all 200
