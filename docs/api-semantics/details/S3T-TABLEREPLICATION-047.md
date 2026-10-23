<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-047: Replacing the destination of a completed table-level config: new entry 'pending' +61-64 s after EVERY Put, completed ~1 s later (n=6)
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-047"></a>**S3T-TABLEREPLICATION-047** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **Replacing the destination of a completed table-level config: new entry 'pending' +61-64 s after EVERY Put, completed ~1 s later (n=6)**
  Empty ICEBERG table, same-region destination buckets, GetTableReplicationStatus polled at 1 s. First Put
  [D0]: D0 entry at +61.06 s, completed at +62.1 s. Replacement Puts (current versionToken, HTTP 200 each):
  Put at +5 s after last change (67 s after first Put): D1 entry at +61.2 s (pending), completed at 62.24 s,
  old gone at 62.24 s; Put at +20 s after last change (149 s after first Put): D2 entry at +61.0 s (pending),
  completed at 62.04 s, old gone at 62.04 s; Put at +40 s after last change (251 s after first Put): D3 entry
  at +61.9 s (pending), completed at 62.94 s, old gone at 61.91 s; Put at +55 s after last change (369 s after
  first Put): D4 entry at +61.9 s (pending), completed at 62.96 s, old gone at 61.93 s; Put at +70 s after
  last change (502 s after first Put): D5 entry at +61.1 s (pending), completed at 62.15 s, old gone at 61.12
  s; Put at +1 s after last change (565 s after first Put): D6 entry at +64.1 s (pending), completed at 65.15
  s, old gone at 65.15 s. The old entry disappears when the new one reaches 'completed' in 3/6 cells.
  - ACK: synced.when, requeue, e2e-timing · ops: PutTableReplication, GetTableReplicationStatus,
    GetTableReplication · fields: destinations[].replicationStatus, destinations[].destinationTableBucketArn,
    configuration.rules[].destinations[]
  - repro: PutTableReplication [D0] on an empty table -> poll status 1 s to completed -> at +5/20/40/55/70/1 s
    after the last status change PutTableReplication [Dk] (fresh bucket, current token) -> poll status 1 s;
    record first Dk entry, Dk completed, old entry gone
  - measurements: first_put_entry_s=61.06, first_put_completed_s=62.1, offsets_after_last_change_s=[4.9,
    19.97, 40.12, 55.08, 70.08, 0.87], new_entry_after_put_s=[61.21, 61.0, 61.91, 61.93, 61.12, 64.12],
    completed_after_put_s=[62.24, 62.04, 62.94, 62.96, 62.15, 65.15], new_entry_since_first_put_s=[128.21,
    210.21, 313.28, 431.32, 563.56, 629.58], n=6
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021),
    [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022) · evidence: tablereplication/state-machine/fu2-04-put-offset-tick

## Notes

FU2-04: 042/021/022 hold; 017's +1.06 s did not reproduce. The delay is a fixed ~61 s per Put, not a periodic
tick: Puts issued 0.9, 5, 20, 40, 55 and 70 s after the last observed status change (67-565 s after the first
Put) all showed the new destination as 'pending' 61.0-64.1 s later and 'completed' 1.0-1.1 s after that (empty
table, same-region buckets); the first Put behaved identically (pending at +61.1 s, completed at +62.1 s). The
positions of the new entries within a 60 s cycle anchored to the first Put were scattered (not clustered),
ruling out a global tick. The replaced destination's entry disappears at the same poll the new entry appears
(3/6) or one poll later when it completes (3/6). A controller should treat ~60 s of 'previous view' after any
PutTableReplication as normal and not as a failed update; the 0.87 s offset cell (closest to 017's situation)
also took 64 s.

Contradiction with [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021),
[S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022): FU2-04 settled by follow-up probe
tablereplication/state-machine/fu2-04-put-offset-tick: 042/021/022 hold; 017's +1.06 s did not reproduce. The
delay is a fixed ~61 s per Put, not a periodic tick: Puts issued 0.9, 5, 20, 40, 55 and 70 s after the last
observed status change (67-565 s after the first Put) all showed the new destination as 'pending' 61.0-64.1 s
later and 'completed' 1.0-1.1 s after that (empty table, same-region buckets); the first Put behaved
identically (pendi Resolution: keep all; [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047) is canonical for the settled rule
