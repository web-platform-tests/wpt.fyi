#!/usr/bin/env node
'use strict';

/**
 * Reproducible data generation & validation pipeline for the wpt.fyi
 * Feature-Based Interop & Browser-Specific Failures (BSF) dataset.
 *
 * Generates:
 *   1. `webapp/scripts/data/web-features-manifest.json.gz`
 *      Canonical `WEB_FEATURES_MANIFEST` mapping, disjoint un-annotated WPT
 *      directory fallback catalog (`validUnannotatedDirs`), and WebStatus BCD
 *      maturity tier labels (`tiers` / `labelTiers`).
 *   2. `webapp/scripts/data/interop-series-bundle.json.gz`
 *      66 historical and daily time-series CSVs (2018–2026) across:
 *        - 2 Channels: `experimental`, `stable`
 *        - 2 Scopes:   `web-features` (857 curated features),
 *                      `hybrid-dirs`  (1,093 features = 857 + 236 `dir:/...` fallbacks)
 *        - 8 Tiers:    `eq3`, `eq2`, `eq1`, `eq0`, `ge3`, `ge2`, `ge1`, `ge0`
 *        - 2 Raw Feature-Count CSVs per (channel, scope, tier):
 *                      `fail.csv` (raw feature-equivalent failure counts + `features` count)
 *                      `bsf.csv`  (raw feature-equivalent BSF counts + `features` count)
 *                      (from which the UI computes both `Σ` counts and `% = Σ / features * 100`)
 *        - Plus `experimental-raw-bsf.csv` and `stable-raw-bsf.csv` (legacy flat-test BSF).
 *
 * 3-Layer Additive Architecture:
 * ------------------------------
 *   1. Base Layer — Official `WEB_FEATURES_MANIFEST` (`web-features` scope):
 *      Matches test paths directly against the canonical WPT
 *      `WEB_FEATURES_MANIFEST.json.gz` (`merge_pr_63299`, 857 features, 53,489
 *      tests) and classifies each feature into a maturity tier (`3/3`, `2/3`,
 *      `1/3`, `0/3`) using `webstatus.dev` BCD implementation counts.
 *   2. Augmentation 1 — 100% WPT Coverage via Disjoint Directory Buckets (`hybrid-dirs` scope):
 *      Extends the 857 curated `web-features` with 236 disjoint `dir:/...`
 *      buckets (`validUnannotatedDirs`) covering the remaining 17,257
 *      un-annotated WPT tests (1,093 total features = 100% of WPT).
 *   3. Augmentation 2 — Pre-2026 Historical Projection (`2018–2025`):
 *      Projects the canonical feature catalog onto aligned historical runs
 *      prior to `2026-01-28` so multi-year trendlines can be inspected
 *      without cataloguing drift.
 *
 * Usage:
 *   cd webapp
 *   npm run interop-validate-data   # Validate bundled .json.gz files (<1s)
 *   npm run interop-build-data      # Rebuild .json.gz files from results-analysis (~60s)
 *
 * Optional flags:
 *   --validate-only                                   (validate bundled data without rebuilding)
 *   --results-analysis-dir=/path/to/results-analysis  (default: ../../results-analysis)
 *   --to=YYYY-MM-DD                                   (default: 2026-10-07, exclusive)
 *   --refresh-webstatus                               (re-fetch feature metadata from api.webstatus.dev)
 */

const childProcess = require('child_process');
const fs = require('fs');
const https = require('https');
const os = require('os');
const path = require('path');
const zlib = require('zlib');

const WEBAPP_DIR = path.resolve(__dirname, '..');
const DATA_DIR = path.join(__dirname, 'data');

function parseArgs(argv) {
  const opts = {
    resultsAnalysisDir: path.resolve(WEBAPP_DIR, '..', '..', 'results-analysis'),
    toDate: '2026-10-07',
    refreshWebstatus: false,
    validateOnly: false,
  };
  for (const arg of argv.slice(2)) {
    if (arg.startsWith('--results-analysis-dir=')) {
      opts.resultsAnalysisDir = path.resolve(arg.split('=')[1]);
    } else if (arg.startsWith('--to=')) {
      const val = arg.split('=')[1];
      if (!/^\d{4}-\d{2}-\d{2}$/.test(val)) {
        throw new Error(`Invalid --to date format (expected YYYY-MM-DD): ${val}`);
      }
      opts.toDate = val;
    } else if (arg === '--refresh-webstatus') {
      opts.refreshWebstatus = true;
    } else if (arg === '--validate-only') {
      opts.validateOnly = true;
    }
  }
  return opts;
}

const OPTS = parseArgs(process.argv);

const PRODUCTS = ['chrome', 'firefox', 'safari'];
const CHANNELS = ['experimental', 'stable'];
const SCOPES = ['web-features', 'hybrid-dirs', 'web-features-tentative', 'hybrid-dirs-tentative'];
const TIERS = ['eq3', 'eq2', 'eq1', 'eq0', 'ge3', 'ge2', 'ge1', 'ge0'];
const ENGINE_SUPPORT_THRESHOLD = 0.50;
const CANONICAL_MANIFEST_TAG = 'merge_pr_63299';
const CANONICAL_REF_DATE = '2026-10-06';

/**
 * Validates the generated `web-features-manifest.json.gz` and
 * `interop-series-bundle.json.gz` for structural completeness, disjoint tier
 * partition identities (`eq3 + eq2 + eq1 + eq0 === ge0`), and additive `Fail Σ`
 * and `BSF Σ` consistency across tiers.
 * Can be run standalone via `npm run interop-validate-data` without requiring
 * `results-analysis` or `wpt` checkouts.
 */
