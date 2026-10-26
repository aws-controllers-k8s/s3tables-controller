<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-082: GetTable/DeleteTable with a missing namespace, table or bucket -> NotFoundException 404 (6/6 cells); DeleteTable messages name the parent
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-082"></a>**S3T-TABLE-082** `error-code` · impact medium · handled · verified 2026-10-13
  **GetTable/DeleteTable with a missing namespace, table or bucket -> NotFoundException 404 (6/6 cells); DeleteTable messages name the parent**
  GetTable: missing_namespace -> NotFoundException (HTTP 404) 'The specified table does not exist.';
  missing_table -> NotFoundException (HTTP 404) 'The specified table does not exist.'; missing_bucket ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'. DeleteTable: missing_namespace ->
  NotFoundException (HTTP 404) 'The specified namespace does not exist.'; missing_table -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'; missing_bucket -> NotFoundException (HTTP 404) 'The
  specified bucket does not exist.'. Distinct GetTable messages: 1; distinct DeleteTable messages: 3.
  - ACK: exceptions.404, requeue · ops: GetTable, DeleteTable · fields: namespace, name, tableBucketARN
  - repro: GetTable and DeleteTable with (a) missing namespace, (b) missing table, (c) missing bucket in an
    existing bucket/namespace
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-027](../table.md#s3t-table-027), [S3T-TABLE-267](../table.md#s3t-table-267), [S3T-TABLE-053](../table.md#s3t-table-053), [S3T-TABLE-130](../table.md#s3t-table-130), [S3T-TABLE-080](../table.md#s3t-table-080), [S3T-TABLE-094](../table.md#s3t-table-094),
    [S3T-TABLE-049](../table.md#s3t-table-049), [S3T-TABLE-326](../table.md#s3t-table-326) · hypotheses: H-T-043 · evidence: table/error-taxonomy/create-validation

## Notes

H-T-043 qualified: one code (NotFoundException) for all six calls; only DeleteTable's message distinguishes
the missing parent, GetTable's does not. A ReadOne cannot tell 'namespace/bucket gone' from 'table gone' (same
code AND message); a controller wanting to requeue on a missing parent must check the parent itself.

Full original title: Missing namespace/table/bucket -> NotFoundException on both ops; GetTable messages are
IDENTICAL ('table does not exist'), DeleteTable messages name the missing parent

Contradiction with [S3T-TABLE-027](../table.md#s3t-table-027): Same code (NotFoundException 404) but the GetTable message for a nonexistent
namespace disagrees: 082 reports 'The specified table does not exist.' for its missing_namespace cell and
concludes GetTable messages are identical across missing namespace/table/bucket; 027 (upper-cased namespace)
and 267 (upper-cased and trailing-space namespace) got 'The specified namespace does not exist.' from
GetTable. Two probes with three cells vs one cell; no finding explains the difference
(deleted-vs-never-created, name shape) Resolution: re-probe: identity - GetTable(tableBucketARN, namespace,
name) with namespace never-created / just-deleted / upper-cased / trailing-space, plus a never-created bucket,
3 trials each, record the exact message per cell; meanwhile keep both, [S3T-TABLE-082](../table.md#s3t-table-082)'s 'identical messages'
clause is dropped from its title (see title_fixes) and 027/267 are the better-supported reading

Contradiction with [S3T-TABLE-326](../table.md#s3t-table-326), [S3T-TABLE-027](../table.md#s3t-table-027), [S3T-TABLE-267](../table.md#s3t-table-267): FU2-06 settled by follow-up probe
table/identity/fu2-06-missing-ns-messages-subread-iam-aroa: both readings are right under different conditions -
the GetTable message depends on the SHAPE of the missing namespace, not on whether it was deleted or never
created: a well-formed lowercase namespace that does not exist (never created 3/3, created-then-deleted 3/3,
never-created bucket ARN 3/3, never-created table 3/3, both missing 3/3) -> 'The specified table does not
exist.' (082's re Resolution: keep all; [S3T-TABLE-326](../table.md#s3t-table-326) is canonical for the settled rule
