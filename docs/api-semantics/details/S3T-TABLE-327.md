<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-327: Role allowed only s3tables:GetTable: every table sub-resource Get -> AccessDeniedException 403 naming its own action (30/30, 3 fresh roles)
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-327"></a>**S3T-TABLE-327** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Role allowed only s3tables:GetTable: every table sub-resource Get -> AccessDeniedException 403 naming its own action (30/30, 3 fresh roles)**
  Three fresh roles with one inline policy Allow s3tables:GetTable on * (trust: this account), first calls
  30.1 s after PutRolePolicy, second pass 238.2 s after; per operation {outcomes, action named in the denial,
  cells}: {"get_table": {"denied_actions": ["None"], "n": 6, "outcomes": ["HTTP 200"]}, "get_table_bucket":
  {"denied_actions": ["s3tables:GetTableBucket"], "n": 6, "outcomes": ["AccessDeniedException (HTTP 403)"]},
  "get_table_encryption": {"denied_actions": ["s3tables:GetTableEncryption"], "n": 6, "outcomes":
  ["AccessDeniedException (HTTP 403)"]}, "get_table_maintenance_configuration": {"denied_actions":
  ["s3tables:GetTableMaintenanceConfiguration"], "n": 6, "outcomes": ["AccessDeniedException (HTTP 403)"]},
  "get_table_metadata_location": {"denied_actions": ["s3tables:GetTableMetadataLocation"], "n": 6, "outcomes":
  ["AccessDeniedException (HTTP 403)"]}, "get_table_policy": {"denied_actions": ["s3tables:GetTablePolicy"],
  "n": 6, "outcomes": ["AccessDeniedException (HTTP 403)"]}, "get_table_storage_class": {"denied_actions":
  ["s3tables:GetTableStorageClass"], "n": 6, "outcomes": ["AccessDeniedException (HTTP 403)"]}}. Missing table
  name: {"get_table": ["NotFoundException (HTTP 404) action=None"], "get_table_encryption":
  ["AccessDeniedException (HTTP 403) action=s3tables:GetTableEncryption"]}. One fresh role allowed only
  s3tables:GetTableBucket: {"get_table": {"denied_actions": ["s3tables:GetTable"], "n": 2, "outcomes":
  ["AccessDeniedException (HTTP 403)"]}, "get_table_bucket": {"denied_actions": ["None"], "n": 2, "outcomes":
  ["HTTP 200"]}, "get_table_bucket_encryption": {"denied_actions": ["s3tables:GetTableBucketEncryption"], "n":
  2, "outcomes": ["AccessDeniedException (HTTP 403)"]}, "get_table_bucket_maintenance_configuration":
  {"denied_actions": ["s3tables:GetTableBucketMaintenanceConfiguration"], "n": 2, "outcomes":
  ["AccessDeniedException (HTTP 403)"]}, "get_table_bucket_policy": {"denied_actions":
  ["s3tables:GetTableBucketPolicy"], "n": 2, "outcomes": ["AccessDeniedException (HTTP 403)"]},
  "get_table_bucket_storage_class": {"denied_actions": ["s3tables:GetTableBucketStorageClass"], "n": 2,
  "outcomes": ["AccessDeniedException (HTTP 403)"]}}.
  - ACK: terminal_codes, hook:sdk_read_one_post_set_output · ops: GetTable, GetTableEncryption,
    GetTableStorageClass, GetTableMetadataLocation, GetTablePolicy, GetTableMaintenanceConfiguration,
    GetTableBucket, GetTableBucketEncryption, GetTableBucketStorageClass, GetTableBucketPolicy,
    GetTableBucketMaintenanceConfiguration
  - repro: CreateRole + PutRolePolicy(Allow s3tables:GetTable only) x3 -> wait >= 30 s -> AssumeRole -> each
    table Get on an existing table and on a missing name; repeat after 60 s; same with Allow
    s3tables:GetTableBucket only for the bucket Gets
  - measurements: roles=3, passes=2, sub_read_cells=30, first_pass_after_put_role_policy_s=30.1,
    second_pass_after_put_role_policy_s=238.2, n=30
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-142](../table-policy.md#s3t-table-142), [S3T-TABLE-132](../table-policy.md#s3t-table-132) · evidence:
    table/identity/fu2-06-missing-ns-messages-subread-iam-aroa

## Notes

FU2-08: 132 holds - each table sub-resource Get (GetTableEncryption, GetTableStorageClass,
GetTableMetadataLocation, GetTablePolicy, GetTableMaintenanceConfiguration) is authorized as its own s3tables
action: AccessDeniedException 403 'not authorized to perform: s3tables:<Op> on resource: <table ARN> because
no identity-based policy allows the s3tables:<Op> action' in 30/30 cells (3 fresh roles x 5 ops x 2 passes at
+30 s and +238 s after PutRolePolicy); the same for the bucket-level Gets with a role allowed only
s3tables:GetTableBucket (8/8). 142's HTTP 200 for GetTableEncryption/GetTableStorageClass did not reproduce
and is attributable to IAM propagation of the policy its probe had just narrowed. Controls: GetTable (allowed)
200 6/6; GetTableBucket (not allowed) 403 6/6; GetTableEncryption on a missing name -> 403 naming
s3tables:GetTableEncryption on '.../table/*' (authorization before existence, as 142 saw), GetTable on a
missing name -> NotFoundException 404.

Contradiction with [S3T-TABLE-142](../table-policy.md#s3t-table-142), [S3T-TABLE-132](../table-policy.md#s3t-table-132): FU2-08 settled by follow-up probe
table/identity/fu2-06-missing-ns-messages-subread-iam-aroa: 132 holds - each table sub-resource Get
(GetTableEncryption, GetTableStorageClass, GetTableMetadataLocation, GetTablePolicy,
GetTableMaintenanceConfiguration) is authorized as its own s3tables action: AccessDeniedException 403 'not
authorized to perform: s3tables:<Op> on resource: <table ARN> because no identity-based policy allows the
s3tables:<Op> action' in 30/30 cells (3 fresh roles x Resolution: keep all; [S3T-TABLE-327](../table-maintenance.md#s3t-table-327) is canonical for the
settled rule
