<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-006: RenameTable(newNamespaceName only) moves the table synchronously (HTTP 204); namespaceId follows the destination, tableARN unchanged
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-006"></a>**S3T-TABLE-006** `identity` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **RenameTable(newNamespaceName only) moves the table synchronously (HTTP 204); namespaceId follows the destination, tableARN unchanged**
  RenameTable(ns1,t2, newNamespaceName=ns2) with no newName -> HTTP 204. GetTable(ns2,t2) -> HTTP 200;
  GetTable(tableArn) -> HTTP 200 with namespace=['<ns2>']. Changed fields vs before the move: modifiedAt,
  namespace, namespaceId; namespaceId equals GetNamespace(ns2).namespaceId: True. ListTables(namespace=ns1) ->
  HTTP 200 with 0 tables; ListTables(namespace=ns2) -> 1 tables.
  - ACK: is_immutable, custom_update, is_arn_primary_key · ops: RenameTable, GetTable, ListTables,
    GetNamespace · fields: newNamespaceName, namespace, namespaceId, tableARN, versionToken, warehouseLocation
  - repro: CreateNamespace ns1, ns2; CreateTable ns1/t; RenameTable(ns1,t,newNamespaceName=ns2);
    GetTable(ns2,t), GetTable(tableArn); ListTables(namespace=ns1)
  - measurements: move_latency_ms=[143], n=1
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLE-022](../table.md#s3t-table-022), [S3T-TABLE-029](../table.md#s3t-table-029), [S3T-TABLE-008](../table.md#s3t-table-008), [S3T-TABLE-068](../table.md#s3t-table-068), [S3T-TABLE-069](../table.md#s3t-table-069), [S3T-TABLE-070](../table.md#s3t-table-070),
    [S3T-TABLE-071](../table.md#s3t-table-071) · hypotheses: H-T-057 · evidence: table/identity/rename-identity

## Notes

H-T-057 confirmed: versionToken changed on move: False; warehouseLocation changed: False; tableBucketId
changed: False. DeleteNamespace of the emptied source ns is exercised later in this probe.

Suspected controller bug confirmed by evidence: not flagged; evidence-driven. (1) After an out-of-band
RenameTable(newNamespaceName) the ARN-keyed read writes the new namespace into latest.Spec.Namespace, the
delta flags Spec.Namespace every reconcile, and customUpdateTable silently ignores it - no RenameTable back,
no condition, Synced=True - so the CR permanently reports a namespace the table is no longer in (not the
feared immutable-field Terminal; a user edit of spec.namespace is simply rejected by the CRD CEL rule); (2)
ConflictException is not a Table terminal code, so a permanent 409 name collision on RenameTable is retried
with backoff forever (the 'loops on collision' sign), while the transient 409s (concurrent modification, stale
versionToken) do recover because each retry re-reads versionToken - either surface the namespace drift
(condition or RenameTable newNamespaceName back to spec) and classify 409 'identical name already exists' as
terminal by message. code: pkg/resource/table/sdk.go:159-161
(templates/hooks/table/sdk_read_one_post_set_output.go.tpl:3-5) sets ko.Spec.Namespace = resp.Namespace[0] on
the latest read by ARN (sdk.go:84-91); pkg/resource/table/ delta.go:146-152 adds Spec.Namespace;
pkg/resource/table/hooks.go:76-112 customUpdateTable handles only Spec.Name, Spec.MetadataLocation, Spec.Tags
and returns desired.DeepCopy() so nothing is written back and no error is raised; hooks.go:126-131 renameTable
never sets NewNamespaceName; apis/v1alpha1/table.go:52 / generator.yaml:34-36 Namespace is_immutable (CEL self ==
oldSelf). generator.yaml:103-104 Table terminal_codes = [BadRequestException] ->
pkg/resource/table/sdk.go:606-611 ConflictException is not terminal; a same-name RenameTable cannot be issued
because delta.go:139-145 fires only when desired.Name != latest.Name and hooks.go:126-134 sends Name=latest,
NewName=desired with latest.Status.VersionToken (re-read each reconcile at sdk.go:145-149). Delete and
metadata-location updates still target the right table because they use the observed spec (runtime
reconciler.go:1453-1473 passes observed to Delete; hooks.go:159-166). finding: [S3T-TABLE-006](../table-metadata-location.md#s3t-table-006) (RenameTable
newNamespaceName moves the table, ARN unchanged, GetTable(tableArn) reports the new namespace), [S3T-TABLE-070](../table.md#s3t-table-070),
[S3T-TABLE-048](../table.md#s3t-table-048) (collision 409 permanent), [S3T-TABLE-050](../table.md#s3t-table-050) (same-name 409), [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193) (concurrent-modification
409 transient), [S3T-TABLE-112](../table-metadata-location.md#s3t-table-112) (stale versionToken 409 transient).