function validateBundledData() {
  const manifestGz = path.join(DATA_DIR, 'web-features-manifest.json.gz');
  const bundleGz = path.join(DATA_DIR, 'interop-series-bundle.json.gz');
  if (!fs.existsSync(manifestGz) || !fs.existsSync(bundleGz)) {
    throw new Error(`Missing bundled data files in ${DATA_DIR}`);
  }
  const manifest = JSON.parse(zlib.gunzipSync(fs.readFileSync(manifestGz)).toString('utf-8'));
  const bundle = JSON.parse(zlib.gunzipSync(fs.readFileSync(bundleGz)).toString('utf-8'));

  const wfCount = (manifest.validWfFeatures || []).length;
  const dirCount = (manifest.validUnannotatedDirs || []).length;
  const bundleKeys = Object.keys(bundle);

  console.log('\n=== Validating Bundled Interop & BSF Data ===');
  console.log(`Manifest catalog : ${wfCount} WEB_FEATURES + ${dirCount} un-annotated WPT dirs = ${wfCount + dirCount} total features`);
  console.log(`Series bundle    : ${bundleKeys.length} CSV series (${(fs.statSync(bundleGz).size / 1024).toFixed(1)} KB gzipped)`);

  if (wfCount !== 857 || dirCount !== 236) {
    throw new Error(`Expected 857 WF + 236 dirs (1,093 total), got ${wfCount} + ${dirCount}`);
  }
  if (!manifest.coverageStats || !manifest.coverageStats.experimental || !manifest.coverageStats.stable) {
    throw new Error('Missing coverageStats (experimental/stable) in web-features-manifest.json.gz');
  }
  if (bundleKeys.length !== 130) {
    throw new Error(`Expected 130 CSV series in bundle, got ${bundleKeys.length}`);
  }

  function parseCsvRows(key) {
    const lines = bundle[key].trim().split('\n');
    const header = lines[0].split(',');
    return lines.slice(1).map(l => {
      const parts = l.split(',');
      const obj = {};
      header.forEach((h, i) => {
        obj[h] = parts[i];
      });
      return obj;
    });
  }

  let identityErrors = 0;
  let maxFailSumDiff = 0;
  let maxBsfSumDiff = 0;

  for (const ch of CHANNELS) {
    for (const scope of SCOPES) {
      const maxAllowedFeats = scope.startsWith('web-features') ? wfCount : (wfCount + dirCount);
      const failByTier = {};
      const bsfByTier = {};
      for (const t of TIERS) {
        failByTier[t] = parseCsvRows(`${ch}--${scope}--${t}--fail.csv`);
        bsfByTier[t] = parseCsvRows(`${ch}--${scope}--${t}--bsf.csv`);
      }
      const nDates = failByTier.ge0.length;
      for (let i = 0; i < nDates; i++) {
        const f3 = Number(failByTier.eq3[i].features);
        const f2 = Number(failByTier.eq2[i].features);
        const f1 = Number(failByTier.eq1[i].features);
        const f0 = Number(failByTier.eq0[i].features);
        const fg2 = Number(failByTier.ge2[i].features);
        const fg1 = Number(failByTier.ge1[i].features);
        const fg0 = Number(failByTier.ge0[i].features);

        if (f3 + f2 + f1 + f0 !== fg0 || f3 + f2 !== fg2 || f3 + f2 + f1 !== fg1 || fg0 > maxAllowedFeats) {
          identityErrors++;
          console.error('Partition identity error:', ch, scope, failByTier.ge0[i].date, { f3, f2, f1, f0, fg2, fg1, fg0 });
        }

        for (const col of ['chrome', 'firefox', 'safari', 'interop']) {
          const failSumParts =
            parseFloat(failByTier.eq3[i][col]) +
            parseFloat(failByTier.eq2[i][col]) +
            parseFloat(failByTier.eq1[i][col]) +
            parseFloat(failByTier.eq0[i][col]);
          const failSumAll = parseFloat(failByTier.ge0[i][col]);
          maxFailSumDiff = Math.max(maxFailSumDiff, Math.abs(failSumParts - failSumAll));

          const bsfSumParts =
            parseFloat(bsfByTier.eq3[i][col]) +
            parseFloat(bsfByTier.eq2[i][col]) +
            parseFloat(bsfByTier.eq1[i][col]) +
            parseFloat(bsfByTier.eq0[i][col]);
          const bsfSumAll = parseFloat(bsfByTier.ge0[i][col]);
          maxBsfSumDiff = Math.max(maxBsfSumDiff, Math.abs(bsfSumParts - bsfSumAll));
        }
      }
    }
  }

  console.log(`Partition identity errors (eq3+eq2+eq1+eq0 == ge0) : ${identityErrors}`);
  console.log(`Max Fail Σ additivity drift across 4 tiers         : ${maxFailSumDiff.toFixed(4)} features`);
  console.log(`Max BSF Σ additivity drift across 4 tiers          : ${maxBsfSumDiff.toFixed(4)} features`);

  const sampleRows = parseCsvRows('experimental--hybrid-dirs--ge2--fail.csv');
  if (sampleRows.length > 0) {
    const first = sampleRows[0];
    const last = sampleRows[sampleRows.length - 1];
    const nFeats = Number(last.features);
    console.log(`Date range       : ${first.date} .. ${last.date} (${sampleRows.length} aligned runs per series)`);
    console.log(
      `Latest (${last.date}, experimental / hybrid-dirs / ge2, ${nFeats} features): ` +
      `Fail Σ (Fail %) -> Chrome ${Number(last.chrome).toFixed(1)} (${((last.chrome / nFeats) * 100).toFixed(1)}%) | ` +
      `Firefox ${Number(last.firefox).toFixed(1)} (${((last.firefox / nFeats) * 100).toFixed(1)}%) | ` +
      `Safari ${Number(last.safari).toFixed(1)} (${((last.safari / nFeats) * 100).toFixed(1)}%) | ` +
      `Interop ${Number(last.interop).toFixed(1)} (${((last.interop / nFeats) * 100).toFixed(1)}%)`
    );
  }

  if (identityErrors > 0 || maxFailSumDiff > 0.1 || maxBsfSumDiff > 0.1) {
    throw new Error('Data validation failed.');
  }
  console.log('All validation checks passed!\n');
}

if (OPTS.validateOnly) {
  validateBundledData();
  process.exit(0);
}

