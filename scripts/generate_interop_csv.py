# Copyright 2026 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Generate all-features interop score CSV from wpt.fyi data.

Fetches aligned test runs, downloads result summaries and the web-features
manifest, then computes per-feature pass rates per browser. Outputs a CSV
suitable for the wpt-interop-graph frontend component.

Usage:
    python scripts/generate_interop_csv.py --channel experimental --max-dates 5
    python scripts/generate_interop_csv.py --manifest path/to/manifest.json --max-dates 10
"""

import argparse
import csv
import gzip
import json
import sys
from collections import defaultdict
from datetime import date as date_type
from urllib.request import urlopen, Request
from urllib.error import HTTPError


WPT_FYI_HOST = 'https://wpt.fyi'
BROWSERS = ['chrome', 'firefox', 'safari']
DEFAULT_OUTPUT = 'webapp/static/all-features-interop-{channel}.csv'
DEFAULT_DETAIL_OUTPUT = 'webapp/static/all-features-detail-{channel}.csv'


def extract_pass_total(summary_value):
    """Extract (passes, total) from a summary value.

    Supports both historical list values and object values like:
    {'s': 'F', 'c': [0, 0]}.

    For tests with subtests, 'c' contains [passes, total] of subtests.
    For tests without subtests, 'c' is [0, 0] and we use the status 's'
    to determine pass (1,1) or fail (0,1).
    """
    if isinstance(summary_value, list) and len(summary_value) >= 2:
        return summary_value[0], summary_value[1]

    if isinstance(summary_value, dict):
        counts = summary_value.get('c')
        if isinstance(counts, list) and len(counts) >= 2:
            passes, total = counts[0], counts[1]
            if total > 0:
                return passes, total
        # No subtests; use the top-level status as a single test result.
        status = summary_value.get('s', '')
        if status in ('O', 'P'):
            return 1, 1
        elif status:
            return 0, 1

    return None, None


def fetch_json(url):
    req = Request(url, headers={
        'Accept': 'application/json',
        'User-Agent': 'wpt-fyi-interop-csv-generator',
    })
    with urlopen(req, timeout=30) as resp:
        return json.loads(resp.read())


def fetch_gzipped_json(url):
    req = Request(url, headers={
        'User-Agent': 'wpt-fyi-interop-csv-generator',
    })
    with urlopen(req, timeout=60) as resp:
        data = resp.read()
    # Try gzip first, then bzip2, then raw JSON.
    try:
        decompressed = gzip.decompress(data)
        return json.loads(decompressed)
    except (gzip.BadGzipFile, OSError):
        pass
    try:
        import bz2
        decompressed = bz2.decompress(data)
        return json.loads(decompressed)
    except Exception:
        pass
    return json.loads(data)


def fetch_aligned_runs(channel, max_count):
    """Fetch aligned runs for all browsers from wpt.fyi API."""
    params = '&'.join([
        'label=master',
        f'label={channel}',
        'aligned',
        f'max-count={max_count}',
    ] + [f'product={b}' for b in BROWSERS])
    url = f'{WPT_FYI_HOST}/api/runs?{params}'
    return fetch_json(url)


def group_runs_by_date(runs):
    """Group runs by date, keeping only dates with all browsers present."""
    groups = defaultdict(list)
    for run in runs:
        date = run['time_start'][:10]
        groups[date].append(run)
    complete = {}
    for date, date_runs in sorted(groups.items()):
        browsers_present = {r['browser_name'] for r in date_runs}
        if all(b in browsers_present for b in BROWSERS):
            complete[date] = date_runs
    return complete


def fetch_web_features_manifest():
    """Fetch WEB_FEATURES_MANIFEST.json.gz from the latest wpt release.

    The manifest is published at github.com/web-platform-tests/wpt/releases
    as WEB_FEATURES_MANIFEST.json.gz.
    """
    # Matches shared/web_features_manifest_github_download.go.
    owner = 'web-platform-tests'
    repo = 'wpt'
    url = f'https://api.github.com/repos/{owner}/{repo}/releases/latest'
    release = fetch_json(url)
    for asset in release.get('assets', []):
        label = (asset.get('label') or asset.get('name', '')).lower()
        if 'web_features_manifest' in label and label.endswith('.json.gz'):
            download_url = asset.get('browser_download_url', '')
            if download_url:
                return fetch_gzipped_json(download_url)
    return None


def build_feature_to_tests(manifest_data):
    """Convert manifest data to {feature: [test_paths]} dict."""
    if isinstance(manifest_data, dict):
        if 'data' in manifest_data and 'version' in manifest_data:
            return manifest_data['data']
        return manifest_data
    return {}


def collect_unlabeled_tests(summaries, feature_tests):
    """Return sorted test paths present in any summary but in no feature."""
    labeled = set()
    for tests in feature_tests.values():
        labeled.update(tests)
    seen = set()
    for summary in summaries.values():
        seen.update(summary.keys())
    return sorted(seen - labeled)


def with_unlabeled(feature_tests, summaries):
    """Return a copy of feature_tests with an 'unlabeled' bucket added."""
    unlabeled = collect_unlabeled_tests(summaries, feature_tests)
    return {**feature_tests, 'unlabeled': unlabeled}


def compute_per_feature_for_date(run_group, feature_tests, include_unlabeled,
                                 debug=False):
    """Compute per-feature, per-browser pass percentages for one date.

    Returns (versions, per_feature) where:
      - versions: {browser: version_string}
      - per_feature: {feature_name: {browser: pct or None}}
    Returns None if any browser's summary can't be fetched.
    """
    summaries = {}
    versions = {}
    for run in run_group:
        browser = run['browser_name']
        if browser not in BROWSERS:
            continue
        versions[browser] = run.get('browser_version', '')
        results_url = run.get('results_url', '')
        if not results_url:
            print(f'    {browser}: no results_url', file=sys.stderr)
            continue
        try:
            summaries[browser] = fetch_gzipped_json(results_url)
        except Exception as e:
            print(f'    Warning: failed to fetch {browser} summary: {e}',
                  file=sys.stderr)
            return None

    if len(summaries) < len(BROWSERS):
        missing = set(BROWSERS) - set(summaries.keys())
        print(f'    Missing browsers: {missing}', file=sys.stderr)
        return None

    if include_unlabeled:
        feature_tests = with_unlabeled(feature_tests, summaries)
        print(f'    Unlabeled bucket: '
              f'{len(feature_tests["unlabeled"])} tests',
              file=sys.stderr)

    if debug:
        for b in BROWSERS:
            print(f'    {b}: {len(summaries[b])} tests in summary',
                  file=sys.stderr)
        all_manifest_paths = set()
        for tests in feature_tests.values():
            all_manifest_paths.update(tests)
        summary_keys = set(summaries[BROWSERS[0]].keys())
        matched = all_manifest_paths & summary_keys
        print(f'    Manifest has {len(all_manifest_paths)} unique paths, '
              f'{len(matched)} match {BROWSERS[0]} summary',
              file=sys.stderr)

    per_feature = {}
    for feature, tests in feature_tests.items():
        if not tests:
            continue
        per_browser = {}
        for browser in BROWSERS:
            summary = summaries[browser]
            total_passes = 0
            total_tests = 0
            for test_path in tests:
                if test_path in summary:
                    passes, total = extract_pass_total(summary[test_path])
                    if passes is not None and total is not None:
                        total_passes += passes
                        total_tests += total
            per_browser[browser] = (total_passes / total_tests * 100
                                    if total_tests > 0 else None)
        per_feature[feature] = per_browser

    return versions, per_feature


def features_with_full_data(per_feature):
    """Return the set of features that have non-None data for every browser."""
    return {f for f, scores in per_feature.items()
            if all(scores.get(b) is not None for b in BROWSERS)}


def aggregate_scores(per_feature, allowed_features):
    """Average per-browser pass rates and interop (min) across allowed features.

    Returns (browser_avgs, interop_avg, count) or None when no feature has
    full-browser data within allowed_features.
    """
    browser_scores = {b: [] for b in BROWSERS}
    interop_scores = []
    for feature in allowed_features:
        scores = per_feature.get(feature)
        if scores is None:
            continue
        if not all(scores.get(b) is not None for b in BROWSERS):
            continue
        for b in BROWSERS:
            browser_scores[b].append(scores[b])
        interop_scores.append(min(scores[b] for b in BROWSERS))

    if not interop_scores:
        return None
    browser_avgs = {b: sum(s) / len(s) for b, s in browser_scores.items()}
    interop_avg = sum(interop_scores) / len(interop_scores)
    return browser_avgs, interop_avg, len(interop_scores)


def aggregate_detail(per_feature, allowed_features):
    """Return per-feature detail rows for features within allowed_features."""
    rows = []
    for feature in sorted(allowed_features):
        scores = per_feature.get(feature)
        if scores is None:
            continue
        if not all(scores.get(b) is not None for b in BROWSERS):
            continue
        interop = min(scores[b] for b in BROWSERS)
        rows.append((
            feature,
            round(scores['chrome'], 1),
            round(scores['firefox'], 1),
            round(scores['safari'], 1),
            round(interop, 1),
        ))
    return rows


def sample_weekly(dates, count, min_days=7):
    """Pick `count` dates roughly `min_days` apart, starting from the latest.

    `dates` is an iterable of ISO date strings. Returns a list of strings
    sorted oldest-first.
    """
    if not dates:
        return []
    sorted_dates = sorted(dates)
    picked = [sorted_dates[-1]]
    for ds in reversed(sorted_dates[:-1]):
        prev = date_type.fromisoformat(picked[-1])
        curr = date_type.fromisoformat(ds)
        if (prev - curr).days >= min_days:
            picked.append(ds)
            if len(picked) >= count:
                break
    return sorted(picked)


def fetch_channel_dates(channel, max_dates, days_between=7):
    """Fetch aligned runs for a channel and pick `max_dates` weekly samples.

    Returns (dates, date_groups) where dates is sorted oldest-first.
    """
    # Pull enough history that 6 weekly samples are reachable even with gaps.
    max_runs = max_dates * days_between * len(BROWSERS) * 3
    print(f'Fetching aligned {channel} runs (up to {max_runs})...',
          file=sys.stderr)
    runs = fetch_aligned_runs(channel, max_runs)
    date_groups = group_runs_by_date(runs)
    print(f'  Found {len(date_groups)} {channel} dates with aligned runs.',
          file=sys.stderr)
    dates = sample_weekly(date_groups.keys(), max_dates, days_between)
    print(f'  Sampled {len(dates)} dates ~{days_between} days apart: '
          f'{dates}', file=sys.stderr)
    return dates, date_groups


def write_trendline_csv(path, channel, channel_dates, channel_versions,
                        channel_aggregates):
    """Write the time-series CSV: date,<browser-version,browser>...,interop."""
    headers = ['date']
    for browser in BROWSERS:
        headers.append(f'{browser}-version')
        headers.append(browser)
    headers.append('interop')

    with open(path, 'w', newline='') as f:
        writer = csv.writer(f, lineterminator='\n')
        writer.writerow(headers)
        for date in channel_dates[channel]:
            agg = channel_aggregates[channel].get(date)
            versions = channel_versions[channel].get(date, {})
            if agg is None:
                continue
            browser_avgs, interop_avg, _ = agg
            row = [date]
            for browser in BROWSERS:
                row.append(versions.get(browser, ''))
                row.append(f'{browser_avgs[browser]:.1f}')
            row.append(f'{interop_avg:.1f}')
            writer.writerow(row)
    print(f'Wrote trendline to {path}.', file=sys.stderr)


def write_detail_csv(path, detail_rows):
    """Write the per-feature detail CSV, sorted by interop score descending."""
    sorted_rows = sorted(detail_rows, key=lambda r: r[4], reverse=True)
    with open(path, 'w', newline='') as f:
        writer = csv.writer(f, lineterminator='\n')
        writer.writerow(['feature', 'chrome', 'firefox', 'safari', 'interop'])
        for row in sorted_rows:
            writer.writerow(row)
    print(f'Wrote {len(sorted_rows)} features to {path}.', file=sys.stderr)


def load_manifest(path):
    """Load web-features manifest from a local file or GitHub release."""
    if path:
        print(f'Loading manifest from {path}...', file=sys.stderr)
        with open(path) as f:
            manifest_raw = json.load(f)
    else:
        print('Fetching web features manifest from GitHub...', file=sys.stderr)
        manifest_raw = fetch_web_features_manifest()
        if manifest_raw is None:
            print('Error: could not fetch web features manifest. '
                  'Try passing --manifest with a local file.',
                  file=sys.stderr)
            sys.exit(1)
    return build_feature_to_tests(manifest_raw)


def main():
    parser = argparse.ArgumentParser(
        description='Generate all-features interop CSVs. When multiple '
                    'channels are given, features are intersected across '
                    'channels for apples-to-apples comparison.')
    parser.add_argument('--channels', nargs='+',
                        default=['experimental', 'stable'],
                        choices=['experimental', 'stable'],
                        help='Channels to process')
    parser.add_argument('--max-dates', type=int, default=6,
                        help='Number of (weekly) date samples per channel')
    parser.add_argument('--days-between', type=int, default=7,
                        help='Minimum days between sampled dates')
    parser.add_argument('--output', default=DEFAULT_OUTPUT,
                        help='Trendline CSV path; must contain {channel}')
    parser.add_argument('--detail-output', default=DEFAULT_DETAIL_OUTPUT,
                        help='Per-feature detail CSV path; must contain {channel}')
    parser.add_argument('--manifest', default=None,
                        help='Path to local WEB_FEATURES_MANIFEST.json')
    parser.add_argument('--no-unlabeled', action='store_true',
                        help='Skip the synthetic "unlabeled" feature bucket')
    args = parser.parse_args()

    if '{channel}' not in args.output or '{channel}' not in args.detail_output:
        parser.error('--output and --detail-output must contain {channel}')

    include_unlabeled = not args.no_unlabeled
    feature_tests = load_manifest(args.manifest)
    print(f'Loaded {len(feature_tests)} features.', file=sys.stderr)

    # Fetch dates and per-feature scores for every (channel, date).
    channel_dates = {}
    channel_groups = {}
    channel_versions = {ch: {} for ch in args.channels}
    channel_per_feature = {ch: {} for ch in args.channels}
    for channel in args.channels:
        dates, date_groups = fetch_channel_dates(
            channel, args.max_dates, args.days_between)
        channel_dates[channel] = dates
        channel_groups[channel] = date_groups
        debug_first = True
        for date in dates:
            print(f'  {channel} {date}: computing per-feature scores...',
                  file=sys.stderr)
            result = compute_per_feature_for_date(
                date_groups[date], feature_tests, include_unlabeled,
                debug=debug_first)
            debug_first = False
            if result is None:
                print(f'    Skipped (incomplete data).', file=sys.stderr)
                continue
            versions, per_feature = result
            channel_versions[channel][date] = versions
            channel_per_feature[channel][date] = per_feature
            full = features_with_full_data(per_feature)
            print(f'    {len(full)} features have full-browser data.',
                  file=sys.stderr)

    # Intersect features across channels using the most recent date of each.
    if len(args.channels) > 1:
        channel_supported = []
        for channel in args.channels:
            for date in reversed(channel_dates[channel]):
                if date in channel_per_feature[channel]:
                    channel_supported.append(features_with_full_data(
                        channel_per_feature[channel][date]))
                    break
            else:
                print(f'Error: no usable dates for channel {channel}.',
                      file=sys.stderr)
                sys.exit(1)
        allowed_features = set.intersection(*channel_supported)
        print(f'\nIntersection across {args.channels}: '
              f'{len(allowed_features)} features '
              f'(individual sets: '
              f'{[len(s) for s in channel_supported]}).',
              file=sys.stderr)
    else:
        only = args.channels[0]
        latest_per_feature = next(
            (channel_per_feature[only][d]
             for d in reversed(channel_dates[only])
             if d in channel_per_feature[only]),
            None)
        if latest_per_feature is None:
            print(f'Error: no usable dates for channel {only}.',
                  file=sys.stderr)
            sys.exit(1)
        allowed_features = features_with_full_data(latest_per_feature)
        print(f'\nUsing {len(allowed_features)} features from latest '
              f'{only} date.', file=sys.stderr)

    # Aggregate trendlines over the intersected feature set.
    channel_aggregates = {ch: {} for ch in args.channels}
    for channel in args.channels:
        for date in channel_dates[channel]:
            per_feature = channel_per_feature[channel].get(date)
            if per_feature is None:
                continue
            agg = aggregate_scores(per_feature, allowed_features)
            if agg is None:
                continue
            channel_aggregates[channel][date] = agg
            _, interop_avg, count = agg
            print(f'  {channel} {date}: {count} features, '
                  f'interop={interop_avg:.1f}%', file=sys.stderr)

        write_trendline_csv(args.output.format(channel=channel),
                            channel, channel_dates,
                            channel_versions, channel_aggregates)

        # Detail uses the latest date in this channel.
        latest_date = next(
            (d for d in reversed(channel_dates[channel])
             if d in channel_per_feature[channel]),
            None)
        if latest_date is not None:
            detail_rows = aggregate_detail(
                channel_per_feature[channel][latest_date],
                allowed_features)
            write_detail_csv(args.detail_output.format(channel=channel),
                             detail_rows)


if __name__ == '__main__':
    main()
