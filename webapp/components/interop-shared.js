/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

/**
 * Shared helpers and coverage metadata for the Feature-Based Interop & BSF
 * views (<wpt-interop-graph>, <wpt-bsf>, and <wpt-results>).
 *
 * Architectural Overview (3-Stage Pipeline):
 * ------------------------------------------
 * 1. Stage 1 — Feature Scope (`scope`):
 *    Groups individual WPT test files into equal-weight features so that a
 *    4,000-test specification folder does not outweigh a 10-test feature 400:1.
 *    - 'web-features': Uses the 857 human-curated WEB_FEATURES.yml definitions
 *      (~53.5k tests, 75.6% of WPT).
 *    - 'hybrid-dirs' (default): Supplements the 857 curated features with 236
 *      subdirectory-derived fallback features (`dir:/...`) for the remaining
 *      ~17.2k un-annotated WPT tests, achieving 100% WPT coverage across
 *      1,093 total potential features.
 *
 * 2. Stage 2 — Maturity Tier Filter (`tier`):
 *    Classifies all 1,093 potential features into 4 disjoint maturity tiers
 *    using official WebDX `web-features` support labels where available and a
 *    >=50% pass-rate engine count fallback for un-annotated `dir:` buckets:
 *    - Tier 3 ('3/3 Baseline'): Shipped in all 3 core browsers (Chrome, Firefox, Safari).
 *    - Tier 2 ('2/3 Partial'):  Shipped in 2 of 3 browsers.
 *    - Tier 1 ('1/3 Frontier'): Single-browser or non-discouraged incubating features.
 *    - Tier 0 ('0/3 Discouraged'): Explicitly discouraged or 0/3 unimplemented features.
 *    Supports cumulative filters ('ge3', 'ge2', 'ge1', 'ge0') or single-slice
 *    filters ('eq3', 'eq2', 'eq1', 'eq0').
 *
 * 3. Stage 3 — Metric (`mode`, all oriented so lower = fewer failures = better):
 *    - 'raw-pass-3of3'      (Fail %): Average % of tests failing per feature (0-100%).
 *    - 'fail-feature-counts' (Fail Σ): Sum of feature failure rates (0-1 per feature)
 *                            without dividing by N (avoids denominator dilution).
 *    - 'bsf-passrate'       (BSF %):  Average % of tests per feature failing uniquely
 *                            in 1 browser while passing in the other 2.
 *    - 'bsf-feature-counts' (BSF Σ):  Sum of feature BSF rates (0-1 per feature).
 */

/**
 * Palette for the 4 disjoint maturity tiers in the left-side coverage bars.
 * Chosen to be distinct from green/red pass/fail assessment colors.
 */
export const TIER_COLORS = {
  3: '#0284c7', // 3/3 Baseline (Sky Blue)
  2: '#6366f1', // 2/3 Partial (Indigo)
  1: '#a855f7', // 1/3 Frontier (Purple)
  0: '#14b8a6', // 0/3 Discouraged (Teal)
};

const COLOR_FILTERED = '#ffffff';

/**
 * Loads `/api/web-features-manifest` once and caches the parsed lookup tables
 * and `coverageStats` on `window._wptWebFeaturesCache`.
 * @returns {Promise<Object|null>}
 */
export async function loadWebFeaturesManifest() {
  if (typeof window === 'undefined') {
    return null;
  }
  if (window._wptWebFeaturesCache) {
    return window._wptWebFeaturesCache;
  }
  if (window._wptWebFeaturesManifestPromise) {
    return window._wptWebFeaturesManifestPromise;
  }
  window._wptWebFeaturesManifestPromise = (async() => {
    try {
      const r = await window.fetch('/api/web-features-manifest');
      if (!r.ok) {
        return null;
      }
      const json = await r.json();
      const testToWf = new Map();
      const data = json.data || {};
      for (const [feat, tests] of Object.entries(data)) {
        for (let i = 0; i < tests.length; i++) {
          const t = tests[i];
          let arr = testToWf.get(t);
          if (!arr) {
            arr = [];
            testToWf.set(t, arr);
          }
          arr.push(feat);
        }
      }
      const webFeatureLabelTiers = json.labelTiers || json.tiers || {};
      const coverageStats = json.coverageStats || null;
      window._wptWebFeaturesCache = {
        testToWebFeatures: testToWf,
        webFeatureLabelTiers,
        coverageStats,
      };
      if (coverageStats) {
        window.dispatchEvent(new CustomEvent('wpt-coverage-stats-updated', {
          detail: coverageStats,
        }));
      }
      return window._wptWebFeaturesCache;
    } catch {
      return null;
    } finally {
      window._wptWebFeaturesManifestPromise = null;
    }
  })();
  return window._wptWebFeaturesManifestPromise;
}

