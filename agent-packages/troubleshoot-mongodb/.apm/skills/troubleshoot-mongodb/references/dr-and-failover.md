# DR: Switchover/Failover Failed

## Description
A switchover or failover operation between the active and standby (DR) sites fails to complete.

## How to solve

1. **Retry the switchover/failover operation:**
   - On the site that failed, edit the ConfigMap `last-applied-configuration-info` — delete the value of the key `summary-spec` and save.
   - Restart the operator pod on the site that failed and wait for that site's status to reach the expected state.
   - Retry the switchover/failover operation.

## Notes
- This clears stale reconciliation state the operator was holding for that site, which is what typically blocks the retry from proceeding.
- If the retry fails again, check the operator logs on the failing site for the underlying reconciliation error before repeating this step — clearing `summary-spec` repeatedly without addressing a root cause (e.g. a replica not reaching PRIMARY/SECONDARY, see `replication-and-consistency.md`) will just fail the same way again.
