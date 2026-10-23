<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-178: Policy/maintenance ops on a missing bucket, namespace or table: NotFoundException/404 sharing 'The specified table does not exist.'
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-178"></a>**S3T-TABLE-178** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Policy/maintenance ops on a missing bucket, namespace or table: NotFoundException/404 sharing 'The specified table does not exist.'**
  Six operations (GetTablePolicy, DeleteTablePolicy, GetTableMaintenanceConfiguration,
  GetTableMaintenanceJobStatus, PutTablePolicy, PutTableMaintenanceConfiguration) against a nonexistent bucket
  (well-formed ARN, own account), a nonexistent namespace and a nonexistent table all return NotFoundException
  (HTTP 404) with the identical message 'The specified table does not exist.' - the missing level is not
  distinguishable. The only different 404 message is the policy-absent case on an existing table: 'The
  specified table policy does not exist.' (an upper-cased namespace additionally yields 'The specified
  namespace does not exist.').
  - ACK: exceptions.404, custom_find · ops: GetTablePolicy, DeleteTablePolicy,
    GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus, PutTablePolicy,
    PutTableMaintenanceConfiguration
  - repro: each op with a bad bucket ARN / bad namespace / bad table name; GetTablePolicy on a real
    policy-less table
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-179](../table-maintenance.md#s3t-table-179), [S3T-TABLE-180](../table-maintenance.md#s3t-table-180), [S3T-TABLE-182](../table-maintenance.md#s3t-table-182), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183) · hypotheses: H-U-019 · evidence:
    table/identity/subresource-identifiers

## Notes

H-U-019 qualified: code and HTTP status are uniform (404 NotFoundException) as hypothesized, but there are
only two messages, not four: bucket/namespace/table missing share 'The specified table does not exist.'; only
the policy-absent case differs.

Full original title: Policy/maintenance ops on a missing bucket, namespace or table: all NotFoundException/404
with ONE message 'The specified table does not exist.'