/**
 * Updates the cached coverage stats for a channel from live `<wpt-results>`
 * data and notifies listening chart components.
 * @param {boolean} isExperimental
 * @param {Object} channelStats
 */
export function updateChannelCoverageStats(isExperimental, channelStats) {
  if (typeof window === 'undefined' || !channelStats) {
    return;
  }
  const chKey = isExperimental ? 'experimental' : 'stable';
  const cache = window._wptWebFeaturesCache || {
    testToWebFeatures: null,
    webFeatureLabelTiers: {},
    coverageStats: {},
  };
  cache.coverageStats = Object.assign({}, cache.coverageStats || {}, {
    [chKey]: channelStats,
  });
  window._wptWebFeaturesCache = cache;
  window.dispatchEvent(new CustomEvent('wpt-coverage-stats-updated', {
    detail: cache.coverageStats,
  }));
}

/**
 * Top-level WPT umbrella directories that are subdivided to a 2nd path level
 * (`/dir/subdir`) when constructing fallback directory feature keys.
 */
const TABLE_UMBRELLA_DIRS = new Set([
  'css', 'html', 'svg', 'dom', 'fetch', 'service-workers', 'workers',
  'shadow-dom', 'custom-elements', 'content-security-policy', 'navigation-api',
  'performance-timeline', 'resource-timing', 'scroll-animations',
  'web-animations', 'webrtc', 'webaudio', 'webnn', 'webgpu', 'wasm',
]);

/**
 * Returns whether a WPT test path is a tentative test (`*.tentative.*`,
 * `*-tentative.*`, `*_tentative.*`, or inside a `/tentative/` directory).
 * @param {string} testPath
 * @returns {boolean}
 */
export function isTentativeTestPath(testPath) {
  const clean = String(testPath || '').split('?')[0];
  return (
    clean.includes('.tentative.') ||
    clean.includes('-tentative.') ||
    clean.includes('_tentative.') ||
    clean.includes('/tentative/')
  );
}

/**
 * Derives the fallback directory feature key (`dir:/...`) for a WPT test path.
 * @param {string} testPath
 * @returns {string}
 */
export function dirFeatureKeyForTable(testPath) {
  const cleanPath = String(testPath || '').split('?')[0];
  const parts = cleanPath.split('/').filter(Boolean);
  if (parts.length <= 1) {
    return `dir:/${parts[0] || 'root'}`;
  }
  if (TABLE_UMBRELLA_DIRS.has(parts[0]) && parts.length >= 3) {
    return `dir:/${parts[0]}/${parts[1]}`;
  }
  return `dir:/${parts[0]}`;
}

/**
 * Returns whether a feature or test in maturity tier `k` (0..3) is included by
 * the active `tier` filter ('ge3'..'ge0' or 'eq3'..'eq0').
 * @param {number} k
 * @param {string} tier
 * @returns {boolean}
 */
export function matchesMaturityTier(k, tier) {
  if (tier === 'eq3' || tier === 'ge3') {
    return k === 3;
  }
  if (tier === 'eq2') {
    return k === 2;
  }
  if (tier === 'eq1') {
    return k === 1;
  }
  if (tier === 'eq0') {
    return k === 0;
  }
  if (tier === 'ge2') {
    return k >= 2;
  }
  if (tier === 'ge1') {
    return k >= 1;
  }
  return true;
}

/**
 * Filters a sorted array of time-series rows to the selected window (`2026`,
 * `1y`, `2y`, `3y`, `5y`, or `all`). `2026` restricts to the official WPT
 * `WEB_FEATURES_MANIFEST` window (`>= 2026-01-28`).
 * @param {Array<Array<*>>} rows
 * @param {number} dateIdx
 * @param {string} timeRange
 * @returns {Array<Array<*>>}
 */
