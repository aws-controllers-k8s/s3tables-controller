<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-253: CreateTable(aws:kms) 400 'Insufficient access to perform table maintenance' flips on the FIRST call after PutKeyPolicy (0.4-0.6 s, n=2)
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-253"></a>**S3T-TABLE-253** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **CreateTable(aws:kms) 400 'Insufficient access to perform table maintenance' flips on the FIRST call after PutKeyPolicy (0.4-0.6 s, n=2)**
  Key created with a root-only policy. CreateTable(aws:kms) at +0.4/+1.8/+3.2 s after CreateKey ->
  BadRequestException (HTTP 400) 'Insufficient access to perform table maintenance. Ensure that the key used
  to encrypt the table still exists, is active, and has a resource policy granting access to the S3 service
  principal 'maintenance.s3tables.amazonaws.com'.'. PutKeyPolicy adding Allow
  maintenance.s3tables.amazonaws.com kms:GenerateDataKey/Decrypt/DescribeKey, then CreateTable(aws:kms, fresh
  name) every 1 s: 400 x 0/0 before the first 200 at 0.5/0.6 s (2 cycles). PutKeyPolicy back to root-only,
  then CreateTable every 1 s: 200 x 0/0 before the first 400 at 0.4/0.4 s (same message). Tables created
  during the window are normal (deleted 204 at once).
  - ACK: requeue, terminal_codes · ops: CreateTable, PutKeyPolicy · fields: encryptionConfiguration.kmsKeyArn
  - repro: CreateKey(root-only policy) -> CreateTable(aws:kms) 400 -> PutKeyPolicy(+maintenance.s3tables
    grant) -> CreateTable every 1 s until 200; PutKeyPolicy(root-only) -> CreateTable every 1 s until 400
  - measurements: grant_to_200_s=[0.51, 0.62], revoke_to_400_s=[0.37, 0.42], attempts_grant=[1, 1],
    attempts_revoke=[1, 1], poll_interval_s=1.0, n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-054](../table-maintenance.md#s3t-table-054), [S3T-TABLE-239](../table-record-expiration.md#s3t-table-239), [S3T-TABLE-301](../table-maintenance.md#s3t-table-301), [S3T-TABLE-254](../table-maintenance.md#s3t-table-254), [S3T-TABLE-246](../table-maintenance.md#s3t-table-246), [S3T-TABLE-248](../table-encryption-storage.md#s3t-table-248),
    [S3T-TABLE-235](../table-maintenance.md#s3t-table-235) · evidence: table/creative/kms-grant-propagation

## Notes

Extends [S3T-TABLE-054](../table-maintenance.md#s3t-table-054) (400 on a key without the grant) and [S3T-TABLE-239](../table-record-expiration.md#s3t-table-239) (400 after revoking) with the
propagation lags: the same 400 BadRequestException text is PERMANENT-SPEC while the key policy is wrong and
clears on the very next CreateTable after the key policy is fixed (no KMS->s3tables propagation lag or
negative caching was observable at 1 s resolution), so a controller can treat it as 'fix the key policy, then
requeue' - the retry succeeds immediately once the referenced key is right; code sequences grant=[['HTTP
200'], ['HTTP 200']] revoke=[['BadRequestException/400'], ['BadRequestException/400']]. n=2 cycles, 1 s
polling.

Full original title: CreateTable(aws:kms) 400 'Insufficient access to perform table maintenance' flips on the
FIRST call after PutKeyPolicy (grant 0.5/0.6 s, revoke 0.4/0.4 s): no propagation window
