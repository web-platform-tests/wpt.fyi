import os


def _is_prod():
    return os.getenv('GOOGLE_CLOUD_PROJECT') == 'wptdashboard'


def raw_results_bucket():
    """Returns the bucket name for storing raw, full results."""
    if _is_prod():
        return 'wptd-results'
    return 'wptd-results-staging'


def results_bucket():
    """Returns the bucket name for storing split results."""
    if _is_prod():
        return 'wptd'
    return 'wptd-staging'
