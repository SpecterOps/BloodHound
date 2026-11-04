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

function shouldRemoveBoilerplate(parsedCodeChunk, parsed) {
    if (ts.isImportDeclaration(parsedCodeChunk)) return true;
    if (
        ts.isExpressionStatement(parsedCodeChunk) &&
        ts.isCallExpression(parsedCodeChunk.expression) &&
        parsedCodeChunk.expression.expression.getText(parsed) === 'test.use'
    )
        return true;
    return (
        ts.isExpressionStatement(parsedCodeChunk) &&
        ts.isAwaitExpression(parsedCodeChunk.expression) &&
        ts.isCallExpression(parsedCodeChunk.expression.expression) &&
        parsedCodeChunk.expression.expression.expression.getText(parsed).endsWith('.routeFromHAR')
    );
}

function makeRelativePageUrlNode(parsedCodeChunk, parsed, origin) {
    if (
        ts.isStringLiteral(parsedCodeChunk) &&
        parsedCodeChunk.text.startsWith(`${origin}/`) &&
        ts.isCallExpression(parsedCodeChunk.parent) &&
        parsedCodeChunk.parent.expression.getText(parsed).endsWith('.goto')
    ) {
        const relativeUrl = parsedCodeChunk.text.slice(origin.length);
        return ts.factory.createStringLiteral(relativeUrl);
    }
}

function makeRenamedSpecCall(parsedCodeChunk, parsed, name, visit) {
    if (
        ts.isCallExpression(parsedCodeChunk) &&
        parsedCodeChunk.expression.getText(parsed) === 'test' &&
        parsedCodeChunk.arguments[0] &&
        ts.isStringLiteral(parsedCodeChunk.arguments[0])
    ) {
        return ts.factory.updateCallExpression(
            parsedCodeChunk,
            parsedCodeChunk.expression,
            parsedCodeChunk.typeArguments,
            [
                ts.factory.createStringLiteral(name),
                ...parsedCodeChunk.arguments.slice(1).map((argument) => ts.visitNode(argument, visit)),
            ]
        );
    }
}

function transformSpecSyntaxTree(parsedSpecFile, originUrl, fileName) {
    const specTransformation = ts.transform(parsedSpecFile, [
        (context) => {
            const visit = (parsedCodeChunk) => {
                if (shouldRemoveBoilerplate(parsedCodeChunk, parsedSpecFile)) return undefined;
                return (
                    makeRelativePageUrlNode(parsedCodeChunk, parsedSpecFile, originUrl) ??
                    makeRenamedSpecCall(parsedCodeChunk, parsedSpecFile, fileName, visit) ??
                    ts.visitEachChild(parsedCodeChunk, visit, context)
                );
            };
            return (parsedCodeChunk) => ts.visitNode(parsedCodeChunk, visit);
        },
    ]);

    return specTransformation;
}

function makeSpecString(specSyntaxTree) {
    const specPrinter = ts.createPrinter();
    const specString = specPrinter.printFile(specSyntaxTree);

    return specString;
}

export function makeMockedSpecString(source, name, baseUrl, licenseHeader) {
    const parsedSpecSyntaxTree = ts.createSourceFile(
        'browserActions.ts',
        source,
        ts.ScriptTarget.Latest,
        true,
        ts.ScriptKind.TS
    );
    const origin = new URL(baseUrl).origin;
    const specTransformation = transformSpecSyntaxTree(parsedSpecSyntaxTree, origin, name);

    try {
        const transformedSpecSyntaxTree = specTransformation.transformed[0];
        const specActionsString = makeSpecString(transformedSpecSyntaxTree);
        const mockedSpecString = [
            licenseHeader,
            "import { test, expect } from 'bh-playwright-testing/har';",
            '',
            `test.use({ replayHar: new URL('./${name}.har', import.meta.url) });`,
            '',
            specActionsString,
        ].join('\n');

        return mockedSpecString;
    } finally {
        specTransformation.dispose();
    }
}
