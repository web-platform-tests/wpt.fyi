import globals from "globals";
import babelParser from "@babel/eslint-parser";
import path from "node:path";
import { fileURLToPath } from "node:url";
import js from "@eslint/js";
import { FlatCompat } from "@eslint/eslintrc";
import html from "eslint-plugin-html";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);


const compat = new FlatCompat({
    baseDirectory: __dirname,
    recommendedConfig: js.configs.recommended,
    allConfig: js.configs.all
});

export default [...compat.extends("eslint:recommended"), {
    languageOptions: {
        globals: {
            ...globals.browser,
            ...globals.mocha,
            assert: true,
            expect: true,
            fixture: true,
            flush: true,
            sandbox: true,
            sinon: true,
        },

        parser: babelParser,
        ecmaVersion: 2023,
        sourceType: "module",

        parserOptions: {
            requireConfigFile: false,

            babelOptions: {
                plugins: ["@babel/plugin-syntax-import-assertions"],
            },
        },
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
},

{
    languageOptions: {
        globals: {
            ...globals.browser,
            ...globals.mocha,
            assert: true,
            expect: true,
            fixture: true,
            flush: true,
            sandbox: true,
            sinon: true,
        },

        parser: babelParser,
        ecmaVersion: 2023,
        sourceType: "module",

        parserOptions: {
            requireConfigFile: false,

            babelOptions: {
                plugins: ["@babel/plugin-syntax-import-assertions"],
            },
        },
    },

    files: ["components/test/*.html"],
    plugins: { html },
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
}];