/**
 * `nodegit` in `results-analysis` is a native C++ addon compiled for a specific
 * Node ABI version (e.g. Node v22). If the current `node` binary has a different
 * ABI version, automatically locate a compatible Node binary in `~/.nvm` and
 * re-exec this script under it so `npm run interop-build-data` works seamlessly.
 */
function ensureCompatibleNodeForNodegit(resultsAnalysisDir) {
  const nodegitPath = path.join(resultsAnalysisDir, 'node_modules', 'nodegit');
  try {
    return require(nodegitPath);
  } catch (err) {
    if (process.env.WPT_INTEROP_REEXEC === '1') {
      throw err;
    }
    const nvmVersionsDir = path.join(os.homedir(), '.nvm', 'versions', 'node');
    if (fs.existsSync(nvmVersionsDir)) {
      const versions = fs.readdirSync(nvmVersionsDir).sort().reverse();
      for (const ver of versions) {
        const candidateNode = path.join(nvmVersionsDir, ver, 'bin', 'node');
        if (!fs.existsSync(candidateNode) || candidateNode === process.execPath) {
          continue;
        }
        const probe = childProcess.spawnSync(
          candidateNode,
          ['-e', `require(${JSON.stringify(nodegitPath)})`],
          { stdio: 'ignore' }
        );
        if (probe.status === 0) {
          console.log(`Re-executing with Node ${ver} (compatible with native nodegit)...`);
          const res = childProcess.spawnSync(
            candidateNode,
            [__filename, ...process.argv.slice(2)],
            {
              stdio: 'inherit',
              env: { ...process.env, WPT_INTEROP_REEXEC: '1' },
            }
          );
          process.exit(res.status || 0);
        }
      }
    }
    throw err;
  }
}

const Git = ensureCompatibleNodeForNodegit(OPTS.resultsAnalysisDir);
const lib = require(path.join(OPTS.resultsAnalysisDir, 'lib'));
const resultTrees = require(path.join(OPTS.resultsAnalysisDir, 'lib', 'result-trees'));
const moment = require(path.join(OPTS.resultsAnalysisDir, 'node_modules', 'moment'));

/**
 * Top-level WPT umbrella directories subdivided to a 2nd path level (`/dir/subdir`)
 * when constructing fallback directory feature keys (`dir:/...`).
 * Must match `TABLE_UMBRELLA_DIRS` in `webapp/components/interop-shared.js`.
 */
const UMBRELLA_DIRS = new Set([
  'css', 'html', 'svg', 'dom', 'fetch', 'service-workers', 'workers',
  'shadow-dom', 'custom-elements', 'content-security-policy', 'navigation-api',
  'performance-timeline', 'resource-timing', 'scroll-animations',
  'web-animations', 'webrtc', 'webaudio', 'webnn', 'webgpu', 'wasm',
]);

function dirFeatureKey(testPath) {
  const cleanPath = String(testPath || '').split('?')[0];
  const parts = cleanPath.split('/').filter(Boolean);
  if (parts.length <= 1) {
    return `dir:/${parts[0] || 'root'}`;
  }
  if (UMBRELLA_DIRS.has(parts[0]) && parts.length >= 3) {
    return `dir:/${parts[0]}/${parts[1]}`;
  }
  return `dir:/${parts[0]}`;
}

// ---------------------------------------------------------------------------
// Canonical WEB_FEATURES_MANIFEST Lookup
// ---------------------------------------------------------------------------

function buildManifestMapper(cacheDir, existingManifest) {
  const canonicalGzPath = path.join(cacheDir, `${CANONICAL_MANIFEST_TAG}-WEB_FEATURES_MANIFEST.json.gz`);
  const rawCanonical = fs.existsSync(canonicalGzPath)
    ? JSON.parse(zlib.gunzipSync(fs.readFileSync(canonicalGzPath)).toString('utf-8'))
    : existingManifest;

  const canonicalData = rawCanonical.data || {};
  const exactMap = new Map();
  const validWfSet = new Set(existingManifest.validWfFeatures || Object.keys(canonicalData));

  for (const [feat, tests] of Object.entries(canonicalData)) {
    if (!validWfSet.has(feat)) continue;
    for (const t of tests) {
      let arr = exactMap.get(t);
      if (!arr) {
        arr = [];
        exactMap.set(t, arr);
      }
      arr.push(feat);
    }
  }

  return {
    canonicalData,
    exactMap,
    validWfSet,
  };
}

// ---------------------------------------------------------------------------
// WebStatus BCD Maturity Tiers
// ---------------------------------------------------------------------------

function fetchJsonHttps(urlStr) {
  return new Promise((resolve, reject) => {
    https.get(urlStr, res => {
      if (res.statusCode < 200 || res.statusCode >= 300) {
        reject(new Error(`HTTP ${res.statusCode} for ${urlStr}`));
        res.resume();
        return;
      }
      const chunks = [];
      res.on('data', d => chunks.push(d));
      res.on('end', () => {
        try {
          resolve(JSON.parse(Buffer.concat(chunks).toString('utf-8')));
        } catch (e) {
          reject(e);
        }
      });
    }).on('error', reject);
  });
}

async function loadWebStatusTiers(existingManifest, refreshWebstatus) {
  const existingTiers = existingManifest.tiers || existingManifest.labelTiers || {};
  if (!refreshWebstatus && Object.keys(existingTiers).length > 0) {
    const copy = { ...existingTiers };
    for (const unmapped of ['masonry', 'plane-detection', 'viewport', 'web-based-payment-handler']) {
      delete copy[unmapped];
    }
    return copy;
  }
  console.log('Fetching feature maturity metadata from https://api.webstatus.dev/v1/features ...');
  const tiers = {};
  let pageToken = '';
  while (true) {
    const url = `https://api.webstatus.dev/v1/features?page_size=100${pageToken ? `&page_token=${encodeURIComponent(pageToken)}` : ''}`;
    const page = await fetchJsonHttps(url);
    for (const f of page.data || []) {
      if (f.discouraged) {
        tiers[f.feature_id] = 0;
      } else {
        const impl = f.browser_implementations || {};
        const k =
          (impl.chrome && impl.chrome.status === 'available' ? 1 : 0) +
          (impl.firefox && impl.firefox.status === 'available' ? 1 : 0) +
          (impl.safari && impl.safari.status === 'available' ? 1 : 0);
        if (k === 3 || (f.baseline && (f.baseline.status === 'widely' || f.baseline.status === 'newly'))) {
          tiers[f.feature_id] = 3;
        } else if (k === 2) {
          tiers[f.feature_id] = 2;
        } else {
          tiers[f.feature_id] = 1;
        }
      }
    }
    if (!page.metadata || !page.metadata.next_page_token) {
      break;
    }
    pageToken = page.metadata.next_page_token;
  }
  return tiers;
}

