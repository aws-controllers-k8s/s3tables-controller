<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-048: RenameTable to a name that already exists in the target namespace -> ConflictException; same name in another namespace is allowed
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-048"></a>**S3T-TABLE-048** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **RenameTable to a name that already exists in the target namespace -> ConflictException; same name in another namespace is allowed**
  RenameTable(ns1,a, newName=b) with ns1/b existing -> ConflictException (HTTP 409) 'A table with an identical
  name already exists in the namespace.'. RenameTable(ns1,a, newNamespaceName=ns2, newName=b) with ns2/b
  existing -> ConflictException (HTTP 409) 'A table with an identical name already exists in the namespace.'.
  GetTable(ns1,a) after each -> HTTP 200 / HTTP 200 (source unchanged: True). RenameTable(ns1,a, newName=zz)
  where zz exists only in ns2 -> HTTP 204 (names are scoped per namespace); renamed back -> HTTP 204.
  - ACK: terminal_codes, custom_update · ops: RenameTable · fields: newName, newNamespaceName
  - repro: Create ns1/a, ns1/b, ns2/b, ns2/zz -> RenameTable(ns1,a,newName=b);
    RenameTable(ns1,a,newNamespaceName=ns2,newName=b); RenameTable(ns1,a,newName=zz)
  - measurements: rename_collision_latency_ms=[55, 11], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-050](../table.md#s3t-table-050), [S3T-TABLE-049](../table.md#s3t-table-049) · hypotheses: H-T-056 · evidence: table/error-taxonomy/rename-errors

## Notes

H-T-056 confirmed (predicted ConflictException 409 for both collisions, success for a name existing only in
another namespace). The 409 message is specific ('A table with an identical name already exists in the
namespace.'), not the generic 'retry the request' shape text; the collision is permanent for this input, so a
retry loop on 409 would never converge.
