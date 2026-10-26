<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-057: Bucket default aws:kms(K): CreateTable without encryptionConfiguration reads aws:kms with kmsKeyArn == K byte-for-byte...
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-057"></a>**S3T-TABLE-057** `server-default` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket default aws:kms(K): CreateTable without encryptionConfiguration reads aws:kms with kmsKeyArn == K byte-for-byte...**
  Key K grants the maintenance principal (no condition). PutTableBucketEncryption(aws:kms, K) -> HTTP 200;
  GetTableBucketEncryption -> {"encryptionConfiguration": {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}}. CreateTable(no
  encryptionConfiguration) -> HTTP 200; GetTableEncryption immediately -> HTTP 200 {"encryptionConfiguration":
  {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}} (kmsKeyArn equals the key ARN
  returned by CreateKey). CreateTable(encryptionConfiguration={AES256}) under the same bucket -> HTTP 200;
  GetTableEncryption -> HTTP 200 {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare · ops: PutTableBucketEncryption, CreateTable,
    GetTableEncryption · fields: encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: PutTableBucketEncryption(aws:kms, K with maintenance grant) -> CreateTable(no config) and
    CreateTable({AES256}) -> GetTableEncryption on both
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-255](../table-encryption-storage.md#s3t-table-255), [S3T-TABLE-259](../table-encryption-storage.md#s3t-table-259), [S3T-TABLE-257](../table-encryption-storage.md#s3t-table-257), [S3T-TABLEBUCKET-120](../tablebucket-encryption-storage.md#s3t-tablebucket-120), [S3T-TABLE-010](../table-encryption-storage.md#s3t-table-010), [S3T-TABLE-054](../table-maintenance.md#s3t-table-054),
    [S3T-TABLE-055](../table-maintenance.md#s3t-table-055), [S3T-TABLE-056](../table-maintenance.md#s3t-table-056), [S3T-TABLE-141](../table-policy.md#s3t-table-141), [S3T-TABLE-021](../table-encryption-storage.md#s3t-table-021), [S3T-TABLE-058](../table-encryption-storage.md#s3t-table-058), [S3T-TABLE-328](../table-policy.md#s3t-table-328) · hypotheses:
    H-T-152, H-T-162 · evidence: table/dependencies/kms-maintenance-principal

## Notes

H-T-152 confirmed; H-T-162 confirmed.

Full original title: Bucket default aws:kms(K): CreateTable without encryptionConfiguration reads aws:kms with
kmsKeyArn == K byte-for-byte; explicit AES256 reads AES256 kmsKeyArn absent

Contradiction with [S3T-TABLE-141](../table-policy.md#s3t-table-141): CreateTable({AES256}) -> HTTP 200 in 057 (lab principal) vs
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

Contradiction with [S3T-TABLE-328](../table-policy.md#s3t-table-328), [S3T-TABLE-141](../table-policy.md#s3t-table-141): FU2-07 settled by follow-up probe
tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny: both readings are right under different
conditions, and 141's attribution was wrong. With kms:* explicitly denied (fresh role, 60 s after
PutRolePolicy, never carried another statement): CreateTable {AES256} -> HTTP 200 and CreateTable without
encryptionConfiguration -> HTTP 200 in both an AES256-default and an aws:kms(K)-default bucket (057's reading;
the KMS-default bucket cell shows the Resolution: keep all; [S3T-TABLE-328](../table-policy.md#s3t-table-328) is canonical for the settled rule