// ---------------------------------------------------------------------------
// Historical & Daily Aligned Run Resolution
// ---------------------------------------------------------------------------

function pickHistoricalSampleDatesFromBsfCsv(csvText) {
  const lines = csvText.trim().split('\n').slice(1);
  const byBucket = new Map();
  for (const line of lines) {
    const parts = line.split(',');
    const sha = parts[0];
    const date = parts[1];
    if (!date || date >= '2026-01-28') continue;
    const year = parseInt(date.slice(0, 4), 10);
    const month = parseInt(date.slice(5, 7), 10);
    // 2018–2022: bi-monthly (Feb, Apr, Jun, Aug, Oct, Dec); 2023–2026: monthly
    if (year <= 2022 && (month % 2 !== 0)) continue;
    const key = `${year}-${String(month).padStart(2, '0')}`;
    if (!byBucket.has(key)) {
      byBucket.set(key, { sha, date });
    }
  }
  return Array.from(byBucket.values());
}

async function loadHistoricalAlignedRunsForChannel(channel, rawBsfCsvText, cacheDir) {
  const targets = pickHistoricalSampleDatesFromBsfCsv(rawBsfCsvText);
  const alignedRuns = new Map();
  for (const t of targets) {
    const d = new Date(`${t.date}T00:00:00Z`);
    const fromStr = new Date(d.getTime() - 86400000).toISOString().slice(0, 10);
    const toStr = new Date(d.getTime() + 86400000).toISOString().slice(0, 10);
    const cacheFile = path.join(cacheDir, `${channel}-chrome-firefox-safari-runs-${fromStr}-${toStr}.json`);
    if (fs.existsSync(cacheFile)) {
      try {
        const runs = JSON.parse(fs.readFileSync(cacheFile, 'utf-8'));
        if (Array.isArray(runs) && runs.length === 3) {
          alignedRuns.set(t.date, runs);
          continue;
        }
      } catch {
        // Fall through to server fetch
      }
    }
    const url = `https://wpt.fyi/api/runs?aligned=true&max-count=1&label=master&label=${channel}&product=chrome&product=firefox&product=safari&from=${fromStr}&to=${toStr}`;
    const runs = await fetchJsonHttps(url);
    if (Array.isArray(runs) && runs.length === 3) {
      fs.writeFileSync(cacheFile, JSON.stringify(runs));
      alignedRuns.set(t.date, runs);
    }
  }
  return alignedRuns;
}

async function buildWptShaToReleaseTagMap(existingBundle) {
  const tagMap = new Map();
  for (const [k, csvText] of Object.entries(existingBundle)) {
    if (!k.endsWith('--fail.csv') && !k.endsWith('--raw-pass-3of3.csv')) continue;
    const lines = csvText.trim().split('\n');
    for (let i = 1; i < lines.length; i++) {
      const parts = lines[i].split(',');
      if (parts.length >= 3 && parts[2] && parts[2] !== 'retro-manifest') {
        tagMap.set(parts[0], parts[2]);
      }
    }
  }
  if (tagMap.size === 0 && lib.runs && typeof lib.runs.fetchWptTagMap === 'function') {
    const fetched = await lib.runs.fetchWptTagMap();
    for (const [sha, tag] of fetched.entries()) {
      tagMap.set(sha, tag);
    }
  }
  return tagMap;
}

// ---------------------------------------------------------------------------
// Per-Test & Per-Feature Scoring Engine
// ---------------------------------------------------------------------------

const flsTestCache = new Map();
const bsfTestCache = new Map();
const subtreeRecordsCache = new Map();

function cachedScoreTestResults(results) {
  if (results.id !== undefined) {
    const cached = flsTestCache.get(results.id);
    if (cached !== undefined) return cached;
  }
  let score = 0;
  try {
    score = resultTrees.scoreTestResults(results);
  } catch {
    score = 0;
  }
  if (results.id !== undefined) {
    flsTestCache.set(results.id, score);
  }
  return score;
}

