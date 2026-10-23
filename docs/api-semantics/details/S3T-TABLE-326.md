<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-326: GetTable NotFound message: well-formed missing namespace -> 'table does not exist'; upper-cased/trailing-space -> 'namespace does not exist'
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-326"></a>**S3T-TABLE-326** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **GetTable NotFound message: well-formed missing namespace -> 'table does not exist'; upper-cased/trailing-space -> 'namespace does not exist'**
  Existing bucket/namespace/table; GetTable(tableBucketARN, namespace, name) 3 trials per cell, outcomes per
  cell: {"bucket_never_created": ["NotFoundException (HTTP 404) 'The specified table does not exist.'"],
  "control_existing": ["HTTP 200"], "ns_and_table_never_created": ["NotFoundException (HTTP 404) 'The
  specified table does not exist.'"], "ns_just_deleted": ["NotFoundException (HTTP 404) 'The specified table
  does not exist.'"], "ns_never_created": ["NotFoundException (HTTP 404) 'The specified table does not
  exist.'"], "ns_trailing_space": ["NotFoundException (HTTP 404) 'The specified namespace does not exist.'"],
  "ns_upper_cased": ["NotFoundException (HTTP 404) 'The specified namespace does not exist.'"],
  "ns_upper_table_never": ["NotFoundException (HTTP 404) 'The specified namespace does not exist.'"],
  "table_never_created": ["NotFoundException (HTTP 404) 'The specified table does not exist.'"]}. DeleteTable
  controls (1 trial): {"bucket_never_created": "NotFoundException (HTTP 404) 'The specified bucket does not
  exist.'", "ns_never_created": "NotFoundException (HTTP 404) 'The specified namespace does not exist.'",
  "table_never_created": "NotFoundException (HTTP 404) 'The specified table does not exist.'"}.
  - ACK: exceptions.404, requeue · ops: GetTable, DeleteTable · fields: namespace, name, tableBucketARN
  - repro: CreateTableBucket -> CreateNamespace -> CreateTable; GetTable with the namespace never created /
    created+deleted / upper-cased / with a trailing space, with a never-created bucket ARN, with a
    never-created table name; 3 trials each
  - measurements: trials_per_cell=3, cells=9, n=27
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-082](../table.md#s3t-table-082), [S3T-TABLE-027](../table.md#s3t-table-027), [S3T-TABLE-267](../table.md#s3t-table-267) · evidence:
    table/identity/fu2-06-missing-ns-messages-subread-iam-aroa

## Notes

FU2-06: both readings are right under different conditions - the GetTable message depends on the SHAPE of the
missing namespace, not on whether it was deleted or never created: a well-formed lowercase namespace that does
not exist (never created 3/3, created-then-deleted 3/3, never-created bucket ARN 3/3, never-created table 3/3,
both missing 3/3) -> 'The specified table does not exist.' (082's reading); a namespace that violates the
[0-9a-z_] shape (upper-cased 3/3, trailing space 3/3, upper-cased + missing table 3/3) -> 'The specified
namespace does not exist.' (027/267's reading). 082's 'identical messages' clause therefore holds for
well-formed names only. DeleteTable controls name the missing parent ('namespace'/'bucket'/'table does not
exist') as 082 recorded.

Contradiction with [S3T-TABLE-082](../table.md#s3t-table-082), [S3T-TABLE-027](../table.md#s3t-table-027), [S3T-TABLE-267](../table.md#s3t-table-267): FU2-06 settled by follow-up probe
table/identity/fu2-06-missing-ns-messages-subread-iam-aroa: both readings are right under different conditions -
the GetTable message depends on the SHAPE of the missing namespace, not on whether it was deleted or never
created: a well-formed lowercase namespace that does not exist (never created 3/3, created-then-deleted 3/3,
never-created bucket ARN 3/3, never-created table 3/3, both missing 3/3) -> 'The specified table does not
exist.' (082's re Resolution: keep all; [S3T-TABLE-326](../table.md#s3t-table-326) is canonical for the settled rule
