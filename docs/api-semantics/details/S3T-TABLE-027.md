<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-027: GetTable/DeleteTable with the existing table's name upper-cased -> NotFoundException / NotFoundException; table untouched
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-027"></a>**S3T-TABLE-027** `request-validation` · impact medium · handled · verified 2026-10-13
  **GetTable/DeleteTable with the existing table's name upper-cased -> NotFoundException / NotFoundException; table untouched**
  For existing table '<lower>' GetTable(name=<UPPER>) -> NotFoundException (HTTP 404) 'The specified table
  does not exist.'; DeleteTable(name=<UPPER>) -> NotFoundException (HTTP 404) 'The specified table does not
  exist.'; GetTable(namespace=<UPPER>, name=<lower>) -> NotFoundException (HTTP 404) 'The specified namespace
  does not exist.'; afterwards GetTable(name=<lower>) -> HTTP 200 (table still exists).
  - ACK: terminal_codes · ops: GetTable, DeleteTable · fields: name, namespace
  - repro: Create t1 (lower-case); GetTable/DeleteTable with name=t1.upper(); GetTable(name=t1)
  - handling: handled via `test/e2e/tests/test_namespace.py:55-57; test/e2e/tests/test_table.py:63-64; generator.yaml:10-13; generator.yaml:31-39`
  - related: [S3T-TABLE-082](../table.md#s3t-table-082), [S3T-TABLE-267](../table.md#s3t-table-267), [S3T-TABLE-053](../table.md#s3t-table-053), [S3T-TABLE-130](../table.md#s3t-table-130), [S3T-TABLE-078](../table.md#s3t-table-078), [S3T-TABLE-052](../table.md#s3t-table-052),
    [S3T-TABLE-143](../table.md#s3t-table-143), [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100), [S3T-TABLE-326](../table.md#s3t-table-326) · hypotheses: H-T-011 · evidence:
    table/identity/get-table-identifier-forms

## Notes

H-T-011 REFUTED on the error code: no case folding (the lower-case table is untouched: True); an upper-case
name is NOT rejected as a pattern violation (no 400); it is simply looked up and yields NotFoundException
(HTTP 404), indistinguishable from a missing table.

Contradiction with [S3T-TABLE-082](../table.md#s3t-table-082): Same code (NotFoundException 404) but the GetTable message for a nonexistent
namespace disagrees: 082 reports 'The specified table does not exist.' for its missing_namespace cell and
concludes GetTable messages are identical across missing namespace/table/bucket; 027 (upper-cased namespace)
and 267 (upper-cased and trailing-space namespace) got 'The specified namespace does not exist.' from
GetTable. Two probes with three cells vs one cell; no finding explains the difference
(deleted-vs-never-created, name shape) Resolution: re-probe: identity - GetTable(tableBucketARN, namespace,
name) with namespace never-created / just-deleted / upper-cased / trailing-space, plus a never-created bucket,
3 trials each, record the exact message per cell; meanwhile keep both, [S3T-TABLE-082](../table.md#s3t-table-082)'s 'identical messages'
clause is dropped from its title (see title_fixes) and 027/267 are the better-supported reading

Contradiction with [S3T-TABLE-326](../table.md#s3t-table-326), [S3T-TABLE-082](../table.md#s3t-table-082), [S3T-TABLE-267](../table.md#s3t-table-267): FU2-06 settled by follow-up probe
table/identity/fu2-06-missing-ns-messages-subread-iam-aroa: both readings are right under different conditions -
the GetTable message depends on the SHAPE of the missing namespace, not on whether it was deleted or never
created: a well-formed lowercase namespace that does not exist (never created 3/3, created-then-deleted 3/3,
never-created bucket ARN 3/3, never-created table 3/3, both missing 3/3) -> 'The specified table does not
exist.' (082's re Resolution: keep all; [S3T-TABLE-326](../table.md#s3t-table-326) is canonical for the settled rule
