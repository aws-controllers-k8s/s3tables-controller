<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-132: IAM identity-policy denials are AccessDeniedException 403 (IAM-style message) for GetTableMetadataLocation, UpdateTableMetadataLocation...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-132"></a>**S3T-TABLE-132** `error-code` · impact high · handled · verified 2026-10-13
  **IAM identity-policy denials are AccessDeniedException 403 (IAM-style message) for GetTableMetadataLocation, UpdateTableMetadataLocation...**
  Same restricted role (CreateTable/GetTable/DeleteTable on bucket X only). GetTable on X's table -> HTTP 200.
  GetTableMetadataLocation (not granted) -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:GetTableMetadataLocation on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4ecaac-iamx/table/718a9891-d591-494b-9e1c-aca67d534eb4
  because no identity-based policy allows the s3tables:GetTableMetadataLocation action'.
  UpdateTableMetadataLocation (not granted) -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:UpdateTableMetadataLocation on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4ecaac-iamx/table/718a9891-d591-494b-9e1c-aca67d534eb4
  because no identity-based policy allows the s3tables:UpdateTableMetadataLocation action'. GetTable on bucket
  Y's table (out of scope) -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:GetTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4ecaac-iamy/table/8fc73291-2cf5-496c-9f0c-5f596914bbc9
  because no identity-based policy allows the s3tables:GetTable action'; GetTable(tableArn) of that table ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4ecaac-iamy/table/8fc73291-2cf5-496c-9f0c-5f596914bbc9
  because no identity-based policy allows the s3tables:GetTable action'. GetTable on a MISSING table in scope
  -> NotFoundException (HTTP 404) 'The specified table does not exist.' vs out of scope ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTable on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4ecaac-iamy/table/*
  because no identity-based policy allows the s3tables:GetTable action'. ListTables on X (not granted) ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:ListTables on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4ecaac-iamx
  because no identity-based policy allows the s3tables:ListTables action'.
  - ACK: terminal_codes, exceptions.404 · ops: GetTable, GetTableMetadataLocation,
    UpdateTableMetadataLocation, ListTables
  - repro: AssumeRole(restricted) -> GetTable / GetTableMetadataLocation / UpdateTableMetadataLocation /
    ListTables on in-scope and out-of-scope tables and missing names
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-109](../table-policy.md#s3t-table-109), [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLE-140](../table-policy.md#s3t-table-140), [S3T-TABLE-142](../table-policy.md#s3t-table-142), [S3T-TABLE-131](../table-policy.md#s3t-table-131), [S3T-TABLE-327](../table-maintenance.md#s3t-table-327) ·
    hypotheses: H-T-146 · evidence: table/error-taxonomy/iam-putdata-denied

## Notes

H-T-146 IAM part QUALIFIED/REFUTED: GetTableMetadataLocation (which models only ForbiddenException) returns
AccessDeniedException (HTTP 403) 'User: <principal> is not authorized to perform:
s3tables:GetTableMetadataLocation on resource: <tableARN> because no identity-based policy allows ...';
GetTable out of scope -> the same AccessDeniedException shape (as hypothesized for GetTable);
UpdateTableMetadataLocation and ListTables -> AccessDeniedException too. GetTable on a MISSING table: in scope
-> NotFoundException 404; out of scope -> AccessDeniedException 403 naming resource '.../table/*'
(authorization precedes the existence check, no existence leak). The ForbiddenException 403 'Access Denied'
seen for CreateTable-with-metadata without PutTableData is a different code path. One attempt per cell.

Full original title: IAM identity-policy denials are AccessDeniedException 403 (IAM-style message) for
GetTableMetadataLocation, UpdateTableMetadataLocation, ListTables and GetTable alike - not the modeled
ForbiddenException; denial is evaluated before existence

Contradiction with [S3T-TABLE-142](../table-policy.md#s3t-table-142): 132: a role lacking the sub-read permission gets AccessDeniedException 403
for GetTableMetadataLocation on an existing in-scope table; 142: a role allowed only s3tables:GetTable gets
HTTP 200 for GetTableEncryption and GetTableStorageClass on an existing table, yet AccessDeniedException
naming s3tables:GetTableEncryption for a missing name - the records disagree on whether table sub-reads need
their own action (142 is n=1 in a probe that had just re-written the role's policy) Resolution: re-probe:
dependencies - fresh role allowed only s3tables:GetTable, wait >=30 s, then
GetTableEncryption/GetTableStorageClass/GetTableMetadataLocation/GetTablePolicy on an existing table x3 roles;
also the bucket-level Gets with only GetTableBucket

Contradiction with [S3T-TABLE-327](../table-maintenance.md#s3t-table-327), [S3T-TABLE-142](../table-policy.md#s3t-table-142): FU2-08 settled by follow-up probe
table/identity/fu2-06-missing-ns-messages-subread-iam-aroa: 132 holds - each table sub-resource Get
(GetTableEncryption, GetTableStorageClass, GetTableMetadataLocation, GetTablePolicy,
GetTableMaintenanceConfiguration) is authorized as its own s3tables action: AccessDeniedException 403 'not
authorized to perform: s3tables:<Op> on resource: <table ARN> because no identity-based policy allows the
s3tables:<Op> action' in 30/30 cells (3 fresh roles x Resolution: keep all; [S3T-TABLE-327](../table-maintenance.md#s3t-table-327) is canonical for the
settled rule