export function filterRowsByTimeRange(rows, dateIdx, timeRange) {
  if (!rows || !rows.length || !timeRange || timeRange === 'all') {
    return rows || [];
  }
  if (timeRange === '2026') {
    const filtered2026 = rows.filter(r => String(r[dateIdx] || '') >= '2026-01-28');
    return filtered2026.length >= 2 ? filtered2026 : rows;
  }
  const yearsMap = { '1y': 1, '2y': 2, '3y': 3, '5y': 5 };
  const yearsBack = yearsMap[timeRange];
  if (!yearsBack) {
    return rows;
  }
  const lastRow = rows[rows.length - 1];
  const lastDateStr = String(lastRow[dateIdx] || '');
  const parts = lastDateStr.split('-').map(x => parseInt(x, 10));
  if (parts.length < 3 || !Number.isFinite(parts[0])) {
    return rows;
  }
  const cutoffYear = parts[0] - yearsBack;
  const cutoffIso = `${cutoffYear}-${String(parts[1]).padStart(2, '0')}-01`;
  const filtered = rows.filter(r => String(r[dateIdx] || '') >= cutoffIso);
  return filtered.length >= 2 ? filtered : rows;
}

/**
 * Builds the view model for the left-side coverage card (`Rows` and `Tests` bars).
 *
 * Both bars use fixed denominators (`1,093` potential features and all WPT
 * tests) divided into 8 fixed-width segments (`WEB_FEATURES` + `WPT Subdir` for
 * each of Tiers 3/3, 2/3, 1/3, and 0/3). Toggling a maturity filter or scope
 * converts segments between their tier color and the white/hatched filtered
 * state in place without rescaling the bar.
 *
 * @param {boolean} isExperimental
 * @param {string} mode
 * @param {string} scope
 * @param {string} tier
 * @param {Object=} coverageStats
 * @returns {Object}
 */