function scoreTestBSFFromResults(res0, res1, res2) {
  const firstId = res0 && res0.id;
  if (
    firstId !== undefined &&
    res1 && res1.id === firstId &&
    res2 && res2.id === firstId
  ) {
    return [0, 0, 0];
  }

  const k0 = res0 ? res0.id : 'm';
  const k1 = res1 ? res1.id : 'm';
  const k2 = res2 ? res2.id : 'm';
  const cacheKey = `${k0}-${k1}-${k2}`;
  const cached = bsfTestCache.get(cacheKey);
  if (cached !== undefined) {
    return cached;
  }

  const resArray = [res0, res1, res2];
  const bsfScores = [0, 0, 0];
  let hasSubtests = false;
  for (let i = 0; i < 3; i++) {
    const r = resArray[i];
    if (r && r.subtests && r.subtests.length > 0) {
      hasSubtests = true;
      break;
    }
  }

  if (!hasSubtests) {
    let failedIdx = -1;
    let failedCount = 0;
    for (let i = 0; i < 3; i++) {
      const r = resArray[i];
      if (!r) {
        failedIdx = i;
        failedCount++;
        continue;
      }
      if (resultTrees.TEST_NEUTRAL_STATUSES.includes(r.status)) {
        failedCount = 0;
        break;
      }
      if (resultTrees.TEST_FAIL_STATUSES.includes(r.status)) {
        failedIdx = i;
        failedCount++;
      }
    }
    if (failedCount === 1) {
      bsfScores[failedIdx] = 1;
    }
    bsfTestCache.set(cacheKey, bsfScores);
    return bsfScores;
  }

  const subtestMaps = [];
  const allSubtestNames = new Set();
  for (let i = 0; i < 3; i++) {
    const r = resArray[i];
    const map = new Map();
    if (r && r.subtests) {
      for (let j = 0; j < r.subtests.length; j++) {
        const st = r.subtests[j];
        if (!map.has(st.name)) {
          map.set(st.name, st.status);
          allSubtestNames.add(st.name);
        }
      }
    }
    subtestMaps.push(map);
  }

  const denominator = allSubtestNames.size;
  if (denominator === 0) {
    bsfTestCache.set(cacheKey, bsfScores);
    return bsfScores;
  }

  for (const name of allSubtestNames) {
    let failedIdx = -1;
    let failedCount = 0;
    let neutral = false;
    for (let i = 0; i < 3; i++) {
      const status = subtestMaps[i].get(name);
      if (status === undefined) {
        failedIdx = i;
        failedCount++;
        continue;
      }
      if (resultTrees.SUBTEST_NEUTRAL_STATUSES.includes(status)) {
        neutral = true;
        break;
      }
      if (resultTrees.SUBTEST_FAIL_STATUSES.includes(status)) {
        failedIdx = i;
        failedCount++;
      }
    }
    if (!neutral && failedCount === 1) {
      bsfScores[failedIdx] += 1 / denominator;
    }
  }

  bsfTestCache.set(cacheKey, bsfScores);
  return bsfScores;
}

function isTentativeTestPath(testPath) {
  const clean = String(testPath || '').split('?')[0];
  return (
    clean.includes('.tentative.') ||
    clean.includes('-tentative.') ||
    clean.includes('_tentative.') ||
    clean.includes('/tentative/')
  );
}

function walkSubtreeTestRecords(t0, t1, t2, currentPath, outList) {
  if (currentPath === '/third_party') {
    return;
  }
  const id0 = t0 ? t0.id : 'm';
  const id1 = t1 ? t1.id : 'm';
  const id2 = t2 ? t2.id : 'm';
  const cacheKey = `${currentPath}:${id0}-${id1}-${id2}`;
  let ownRecords = subtreeRecordsCache.get(cacheKey);
  if (ownRecords === undefined) {
    ownRecords = [];
    const testNames = new Set();
    if (t0 && t0.tests) for (const k of Object.keys(t0.tests)) testNames.add(k);
    if (t1 && t1.tests) for (const k of Object.keys(t1.tests)) testNames.add(k);
    if (t2 && t2.tests) for (const k of Object.keys(t2.tests)) testNames.add(k);

    for (const name of testNames) {
      const r0 = t0 && t0.tests ? t0.tests[name] : undefined;
      const r1 = t1 && t1.tests ? t1.tests[name] : undefined;
      const r2 = t2 && t2.tests ? t2.tests[name] : undefined;
      if (!r0 && !r1 && !r2) continue;

      const fullPath = `${currentPath}/${name}`;
      const s0 = r0 ? cachedScoreTestResults(r0) : 0;
      const s1 = r1 ? cachedScoreTestResults(r1) : 0;
      const s2 = r2 ? cachedScoreTestResults(r2) : 0;
      const minS = Math.min(s0, s1, s2);
      const bsf = scoreTestBSFFromResults(r0, r1, r2);
      const inAllThree = Boolean(r0 && r1 && r2);
      const rawBsf = inAllThree ? bsf : [0, 0, 0];

      ownRecords.push({
        path: fullPath,
        dirKey: dirFeatureKey(fullPath),
        isTentative: isTentativeTestPath(fullPath),
        s0,
        s1,
        s2,
        minS,
        bsf0: rawBsf[0],
        bsf1: rawBsf[1],
        bsf2: rawBsf[2],
      });
    }
    subtreeRecordsCache.set(cacheKey, ownRecords);
  }

  for (let i = 0; i < ownRecords.length; i++) {
    outList.push(ownRecords[i]);
  }

  const dirNames = new Set();
  if (t0 && t0.trees) for (const k of Object.keys(t0.trees)) dirNames.add(k);
  if (t1 && t1.trees) for (const k of Object.keys(t1.trees)) dirNames.add(k);
  if (t2 && t2.trees) for (const k of Object.keys(t2.trees)) dirNames.add(k);

  for (const d of dirNames) {
    const c0 = t0 && t0.trees ? t0.trees[d] : undefined;
    const c1 = t1 && t1.trees ? t1.trees[d] : undefined;
    const c2 = t2 && t2.trees ? t2.trees[d] : undefined;
    walkSubtreeTestRecords(c0, c1, c2, `${currentPath}/${d}`, outList);
  }
}

function createFeatureAccumulator() {
  return {
    s0: 0,
    s1: 0,
    s2: 0,
    interop3: 0,
    count: 0,
    bsf0: 0,
    bsf1: 0,
    bsf2: 0,
  };
}

function addRecordToAccumulator(acc, rec) {
  acc.s0 += rec.s0;
  acc.s1 += rec.s1;
  acc.s2 += rec.s2;
  acc.interop3 += rec.minS;
  acc.count += 1;
  acc.bsf0 += rec.bsf0;
  acc.bsf1 += rec.bsf1;
  acc.bsf2 += rec.bsf2;
}

function computeTierFromPassRates(r0, r1, r2) {
  return (
    (r0 >= ENGINE_SUPPORT_THRESHOLD ? 1 : 0) +
    (r1 >= ENGINE_SUPPORT_THRESHOLD ? 1 : 0) +
    (r2 >= ENGINE_SUPPORT_THRESHOLD ? 1 : 0)
  );
}

