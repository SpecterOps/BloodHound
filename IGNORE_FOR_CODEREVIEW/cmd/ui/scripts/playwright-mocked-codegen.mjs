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

export function sanitizeRecording(recording, storageState) {
    const document = structuredClone(recording);
    const secrets = storageState.origins.flatMap(({ localStorage }) =>
        localStorage.flatMap(({ name, value }) => {
            if (name !== 'persistedState') return [];
            const token = JSON.parse(value).auth?.sessionToken;
            return token ? [token] : [];
        })
    );
    for (const entry of document.log.entries) {
        // HAR replay breaks ties using matching headers. Keep authenticated requests
        // distinct from anonymous requests without retaining real credentials.
        entry.request.headers = (entry.request.headers ?? []).map((header) =>
            /^authorization$/i.test(header.name) ? { ...header, value: 'Bearer playwright-mock-session' } : header
        );
        for (const message of [entry.request, entry.response]) {
            message.headers = (message.headers ?? []).filter(
                ({ name }) =>
                    !/^(cookie|set-cookie)$/i.test(name) &&
                    (message === entry.request || !/^authorization$/i.test(name))
            );
            message.cookies = [];
        }
    }
    let json = JSON.stringify(document);
    for (const secret of secrets) json = json.replaceAll(secret, 'playwright-mock-session');
    return JSON.parse(json);
}

export function makeMockedTest(source, name, baseUrl, licenseHeader) {
    // Use the TS parser so storage-state configuration and quoted URLs are handled regardless
    // of the quote style or formatting emitted by the installed Playwright recorder.
    const parsed = ts.createSourceFile('recording.ts', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
    const origin = new URL(baseUrl).origin;
    const transformed = ts.transform(parsed, [
        (context) => {
            const visit = (node) => {
                if (ts.isImportDeclaration(node)) return undefined;
                if (
                    ts.isExpressionStatement(node) &&
                    ts.isAwaitExpression(node.expression) &&
                    ts.isCallExpression(node.expression.expression) &&
                    node.expression.expression.expression.getText(parsed).endsWith('.routeFromHAR')
                )
                    return undefined;
                if (
                    ts.isExpressionStatement(node) &&
                    ts.isCallExpression(node.expression) &&
                    node.expression.expression.getText(parsed) === 'test.use'
                )
                    return undefined;
                if (
                    ts.isStringLiteral(node) &&
                    node.text.startsWith(`${origin}/`) &&
                    ts.isCallExpression(node.parent) &&
                    node.parent.expression.getText(parsed).endsWith('.goto')
                ) {
                    return ts.factory.createStringLiteral(node.text.slice(origin.length));
                }
                if (
                    ts.isCallExpression(node) &&
                    node.expression.getText(parsed) === 'test' &&
                    node.arguments[0] &&
                    ts.isStringLiteral(node.arguments[0])
                ) {
                    return ts.factory.updateCallExpression(node, node.expression, node.typeArguments, [
                        ts.factory.createStringLiteral(name),
                        ...node.arguments.slice(1).map((arg) => ts.visitNode(arg, visit)),
                    ]);
                }
                return ts.visitEachChild(node, visit, context);
            };
            return (node) => ts.visitNode(node, visit);
        },
    ]);
    const actions = ts.createPrinter().printFile(transformed.transformed[0]);
    transformed.dispose();
    return `${licenseHeader}
import { test, expect } from 'bh-playwright-testing/mocked';

test.use({ replayHar: new URL('./${name}.har', import.meta.url) });

${actions}`;
}
