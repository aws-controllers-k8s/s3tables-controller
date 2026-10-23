<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-030: SSE-KMS source -> AES256 dest and AES256 source -> SSE-KMS dest: both Put 200, 'completed' ~62-63s; replica encryption follows dest bucket
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-030"></a>**S3T-TABLEREPLICATION-030** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **SSE-KMS source -> AES256 dest and AES256 source -> SSE-KMS dest: both Put 200, 'completed' ~62-63s; replica encryption follows dest bucket** (hypothesis refuted; behavior confirmed)
  Source table tk in a bucket with PutTableBucketEncryption(aws:kms, customer key; key policy grants the
  replication role and the S3 Tables service principals): GetTableEncryption(tk) -> {"kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/6c06ac3e-922e-4619-97f6-71c7ce6ee948", "sseAlgorithm": "aws:kms"}.
  PutTableReplication(tk -> EA, EA has {"sseAlgorithm": "AES256"}, same region) -> HTTP 200; status
  transitions [(0.1, 'no-destinations'), (62.2, 'EA:completed')] (failure message None).
  GetTableEncryption(replica of tk in EA) -> HTTP 200 {"sseAlgorithm": "AES256"}; GetTable(replica)
  {"managedByService": "replication.s3tables.amazonaws.com", "type": "customer"}. Source table tp (AES256
  bucket) -> EK (bucket with aws:kms, same key, same region): Put -> HTTP 200; transitions [(1.1,
  'no-destinations'), (61.1, 'EK:pending'), (63.2, 'EK:completed')] (failure message None);
  GetTableEncryption(replica of tp in EK) -> HTTP 200 {"kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/6c06ac3e-922e-4619-97f6-71c7ce6ee948", "sseAlgorithm": "aws:kms"};
  GetTable(replica) {"managedByService": "replication.s3tables.amazonaws.com", "type": "customer"}.
  - ACK: requeue, synced.when, late_initialize · ops: PutTableReplication, GetTableReplicationStatus,
    GetTableEncryption, PutTableBucketEncryption · fields: encryptionConfiguration.sseAlgorithm,
    encryptionConfiguration.kmsKeyArn, destinations[].replicationStatus, destinations[].failureMessage
  - repro: KMS key (role + service principals) -> SSE-KMS source bucket + table -> Put to an AES256 bucket;
    AES256 table -> Put to an SSE-KMS bucket; poll; GetTableEncryption(replicas)
  - measurements: tk_terminal_s=[62.2], tp_terminal_s=[63.2], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019), [S3T-TABLEREPLICATION-031](../table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEREPLICATION-038](../table-replication.md#s3t-tablereplication-038),
    [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025), [S3T-TABLEREPLICATION-043](../table-replication-status.md#s3t-tablereplication-043), [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005),
    [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-041](../table-replication-status.md#s3t-tablereplication-041), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224) ·
    hypotheses: H-U-192 · evidence: table/cross-region/repl-kms-parity

## Notes

H-U-192 REFUTED: no synchronous encryption-parity check (both Puts 200). Replica encryption follows the
DESTINATION bucket setting (AES256 / aws:kms); empty tables, one key, n=1 per direction (run 2; in run 1 the
AES->KMS direction also completed, and the KMS->AES direction toward the shard's us-east-1 bucket ended
'failed: The specified destination bucket does not exist.' at +62 s because that bucket was deleted by the
teardown probe while the replication was in flight).

Full original title: SSE-KMS source -> AES256 destination: Put 200, status 'EA:completed'; AES256 source ->
SSE-KMS destination: Put 200, status 'EK:completed'; replica encryption "AES256" / "aws:kms"
