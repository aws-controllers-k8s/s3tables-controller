<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-086: CreateTable(aws:kms): missing key (own acct or other region), malformed ARN -> BadRequestException 400; foreign-account key -> Forbidden 403
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-086"></a>**S3T-TABLE-086** `error-code` · impact high · handled · verified 2026-10-13
  **CreateTable(aws:kms): missing key (own acct or other region), malformed ARN -> BadRequestException 400; foreign-account key -> Forbidden 403**
  zero-UUID key in this account/region -> BadRequestException (HTTP 400) 'The specified KMS key ARN does not
  exist.'; random-UUID key in this account/region -> BadRequestException (HTTP 400) 'The specified KMS key ARN
  does not exist.'; key in another (non-owned) account, same region -> ForbiddenException (HTTP 403) 'User:
  arn:aws:sts::<acct>:assumed-role/<PRINCIPAL> is not authorized to perform: kms:DescribeKey on this resource
  because the resource does not exist in this Region, no resource-based policies allow access, or a
  resource-based policy explicitly denies access'; non-existent key ARN in us-east-1 (other region) ->
  BadRequestException (HTTP 400) 'Cross region KMS keys are not supported.'; malformed ARN (no key/ prefix) ->
  BadRequestException (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'. None of these is s3tables
  NotFoundException; no table is created.
  - ACK: terminal_codes, exceptions.404 · ops: CreateTable, GetTable · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTable(aws:kms, kmsKeyArn=<non-existent ARN variant>) -> GetTable on the triple
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-035](../tablebucket-encryption-storage.md#s3t-tablebucket-035), [S3T-TABLEBUCKET-034](../tablebucket-encryption-storage.md#s3t-tablebucket-034), [S3T-TABLEBUCKET-181](../tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-247](../table-policy.md#s3t-table-247) · hypotheses:
    H-T-161, H-T-160 · evidence: table/weird-inputs/encryption-storage-enums

## Notes

H-T-161 confirmed (not NotFoundException). H-T-160 only partially touched: a well-formed but non-existent key
ARN in us-east-1 gives the same class of error as a non-existent same-region key (no real cross-region key was
created - cross-region is outside this shard).

Full original title: CreateTable(aws:kms) with a key ARN that does not exist (own account, foreign account,
other region, malformed) -> BadRequestException/400, ForbiddenException/403
