<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-028: warehouseLocation host = first 3 uuid groups (18-19 chars) of the table uuid + opaque suffix + '--table-s3'; no names in it; stable, unique
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-028"></a>**S3T-TABLE-028** `identity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **warehouseLocation host = first 3 uuid groups (18-19 chars) of the table uuid + opaque suffix + '--table-s3'; no names in it; stable, unique**
  GetTable warehouseLocation for t1 is 's3://ef44dd0a-f123-4ce4-nuqppejam7deeueahexkquthetdnqusw2b--table-s3'
  (host length 63): it shares only its first 19 characters with the table uuid from the ARN (full uuid prefix:
  False), ends with '--table-s3': True, contains bucket name: False, namespace: False, table name: False; t2
  likewise shares 19 chars with its uuid. Stable across two reads: t1=True t2=True; different between t1 and
  t2: True.
  - ACK: is_read_only · ops: GetTable · fields: warehouseLocation
  - repro: Create t1,t2; GetTable twice each; inspect warehouseLocation host
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-018](../table-schema.md#s3t-table-018), [S3T-TABLE-310](../table-schema.md#s3t-table-310), [S3T-TABLE-313](../table-schema.md#s3t-table-313), [S3T-TABLE-090](../table-metadata-location.md#s3t-table-090), [S3T-TABLE-023](../table-metadata-location.md#s3t-table-023), [S3T-TABLE-113](../table.md#s3t-table-113),
    [S3T-TABLE-067](../table-metadata-location.md#s3t-table-067), [S3T-TABLE-122](../table-metadata-location.md#s3t-table-122), [S3T-TABLE-129](../table-metadata-location.md#s3t-table-129), [S3T-TABLE-064](../table-metadata-location.md#s3t-table-064), [S3T-TABLE-089](../table-metadata-location.md#s3t-table-089) · hypotheses: H-T-031 ·
    evidence: table/identity/get-table-identifier-forms

## Notes

H-T-031 qualified: the host is NOT the full table uuid (only the first three uuid groups, an 18-19 char shared
prefix across two runs), the rest is opaque; the other claims (no names in it, --table-s3 suffix, unique,
stable) hold. The controller must copy warehouseLocation from GetTable; it cannot be constructed from spec
fields or from the ARN.
