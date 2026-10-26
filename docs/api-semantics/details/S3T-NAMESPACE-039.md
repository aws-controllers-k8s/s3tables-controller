<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-039: Parallel finalizers: DeleteTable||DeleteNamespace -> 400 'not empty' 6/8; DeleteNamespace||DeleteTableBucket -> 400 6/8; +0 ms retry 204
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-039"></a>**S3T-NAMESPACE-039** `delete-semantics` · impact high · handled · verified 2026-10-14
  **Parallel finalizers: DeleteTable||DeleteNamespace -> 400 'not empty' 6/8; DeleteNamespace||DeleteTableBucket -> 400 6/8; +0 ms retry 204**
  Fired from two threads at the same instant, n=8 each. Set 1 (namespace holding one table): DeleteTable ->
  {"HTTP 204": 8}, DeleteNamespace -> {"BadRequestException/400": 6, "HTTP 204": 2} (messages ["The namespace
  that you tried to delete is not empty."]); the pair alone removed the namespace in 2/8 trials; after the
  retry below 0 namespaces and 0 orphan tables were left. Immediate controller retry of the failed
  DeleteNamespace -> {"HTTP 204": 6}. Set 2 (bucket holding one empty namespace): DeleteNamespace -> {"HTTP
  204": 8}, DeleteTableBucket -> {"HTTP 204": 2, "BadRequestException/400": 6} (messages ["The bucket that you
  tried to delete is not empty."]); bucket left in 0/8; immediate retry of the failed DeleteTableBucket ->
  {"HTTP 204": 6}. The parent 'not empty' rejection is BadRequestException 400 and clears on the very next
  call once the child delete has returned.
  - ACK: requeue, terminal_codes, pre-delete-cleanup · ops: DeleteTable, DeleteNamespace, DeleteTableBucket
  - repro: 8x (namespace+table): DeleteTable and DeleteNamespace from two threads; retry the failed parent at
    +0 ms; 8x (bucket+namespace): DeleteNamespace and DeleteTableBucket likewise
  - measurements: set1_parent_rejected=6, set2_parent_rejected=6, set1_namespace_left=0, set2_bucket_left=0,
    n=8
  - handling: handled via `test/e2e/conftest.py:102-106; test/e2e/conftest.py:141-142; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-024](../namespace.md#s3t-namespace-024), [S3T-NAMESPACE-027](../namespace.md#s3t-namespace-027), [S3T-TABLEBUCKET-019](../tablebucket.md#s3t-tablebucket-019), [S3T-NAMESPACE-028](../namespace.md#s3t-namespace-028), [S3T-NAMESPACE-021](../namespace.md#s3t-namespace-021) ·
    evidence: namespace/creative/parallel-finalizers

## Notes

extends [S3T-NAMESPACE-027](../namespace.md#s3t-namespace-027) (sequential +0 ms -> 204) and -024/[S3T-TABLEBUCKET-019](../tablebucket.md#s3t-tablebucket-019) (400 'not empty') with the
concurrent case two controller workers produce. Decision rule: a 400 BadRequestException whose message says
'not empty' on DeleteNamespace/DeleteTableBucket is a requeue (the child finalizer is in flight), not a
terminal error; retrying immediately after the child delete returned succeeds.

Full original title: Parallel finalizers: DeleteTable||DeleteNamespace -> namespace
{"BadRequestException/400": 6, "HTTP 204": 2}; DeleteNamespace||DeleteTableBucket -> bucket {"HTTP 204": 2,
"BadRequestException/400": 6}; +0 ms retry {"HTTP 204": 6}
