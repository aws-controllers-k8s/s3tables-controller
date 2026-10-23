<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-159: CreateTableBucket ConflictException has two meanings (exists vs post-delete transitional); only the message and a Get tell them apart
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-159"></a>**S3T-TABLEBUCKET-159** `error-code` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-15, re-verified
  **CreateTableBucket ConflictException has two meanings (exists vs post-delete transitional); only the message and a Get tell them apart**
  CreateTableBucket for an existing name -> ConflictException HTTP 409 'The bucket that you tried to create
  already exists, and you own it.'. CreateTableBucket for a name deleted < ~10 min ago -> ConflictException
  HTTP 409 ["The bucket is in a transitional state because of a previous deletion attempt. Try again later."],
  while in that window GetTableBucket -> NotFoundException; DeleteTableBucket -> NotFoundException;
  TagResource -> NotFoundException; ListTableBuckets(prefix) -> 0 item(s), and the same name in the other
  region -> HTTP 200 (created, distinct ARN region) at +2.54 s after the delete. Same code, same HTTP status;
  the stored resource exists only in the first case.
  - ACK: terminal_codes, requeue, custom_create · ops: CreateTableBucket, GetTableBucket, DeleteTableBucket,
    TagResource, ListTableBuckets · fields: name
  - repro: CreateTableBucket(existing) ; DeleteTableBucket -> CreateTableBucket(same) at +15..+600 s ;
    Get/Delete/Tag/List on the ARN inside the window
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-013](../tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-016](../tablebucket.md#s3t-tablebucket-016), [S3T-TABLEBUCKET-014](../tablebucket.md#s3t-tablebucket-014), [S3T-TABLEBUCKET-015](../tablebucket.md#s3t-tablebucket-015),
    [S3T-TABLEBUCKET-115](../tablebucket.md#s3t-tablebucket-115), [S3T-TABLEBUCKET-127](../tablebucket.md#s3t-tablebucket-127), [S3T-TABLEBUCKET-017](../tablebucket.md#s3t-tablebucket-017), [S3T-TABLEBUCKET-070](../tablebucket.md#s3t-tablebucket-070), [S3T-TABLEBUCKET-125](../tablebucket.md#s3t-tablebucket-125) ·
    evidence: tablebucket/creative/delete-recreate-window

## Notes

extends [S3T-TABLEBUCKET-013](../tablebucket-encryption-storage.md#s3t-tablebucket-013)/-016 with the side-by-side of both 409 messages and what every other bucket op
returns inside the window. A controller mapping ConflictException -> 'already exists, adopt' would loop on
GetTableBucket 404; mapping it -> terminal would never create. Rule: 409 + Get 200 => exists; 409 + Get 404 =>
transitional, requeue. Messages observed in one run; cross-region create n=1.

Suspected controller bug confirmed by evidence: not flagged; evidence-driven: while the status ARN is unset
the read path reports NotFound and the reconciler creates, but CreateTableBucket for a name that already
exists is ConflictException 409 'already exists, and you own it' (-017, -159) and for a name deleted <10 min
ago 409 'transitional state' (-115), both recoverable, so a CR for a pre-existing bucket (or one re-created
after a controller restart before the status patch, or deleted and re-created within 10 min) requeues on
Create indefinitely (or ~10 min) and never reaches Get; derive the deterministic ARN
arn:<partition>:s3tables:<region>:<account>:bucket/<name> (-003) in the read path, or on this 409 Get the
derived ARN and adopt on 200 / requeue on 404.
