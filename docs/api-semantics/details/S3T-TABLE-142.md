<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-142: Role allowed only s3tables:GetTable: GetTableEncryption/StorageClass of an existing table -> 200; missing name -> AccessDeniedException 403
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-142"></a>**S3T-TABLE-142** `error-code` · impact high · handled · verified 2026-10-13
  **Role allowed only s3tables:GetTable: GetTableEncryption/StorageClass of an existing table -> 200; missing name -> AccessDeniedException 403**
  Role allowed only s3tables:GetTable. GetTable(existing KMS table) -> HTTP 200; GetTableEncryption -> HTTP
  200; GetTableStorageClass -> HTTP 200; GetTableEncryption on a non-existent name -> AccessDeniedException
  (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:GetTableEncryption on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam/table/*
  because no identity-based policy allows the s3tables:GetTableEncryption action' (authorization is evaluated
  before existence); CreateTable with the existing name -> ConflictException (HTTP 409) 'A table with an
  identical name already exists in the namespace.'; CreateTable with a new name -> AccessDeniedException (HTTP
  403, NOT in the modeled error list) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized
  to perform: s3tables:CreateTable on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam
  because no identity-based policy allows the s3tables:CreateTable action'.
  - ACK: terminal_codes, hook:sdk_read_one_post_set_output · ops: GetTable, GetTableEncryption,
    GetTableStorageClass, CreateTable
  - repro: Assume a role with only s3tables:GetTable; call GetTable, GetTableEncryption, GetTableStorageClass,
    CreateTable (existing and new name)
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-109](../table-policy.md#s3t-table-109), [S3T-TABLE-132](../table-policy.md#s3t-table-132), [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLE-140](../table-policy.md#s3t-table-140), [S3T-TABLE-131](../table-policy.md#s3t-table-131), [S3T-TABLE-327](../table-maintenance.md#s3t-table-327) ·
    hypotheses: H-T-170 · evidence: table/dependencies/iam-permissions

## Notes

H-T-170 qualified/refuted - see behavior: sub-read deny code None (modeled on those operations: None);
CreateTable deny code AccessDeniedException (modeled: False); x-amzn-ErrorType headers: enc 'None', create
'AccessDeniedException:http://internal.amazon.com/coral/com.amazon.coral.service/'.

Full original title: IAM deny: GetTableEncryption/GetTableStorageClass -> HTTP 200 while GetTable -> HTTP 200;
CreateTable denied -> AccessDeniedException 403 (existing name: ConflictException 409)

Contradiction with [S3T-TABLE-132](../table-policy.md#s3t-table-132): 132: a role lacking the sub-read permission gets AccessDeniedException 403
for GetTableMetadataLocation on an existing in-scope table; 142: a role allowed only s3tables:GetTable gets
HTTP 200 for GetTableEncryption and GetTableStorageClass on an existing table, yet AccessDeniedException
naming s3tables:GetTableEncryption for a missing name - the records disagree on whether table sub-reads need
their own action (142 is n=1 in a probe that had just re-written the role's policy) Resolution: re-probe:
dependencies - fresh role allowed only s3tables:GetTable, wait >=30 s, then
GetTableEncryption/GetTableStorageClass/GetTableMetadataLocation/GetTablePolicy on an existing table x3 roles;
also the bucket-level Gets with only GetTableBucket

Contradiction with [S3T-TABLE-327](../table-maintenance.md#s3t-table-327), [S3T-TABLE-132](../table-policy.md#s3t-table-132): FU2-08 settled by follow-up probe
table/identity/fu2-06-missing-ns-messages-subread-iam-aroa: 132 holds - each table sub-resource Get
(GetTableEncryption, GetTableStorageClass, GetTableMetadataLocation, GetTablePolicy,
GetTableMaintenanceConfiguration) is authorized as its own s3tables action: AccessDeniedException 403 'not
authorized to perform: s3tables:<Op> on resource: <table ARN> because no identity-based policy allows the
s3tables:<Op> action' in 30/30 cells (3 fresh roles x Resolution: keep all; [S3T-TABLE-327](../table-maintenance.md#s3t-table-327) is canonical for the
settled rule

Consolidation note: Follow-up re-probe outcome: the 200 on sub-resource Gets with only s3tables:GetTable did
not reproduce (30/30 AccessDeniedException, 3 fresh roles); attributed to IAM propagation lag; [S3T-TABLE-327](../table-maintenance.md#s3t-table-327)
settles it Resolution: keep as evidence; superseded reading - [S3T-TABLE-327](../table-maintenance.md#s3t-table-327) settles it
