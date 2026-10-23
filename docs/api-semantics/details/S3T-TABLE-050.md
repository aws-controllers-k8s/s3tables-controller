<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-050: RenameTable to the CURRENT name/namespace is not a no-op: ConflictException 409 'identical name already exists'; neither field -> 400
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-050"></a>**S3T-TABLE-050** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **RenameTable to the CURRENT name/namespace is not a no-op: ConflictException 409 'identical name already exists'; neither field -> 400**
  On ns1/a: RenameTable(newName=a) -> ConflictException (HTTP 409) 'A table with an identical name already
  exists in the namespace.' (modifiedAt changed: False); RenameTable(newNamespaceName=ns1) ->
  ConflictException (HTTP 409) 'A table with an identical name already exists in the namespace.' (modifiedAt
  changed: False); RenameTable(newName=a, newNamespaceName=ns1) -> ConflictException (HTTP 409) 'A table with
  an identical name already exists in the namespace.' (modifiedAt changed: False); RenameTable with neither
  newName nor newNamespaceName -> BadRequestException (HTTP 400) 'Neither a new namespace name nor a new table
  name is specified.'.
  - ACK: custom_update, compare.is_ignored+delta_pre_compare · ops: RenameTable · fields: newName,
    newNamespaceName
  - repro: RenameTable(ns1,a,newName=a); RenameTable(ns1,a,newNamespaceName=ns1);
    RenameTable(ns1,a,newName=a,newNamespaceName=ns1); RenameTable(ns1,a)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-271](../table.md#s3t-table-271), [S3T-TABLE-272](../table.md#s3t-table-272), [S3T-TABLE-273](../table.md#s3t-table-273), [S3T-TABLE-274](../table.md#s3t-table-274), [S3T-TABLE-275](../table-metadata-location.md#s3t-table-275), [S3T-TABLE-276](../table.md#s3t-table-276),
    [S3T-NAMESPACE-036](../namespace.md#s3t-namespace-036), [S3T-NAMESPACE-021](../namespace.md#s3t-namespace-021), [S3T-TABLE-064](../table-metadata-location.md#s3t-table-064), [S3T-TABLE-048](../table.md#s3t-table-048), [S3T-TABLE-049](../table.md#s3t-table-049) · hypotheses: H-T-059,
    H-T-060 · evidence: table/error-taxonomy/rename-errors

## Notes

H-T-059 qualified: the contrarian claim 'same-name rename is rejected, not a harmless no-op' holds, but the
code is ConflictException (HTTP 409) 'A table with an identical name already exists in the namespace.' rather
than BadRequestException 400; the same 409 for newNamespaceName=<current ns> alone and for both. modifiedAt
did not change in any of the three calls. H-T-060 first half confirmed: neither field -> BadRequestException
400 'Neither a new namespace name nor a new table name is specified.'
