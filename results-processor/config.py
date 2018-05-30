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


def project_baseurl():
    """Returns the base URL of the current project."""
    # Defaults to staging to prevent accidental access of prod.
    # TODO(Hexcles): Support local dev_appserver.
    project = os.getenv('GOOGLE_CLOUD_PROJECT') or 'wptdashboard-staging'
    return 'https://%s.appspot.com' % project
