<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-140: CreateTable with encryptionConfiguration needs s3tables:PutTableEncryption (deny -> AccessDeniedException 403, even for AES256...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-140"></a>**S3T-TABLE-140** `prerequisite` · impact high · handled · verified 2026-10-13
  **CreateTable with encryptionConfiguration needs s3tables:PutTableEncryption (deny -> AccessDeniedException 403, even for AES256...**
  Role policy Allow s3tables:* + kms use on K, explicit Deny s3tables:PutTableEncryption: CreateTable {AES256}
  -> AccessDeniedException (HTTP 403, NOT in the modeled error list) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableEncryption on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an explicit deny in an
  identity-based policy'; {aws:kms K} -> AccessDeniedException (HTTP 403, NOT in the modeled error list)
  'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:PutTableEncryption on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an
  explicit deny in an identity-based policy'; no encryptionConfiguration -> HTTP 200 (GetTableEncryption
  {"sseAlgorithm": "AES256"}); storageClassConfiguration={INTELLIGENT_TIERING} only -> HTTP 200
  (GetTableEncryption {"sseAlgorithm": "AES256"}). With the Deny switched to s3tables:PutTableStorageClass
  instead: {INTELLIGENT_TIERING} -> HTTP 200 (GetTableEncryption {"sseAlgorithm": "AES256"}); no config ->
  HTTP 200 (GetTableEncryption {"sseAlgorithm": "AES256"}).
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: encryptionConfiguration,
    storageClassConfiguration
  - repro: Assume a role with s3tables:* and an explicit Deny on s3tables:PutTableEncryption (then on
    s3tables:PutTableStorageClass); CreateTable with/without each struct
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-109](../table-policy.md#s3t-table-109), [S3T-TABLE-132](../table-policy.md#s3t-table-132), [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLE-142](../table-policy.md#s3t-table-142), [S3T-TABLE-131](../table-policy.md#s3t-table-131), [S3T-TABLEBUCKET-129](../tablebucket-policy.md#s3t-tablebucket-129) ·
    hypotheses: H-T-169 · evidence: table/dependencies/iam-permissions

## Notes

H-T-169 confirmed. The controller's IAM policy must include s3tables:PutTableEncryption (storage class needs
no extra action) whenever spec carries the struct.

Full original title: CreateTable with encryptionConfiguration needs s3tables:PutTableEncryption (deny ->
AccessDeniedException 403, even for AES256: AccessDeniedException 403); storageClassConfiguration with
PutTableStorageClass denied -> HTTP 200
