<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-024: DeleteNamespace with one table inside -> BadRequestException 400 'not empty' on 6/6 retries over 10 s; table and namespace untouched
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-024"></a>**S3T-NAMESPACE-024** `delete-semantics` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **DeleteNamespace with one table inside -> BadRequestException 400 'not empty' on 6/6 retries over 10 s; table and namespace untouched**
  DeleteNamespace on a namespace holding one ICEBERG table returned BadRequestException with message ['The
  namespace that you tried to delete is not empty.'] on 6/6 attempts at 2 s intervals (latency {'min': 11,
  'max': 59, 'n': 6} ms). Afterwards GetTable -> HTTP 200 and GetNamespace -> HTTP 200 (nothing cascaded). The
  duplicate-create rejection on the same namespace is ConflictException 'A namespace with an identical name
  already exists in the bucket.', so the two rejections are distinguishable by code and message.
  - ACK: requeue, pre-delete-cleanup, terminal_codes · ops: DeleteNamespace, CreateNamespace, GetTable,
    GetNamespace · fields: namespace
  - repro: CreateNamespace -> CreateTable(format=ICEBERG) -> DeleteNamespace x6 at 2 s -> GetTable,
    GetNamespace -> CreateNamespace(same name)
  - measurements: delete_nonempty_latency_ms.min=11, delete_nonempty_latency_ms.max=59,
    delete_nonempty_latency_ms.n=6, attempts=6, n=6
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-019](../tablebucket.md#s3t-tablebucket-019), [S3T-TABLEBUCKET-044](../tablebucket.md#s3t-tablebucket-044), [S3T-TABLEBUCKET-132](../tablebucket.md#s3t-tablebucket-132), [S3T-TABLEBUCKET-070](../tablebucket.md#s3t-tablebucket-070),
    [S3T-NAMESPACE-037](../namespace.md#s3t-namespace-037), [S3T-NAMESPACE-039](../namespace.md#s3t-namespace-039), [S3T-NAMESPACE-027](../namespace.md#s3t-namespace-027), [S3T-NAMESPACE-028](../namespace.md#s3t-namespace-028), [S3T-NAMESPACE-021](../namespace.md#s3t-namespace-021) ·
    hypotheses: H-N-015, H-N-016, H-N-010 · evidence: namespace/dependencies/delete-with-tables

## Notes

H-N-015 refuted; H-N-016 (contrarian: BadRequestException) confirmed; H-N-010 qualified: the delete-non-empty
rejection is BadRequestException (not 409 as hypothesized), it is deterministic (not transient) and its
message differs from the duplicate-create ConflictException (['The namespace that you tried to delete is not
empty.'] vs 'A namespace with an identical name already exists in the bucket.').

Full original title: DeleteNamespace with one table inside -> BadRequestException (HTTP 400) 'The namespace
that you tried to delete is not empty.'; 6/6 retries over 10 s identical; table and namespace untouched

Suspected controller bug confirmed by evidence: DeleteNamespace on a non-empty namespace returns
BadRequestException 400 'The namespace that you tried to delete is not empty.' deterministically (6/6 over 10
s; same code and pattern for DeleteTableBucket with a namespace inside, [S3T-NAMESPACE-039](../namespace.md#s3t-namespace-039)), so with
terminal_codes: [BadRequestException] a Namespace or TableBucket CR deleted together with its children goes
Terminal on the first delete attempt while the child finalizer is still in flight (6/8 concurrent trials in
[S3T-NAMESPACE-039](../namespace.md#s3t-namespace-039)) and never retries, although the identical call returns 204 at +0 ms once the child delete
has returned ([S3T-NAMESPACE-027](../namespace.md#s3t-namespace-027), -039); the 'not empty' BadRequestException must be requeued, not treated as
terminal, and no delete path ever returned ConflictException.
