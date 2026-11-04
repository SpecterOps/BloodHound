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

function checkIsBoilerplate(parsedCodeChunk, parsed) {
    const isImportDeclaration = ts.isImportDeclaration(parsedCodeChunk);
    const isExpressionStatement = ts.isExpressionStatement(parsedCodeChunk);
    const isCallExpression = isExpressionStatement && ts.isCallExpression(parsedCodeChunk.expression);
    const isTestConfiguration =
        isCallExpression && parsedCodeChunk.expression.expression.getText(parsed) === 'test.use';
    const isAwaitExpression = isExpressionStatement && ts.isAwaitExpression(parsedCodeChunk.expression);
    const isAwaitedCall = isAwaitExpression && ts.isCallExpression(parsedCodeChunk.expression.expression);
    const isHarRoutingCall =
        isAwaitedCall && parsedCodeChunk.expression.expression.expression.getText(parsed).endsWith('.routeFromHAR');

    const isTestConfigurationStatement = isExpressionStatement && isCallExpression && isTestConfiguration;
    const isHarRoutingStatement = isExpressionStatement && isAwaitExpression && isAwaitedCall && isHarRoutingCall;
    const isBoilerplate = isImportDeclaration || isTestConfigurationStatement || isHarRoutingStatement;

    return isBoilerplate;
}

function makeRelativePageUrlNode(parsedCodeChunk, parsed, origin) {
    const isStringLiteral = ts.isStringLiteral(parsedCodeChunk);
    const hasMatchingOrigin = isStringLiteral && parsedCodeChunk.text.startsWith(`${origin}/`);
    const isCallArgument = hasMatchingOrigin && ts.isCallExpression(parsedCodeChunk.parent);
    const isPageNavigation = isCallArgument && parsedCodeChunk.parent.expression.getText(parsed).endsWith('.goto');
    const isMatchingPageUrl = isStringLiteral && hasMatchingOrigin && isCallArgument && isPageNavigation;

    if (isMatchingPageUrl) {
        const relativeUrl = parsedCodeChunk.text.slice(origin.length);
        const relativePageUrlNode = ts.factory.createStringLiteral(relativeUrl);

        return relativePageUrlNode;
    }
}

function makeRenamedSpecCall(parsedCodeChunk, parsed, name, visit) {
    const isCallExpression = ts.isCallExpression(parsedCodeChunk);
    const isTestCall = isCallExpression && parsedCodeChunk.expression.getText(parsed) === 'test';
    const hasFirstArgument = isTestCall && parsedCodeChunk.arguments[0] !== undefined;
    const hasStringName = hasFirstArgument && ts.isStringLiteral(parsedCodeChunk.arguments[0]);
    const isNamedTestCall = isCallExpression && isTestCall && hasFirstArgument && hasStringName;

    if (isNamedTestCall) {
        const renamedSpecCall = ts.factory.updateCallExpression(
            parsedCodeChunk,
            parsedCodeChunk.expression,
            parsedCodeChunk.typeArguments,
            [
                ts.factory.createStringLiteral(name),
                ...parsedCodeChunk.arguments.slice(1).map((argument) => ts.visitNode(argument, visit)),
            ]
        );

        return renamedSpecCall;
    }
}

function makeTransformedSpec(parsedSpecFile, originUrl, fileName) {
    const transformedSpec = ts.transform(parsedSpecFile, [
        (context) => {
            const visit = (parsedCodeChunk) => {
                const isBoilerplate = checkIsBoilerplate(parsedCodeChunk, parsedSpecFile);

                if (isBoilerplate) {
                    // do not pack boilerplate into new spec
                    return undefined;
                }

                const relativePageUrlNode = makeRelativePageUrlNode(parsedCodeChunk, parsedSpecFile, originUrl);

                if (relativePageUrlNode !== undefined) {
                    return relativePageUrlNode;
                }

                const renamedSpecCall = makeRenamedSpecCall(parsedCodeChunk, parsedSpecFile, fileName, visit);

                if (renamedSpecCall !== undefined) {
                    return renamedSpecCall;
                }

                const transformedCodeChunk = ts.visitEachChild(parsedCodeChunk, visit, context);

                return transformedCodeChunk;
            };
            return (parsedCodeChunk) => ts.visitNode(parsedCodeChunk, visit);
        },
    ]);

    return transformedSpec;
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
    const transformedSpec = makeTransformedSpec(parsedSpecSyntaxTree, origin, name);

    try {
        const specSyntaxTree = transformedSpec.transformed[0];
        const specActionsString = makeSpecString(specSyntaxTree);
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
        transformedSpec.dispose();
    }
}
