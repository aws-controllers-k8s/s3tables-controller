<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-004: CreateTableBucket is synchronous: GetTableBucket succeeds on the first poll ~0 ms after Create, no status/state field exists
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-004"></a>**S3T-TABLEBUCKET-004** `async-state-machine` · impact high · partially handled · verified 2026-10-13
  **CreateTableBucket is synchronous: GetTableBucket succeeds on the first poll ~0 ms after Create, no status/state field exists**
  For 5 fresh buckets GetTableBucket issued immediately after CreateTableBucket returned (poll interval 50 ms)
  returned HTTP 200 on attempt sequence ["['OK']"]; the body always had exactly ['arn', 'createdAt', 'name',
  'ownerAccountId', 'tableBucketId', 'type'] (no status/state enum). type=['customer'], ownerAccountId ==
  caller account in 5/5, tableBucketId is a UUID in 5/5, createdAt minus the local pre-call time was [1.5,
  1.51, 1.43, 1.42, 1.47] s.
  - ACK: none, hook:sdk_create_post_set_output · ops: CreateTableBucket, GetTableBucket · fields: createdAt,
    ownerAccountId, tableBucketId, type
  - repro: t0=now; CreateTableBucket(name); GetTableBucket(arn) every 50 ms until 200; inspect body
  - measurements: first_get_ok_s=[0.04, 0.03, 0.03, 0.03, 0.03], createdAt_minus_t0_s=[1.504, 1.509, 1.432,
    1.419, 1.468], n=5
  - handling: partially handled via `generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; generator.yaml:45-46; generator.yaml:60-65; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304` - see Handling gaps
  - related: [S3T-TABLEBUCKET-003](../tablebucket.md#s3t-tablebucket-003), [S3T-TABLEBUCKET-121](../tablebucket.md#s3t-tablebucket-121), [S3T-TABLEBUCKET-122](../tablebucket.md#s3t-tablebucket-122), [S3T-TABLEBUCKET-002](../tablebucket.md#s3t-tablebucket-002),
    [S3T-TABLEBUCKET-130](../tablebucket.md#s3t-tablebucket-130), [S3T-TABLEBUCKET-042](../tablebucket.md#s3t-tablebucket-042) · hypotheses: H-B-002 · evidence:
    tablebucket/round-trip/create-identity-defaults

## Notes

Confirms H-B-002: no CREATING/DELETING phase was observable; the controller needs no requeue after create and
can fill Status (ownerAccountId, createdAt, tableBucketId, type) from one GetTableBucket right after Create.

Partially handled in controller via: [GT-S3T-040](../service.md#gt-s3t-040) (controller hooks catalog entry): sdk_read_one_post_set_output hooks on all three resources
assign Status.NamespaceID / Status.TableBucketID from resp.NamespaceId / resp.TableBucketId by hand. (;
[GT-S3T-041](../service.md#gt-s3t-041) (controller hooks catalog entry): go_tag: json:"type,omitempty" overrides on Table.Type, TableBucket.Type and
Metadata.Iceberg.Schema.Fields.Type so the CRD exposes `type`.; [GT-S3T-042](../service.md#gt-s3t-042) (controller hooks catalog entry): Generated setStatusDefaults fills
ACKResourceMetadata.OwnerAccountID from rm.awsAccountID; no `from:` field is declared for ownerAccountId. e2e
asserts status.a (scorer verdict: partial match)
