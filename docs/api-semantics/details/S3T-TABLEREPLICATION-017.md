<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-017: Switch [D]->[E] (same region, empty table, n=1): Get reflects it at once; status showed E at +1.06s, dropped D and E 'completed' at +2.1s
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-017"></a>**S3T-TABLEREPLICATION-017** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Switch [D]->[E] (same region, empty table, n=1): Get reflects it at once; status showed E at +1.06s, dropped D and E 'completed' at +2.1s**
  With D's entry D:completed (63 s after the first Put), PutTableReplication [E] with the current token ->
  HTTP 200; GetTableReplication 181 ms later showed ['E']. GetTableReplicationStatus polled at 1 s:
  transitions [(0.02, 'D:completed'), (1.06, 'D:completed|E:pending'), (2.1, 'E:completed')] (D entry gone at
  +2.1s, E first seen at +1.06s, E completed at +2.1s; final view 'E:completed').
  - ACK: synced.when, custom_update · ops: PutTableReplication, GetTableReplication, GetTableReplicationStatus
    · fields: destinations[].replicationStatus, destinations[].destinationTableBucketArn
  - repro: config [D] completed -> Put [E] (token) -> Get -> poll Status at 1 s for <= 7 min
  - measurements: status_drops_old_dest_s=[2.1], status_shows_new_dest_s=[1.06], new_dest_completed_s=[2.1],
    n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-033](../table-replication-status.md#s3t-tablereplication-033), [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005), [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019),
    [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028), [S3T-TABLEREPLICATION-032](../table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEBUCKETREPLICATION-022](../tablebucket-replication.md#s3t-tablebucketreplication-022),
    [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009), [S3T-TABLEBUCKETREPLICATION-011](../tablebucket-replication.md#s3t-tablebucketreplication-011),
    [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022), [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047) ·
    hypotheses: H-U-160 · evidence: table/dependencies/repl-prereqs-token-matrix

## Notes

H-U-160 second half confirmed (hypothesis: old set visible up to 60 s). Empty table, same-region new
destination; n=1.

Contradiction with [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042): time until a newly configured destination appears in
GetTableReplicationStatus after a Put on an existing completed table-level config: 017 (replace [D]->[E], same
region, empty table) saw E:pending at +1.06 s and E:completed at +2.1 s; 042 (replace [dead D]->[D2]) saw
D2:pending at +60.68 s and completed at +65.73 s, and 021 (+61.04 s) / 022 (+61.17 s) / 031 ('previous entry
shown for ~60 s after any Put') agree with 042; all n=1 - 017's Put landed 63 s after the first Put, i.e.
possibly right at a ~60 s evaluation tick Resolution: re-probe: state-machine - completed table-level config
on an empty table, Put [E] replacing [D] issued at varied offsets after the last status change (5/20/40/55/70
s, n>=5), poll GetTableReplicationStatus at 1 s, record first E entry and E:completed; test whether the ~60 s
delay is a periodic evaluation tick rather than a fixed latency

Contradiction with [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047), [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021),
[S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022): FU2-04 settled by follow-up probe
tablereplication/state-machine/fu2-04-put-offset-tick: 042/021/022 hold; 017's +1.06 s did not reproduce. The
delay is a fixed ~61 s per Put, not a periodic tick: Puts issued 0.9, 5, 20, 40, 55 and 70 s after the last
observed status change (67-565 s after the first Put) all showed the new destination as 'pending' 61.0-64.1 s
later and 'completed' 1.0-1.1 s after that (empty table, same-region buckets); the first Put behaved
identically (pendi Resolution: keep all; [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047) is canonical for the settled rule

Consolidation note: Follow-up re-probe outcome: the +1 s appearance of a replacement destination did not
reproduce (6/6 at +61-64 s regardless of Put offset); [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047) settles it Resolution: keep as
evidence; superseded reading - [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047) settles it
