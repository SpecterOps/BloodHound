// Copyright 2025 Specter Ops, Inc.
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

// .eslint-rules/rules/no-sx-prop.js
/* eslint-env node */
module.exports = {
  meta: {
    type: "problem",
    docs: {
      description: "Disallow the use of the `sx` prop.",
      category: "Possible Errors",
      recommended: false,
    },
    schema: [], // No options for this rule
  },
  create(context) {
    return {
      JSXAttribute(node) {
        if (node.name.name === "sx") {
          context.report({
            node: node,
            message:
              "The `sx` prop is not allowed. Please use tailwind classes instead.",
          });
        }
      },
    };
  },
};
