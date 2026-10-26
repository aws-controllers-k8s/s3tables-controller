<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-117: CreateTableBucket accepts storageClassConfiguration + encryptionConfiguration (HTTP 200)...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-117"></a>**S3T-TABLEBUCKET-117** `sub-resource-api` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTableBucket accepts storageClassConfiguration + encryptionConfiguration (HTTP 200)...**
  CreateTableBucket(name, tags, storageClassConfiguration=INTELLIGENT_TIERING, encryptionConfiguration=AES256)
  -> HTTP 200; GetTableBucketStorageClass -> {'storageClassConfiguration': {'storageClass':
  'INTELLIGENT_TIERING'}}, GetTableBucketEncryption -> {'encryptionConfiguration': {'sseAlgorithm':
  'AES256'}}. A plain bucket then PutTableBucketStorageClass(INTELLIGENT_TIERING) (HTTP 200) and
  PutTableBucketEncryption(AES256) (HTTP 200) reads {'storageClassConfiguration': {'storageClass':
  'INTELLIGENT_TIERING'}} / {'encryptionConfiguration': {'sseAlgorithm': 'AES256'}}. The CreateTableBucket
  input shape has members ['encryptionConfiguration', 'name', 'storageClassConfiguration', 'tags']: policy,
  maintenance, metrics and replication cannot be set at create.
  - ACK: custom_create, scope:field-on-parent · ops: CreateTableBucket, GetTableBucketStorageClass,
    GetTableBucketEncryption, PutTableBucketStorageClass, PutTableBucketEncryption · fields:
    storageClassConfiguration, encryptionConfiguration
  - repro: CreateTableBucket with both configs -> Gets; CreateTableBucket plain -> Puts -> Gets; diff; inspect
    the model
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-146](../tablebucket-encryption-storage.md#s3t-tablebucket-146), [S3T-TABLEBUCKET-147](../tablebucket-encryption-storage.md#s3t-tablebucket-147), [S3T-TABLEBUCKET-084](../tablebucket-encryption-storage.md#s3t-tablebucket-084), [S3T-TABLEBUCKET-009](../tablebucket-encryption-storage.md#s3t-tablebucket-009),
    [S3T-TABLEBUCKET-059](../tablebucket-encryption-storage.md#s3t-tablebucket-059), [S3T-TABLEBUCKET-008](../tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLEBUCKET-060](../tablebucket.md#s3t-tablebucket-060), [S3T-TABLEBUCKET-116](../tablebucket-replication.md#s3t-tablebucket-116) · hypotheses: H-S-041,
    H-S-012 · evidence: tablebucket/identity/subresources-across-recreate

## Notes

H-S-041 confirmed. H-S-012 qualified: only the AES256 cell was run (create-time and post-create bodies
identical); the aws:kms-without-maintenance-access cell needs a KMS key and was not run in this probe.

Full original title: CreateTableBucket accepts storageClassConfiguration + encryptionConfiguration (HTTP 200);
Get bodies identical to a bucket created plain then Put; create input has only ['encryptionConfiguration',
'name', 'storageClassConfiguration', 'tags']
