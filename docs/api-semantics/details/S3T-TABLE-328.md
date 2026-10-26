<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-328: CreateTable {aws:kms K} needs the caller's kms:DescribeKey (ForbiddenException 403 naming it); {AES256}/no config -> 200 with kms:* denied
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-328"></a>**S3T-TABLE-328** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **CreateTable {aws:kms K} needs the caller's kms:DescribeKey (ForbiddenException 403 naming it); {AES256}/no config -> 200 with kms:* denied**
  Key K: policy = account root kms:* + maintenance.s3tables.amazonaws.com GenerateDataKey/Decrypt. Buckets:
  AES256-default and aws:kms(K)-default. Fresh roles (trust this account, one inline policy, never carried any
  other statement), first CreateTable {'denykms': 60.8, 'allowkms': 64.3, 'denyput': 67.6} s after
  PutRolePolicy: {"allowkms": {"aesbucket/AES256": "HTTP 200", "aesbucket/kms_K": "ForbiddenException (HTTP 403)
  [kms:DescribeKey]", "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "HTTP 200", "kmsbucket/kms_K":
  "ForbiddenException (HTTP 403) [kms:DescribeKey]", "kmsbucket/none": "HTTP 200"}, "denykms":
  {"aesbucket/AES256": "HTTP 200", "aesbucket/kms_K": "ForbiddenException (HTTP 403) [kms:DescribeKey]",
  "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "HTTP 200", "kmsbucket/kms_K": "ForbiddenException (HTTP 403)
  [kms:DescribeKey]", "kmsbucket/none": "HTTP 200"}, "denyput": {"aesbucket/AES256": "AccessDeniedException
  (HTTP 403) [s3tables:PutTableEncryption]", "aesbucket/kms_K": "AccessDeniedException (HTTP 403)
  [s3tables:PutTableEncryption]", "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "AccessDeniedException
  (HTTP 403) [s3tables:PutTableEncryption]", "kmsbucket/kms_K": "AccessDeniedException (HTTP 403)
  [s3tables:PutTableEncryption]", "kmsbucket/none": "HTTP 200"}, "lab": {"aesbucket/AES256": "HTTP 200",
  "aesbucket/kms_K": "HTTP 200", "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "HTTP 200",
  "kmsbucket/kms_K": "HTTP 200", "kmsbucket/none": "HTTP 200"}}. Full cells incl. messages:
  denykms={"aesbucket/AES256": "HTTP 200", "aesbucket/kms_K": "ForbiddenException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/02a3c4ea-41c2-4ece-8830-303258e701f4 with an explicit deny in an
  identity-based policy'", "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "HTTP 200", "kmsbucket/kms_K":
  "ForbiddenException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/02a3c4ea-41c2-4ece-8830-303258e701f4 with an explicit deny in an
  identity-based policy'", "kmsbucket/none": "HTTP 200"}; allowkms={"aesbucket/AES256": "HTTP 200",
  "aesbucket/kms_K": "ForbiddenException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is
  not authorized to perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/02a3c4ea-41c2-4ece-8830-303258e701f4 because no identity-based policy
  allows the kms:DescribeKey action'", "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "HTTP 200",
  "kmsbucket/kms_K": "ForbiddenException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is
  not authorized to perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/02a3c4ea-41c2-4ece-8830-303258e701f4 because no identity-based policy
  allows the kms:DescribeKey action'", "kmsbucket/none": "HTTP 200"}; denyput={"aesbucket/AES256":
  "AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized
  to perform: s3tables:PutTableEncryption on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fda88c-fu2aes with an explicit deny in an identity-based
  policy'", "aesbucket/kms_K": "AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableEncryption on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fda88c-fu2aes with an explicit deny in an
  identity-based policy'", "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "AccessDeniedException (HTTP 403)
  'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:PutTableEncryption on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fda88c-fu2kms with
  an explicit deny in an identity-based policy'", "kmsbucket/kms_K": "AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableEncryption on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fda88c-fu2kms with an explicit deny in an
  identity-based policy'", "kmsbucket/none": "HTTP 200"}.
  - ACK: terminal_codes, docs-only · ops: CreateTable, GetTableEncryption · fields:
    encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: CreateKey(K) -> CreateTableBucket x2 (AES256 default, aws:kms K default) -> CreateRole x3 (Allow
    s3tables:* + Deny kms:* | + Allow kms:GenerateDataKey/Decrypt on K | + Deny s3tables:PutTableEncryption)
    -> wait 60 s -> AssumeRole -> CreateTable {aws:kms K}, {AES256}, no config in each bucket
  - measurements: roles=3, cells_per_role=6, wait_after_put_role_policy_s=[60.8, 64.3, 67.6], n=18
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-141](../table-policy.md#s3t-table-141), [S3T-TABLE-057](../table-encryption-storage.md#s3t-table-057) · evidence:
    tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny

## Notes

FU2-07: both readings are right under different conditions, and 141's attribution was wrong. With kms:*
explicitly denied (fresh role, 60 s after PutRolePolicy, never carried another statement): CreateTable
{AES256} -> HTTP 200 and CreateTable without encryptionConfiguration -> HTTP 200 in both an AES256-default and
an aws:kms(K)-default bucket (057's reading; the KMS-default bucket cell shows the table inherits aws:kms K
without any caller KMS permission); CreateTable {aws:kms K} -> ForbiddenException (HTTP 403, a MODELED code)
'User: ... is not authorized to perform: kms:DescribeKey on resource: <key ARN> with an explicit deny in an
identity-based policy' (2/2 buckets). A role allowed only kms:GenerateDataKey+kms:Decrypt on K gets the same
ForbiddenException naming kms:DescribeKey (implicit deny) - the caller needs kms:DescribeKey, not the data-key
actions, to name a key at CreateTable. A fresh role with Deny s3tables:PutTableEncryption reproduces 141's
exact message shape (AccessDeniedException 403 naming s3tables:PutTableEncryption on the BUCKET ARN) for
{AES256} and {aws:kms K} alike while no-config -> 200: CreateTable with an encryptionConfiguration is
additionally authorized as s3tables:PutTableEncryption, and 141's identical 403 across all its cells came from
the lingering PutTableEncryption deny of its earlier phase, not from KMS permissions. GetTableEncryption on a
KMS table needs no KMS permission (200 for all three roles, as 141 noted).

Contradiction with [S3T-TABLE-141](../table-policy.md#s3t-table-141), [S3T-TABLE-057](../table-encryption-storage.md#s3t-table-057): FU2-07 settled by follow-up probe
tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny: both readings are right under different
conditions, and 141's attribution was wrong. With kms:* explicitly denied (fresh role, 60 s after
PutRolePolicy, never carried another statement): CreateTable {AES256} -> HTTP 200 and CreateTable without
encryptionConfiguration -> HTTP 200 in both an AES256-default and an aws:kms(K)-default bucket (057's reading;
the KMS-default bucket cell shows the Resolution: keep all; [S3T-TABLE-328](../table-policy.md#s3t-table-328) is canonical for the settled rule
