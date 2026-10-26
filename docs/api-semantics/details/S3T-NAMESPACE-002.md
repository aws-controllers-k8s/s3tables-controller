<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-002: CreateNamespace 200 body = {tableBucketARN, namespace:[name]} only; GetNamespace adds 5 read-only members, no ARN/tags/status
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-002"></a>**S3T-NAMESPACE-002** `response-fidelity` · impact medium · handled · verified 2026-10-13
  **CreateNamespace 200 body = {tableBucketARN, namespace:[name]} only; GetNamespace adds 5 read-only members, no ARN/tags/status**
  CreateNamespace(tableBucketARN, namespace=[name]) returned HTTP 200 with raw body
  {"namespace":["ackq_86ff1a_ns_a"],"tableBucketARN":"arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-86ff1a-b1"}:
  members ['namespace', 'tableBucketARN']; tableBucketARN byte-identical to the input and namespace is a
  one-element list. GetNamespace at +12 ms returned HTTP 200 with members ['createdAt', 'createdBy',
  'namespace', 'namespaceId', 'ownerAccountId', 'tableBucketId'] (namespace=list[str], createdAt=timestamp,
  createdBy=12-digit account id, ownerAccountId=12-digit account id, namespaceId=36-char uuid,
  tableBucketId=36-char uuid); no member contains an ARN, no tags, no status/state.
  - ACK: hook:sdk_create_post_set_output, is_read_only · ops: CreateNamespace, GetNamespace · fields:
    namespace, tableBucketARN, createdAt, createdBy, ownerAccountId, namespaceId, tableBucketId
  - repro: CreateTableBucket -> CreateNamespace(arn, [name]) -> GetNamespace(arn, name) immediately
  - measurements: get_after_create_latency_ms=12, created_at_minus_client_clock_s=0.014, n=1
  - handling: handled via `generator.yaml:4-9; generator.yaml:123-131; pkg/resource/namespace/sdk.go:122-147; pkg/resource/namespace/resource.go:88-116; generator.yaml:188-189; pkg/resource/namespace/manager.go:284-291; generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304; generator.yaml:57-98`
  - related: [S3T-NAMESPACE-041](../namespace.md#s3t-namespace-041), [S3T-NAMESPACE-005](../namespace.md#s3t-namespace-005), [S3T-NAMESPACE-003](../namespace.md#s3t-namespace-003), [S3T-NAMESPACE-004](../namespace.md#s3t-namespace-004), [S3T-NAMESPACE-025](../namespace.md#s3t-namespace-025) ·
    hypotheses: H-N-001, H-N-002 · evidence: namespace/round-trip/create-get-list-shape

## Notes

H-N-001 confirmed (create echo identical; GetNamespace members exactly {namespace, createdAt, createdBy,
ownerAccountId, namespaceId, tableBucketId}: True). H-N-002 createdAt-vs-client-clock part: createdAt is 0.01
s from the client clock (single sample; the 20-trial part is in consistency-windows/create-delete-visibility).
