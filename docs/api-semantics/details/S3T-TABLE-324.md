<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-324: UntagResource on any nonexistent table ARN (deleted or never-created) -> InternalServerErrorException 500 for 12 min; Tag/ListTags -> 404
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-324"></a>**S3T-TABLE-324** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **UntagResource on any nonexistent table ARN (deleted or never-created) -> InternalServerErrorException 500 for 12 min; Tag/ListTags -> 404**
  Three clean runs. Two tables per run created with tags {a:1}, DeleteTable -> HTTP 204; UntagResource(old
  tableARN, ['a']) at t+0 and t+5 s -> InternalServerErrorException (HTTP 500) 'An internal error occurred.
  Try again.' in 12/12 (x-amzn-ErrorType
  'InternalServerErrorException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/', raw body
  {"message":"An internal error occurred. Try again."}, 314-388 ms), while TagResource and ListTagsForResource
  on the same ARN -> NotFoundException (HTTP 404) 'The specified table does not exist.' in 24/24. The same 500
  came back for UntagResource with a key the table never had (['zzz'], 3/3), for a well-formed table ARN whose
  uuid never existed in the live bucket (3/3; TagResource/ListTagsForResource on it -> 404), 180 s after the
  delete (2/2; +36 s in run 1), and for ARNs deleted 8.3 and 11.7 min earlier whose table BUCKET had been
  deleted too (2/2, ListTagsForResource -> 404 'The specified table does not exist.'). In total 23/23
  UntagResource calls on a nonexistent table id -> 500; 35/35 TagResource/ListTagsForResource -> 404.
  - ACK: exceptions.404, terminal_codes, tags.custom-sync, pre-delete-cleanup · ops: UntagResource,
    TagResource, ListTagsForResource, DeleteTable · fields: resourceArn, tagKeys
  - repro: CreateTable(tags {a:1}) x2 -> DeleteTable -> UntagResource/TagResource/ListTagsForResource on the
    old ARN at t+0, t+5 s and minutes later; UntagResource on <bucketARN>/table/<random uuid>
  - measurements: untag_500_latency_ms.min=314, untag_500_latency_ms.max=388, untag_500_latency_ms.n=23,
    late_samples_after_delete_s=[36, 181, 181], bucket_deleted_samples_age_min=[11.7, 8.3], n=23
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-130](../table.md#s3t-table-130) · evidence: table/creative/reverify-set-c

## Notes

Precision gain on [S3T-TABLE-130](../table.md#s3t-table-130) (which reproduced: 500 at t+0 and t+5 s, Tag/ListTags 404). The 500 is not a
transient post-delete artefact: it is the UntagResource not-found path itself (any unknown table uuid, any
key, minutes later, even after the bucket is gone), so a controller must treat UntagResource 500 on a table it
just failed to find as terminal not-found, not as a retryable server error.
