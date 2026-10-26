<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-115: CreateTableBucket with a just-deleted name -> ConflictException 409 'transitional state ... previous deletion attempt' ~10 min (OK +606 s)
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-115"></a>**S3T-TABLEBUCKET-115** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTableBucket with a just-deleted name -> ConflictException 409 'transitional state ... previous deletion attempt' ~10 min (OK +606 s)**
  After DeleteTableBucket returned HTTP 204 (GetTableBucket -> NotFoundException on the first poll),
  CreateTableBucket with the same name returned ConflictException (HTTP 409) 'The bucket is in a transitional
  state because of a previous deletion attempt. Try again later.' on all 44 attempts between +0.03 s and
  +594.6 s after the delete (every 5 s for the first minute, then every 10 s; a gap of ~4 min while the probe
  was restarted) and succeeded (HTTP 200) at +606.2 s. The name stays reserved in a server-side 'transitional
  state' that is invisible to GetTableBucket (404) and to the sub-resource Gets (404 'The specified bucket
  does not exist.').
  - ACK: requeue, deletable.when · ops: DeleteTableBucket, CreateTableBucket, GetTableBucket · fields: name
  - repro: DeleteTableBucket -> GetTableBucket until 404 -> CreateTableBucket(same name) every 5-10 s until it
    succeeds
  - measurements: recreate_409_last_seen_s=594.6, recreate_success_s=606.2, attempts=45, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-016](../tablebucket.md#s3t-tablebucket-016), [S3T-TABLEBUCKET-014](../tablebucket.md#s3t-tablebucket-014), [S3T-TABLEBUCKET-015](../tablebucket.md#s3t-tablebucket-015), [S3T-TABLEBUCKET-159](../tablebucket.md#s3t-tablebucket-159),
    [S3T-TABLEBUCKET-127](../tablebucket.md#s3t-tablebucket-127), [S3T-TABLEBUCKET-017](../tablebucket.md#s3t-tablebucket-017), [S3T-TABLEBUCKET-070](../tablebucket.md#s3t-tablebucket-070), [S3T-TABLEBUCKET-013](../tablebucket-encryption-storage.md#s3t-tablebucket-013),
    [S3T-TABLEBUCKETREPLICATION-029](../tablebucket-replication.md#s3t-tablebucketreplication-029), [S3T-TABLEBUCKETREPLICATION-030](../tablebucket-replication.md#s3t-tablebucketreplication-030), [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005) · hypotheses:
    H-S-071, H-S-022, H-S-099 · evidence: tablebucket/identity/subresources-across-recreate

## Notes

Qualifies H-S-071: no DELETING state is observable on the read path, but the create path exposes one through
ConflictException 409 for ~10 minutes - a controller re-creating a deleted TableBucket (or an e2e test reusing
names) must retry this 409 for at least that long; the window may be longer for non-empty buckets (this one
had 5 sub-resources, no tables). Single run.

Full original title: CreateTableBucket with a just-deleted name -> ConflictException 409 'transitional state
because of a previous deletion attempt' for ~10 min (last 409 at +595 s, success at +606 s)
