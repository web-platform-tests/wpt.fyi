#!/usr/bin/env python3
import datetime
import logging
from processor import Processor

_log = logging.getLogger(__name__)


def cleanup_stuck_runs(days_threshold: int = 14) -> None:
    """
    Identifies PendingTestRun entities that have been in a non-terminal state
    for longer than the given threshold and marks them as INVALID.
    """
    p = Processor()
    cutoff = (datetime.datetime.now(datetime.timezone.utc)
              - datetime.timedelta(days=days_threshold))

    _log.info("Querying for stuck runs updated before %s", cutoff)

    # Query for runs with Stage < 800 (StageValid)
    q = p.datastore.query(kind='PendingTestRun')
    q.add_filter('Stage', '<', 800)

    stuck_runs = list(q.fetch())
    _log.info("Found %d potential stuck runs", len(stuck_runs))

    count = 0
    for run in stuck_runs:
        # Check if the Updated time is before our cutoff
        updated = run.get('Updated')
        if updated and updated < cutoff:
            run_id = str(run.key.id)
            _log.info("Marking run %s as INVALID (last updated %s)",
                      run_id, updated)
            try:
                # Use the update_status method from Processor which handles
                # the API call.
                p.update_status(
                    run_id, 'INVALID',
                    error='Run timed out after {} days'.format(days_threshold))
                count += 1
            except Exception as e:
                _log.exception("Failed to update run %s: %s", run_id, str(e))

    _log.info("Successfully updated %d stuck runs", count)


if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO)
    cleanup_stuck_runs()