function matchesTierFilter(k, tier) {
  if (tier === 'eq3' || tier === 'ge3') return k === 3;
  if (tier === 'eq2') return k === 2;
  if (tier === 'eq1') return k === 1;
  if (tier === 'eq0') return k === 0;
  if (tier === 'ge2') return k >= 2;
  if (tier === 'ge1') return k >= 1;
  return true;
}

function summarizeFeatureMapForTier(accMap, fixedTierMap, sha, manifestTag, versions, tier) {
  let failSum = [0, 0, 0];
  let failInteropSum = 0;
  let bsfFailSum = [0, 0, 0];
  let featCount = 0;

  for (const [featureKey, acc] of accMap.entries()) {
    if (acc.count === 0) continue;
    const r0 = acc.s0 / acc.count;
    const r1 = acc.s1 / acc.count;
    const r2 = acc.s2 / acc.count;
    const k = fixedTierMap.has(featureKey)
      ? fixedTierMap.get(featureKey)
      : computeTierFromPassRates(r0, r1, r2);
    if (!matchesTierFilter(k, tier)) continue;

    featCount++;
    failSum[0] += 1 - r0;
    failSum[1] += 1 - r1;
    failSum[2] += 1 - r2;
    failInteropSum += 1 - (acc.interop3 / acc.count);

    bsfFailSum[0] += acc.bsf0 / acc.count;
    bsfFailSum[1] += acc.bsf1 / acc.count;
    bsfFailSum[2] += acc.bsf2 / acc.count;
  }

  return {
    fail: {
      sha,
      manifest: manifestTag,
      versions,
      scores: new Map([
        ['chrome', failSum[0]],
        ['firefox', failSum[1]],
        ['safari', failSum[2]],
      ]),
      interop: failInteropSum,
      features: featCount,
    },
    bsf: {
      sha,
      manifest: manifestTag,
      versions,
      scores: new Map([
        ['chrome', bsfFailSum[0]],
        ['firefox', bsfFailSum[1]],
        ['safari', bsfFailSum[2]],
      ]),
      interop: bsfFailSum[0] + bsfFailSum[1] + bsfFailSum[2],
      features: featCount,
    },
  };
}

/**
 * Formats an aggregate feature-count CSV storing raw feature-equivalent
 * failure counts (`Fail Σ` or `BSF Σ`) to 3 decimal places plus the total
 * active feature count (`features` = N). The frontend UI directly plots these
 * counts in `Σ` mode or divides by `features` (`Σ / features * 100`) in `%` mode.
 */
function formatFeatureCountsCsv(products, dateToCounts) {
  let csv = 'sha,date,manifest';
  for (const product of products) {
    csv += `,${product}-version,${product}`;
  }
  csv += ',interop,features\n';

  for (const [date, entry] of dateToCounts.entries()) {
    const row = [entry.sha, date, entry.manifest];
    for (const product of products) {
      const ver = String(entry.versions.get(product) || '').replace(/,/g, ' ');
      row.push(ver);
      row.push(entry.scores.get(product).toFixed(3));
    }
    row.push(entry.interop.toFixed(3));
    row.push(String(entry.features));
    csv += row.join(',') + '\n';
  }
  return csv;
}

// ---------------------------------------------------------------------------
// Main Pipeline Execution & Validation
// ---------------------------------------------------------------------------

