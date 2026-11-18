# Feature-Based Interop & Browser-Specific Failures (BSF)

Raw WPT test counts weight every test file equally—meaning a 4,000-test specification folder outweighs a 10-test feature by `400:1`, and newly landed single-engine or `.tentative` tests immediately penalize other browsers' failure rates.

This proposal normalizes WPT results by **equal-weight Web Features** ($1/N$) and partitions features into **maturity tiers** (`Baseline`, `Quorum`, `Frontier`, `Discouraged`).

## Trying It Locally

Pre-built data bundles (`webapp/scripts/data/*.json.gz`) are checked into the repository and served automatically—no extra flags or setup are required beyond standard `wpt.fyi` local server instructions:

```bash
# 1. Start the local wpt.fyi web server (follows standard main branch setup)
go build ./webapp/web && ./web
# Then open http://localhost:8080/results/

# 2. (Optional) Validate or rebuild the bundled .json.gz dataset from scratch
cd webapp
npm run interop-validate-data   # Verify bundled data integrity (<1s)
npm run interop-build-data      # Rebuild from sibling ../results-analysis checkout (~90s)
```
*(See [`webapp/scripts/README.md`](webapp/scripts/README.md) for full data pipeline and bundle format details.)*

---

## 1. The 3 Additive Data Layers (Plus Tentative Filtering)

The model is structured into decoupled layers that can each be toggled independently in the UI:

### Layer 1 (Base) — Official `WEB_FEATURES_MANIFEST` & Maturity Tiers (`2026+`)
- **Feature Mapping**: Matches WPT test paths against the official WPT `WEB_FEATURES_MANIFEST.json.gz` (`857` curated `WEB_FEATURES.yml` features covering `~75.6%` of WPT).
- **Maturity Tiers (`Baseline`, `Quorum`, `Frontier`, `Discouraged`)**: Classifies each feature into one of 4 mutually exclusive tiers (`eq3 + eq2 + eq1 + eq0 = ge0`) using `webstatus.dev` BCD browser implementation counts (with a $\ge 50\%$ engine pass-rate fallback for newly minted IDs not yet indexed by `webstatus.dev`):
  - **`Baseline` (`3/3`)**: Shipped in all 3 core engines (Chrome, Firefox, Safari). Every failure here is a genuine interop bug on a consensus feature.
  - **`Quorum` (`2/3`)**: Shipped in 2 of 3 engines (cumulative `≥2/3` = `Baseline + Quorum`). Highlights where a 3rd browser is catching up.
  - **`Frontier` (`1/3`)**: Single-engine and incubating features.
  - **`Discouraged` (`0/3`)**: Explicitly discouraged features or `0/3` unimplemented buckets.

### Layer 2 (`+ WPT Subdirs`, on by default) — 100% WPT Coverage via Disjoint Directory Buckets
- **Why**: `~24.4%` of WPT tests (`236` directories) are not yet annotated in `WEB_FEATURES.yml`.
- **How**: Un-annotated tests are grouped into `236` disjoint `dir:/...` feature buckets (subdividing 20 large umbrella directories like `css/`, `html/`, `svg/`, `dom/`, `fetch/`, `webgpu/`, `wasm/` at depth 2 `/dir/subdir`, and standalone spec folders at depth 1 `/dir`).
- **Strictly Additive**: Because the `236` `dir:/...` buckets only cover un-annotated tests, enabling `+ WPT Subdirs` leaves all `857` curated `WEB_FEATURES` untouched (`857 + 236 = 1,093` total features = `100%` of WPT) and classifies each `dir:` bucket into tiers `0..3` using the $\ge 50\%$ engine pass-rate threshold.

### Layer 3 (`Range: 1Y / 3Y / 5Y / All`) — Pre-2026 Historical Projection (`2018–2025`)
- **Why**: Official WPT `WEB_FEATURES_MANIFEST.json.gz` release assets only stabilized on `2026-01-28`.
- **How**: Prior to `2026-01-28`, the canonical feature catalog and cohort tiers are projected onto aligned historical WPT runs (`2018–2025`) and rendered as a **dashed line** (with `>= 2026-01-28` rendered as a **solid line**) so multi-year progress can be inspected without catalog drift.

### Tentative Filtering (`+ .tentative`, off by default)
- **Why**: `~6.3%` of WPT tests (`~4,600` tests) are `.tentative` (`*.tentative.*`, `*-tentative.*`, `*_tentative.*`, `/tentative/`) for unmerged spec proposals. Broad `files: "**"` globs in `WEB_FEATURES.yml` sweep `~1,400` tentative tests into `3/3 Baseline` features (where tentative tests fail at `62.7%` vs `10.2%` for standard tests). Additionally, several incubating features contain dozens of `.tentative` tests alongside a single non-tentative crash or `idlharness` test (e.g. `interest-invokers` `53/54` tentative, `soft-navigations` `82/83`, `edit-context` `14/15`).
- **How**: By default (`+ .tentative` unchecked), both individual `.tentative` tests **and** majority-tentative features (`> 50%` tentative tests) are excluded (`818` non-tentative `WEB_FEATURES` + `206` `dir:` features = `1,024` features). Checking `+ .tentative` includes all tentative tests and features (`1,093` features).

---

## 2. Metrics: `Fail` vs `BSF` and `%` vs `Σ`

Within any selected scope and tier of $N$ active features, each feature $f$ has equal weight. Let $r_{f,b} \in [0, 1]$ be browser $b$'s average test pass rate in feature $f$, $r_{f,\text{interop}} = \frac{1}{|T_f|}\sum_{t \in T_f} \min_b s_{t,b}$ be the 3-browser interop pass rate, and $\text{bsf}_{f,b} \in [0, 1]$ be browser $b$'s average unique-failure rate in feature $f$:

- **Feature-Equivalent Failure & BSF Sums (`Fail Σ` and `BSF Σ`, in `[0, N]` features)**:
  $$\text{Fail }\Sigma_b = \sum_{f=1}^{N} (1 - r_{f,b}), \qquad \text{BSF }\Sigma_b = \sum_{f=1}^{N} \text{bsf}_{f,b}$$
  Because `Σ` does not divide by $N$, expanding the maturity filter (e.g. from `3/3` to `≥2/3` or `≥1/3`) is **strictly additive**: $\text{Fail }\Sigma(\text{eq3}) + \text{Fail }\Sigma(\text{eq2}) + \text{Fail }\Sigma(\text{eq1}) + \text{Fail }\Sigma(\text{eq0}) = \text{Fail }\Sigma(\text{ge0})$, avoiding denominator dilution when new features enter the catalog.

- **Average Feature Failure & BSF Rates (`Fail %` and `BSF %`, in `0–100%`)**:
  $$\text{Fail \%}_b = \frac{\text{Fail }\Sigma_b}{N} \times 100, \qquad \text{BSF \%}_b = \frac{\text{BSF }\Sigma_b}{N} \times 100$$

---

*For data generation, bundle schema, and validation commands, see [`webapp/scripts/README.md`](webapp/scripts/README.md).*