export function buildCoverageViewModel(isExperimental, mode, scope, tier, coverageStats) {
  const chKey = isExperimental ? 'experimental' : 'stable';
  const stats = coverageStats
    || (typeof window !== 'undefined' && window._wptWebFeaturesCache && window._wptWebFeaturesCache.coverageStats)
    || null;
  const ch = stats ? stats[chKey] : null;
  if (!ch || !ch.scopes) {
    return {
      rowUnit: (!mode || mode === 'raw-bsf') ? 'flat tests' : 'features',
      includedRowsFmt: '—',
      totalRowsFmt: '—',
      filteredRowsFmt: '0',
      hasFilteredRows: false,
      includedRowsPct: '0.0',
      filteredRowsPct: '0.0',
      rowSlices: [],
      testSlices: [],
      includedTestsFmt: '—',
      totalTestsFmt: '—',
      filteredTestsFmt: '0',
      hasFilteredTests: false,
      includedTestsPct: '0.0',
      filteredTestsPct: '0.0',
      denomBadge: '',
    };
  }

  const fmt = n => Number(n || 0).toLocaleString('en-US');
  const pct = (a, b) => (b > 0 ? ((a / b) * 100).toFixed(1) : '0.0');

  const sliceDefs = [
    { num: 3, key: 'eq3', label: 'Baseline', color: TIER_COLORS[3] },
    { num: 2, key: 'eq2', label: 'Quorum', color: TIER_COLORS[2] },
    { num: 1, key: 'eq1', label: 'Frontier', color: TIER_COLORS[1] },
    { num: 0, key: 'eq0', label: 'Discouraged', color: TIER_COLORS[0] },
  ];

  // Legacy Old BSF mode: all flat WPT tests are included without feature grouping.
  if (!mode || mode === 'raw-bsf') {
    const sc = ch.scopes['no-grouping'];
    if (!sc) {
      return {
        rowUnit: 'flat tests',
        includedRowsFmt: fmt(ch.totalTests),
        totalRowsFmt: fmt(ch.totalTests),
        filteredRowsFmt: '0',
        hasFilteredRows: false,
        includedRowsPct: '100.0',
        filteredRowsPct: '0.0',
        rowSlices: [],
        testSlices: [],
        includedTestsFmt: fmt(ch.totalTests),
        totalTestsFmt: fmt(ch.totalTests),
        filteredTestsFmt: '0',
        hasFilteredTests: false,
        includedTestsPct: '100.0',
        filteredTestsPct: '0.0',
        denomBadge: `${fmt(ch.totalTests)} / ${fmt(ch.totalTests)} tests (100%)`,
      };
    }
    const totalRows = sc.totalRows;
    const rowSlices = sliceDefs.map(sd => ({
      pct: pct(sc.rows[sd.key], totalRows),
      color: sd.color,
      cls: 'inc',
      tooltip: `${sd.label}: ${fmt(sc.rows[sd.key])} tests (Included)`,
    }));
    const testSlices = sliceDefs.map(sd => ({
      pct: pct(sc.tests[sd.key], ch.totalTests),
      color: sd.color,
      cls: 'inc',
      tooltip: `${sd.label}: ${fmt(sc.tests[sd.key])} tests (Included)`,
    }));
    return {
      rowUnit: 'flat tests',
      includedRowsFmt: fmt(totalRows),
      totalRowsFmt: fmt(totalRows),
      filteredRowsFmt: '0',
      hasFilteredRows: false,
      includedRowsPct: '100.0',
      filteredRowsPct: '0.0',
      rowSlices,
      testSlices,
      includedTestsFmt: fmt(ch.totalTests),
      totalTestsFmt: fmt(ch.totalTests),
      filteredTestsFmt: '0',
      hasFilteredTests: false,
      includedTestsPct: '100.0',
      filteredTestsPct: '0.0',
      denomBadge: `${fmt(ch.totalTests)} / ${fmt(ch.totalTests)} tests (100%)`,
    };
  }

  const includeTentative = String(scope || '').endsWith('-tentative');
  const wfKey = includeTentative && ch.scopes['web-features-tentative'] ? 'web-features-tentative' : 'web-features';
  const hybKey = includeTentative && ch.scopes['hybrid-dirs-tentative'] ? 'hybrid-dirs-tentative' : 'hybrid-dirs';
  const wfSc = ch.scopes[wfKey];
  const hybSc = ch.scopes[hybKey];
  const allHybSc = ch.scopes['hybrid-dirs-tentative'] || hybSc;
  const potentialRows = allHybSc.totalRows;

  // Derive disjoint test counts per maturity tier from cumulative ge3..ge0 counts.
  const wfTestsByTier = {
    eq3: wfSc.tests.ge3,
    eq2: Math.max(0, wfSc.tests.ge2 - wfSc.tests.ge3),
    eq1: Math.max(0, wfSc.tests.ge1 - wfSc.tests.ge2),
    eq0: Math.max(0, wfSc.tests.ge0 - wfSc.tests.ge1),
  };
  const hybTestsByTier = {
    eq3: hybSc.tests.ge3,
    eq2: Math.max(0, hybSc.tests.ge2 - hybSc.tests.ge3),
    eq1: Math.max(0, hybSc.tests.ge1 - hybSc.tests.ge2),
    eq0: Math.max(0, hybSc.tests.ge0 - hybSc.tests.ge1),
  };

  // Proposed Feature-Based modes (hybrid-dirs vs web-features, with or without tentative)
  const scKey = ch.scopes[scope] ? scope : hybKey;
  const sc = ch.scopes[scKey];
  const tKey = sc.rows[tier] !== undefined ? tier : 'ge2';
  const isHybrid = !scKey.startsWith('web-features');
  const incRows = sc.rows[tKey];
  const filtRows = potentialRows - incRows;

  const rowSlices = [];
  const testSlices = [];
  for (let idx = 0; idx < sliceDefs.length; idx++) {
    const sd = sliceDefs[idx];
    const inTier = matchesMaturityTier(sd.num, tKey);
    const wfR = wfSc.rows[sd.key];
    const dirR = Math.max(0, hybSc.rows[sd.key] - wfR);
    const wfInc = inTier;
    const dirInc = isHybrid && inTier;

    rowSlices.push({
      pct: pct(wfR, potentialRows),
      color: wfInc ? sd.color : COLOR_FILTERED,
      cls: wfInc ? 'inc' : 'excl',
      tooltip: `${sd.label} (WEB_FEATURES): ${fmt(wfR)} features (${wfInc ? 'Included' : 'Filtered by Tier'})`,
    });
    rowSlices.push({
      pct: pct(dirR, potentialRows),
      color: dirInc ? sd.color : COLOR_FILTERED,
      cls: dirInc ? 'inc' : 'excl',
      tooltip: `${sd.label} (WPT Subdir Features): ${fmt(dirR)} features (${!isHybrid ? 'Filtered: Un-annotated' : (inTier ? 'Included' : 'Filtered by Tier')})`,
    });

    const wfT = wfTestsByTier[sd.key];
    const dirT = Math.max(0, hybTestsByTier[sd.key] - wfT);
    testSlices.push({
      pct: pct(wfT, ch.totalTests),
      color: wfInc ? sd.color : COLOR_FILTERED,
      cls: wfInc ? 'inc' : 'excl',
      tooltip: `${sd.label} (WEB_FEATURES): ${fmt(wfT)} tests (${wfInc ? 'Included' : 'Filtered by Tier'})`,
    });
    testSlices.push({
      pct: pct(dirT, ch.totalTests),
      color: dirInc ? sd.color : COLOR_FILTERED,
      cls: dirInc ? 'inc' : 'excl',
      tooltip: `${sd.label} (WPT Subdir Features): ${fmt(dirT)} tests (${!isHybrid ? 'Filtered: Un-annotated' : (inTier ? 'Included' : 'Filtered by Tier')})`,
    });
  }

  const tentRows = Math.max(0, potentialRows - hybSc.totalRows);
  if (tentRows > 0) {
    rowSlices.push({
      pct: pct(tentRows, potentialRows),
      color: COLOR_FILTERED,
      cls: 'excl',
      tooltip: `Tentative Features (.tentative): ${fmt(tentRows)} features (Filtered: Tentative)`,
    });
  }
  const tentTests = Math.max(0, ch.totalTests - hybSc.tests.ge0);
  if (tentTests > 0) {
    testSlices.push({
      pct: pct(tentTests, ch.totalTests),
      color: COLOR_FILTERED,
      cls: 'excl',
      tooltip: `Tentative WPT Tests (.tentative): ${fmt(tentTests)} tests (Filtered: Tentative)`,
    });
  }

  const incTests = sc.tests[tKey];
  const filtTests = ch.totalTests - incTests;

  return {
    rowUnit: 'features',
    includedRowsFmt: fmt(incRows),
    totalRowsFmt: fmt(potentialRows),
    filteredRowsFmt: fmt(filtRows),
    hasFilteredRows: filtRows > 0,
    includedRowsPct: pct(incRows, potentialRows),
    filteredRowsPct: pct(filtRows, potentialRows),
    rowSlices,
    testSlices,
    includedTestsFmt: fmt(incTests),
    totalTestsFmt: fmt(ch.totalTests),
    filteredTestsFmt: fmt(filtTests),
    hasFilteredTests: filtTests > 0,
    includedTestsPct: pct(incTests, ch.totalTests),
    filteredTestsPct: pct(filtTests, ch.totalTests),
    denomBadge: `${fmt(incRows)}/${fmt(potentialRows)} feat (${Math.round((incRows / potentialRows) * 100)}%) • ${fmt(incTests)} tests (${pct(incTests, ch.totalTests)}%)`,
  };
}

