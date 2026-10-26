<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-006: Namespace identity is (tableBucketARN, name): same name in 2 buckets -> 2 namespaceIds; namespaceId or ARN as namespace -> NotFoundException
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-006"></a>**S3T-NAMESPACE-006** `identity` · impact medium · handled · verified 2026-10-13
  **Namespace identity is (tableBucketARN, name): same name in 2 buckets -> 2 namespaceIds; namespaceId or ARN as namespace -> NotFoundException**
  CreateNamespace(same name) in two buckets -> HTTP 200 and HTTP 200; namespaceIds are 36-char uuid and
  36-char uuid (different). GetNamespace(namespace=<namespaceId>) -> NotFoundException (HTTP 404) 'The
  specified namespace does not exist.'; DeleteNamespace(namespace=<namespaceId>) -> NotFoundException (HTTP 404)
  'The specified namespace does not exist.'; GetNamespace(namespace='arn:...:bucket/<b>/namespace/<name>') ->
  NotFoundException (HTTP 404) 'The specified namespace does not exist.'.
  - ACK: is_arn_primary_key, custom_find, references · ops: CreateNamespace, GetNamespace, DeleteNamespace ·
    fields: namespace, namespaceId, tableBucketARN
  - repro: Two buckets; CreateNamespace([same]) in each; GetNamespace both; GetNamespace/DeleteNamespace with
    the namespaceId string and a synthesized ARN as namespace
  - handling: handled via `pkg/resource/namespace/sdk.go:122-147; pkg/resource/namespace/resource.go:88-116; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-023](../namespace.md#s3t-namespace-023), [S3T-NAMESPACE-003](../namespace.md#s3t-namespace-003), [S3T-NAMESPACE-007](../namespace.md#s3t-namespace-007), [S3T-NAMESPACE-022](../namespace.md#s3t-namespace-022) · hypotheses: H-N-027
    · evidence: namespace/round-trip/create-get-list-shape

## Notes

H-N-027 confirmed: no ARN form exists for a namespace; namespaceId is not accepted as a lookup key
(NotFoundException); the two namespaces have distinct uuid namespaceIds.

Full original title: Namespace identity is (tableBucketARN, name): same name in two buckets -> two distinct
namespaceIds; namespaceId/ARN as namespace param -> NotFoundException/NotFoundException
