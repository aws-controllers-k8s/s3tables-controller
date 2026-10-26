<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-124: TagResource/UntagResource bump GetTable.modifiedAt (not versionToken); the pre-tag versionToken still deletes the table
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-124"></a>**S3T-TABLE-124** `tag-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **TagResource/UntagResource bump GetTable.modifiedAt (not versionToken); the pre-tag versionToken still deletes the table**
  GetTable before, after TagResource({z:1}) (-> HTTP 200) and after UntagResource(['z']) (-> HTTP 204):
  modifiedAt CHANGED 2026-10-13 22:37:45.178044+00:00 -> 2026-10-13 22:37:45.660173+00:00 -> 2026-10-13
  22:37:45.807230+00:00, versionToken identical, modifiedBy identical (before: modifiedAt=2026-10-13
  22:37:45.178044+00:00 versionToken=<20 chars>).
  - ACK: compare.is_ignored+delta_pre_compare, tags.custom-sync · ops: TagResource, UntagResource, GetTable ·
    fields: modifiedAt, versionToken, modifiedBy, tags
  - repro: GetTable; TagResource({z:1}); GetTable; UntagResource(['z']); GetTable; compare
    modifiedAt/versionToken (DeleteTable with the original token is at the end of this probe)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-007](../table-metadata-location.md#s3t-table-007), [S3T-TABLE-101](../table-metadata-location.md#s3t-table-101), [S3T-TABLE-104](../table-metadata-location.md#s3t-table-104), [S3T-TABLE-316](../table-metadata-location.md#s3t-table-316), [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320), [S3T-TABLE-019](../table-metadata-location.md#s3t-table-019),
    [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193) · hypotheses: H-T-096 · evidence: table/idempotency/untag-and-post-delete

## Notes

H-T-096 qualified: the contrarian claim holds for modifiedAt (advanced by ~0.5 s after TagResource and again
after UntagResource, 1 trial each) but not for versionToken (byte-identical, 20 chars) nor modifiedBy;
DeleteTable guarded with the CreateTable-time versionToken after these tag writes -> HTTP 204 (see the
post-delete finding of this probe). A controller using modifiedAt as an external-drift signal sees its own tag
writes as changes; versionToken-guarded calls are unaffected.
