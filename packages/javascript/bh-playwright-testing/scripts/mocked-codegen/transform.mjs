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

import ts from 'typescript';

function removeRecorderSetup(node, parsed) {
    if (ts.isImportDeclaration(node)) return true;
    if (
        ts.isExpressionStatement(node) &&
        ts.isCallExpression(node.expression) &&
        node.expression.expression.getText(parsed) === 'test.use'
    )
        return true;
    return (
        ts.isExpressionStatement(node) &&
        ts.isAwaitExpression(node.expression) &&
        ts.isCallExpression(node.expression.expression) &&
        node.expression.expression.expression.getText(parsed).endsWith('.routeFromHAR')
    );
}

function rewritePageUrl(node, parsed, origin) {
    if (
        ts.isStringLiteral(node) &&
        node.text.startsWith(`${origin}/`) &&
        ts.isCallExpression(node.parent) &&
        node.parent.expression.getText(parsed).endsWith('.goto')
    ) {
        return ts.factory.createStringLiteral(node.text.slice(origin.length));
    }
}

function rewriteTestName(node, parsed, name, visit) {
    if (
        ts.isCallExpression(node) &&
        node.expression.getText(parsed) === 'test' &&
        node.arguments[0] &&
        ts.isStringLiteral(node.arguments[0])
    ) {
        return ts.factory.updateCallExpression(node, node.expression, node.typeArguments, [
            ts.factory.createStringLiteral(name),
            ...node.arguments.slice(1).map((argument) => ts.visitNode(argument, visit)),
        ]);
    }
}

function transformRecording(parsed, origin, name) {
    const transformed = ts.transform(parsed, [
        (context) => {
            const visit = (node) => {
                if (removeRecorderSetup(node, parsed)) return undefined;
                return (
                    rewritePageUrl(node, parsed, origin) ??
                    rewriteTestName(node, parsed, name, visit) ??
                    ts.visitEachChild(node, visit, context)
                );
            };
            return (node) => ts.visitNode(node, visit);
        },
    ]);
    const actions = ts.createPrinter().printFile(transformed.transformed[0]);
    transformed.dispose();
    return actions;
}

export function makeMockedTest(source, name, baseUrl, licenseHeader) {
    const parsed = ts.createSourceFile('recording.ts', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
    const actions = transformRecording(parsed, new URL(baseUrl).origin, name);
    return `${licenseHeader}
import { test, expect } from 'bh-playwright-testing/har';

test.use({ replayHar: new URL('./${name}.har', import.meta.url) });

${actions}`;
}
