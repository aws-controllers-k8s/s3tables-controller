<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-122: ListTableBuckets summary at +0 ms equals GetTableBucket field by field (5/5; same 6 fields, raw createdAt strings identical)
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-122"></a>**S3T-TABLEBUCKET-122** `response-fidelity` · impact medium · partially handled · verified 2026-10-14
  **ListTableBuckets summary at +0 ms equals GetTableBucket field by field (5/5; same 6 fields, raw createdAt strings identical)**
  On the first ListTableBuckets(prefix=<name>) after CreateTableBucket (the bucket present on poll #1 in 5/5),
  the TableBucketSummary had fields ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type']
  and GetTableBucket had ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type']; parsed
  values equal on every field in 5/5 trials. Raw HTTP bodies: createdAt string identical between List and Get
  in 5/5 (sample precision '2026-10-14T00:43:19.929801410Z', 9 fractional digits).
  - ACK: list_operation.match_fields, custom_find · ops: ListTableBuckets, GetTableBucket · fields: arn, name,
    ownerAccountId, createdAt, tableBucketId, type, prefix
  - repro: CreateTableBucket -> ListTableBuckets(prefix=name) at +0 ms -> GetTableBucket(arn) -> compare
    summary to body (parsed and raw JSON)
  - measurements: trials_equal=5, raw_createdAt_equal=5, n=5
  - handling: partially handled via `generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; generator.yaml:45-46; generator.yaml:60-65; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304` - see Handling gaps
  - related: [S3T-TABLEBUCKET-010](../tablebucket-encryption-storage.md#s3t-tablebucket-010), [S3T-TABLEBUCKET-003](../tablebucket.md#s3t-tablebucket-003), [S3T-TABLEBUCKET-004](../tablebucket.md#s3t-tablebucket-004), [S3T-TABLEBUCKET-121](../tablebucket.md#s3t-tablebucket-121),
    [S3T-TABLEBUCKET-002](../tablebucket.md#s3t-tablebucket-002), [S3T-TABLEBUCKET-130](../tablebucket.md#s3t-tablebucket-130), [S3T-TABLEBUCKET-042](../tablebucket.md#s3t-tablebucket-042), [S3T-TABLEBUCKET-160](../tablebucket.md#s3t-tablebucket-160), [S3T-TABLEBUCKET-167](../tablebucket.md#s3t-tablebucket-167),
    [S3T-TABLEBUCKET-013](../tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-017](../tablebucket.md#s3t-tablebucket-017), [S3T-NAMESPACE-019](../namespace.md#s3t-namespace-019) · hypotheses: H-B-004 · evidence:
    tablebucket/consistency-windows/create-read-paths-tags-n5

## Notes

H-B-004 confirmed: a controller can adopt/de-duplicate by name from one ListTableBuckets(prefix) call without
a follow-up Get; no field is dropped or reformatted in the summary.

Partially handled in controller via: [GT-S3T-040](../service.md#gt-s3t-040) (controller hooks catalog entry): sdk_read_one_post_set_output hooks on all three resources
assign Status.NamespaceID / Status.TableBucketID from resp.NamespaceId / resp.TableBucketId by hand. (;
[GT-S3T-041](../service.md#gt-s3t-041) (controller hooks catalog entry): go_tag: json:"type,omitempty" overrides on Table.Type, TableBucket.Type and
Metadata.Iceberg.Schema.Fields.Type so the CRD exposes `type`.; [GT-S3T-042](../service.md#gt-s3t-042) (controller hooks catalog entry): Generated setStatusDefaults fills
ACKResourceMetadata.OwnerAccountID from rm.awsAccountID; no `from:` field is declared for ownerAccountId. e2e
asserts status.a (scorer verdict: partial match)
