<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-186: 10 concurrent TagResource on one table -> 8-9/10 InternalServerErrorException 500 (no 409; 3 rounds), 3-6 s wall; concurrent Untag 9/10 500
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-186"></a>**S3T-TABLE-186** `error-code` · impact high · handled · verified 2026-10-13
  **10 concurrent TagResource on one table -> 8-9/10 InternalServerErrorException 500 (no 409; 3 rounds), 3-6 s wall; concurrent Untag 9/10 500**
  Three rounds of 10 concurrent TagResource (distinct keys) on one table: per-round codes [{'200': 2,
  'InternalServerErrorException/500': 8}, {'200': 1, 'InternalServerErrorException/500': 9}, {'200': 2,
  'InternalServerErrorException/500': 8}] (wall [6.18, 3.15, 3.16] s); ListTagsForResource afterwards showed
  [2, 1, 2] keys present per round, of which [[], [], []] were applied by calls that returned an error (0) and
  [[], [], []] successful calls lost their key. Error messages: {"InternalServerErrorException": "An internal
  error occurred. Try again."}. 10 concurrent UntagResource (one existing key each) -> {"204": 1,
  "InternalServerErrorException/500": 9} (wall 4.14s); 9 keys remained, [] removed despite an error, [] still
  present despite a 2xx. 30 sequential UntagResource with no pause -> {"204": 30} in 3.08s (first error at
  index None), 0 of the 30 keys left.
  - ACK: tags.custom-sync, requeue, one-per-reconcile · ops: TagResource, UntagResource, ListTagsForResource ·
    fields: tags, tagKeys
  - repro: 10 threads x TagResource(tableArn, {k_i: v}) x3; TagResource(10 keys) then 10 threads x
    UntagResource([k_i]); TagResource(30 keys) then 30 sequential UntagResource
  - measurements: tag_race_wall_s=[6.18, 3.15, 3.16], untag_race_wall_s=4.14, untag_sequential_wall_s=3.08,
    untag_sequential_latency_ms_min_max=[88, 133], n=70
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-188](../table.md#s3t-table-188), [S3T-TABLE-218](../table.md#s3t-table-218), [S3T-TABLE-217](../table.md#s3t-table-217) · hypotheses: H-T-098, H-T-209 · evidence:
    table/limits/tag-count-and-read-bursts

## Notes

H-T-098 refuted on the code: concurrent tag writes to one table fail with InternalServerErrorException 500 'An
internal error occurred. Try again.' (8-9 of 10 in each of 3 rounds), never ConflictException; in these 3
rounds the losing keys were NOT applied and no successful key was lost (on a table BUCKET the same experiment
did apply some keys of 500-failed calls - see the TableBucket finding). Each 10-call burst took 3.1-6.2 s wall
with individual latencies up to 6.2 s, i.e. the calls queue server-side before most of them fail. H-T-209
refuted on the code: 10 concurrent UntagResource -> 1 x 204, 9 x 500, the 9 keys remained; 30 sequential
UntagResource with no pause -> 30 x 204 in 3.08 s (88-133 ms each), no 429. 2-3 other probes running.

Full original title: 10 concurrent TagResource on one table -> 8-9/10 InternalServerErrorException 500 (never
409), burst serialized to 3-6 s; 10 concurrent UntagResource -> 9/10 500