async function main() {
  const startTime = Date.now();
  const cacheDir = path.join(OPTS.resultsAnalysisDir, 'cache');
  const gitRepoPath = path.join(OPTS.resultsAnalysisDir, 'results-analysis-cache.git');

  const existingManifestGz = path.join(DATA_DIR, 'web-features-manifest.json.gz');
  const existingManifest = fs.existsSync(existingManifestGz)
    ? JSON.parse(zlib.gunzipSync(fs.readFileSync(existingManifestGz)).toString('utf-8'))
    : {};

  const existingBundleGz = path.join(DATA_DIR, 'interop-series-bundle.json.gz');
  const existingBundle = fs.existsSync(existingBundleGz)
    ? JSON.parse(zlib.gunzipSync(fs.readFileSync(existingBundleGz)).toString('utf-8'))
    : {};

  // 1. Load WebStatus maturity tiers & canonical WEB_FEATURES_MANIFEST
  const webStatusTiers = await loadWebStatusTiers(existingManifest, OPTS.refreshWebstatus);
  const mapper = buildManifestMapper(cacheDir, existingManifest);
  const tagMap = await buildWptShaToReleaseTagMap(existingBundle);

  process.chdir(OPTS.resultsAnalysisDir);
  const repo = await Git.Repository.open(gitRepoPath);

  const newBundle = {};
  if (existingBundle['experimental-raw-bsf.csv']) {
    newBundle['experimental-raw-bsf.csv'] = existingBundle['experimental-raw-bsf.csv'];
  }
  if (existingBundle['stable-raw-bsf.csv']) {
    newBundle['stable-raw-bsf.csv'] = existingBundle['stable-raw-bsf.csv'];
  }

  let canonicalUnannotatedDirs = new Set(
    (existingManifest.validUnannotatedDirs || []).filter(d => d !== 'dir:/aria' && d !== 'dir:/third_party')
  );
  const coverageAudit = {};

  for (const channel of CHANNELS) {
    const isExperimental = channel === 'experimental';
    console.log(`\n=== Scoring ${channel.toUpperCase()} channel (2018..2026) ===`);
    subtreeRecordsCache.clear();

    // Load historical sample dates (2018..2026-01-02) + daily 2026 runs (2026-01-28..toDate)
    const rawBsfCsv = existingBundle[`${channel}-raw-bsf.csv`] || '';
    const alignedRuns = await loadHistoricalAlignedRunsForChannel(channel, rawBsfCsv, cacheDir);
    const runs2026 = await lib.runs.fetchAlignedRunsFromServer(
      PRODUCTS,
      moment('2026-01-28'),
      moment(OPTS.toDate),
      isExperimental
    );
    for (const [date, runs] of runs2026.entries()) {
      if (OPTS.toDate === '2026-10-07' && date > CANONICAL_REF_DATE) {
        continue;
      }
      alignedRuns.set(date, runs);
    }

    await lib.results.loadRunTrees(repo, alignedRuns);

    const sortedDates = Array.from(alignedRuns.keys()).sort();
    const refDate = alignedRuns.has(CANONICAL_REF_DATE)
      ? CANONICAL_REF_DATE
      : sortedDates[sortedDates.length - 1];

    // Step A: Score the canonical reference date first to establish:
    //   1. The exact set of 236 `validUnannotatedDirs` (from experimental reference run)
    //   2. The fixed cohort maturity tier (`0..3`) for all 1,093 features in this channel
    const refRuns = alignedRuns.get(refDate);
    const refRecords = [];
    walkSubtreeTestRecords(
      (refRuns.find(r => r.browser_name === 'chrome') || refRuns[0]).tree,
      (refRuns.find(r => r.browser_name === 'firefox') || refRuns[1]).tree,
      (refRuns.find(r => r.browser_name === 'safari') || refRuns[2]).tree,
      '',
      refRecords
    );

    if (isExperimental) {
      const discoveredUnannotated = new Set();
      for (let i = 0; i < refRecords.length; i++) {
        const rec = refRecords[i];
        const wf = mapper.exactMap.get(rec.path);
        if (!wf || wf.length === 0) {
          discoveredUnannotated.add(rec.dirKey);
        }
      }
      canonicalUnannotatedDirs = discoveredUnannotated;
    }

    const refWfAccs = new Map();
    const refHybAccs = new Map();
    const refWfTests = new Map();
    const refHybTests = new Map();
    const refWfTentAccs = new Map();
    const refHybTentAccs = new Map();
    const refWfTentTests = new Map();
    const refHybTentTests = new Map();
    const getAcc = (m, k) => {
      let a = m.get(k);
      if (!a) {
        a = createFeatureAccumulator();
        m.set(k, a);
      }
      return a;
    };

    for (let i = 0; i < refRecords.length; i++) {
      const rec = refRecords[i];
      const wf = mapper.exactMap.get(rec.path);
      if (wf && wf.length > 0) {
        for (let j = 0; j < wf.length; j++) {
          const fKey = `wf:${wf[j]}`;
          addRecordToAccumulator(getAcc(refWfTentAccs, fKey), rec);
          addRecordToAccumulator(getAcc(refHybTentAccs, fKey), rec);
          if (!refWfTentTests.has(fKey)) refWfTentTests.set(fKey, new Set());
          refWfTentTests.get(fKey).add(rec.path);
          if (!refHybTentTests.has(fKey)) refHybTentTests.set(fKey, new Set());
          refHybTentTests.get(fKey).add(rec.path);
          if (!rec.isTentative) {
            addRecordToAccumulator(getAcc(refWfAccs, fKey), rec);
            addRecordToAccumulator(getAcc(refHybAccs, fKey), rec);
            if (!refWfTests.has(fKey)) refWfTests.set(fKey, new Set());
            refWfTests.get(fKey).add(rec.path);
            if (!refHybTests.has(fKey)) refHybTests.set(fKey, new Set());
            refHybTests.get(fKey).add(rec.path);
          }
        }
      } else {
        const uKey = `unannotated-${rec.dirKey}`;
        addRecordToAccumulator(getAcc(refHybTentAccs, uKey), rec);
        if (!refHybTentTests.has(uKey)) refHybTentTests.set(uKey, new Set());
        refHybTentTests.get(uKey).add(rec.path);
        if (!rec.isTentative) {
          addRecordToAccumulator(getAcc(refHybAccs, uKey), rec);
          if (!refHybTests.has(uKey)) refHybTests.set(uKey, new Set());
          refHybTests.get(uKey).add(rec.path);
        }
      }
    }

    // Build fixed cohort maturity tier map (`0..3`) and identify majority-tentative
    // features (>50% tentative tests) for this channel.
    const fixedTierMap = new Map();
    const tentativeFeatureSet = new Set();
    for (const [fKey, tentAcc] of refHybTentAccs.entries()) {
      const nonTentAcc = refHybAccs.get(fKey);
      const nonTentCount = nonTentAcc ? nonTentAcc.count : 0;
      if (nonTentCount * 2 < tentAcc.count) {
        tentativeFeatureSet.add(fKey);
        refWfAccs.delete(fKey);
        refHybAccs.delete(fKey);
        refWfTests.delete(fKey);
        refHybTests.delete(fKey);
      }
      const baseAcc = nonTentCount > 0 ? nonTentAcc : tentAcc;
      const r0 = baseAcc.s0 / baseAcc.count;
      const r1 = baseAcc.s1 / baseAcc.count;
      const r2 = baseAcc.s2 / baseAcc.count;
      if (fKey.startsWith('wf:')) {
        const fid = fKey.slice(3);
        const wsTier = webStatusTiers[fid];
        const tier = wsTier !== undefined ? wsTier : computeTierFromPassRates(r0, r1, r2);
        fixedTierMap.set(fKey, tier);
      } else {
        fixedTierMap.set(fKey, computeTierFromPassRates(r0, r1, r2));
      }
    }

    // Record coverage audit for `refDate`
    coverageAudit[channel] = { totalTests: refRecords.length, scopes: {} };
    for (const [scopeName, accMap, testsMap] of [
      ['web-features', refWfAccs, refWfTests],
      ['hybrid-dirs', refHybAccs, refHybTests],
      ['web-features-tentative', refWfTentAccs, refWfTentTests],
      ['hybrid-dirs-tentative', refHybTentAccs, refHybTentTests],
    ]) {
      const rows = {};
      const tests = {};
      for (const t of TIERS) {
        let rCount = 0;
        const tSet = new Set();
        for (const [fKey, acc] of accMap.entries()) {
          if (acc.count === 0) continue;
          const k = fixedTierMap.get(fKey);
          if (matchesTierFilter(k, t)) {
            rCount++;
            for (const p of testsMap.get(fKey) || []) tSet.add(p);
          }
        }
        rows[t] = rCount;
        tests[t] = tSet.size;
      }
      coverageAudit[channel].scopes[scopeName] = { totalRows: rows.ge0, rows, tests };
    }
    const flatRows = { eq3: 0, eq2: 0, eq1: 0, eq0: 0 };
    for (let i = 0; i < refRecords.length; i++) {
      const rec = refRecords[i];
      const k = computeTierFromPassRates(rec.s0, rec.s1, rec.s2);
      flatRows[`eq${k}`]++;
    }
    flatRows.ge3 = flatRows.eq3;
    flatRows.ge2 = flatRows.ge3 + flatRows.eq2;
    flatRows.ge1 = flatRows.ge2 + flatRows.eq1;
    flatRows.ge0 = flatRows.ge1 + flatRows.eq0;
    coverageAudit[channel].scopes['no-grouping'] = {
      totalRows: refRecords.length,
      rows: flatRows,
      tests: { ...flatRows },
    };

    // Step B: Score all historical and 2026 dates using the canonical manifest + fixed cohort tiers
    const series = {};
    for (const scope of SCOPES) {
      series[scope] = {};
      for (const tier of TIERS) {
        series[scope][tier] = {
          fail: new Map(),
          bsf: new Map(),
        };
      }
    }

    for (const date of sortedDates) {
      if (subtreeRecordsCache.size > 25000) {
        subtreeRecordsCache.clear();
      }
      const runs = alignedRuns.get(date);
      const t0 = (runs.find(r => r.browser_name === 'chrome') || runs[0]).tree;
      const t1 = (runs.find(r => r.browser_name === 'firefox') || runs[1]).tree;
      const t2 = (runs.find(r => r.browser_name === 'safari') || runs[2]).tree;
      if (!t0 || !t1 || !t2) continue;

      const sha = runs[0].full_revision_hash;
      const manifestTag = date < '2026-01-28'
        ? 'retro-manifest'
        : (tagMap.get(sha) || CANONICAL_MANIFEST_TAG);
      const versions = new Map(runs.map(r => [r.browser_name, r.browser_version]));

      const allRecords = [];
      walkSubtreeTestRecords(t0, t1, t2, '', allRecords);

      const wfAccs = new Map();
      const hybAccs = new Map();
      const wfTentAccs = new Map();
      const hybTentAccs = new Map();

      for (let i = 0; i < allRecords.length; i++) {
        const rec = allRecords[i];
        const wf = mapper.exactMap.get(rec.path);
        if (wf && wf.length > 0) {
          for (let j = 0; j < wf.length; j++) {
            const fKey = `wf:${wf[j]}`;
            addRecordToAccumulator(getAcc(wfTentAccs, fKey), rec);
            addRecordToAccumulator(getAcc(hybTentAccs, fKey), rec);
            if (!rec.isTentative && !tentativeFeatureSet.has(fKey)) {
              addRecordToAccumulator(getAcc(wfAccs, fKey), rec);
              addRecordToAccumulator(getAcc(hybAccs, fKey), rec);
            }
          }
        } else if (canonicalUnannotatedDirs.has(rec.dirKey)) {
          const uKey = `unannotated-${rec.dirKey}`;
          addRecordToAccumulator(getAcc(hybTentAccs, uKey), rec);
          if (!rec.isTentative && !tentativeFeatureSet.has(uKey)) {
            addRecordToAccumulator(getAcc(hybAccs, uKey), rec);
          }
        }
      }

      const scopeMaps = {
        'web-features': wfAccs,
        'hybrid-dirs': hybAccs,
        'web-features-tentative': wfTentAccs,
        'hybrid-dirs-tentative': hybTentAccs,
      };

      for (const scope of SCOPES) {
        const accMap = scopeMaps[scope];
        for (const tier of TIERS) {
          const sum = summarizeFeatureMapForTier(accMap, fixedTierMap, sha, manifestTag, versions, tier);
          series[scope][tier].fail.set(date, sum.fail);
          series[scope][tier].bsf.set(date, sum.bsf);
        }
      }
    }

    // Format CSV strings into `newBundle`
    for (const scope of SCOPES) {
      for (const tier of TIERS) {
        const prefix = `${channel}--${scope}--${tier}`;
        newBundle[`${prefix}--fail.csv`] = formatFeatureCountsCsv(
          PRODUCTS,
          series[scope][tier].fail
        );
        newBundle[`${prefix}--bsf.csv`] = formatFeatureCountsCsv(
          PRODUCTS,
          series[scope][tier].bsf
        );
      }
    }
  }

  // 2. Write updated `web-features-manifest.json.gz`
  const updatedManifest = {
    version: 1,
    data: mapper.canonicalData,
    validWfFeatures: Array.from(mapper.validWfSet).sort(),
    validUnannotatedDirs: Array.from(canonicalUnannotatedDirs).sort(),
    tiers: webStatusTiers,
    coverageStats: coverageAudit,
  };
  fs.mkdirSync(DATA_DIR, { recursive: true });
  const manifestGzOut = path.join(DATA_DIR, 'web-features-manifest.json.gz');
  fs.writeFileSync(manifestGzOut, zlib.gzipSync(Buffer.from(JSON.stringify(updatedManifest), 'utf-8')));

  // 3. Write updated `interop-series-bundle.json.gz`
  const bundleGzOut = path.join(DATA_DIR, 'interop-series-bundle.json.gz');
  fs.writeFileSync(bundleGzOut, zlib.gzipSync(Buffer.from(JSON.stringify(newBundle), 'utf-8')));

  console.log('\n=== Coverage Audit Summary (Reference Date: ' + CANONICAL_REF_DATE + ') ===');
  console.log(JSON.stringify(coverageAudit, null, 2));
  console.log(`\nWrote ${Object.keys(newBundle).length} series to ${bundleGzOut}`);
  console.log(`Wrote manifest (${updatedManifest.validWfFeatures.length} WF + ${updatedManifest.validUnannotatedDirs.length} dirs) to ${manifestGzOut}`);
  console.log(`Completed in ${((Date.now() - startTime) / 1000).toFixed(1)}s.`);
  validateBundledData();
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
