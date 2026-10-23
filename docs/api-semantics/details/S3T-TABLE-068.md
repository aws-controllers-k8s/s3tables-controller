<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-068: RenameTable visibility at 150 ms polling: both-200 overlap never, both-404 gap never, ARN read always 200 across 5 renames
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-068"></a>**S3T-TABLE-068** `eventual-consistency` · impact high · handled · verified 2026-10-13
  **RenameTable visibility at 150 ms polling: both-200 overlap never, both-404 gap never, ARN read always 200 across 5 renames**
  Three pollers (GetTable old triple, new triple, tableArn) at 150 ms around RenameTable(newName) in 5 trials
  (27 samples/trial): RenameTable -> HTTP 204 in [153, 146, 178, 147, 150] ms. First new-triple 200 completed
  [-0.04, -0.04, -0.07, -0.04, -0.05] s relative to the rename call returning (negative = before); first
  old-triple 404 completed [-0.05, -0.04, -0.07, -0.04, -0.05] s relative to the return. Samples where both
  triples returned 200: none; where both returned 404: none. The tableArn read returned ['200'] throughout and
  the name it reported was one of old/new (names seen per trial: [2, 2, 2, 2, 2]).
  - ACK: requeue, is_arn_primary_key · ops: RenameTable, GetTable · fields: name, tableArn
  - repro: Start 3 GetTable pollers at 150 ms (old triple, new triple, tableArn); RenameTable(newName); stop 1
    s later; repeat 5x alternating names
  - measurements: rename_latency_ms=[153, 146, 178, 147, 150], new_visible_rel_rename_end_s=[-0.045, -0.042,
    -0.071, -0.042, -0.046], old_gone_rel_rename_end_s=[-0.047, -0.041, -0.071, -0.041, -0.046],
    both_200_overlap_s=[0, 0, 0, 0, 0], both_404_gap_s=[0, 0, 0, 0, 0], poll_interval_s=0.15, n=5
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122`
  - related: [S3T-TABLE-069](../table.md#s3t-table-069), [S3T-TABLE-070](../table.md#s3t-table-070), [S3T-TABLE-071](../table.md#s3t-table-071), [S3T-TABLE-006](../table-metadata-location.md#s3t-table-006) · hypotheses: H-T-051, H-T-198,
    H-T-199 · evidence: table/consistency-windows/rename-visibility

## Notes

H-T-051 (contrarian: old triple still 200 / new triple 404 after the call returns) refuted at this resolution.
H-T-198 (sync, atomic, no transitional view) confirmed at this resolution. H-T-199 (contrarian: both-404
window) refuted at this resolution. Resolution is the 150 ms poll interval; windows shorter than that are not
excluded.
