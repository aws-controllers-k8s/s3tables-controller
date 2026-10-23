<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-141: CreateTable({aws:kms K} or {AES256}) by the Deny-kms:* probe role -> AccessDeniedException 403 (unmodeled) on s3tables:PutTableEncryption
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-141"></a>**S3T-TABLE-141** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable({aws:kms K} or {AES256}) by the Deny-kms:* probe role -> AccessDeniedException 403 (unmodeled) on s3tables:PutTableEncryption**
  Key policy grants the maintenance principal and delegates IAM-user permissions to the account. Role with
  Allow s3tables:* and Deny kms:*: CreateTable {aws:kms K} -> AccessDeniedException (HTTP 403, NOT in the
  modeled error list) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:PutTableEncryption on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an
  explicit deny in an identity-based policy'; CreateTable {AES256} -> AccessDeniedException (HTTP 403, NOT in
  the modeled error list) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:PutTableEncryption on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an
  explicit deny in an identity-based policy'; GetTableEncryption on an existing KMS table -> HTTP 200. Allow
  only kms:GenerateDataKey on K -> AccessDeniedException (HTTP 403, NOT in the modeled error list) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableEncryption on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an explicit deny in an
  identity-based policy'. Allow only kms:Decrypt on K -> AccessDeniedException (HTTP 403, NOT in the modeled
  error list) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:PutTableEncryption on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an
  explicit deny in an identity-based policy'. Allow kms:GenerateDataKey+kms:Decrypt on K ->
  AccessDeniedException (HTTP 403, NOT in the modeled error list) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableEncryption on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an explicit deny in an
  identity-based policy'.
  - ACK: terminal_codes, docs-only · ops: CreateTable, GetTableEncryption · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: Assume a role with s3tables:* and (Deny kms:* | Allow kms:GenerateDataKey | Allow kms:Decrypt |
    Allow both) on K; CreateTable(aws:kms K)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-054](../table-maintenance.md#s3t-table-054), [S3T-TABLE-055](../table-maintenance.md#s3t-table-055), [S3T-TABLE-056](../table-maintenance.md#s3t-table-056), [S3T-TABLE-057](../table-encryption-storage.md#s3t-table-057), [S3T-TABLE-179](../table-maintenance.md#s3t-table-179), [S3T-TABLE-216](../table-policy.md#s3t-table-216),
    [S3T-TABLE-328](../table-policy.md#s3t-table-328) · hypotheses: H-T-168 · evidence: table/dependencies/iam-permissions

## Notes

H-T-168 qualified/refuted - see behavior. Reading the encryption configuration needs no KMS permission
(GetTableEncryption succeeded with kms:* denied).

Full original title: CreateTable(aws:kms K) by a caller without kms permissions on K -> AccessDeniedException
403; with kms:GenerateDataKey+kms:Decrypt -> AccessDeniedException 403 (GenerateDataKey only:
AccessDeniedException, Decrypt only: AccessDeniedException)

Contradiction with [S3T-TABLE-057](../table-encryption-storage.md#s3t-table-057): CreateTable({AES256}) -> HTTP 200 in 057 (lab principal) vs
AccessDeniedException 403 in 141 (restricted role). The caller differs, but 141's described identity policy
(Allow s3tables:*, Deny kms:*) cannot produce its own error text 'not authorized to perform:
s3tables:PutTableEncryption ... with an explicit deny in an identity-based policy'; every 141 cell - including
the AES256 control (no KMS involved) and the kms:GenerateDataKey+kms:Decrypt-allowed cell - returns that
identical message, so the digest cannot attribute the 403 to missing KMS permissions (an earlier phase of the
same probe had put an explicit Deny s3tables:PutTableEncryption on the same role ~90 s before; IAM propagation
lag is the likely confound) Resolution: re-probe: dependencies - fresh role that never carried a
PutTableEncryption Deny, policy = Allow s3tables:* + Deny kms:*; wait >= 60 s after PutRolePolicy; CreateTable
{aws:kms K}, {AES256}, no config; record code (400 BadRequest vs 403 AccessDenied) and the action named in the
message; repeat with Allow kms:GenerateDataKey+Decrypt

Contradiction with [S3T-TABLE-328](../table-policy.md#s3t-table-328), [S3T-TABLE-057](../table-encryption-storage.md#s3t-table-057): FU2-07 settled by follow-up probe
tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny: both readings are right under different
conditions, and 141's attribution was wrong. With kms:* explicitly denied (fresh role, 60 s after
PutRolePolicy, never carried another statement): CreateTable {AES256} -> HTTP 200 and CreateTable without
encryptionConfiguration -> HTTP 200 in both an AES256-default and an aws:kms(K)-default bucket (057's reading;
the KMS-default bucket cell shows the Resolution: keep all; [S3T-TABLE-328](../table-policy.md#s3t-table-328) is canonical for the settled rule

Consolidation note: Follow-up re-probe outcome: the attribution to missing KMS permissions is refuted: the 403
came from a Deny s3tables:PutTableEncryption on the same role; a fresh Deny kms:* role gets 200 for AES256 and
ForbiddenException naming kms:DescribeKey for aws:kms ([S3T-TABLE-328](../table-policy.md#s3t-table-328)) Resolution: keep as evidence; superseded
reading - a fresh Deny kms:* role gets 200 for AES256 and ForbiddenException naming kms:DescribeKey for
aws:kms ([S3T-TABLE-328](../table-policy.md#s3t-table-328))
