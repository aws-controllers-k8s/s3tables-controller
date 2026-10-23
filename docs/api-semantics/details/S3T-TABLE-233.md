<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-233: Role lacking the IAM action: Get/Put/DeleteTablePolicy, GetTableMaintenanceJobStatus, DeleteTable -> AccessDeniedException 403, IAM message
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-233"></a>**S3T-TABLE-233** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **Role lacking the IAM action: Get/Put/DeleteTablePolicy, GetTableMaintenanceJobStatus, DeleteTable -> AccessDeniedException 403, IAM message**
  Assumed role with an inline Allow for only ['s3tables:GetTable',
  's3tables:GetTableMaintenanceConfiguration', 's3tables:PutTableMaintenanceConfiguration'] (effective 0.0 s
  after put_role_policy): GetTable -> HTTP 200; GetTablePolicy -> AccessDeniedException/403; PutTablePolicy ->
  AccessDeniedException/403; DeleteTablePolicy -> AccessDeniedException/403; GetTableMaintenanceConfiguration
  -> HTTP 200; PutTableMaintenanceConfiguration(binpack,256) -> HTTP 204;
  PutTableMaintenanceConfiguration(128) -> HTTP 204; PutTableMaintenanceConfiguration(sort) ->
  AccessDeniedException/403; PutTableMaintenanceConfiguration(z-order) -> AccessDeniedException/403;
  GetTableMaintenanceJobStatus -> AccessDeniedException/403; DeleteTable -> AccessDeniedException/403. Message
  form: 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:GetTablePolicy on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-984'.
  - ACK: terminal_codes · ops: GetTablePolicy, PutTablePolicy, DeleteTablePolicy,
    GetTableMaintenanceJobStatus, PutTableMaintenanceConfiguration, DeleteTable
  - repro: iam CreateRole + inline policy (GetTable, Get/PutTableMaintenanceConfiguration) -> sts AssumeRole
    -> each op
  - measurements: assume_role_after_s=9.1, policy_effective_after_s=0.0, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-232](../table-maintenance.md#s3t-table-232), [S3T-TABLE-234](../table-maintenance.md#s3t-table-234), [S3T-TABLE-230](../table-maintenance.md#s3t-table-230), [S3T-TABLE-251](../table-policy.md#s3t-table-251) · hypotheses: H-U-017 · evidence:
    table/error-taxonomy/iam-deny-subresource-codes

## Notes

H-U-017 confirmed: identity-policy denials on GetTablePolicy, PutTablePolicy, DeleteTablePolicy,
GetTableMaintenanceJobStatus and DeleteTable all surface as AccessDeniedException (HTTP 403) with the standard
IAM message, although these operations declare only ForbiddenException; ForbiddenException/403 'Unauthorized'
is what a foreign-account tableBucketARN returns (identifier probe). AssumeRole worked 9 s after CreateRole;
the inline policy was effective at the first call.

Full original title: IAM-denied policy/maintenance ops (assumed role without the permissions) ->
AccessDeniedException; allowed ops ['GetTable', 'GetTableMaintenanceConfiguration',
'PutTableMaintenanceConfiguration(binpack,256)', 'PutTableMaintenanceConfiguration(128)']

Re-verification CORRECTED (2026-10-15, independent probe table/creative/reverify-set-c): 3/3 runs reproduce
every clause but one: GetTable 200; Get/Put/DeleteTablePolicy, GetTableMaintenanceJobStatus, DeleteTable ->
AccessDeniedException 403 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
s3tables:<Op> on resource: <full table ARN bucket/<name>/table/<uuid>> because no identity-based policy allows
...'; GetTableMaintenanceConfiguration 200; Put binpack/256 and 128 -> 204; role assumable and policy
effective at the first call. DIFFERS: Put strategy sort / z-order on a table created without metadata ->
BadRequestException (HTTP 400) 'The table metadata location is not specified.' (role and admin alike, 6/6),
not AccessDeniedException; the 403 appears only on a table WITH a metadata location, with the service message
'Insufficient access to update the table maintenance configuration' (4/4), and the missing action is
s3tables:GetTableData (adding it yields the admin's 400 'requires sort order in the table metadata'; adding
GetTableMetadataLocation does not). Recorded as [S3T-TABLEMAINTENANCECONFIGURATION-013](../table-maintenance.md#s3t-tablemaintenanceconfiguration-013).