/**
 * Sorts right-of-line series labels top-to-bottom by their natural Y position
 * (`rawY`) and spaces them at least `minGap` pixels apart within `area`.
 * @param {Array<Object>} items
 * @param {{top: number, height: number}} area
 */
export function layoutEndLabels(items, area) {
  items.sort((a, b) => a.rawY - b.rawY);
  const minGap = 25;
  const minY = Math.max(24, Math.round(area.top));
  const maxY = Math.round(area.top + area.height);
  for (let i = 0; i < items.length; i++) {
    if (i === 0) {
      items[i].labelY = Math.max(minY, items[i].rawY);
    } else {
      items[i].labelY = Math.max(items[i].rawY, items[i - 1].labelY + minGap);
    }
  }
  if (items.length && items[items.length - 1].labelY > maxY) {
    items[items.length - 1].labelY = maxY;
    for (let i = items.length - 2; i >= 0; i--) {
      items[i].labelY = Math.min(items[i].labelY, items[i + 1].labelY - minGap);
    }
  }
}

/**
 * Renders SVG endpoint dots and dashed leader lines connecting each series'
 * rightmost point `(dotX, rawY)` to its de-overlapped label `(lineEndX, labelY)`.
 * Uses `createElementNS` and `replaceChildren` rather than `innerHTML`.
 * @param {SVGElement} svgEl
 * @param {Array<Object>} items
 */
export function renderEndConnectorsSvg(svgEl, items) {
  if (!svgEl) {
    return;
  }
  const svgNs = 'http://www.w3.org/2000/svg';
  const nodes = [];
  for (let i = 0; i < items.length; i++) {
    const it = items[i];
    const circle = document.createElementNS(svgNs, 'circle');
    circle.setAttribute('cx', String(it.dotX));
    circle.setAttribute('cy', String(it.rawY));
    circle.setAttribute('r', '4');
    circle.setAttribute('fill', String(it.color));
    nodes.push(circle);

    const line = document.createElementNS(svgNs, 'line');
    line.setAttribute('x1', String(it.dotX));
    line.setAttribute('y1', String(it.rawY));
    line.setAttribute('x2', String(it.lineEndX));
    line.setAttribute('y2', String(it.labelY));
    line.setAttribute('stroke', String(it.color));
    line.setAttribute('stroke-width', '1.5');
    line.setAttribute('stroke-dasharray', '2,2');
    nodes.push(line);
  }
  svgEl.replaceChildren(...nodes);
}
