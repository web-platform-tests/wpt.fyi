# Interop & BSF Data Generation Pipeline

[`generate-interop-data.js`](generate-interop-data.js) builds and validates the precomputed Feature-Based Interop and Browser-Specific Failures (BSF) data bundles in [`data/`](data/) served by [`api/feature_interop_handler.go`](../../api/feature_interop_handler.go) and [`shared/fetch_feature_interop.go`](../../shared/fetch_feature_interop.go).

## Commands

Run from `webapp/`:

```sh
# 1. Validate bundled .json.gz files offline (< 1s, no external repos required)
npm run interop-validate-data

# 2. Rebuild both .json.gz bundles from a sibling results-analysis checkout (~2 min)
npm run interop-build-data
```

### Optional Flags (`generate-interop-data.js`)
- `--validate-only`: Validate the existing `.json.gz` files in `webapp/scripts/data/` without rebuilding.
- `--results-analysis-dir=/path/to/results-analysis`: Path to the `results-analysis` checkout containing `results-analysis-cache.git` (default: `../../results-analysis`).
- `--to=YYYY-MM-DD`: Exclusive end date for 2026 daily aligned runs (default: `2026-10-07`).
- `--refresh-webstatus`: Re-fetch feature maturity labels from `https://api.webstatus.dev/v1/features`.

---

## Generated Files (`webapp/scripts/data/`)

1. **`web-features-manifest.json.gz`**:
   - `data`: Canonical `WEB_FEATURES_MANIFEST.json` mapping (`feature_id -> [test paths]`).
   - `validWfFeatures`: Sorted list of `857` `WEB_FEATURES` IDs with WPT tests.
   - `validUnannotatedDirs`: Sorted list of `236` disjoint `dir:/...` fallback buckets covering un-annotated WPT tests.
   - `tiers`: Maturity tier classification (`3`, `2`, `1`, `0`) for each `WEB_FEATURES` ID from `webstatus.dev`.
   - `coverageStats`: Reference-run feature (`rows`) and test (`tests`) counts per channel (`experimental`, `stable`), scope (`web-features`, `hybrid-dirs`, `web-features-tentative`, `hybrid-dirs-tentative`, `no-grouping`), and tier (`eq3..eq0`, `ge3..ge0`).

2. **`interop-series-bundle.json.gz`**:
   - Contains `130` CSV time series (`2018–2026`, 260 aligned Chrome/Firefox/Safari runs for `experimental` and 208 for `stable`):
     - **2 Channels**: `experimental`, `stable`
     - **4 Scopes**: `web-features`, `hybrid-dirs` (excluding `.tentative` tests and `> 50%` tentative features), plus `web-features-tentative`, `hybrid-dirs-tentative` (all tests and features)
     - **8 Tiers**: `eq3`, `eq2`, `eq1`, `eq0`, `ge3`, `ge2`, `ge1`, `ge0`
     - **2 Series Kinds per `(channel, scope, tier)`**:
       - `${channel}--${scope}--${tier}--fail.csv`: columns `sha,date,manifest,chrome-version,chrome,firefox-version,firefox,safari-version,safari,interop,features` storing `Fail Σ` to 3 decimal places and active feature count `features` ($N$).
       - `${channel}--${scope}--${tier}--bsf.csv`: same schema storing `BSF Σ` and `features` ($N$).
     - Plus `experimental-raw-bsf.csv` and `stable-raw-bsf.csv` (legacy flat-test BSF).

---

## Validation Invariants (`npm run interop-validate-data`)

`validateBundledData()` verifies across every date, channel, and scope:
1. **Catalog Completeness**: `857` `validWfFeatures` + `236` `validUnannotatedDirs` = `1,093` total features, and `130` CSV series in `interop-series-bundle.json.gz`.
2. **Disjoint Tier Partition Identities**:
   - `features(eq3) + features(eq2) + features(eq1) + features(eq0) === features(ge0)`
   - `features(eq3) + features(eq2) === features(ge2)`
   - `features(eq3) + features(eq2) + features(eq1) === features(ge1)`
3. **Additivity of `Fail Σ` and `BSF Σ`**:
   - For every browser column (`chrome`, `firefox`, `safari`, `interop`), the sum of `eq3 + eq2 + eq1 + eq0` equals `ge0` within floating-point rounding (`<= 0.002` features).
