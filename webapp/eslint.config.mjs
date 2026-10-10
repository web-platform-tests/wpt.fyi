import globals from "globals";
import path from "node:path";
import { fileURLToPath } from "node:url";
import js from "@eslint/js";
import { FlatCompat } from "@eslint/eslintrc";
import { configs as litConfigs } from "eslint-plugin-lit";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);


const compat = new FlatCompat({
    baseDirectory: __dirname,
    recommendedConfig: js.configs.recommended,
    allConfig: js.configs.all
});

// Files that still import Polymer or Vaadin, which are being replaced with Lit
// and native HTML (#2986). Each file leaves this list when it stops importing
// them. Don't add files to it.
const polymerFiles = [
    "components/browser-picker.js",
    "components/channel-picker.js",
    "components/compat-2021.js",
    "components/display-logo.js",
    "components/github-login.js",
    "components/info-banner.js",
    "components/interop-dashboard.js",
    "components/interop-feature-chart.js",
    "components/interop-summary.js",
    "components/path.js",
    "components/product-builder.js",
    "components/reftest-analyzer.js",
    "components/results-navigation.js",
    "components/self-navigator.js",
    "components/test-file-results-table.js",
    "components/test-file-results.js",
    "components/test-results-history-timeline.js",
    "components/test-run.js",
    "components/test-runs-query-builder.js",
    "components/test-runs.js",
    "components/test-search.js",
    "components/test/loading-state.test.js",
    "components/test/path.test.js",
    "components/test/product-info.test.js",
    "components/test/test-runs-query.test.js",
    "components/test/wpt-amend-metadata.test.js",
    "components/test/wpt-flags.test.js",
    "components/wpt-amend-metadata.js",
    "components/wpt-bsf.js",
    "components/wpt-colors.js",
    "components/wpt-flags.js",
    "components/wpt-header.js",
    "components/wpt-insights.js",
    "components/wpt-metadata.js",
    "components/wpt-permalinks.js",
    "components/wpt-processor.js",
    "components/wpt-runs.js",
];

export default [...compat.extends("eslint:recommended"), {
    languageOptions: {
        globals: {
            ...globals.browser,
            ...globals.mocha,
        },

        ecmaVersion: 2024,
        sourceType: "module",
    },

    files: ["components/**/*.js"],
    rules: {
        "brace-style": ["error", "1tbs"],
        curly: ["error", "all"],
        eqeqeq: ["error", "always"],
        "func-call-spacing": ["error", "never"],
        indent: ["error", 2],
        "linebreak-style": ["error", "unix"],

        "no-console": ["error", {
            allow: ["assert"],
        }],

        "no-mixed-spaces-and-tabs": ["error"],
        "no-redeclare": ["error"],
        "no-trailing-spaces": ["error"],
        quotes: ["error", "single"],
        semi: ["error", "always"],
        "space-before-function-paren": ["error", "never"],
        strict: ["error", "global"],
        yoda: ["error"],
    },
}, {
    // Files that don't use Polymer: check their Lit templates, and keep Polymer
    // and Vaadin out.
    files: ["components/**/*.js"],
    ignores: polymerFiles,
    plugins: litConfigs["flat/recommended"].plugins,
    rules: {
        ...litConfigs["flat/recommended"].rules,
        "no-restricted-imports": ["error", {
            patterns: [{
                regex: "(^|/)@(polymer|vaadin)/",
                message: "Polymer and Vaadin are being replaced with Lit and native HTML (#2986). Don't add new uses.",
            }],
        }],
    },
}];