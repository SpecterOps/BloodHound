// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
const {
    defineConfig,
    globalIgnores,
} = require("eslint/config");

const globals = require("globals");

const {
    fixupConfigRules,
} = require("@eslint/compat");

const tsParser = require("@typescript-eslint/parser");
const sxProps = require("eslint-plugin-sx-props");
const js = require("@eslint/js");

const {
    FlatCompat,
} = require("@eslint/eslintrc");

const compat = new FlatCompat({
    baseDirectory: __dirname,
    recommendedConfig: js.configs.recommended,
    allConfig: js.configs.all
});

module.exports = defineConfig([{
    languageOptions: {
        globals: {
            ...globals.browser,
        },

        parser: tsParser,
        ecmaVersion: "latest",
        sourceType: "module",
        parserOptions: {},
    },

    extends: fixupConfigRules(compat.extends(
        "eslint:recommended",
        "plugin:@typescript-eslint/recommended",
        "plugin:react/recommended",
        "plugin:react/jsx-runtime",
        "plugin:react-hooks/recommended",
        "plugin:jsx-a11y/recommended",
        "prettier",
    )),

    plugins: {
        "sx-props": sxProps,
    },

    rules: {
        "@typescript-eslint/ban-ts-comment": "off",
        "@typescript-eslint/no-non-null-assertion": "off",
        "@typescript-eslint/no-explicit-any": "off",
        "@typescript-eslint/no-inferrable-types": "off",
        "@typescript-eslint/no-empty-function": "off",

        "@typescript-eslint/no-unused-vars": ["error", {
            ignoreRestSiblings: true,
        }],

        "react/prop-types": "off",
        "react/no-unescaped-entities": "off",
        "jsx-a11y/no-autofocus": "off",
        "sx-props/no-sx-props": "error",
    },

    settings: {
        react: {
            version: "detect",
        },
    },
}, globalIgnores(["**/rollup.config.js"])]);
