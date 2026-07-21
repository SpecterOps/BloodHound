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
import { Divider } from '@mui/material';
import { Alert, Link, Typography } from 'doodle-ui';
import React, { useMemo } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import ReactMarkdown, { Components } from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Theme, useTheme } from '../hooks';

const ErrorFallback = () => {
    return (
        <div className='py-4 flex justify-end' data-testid='error-boundary'>
            <Alert variant='error' title='Error'>
                An unexpected error has occurred. Please refresh the page and try again.
            </Alert>
        </div>
    );
};

const applyThemeToComponents = (theme: Theme) => {
    return {
        h1: ({ children }: any) => {
            return (
                <Typography variant='h1' style={{ margin: '1rem 0' }}>
                    {children}
                </Typography>
            );
        },
        h2: ({ children }: any) => {
            return (
                <Typography variant='h2' style={{ margin: '1rem 0' }}>
                    {children}
                </Typography>
            );
        },
        h3: ({ children }: any) => {
            return (
                <Typography variant='h3' style={{ margin: '1rem 0' }}>
                    {children}
                </Typography>
            );
        },
        h4: ({ children }: any) => {
            return (
                <Typography variant='h4' style={{ margin: '1rem 0' }}>
                    {children}
                </Typography>
            );
        },
        h5: ({ children }: any) => {
            return (
                <Typography variant='h5' style={{ margin: '1rem 0' }}>
                    {children}
                </Typography>
            );
        },
        h6: ({ children }: any) => {
            return (
                <Typography variant='h6' style={{ margin: '1rem 0' }}>
                    {children}
                </Typography>
            );
        },
        a: ({ node, ...props }: any) => (
            <Link to='#' color='primary' target='_blank' rel='noopener noreferrer' {...props} />
        ),
        blockquote: ({ node, ...props }: any) => <blockquote style={{ margin: '1rem 0' }} {...props} />,
        code: ({ node, inline, ...props }: any) => (
            <code
                style={{
                    backgroundColor: theme.neutral.quinary,
                    borderRadius: '4px',
                    padding: inline ? '0 0.25em' : '0',
                }}
                {...props}
            />
        ),
        em: ({ node, ...props }: any) => <em {...props} />,
        hr: ({ node, ...props }: any) => <Divider {...props} />,
        li: ({ node, ordered, ...props }: any) => <li {...props} />,
        ol: ({ node, ordered, ...props }: any) => (
            <ol
                style={{
                    paddingLeft: '1em',
                    marginBottom: '1em',
                    listStyle: 'decimal',
                }}
                {...props}
            />
        ),
        p: ({ node, ...props }: any) => <p style={{ margin: '1rem 0' }} {...props} />,
        pre: ({ node, ...props }: any) => (
            <pre
                style={{
                    fontSize: '0.875rem',
                    backgroundColor: theme.neutral.quinary,
                    padding: '1rem',
                    borderRadius: '4px',
                    overflow: 'auto',
                    margin: '1rem 0',
                }}
                {...props}
            />
        ),
        strong: ({ node, ...props }: any) => <strong {...props} />,
        ul: ({ node, ordered, ...props }: any) => (
            <ul
                style={{
                    paddingLeft: '1em',
                    marginBottom: '1em',
                    listStyle: 'disc',
                }}
                {...props}
            />
        ),
    };
};

export const useMarkdownComponents = () => {
    const theme = useTheme();

    return useMemo(() => applyThemeToComponents(theme), [theme]);
};

const MarkdownContent: React.FC<{
    markdown: string;
    components?: Components;
}> = ({ markdown, components }) => {
    const defaultComponents = useMarkdownComponents();
    return (
        <ReactMarkdown remarkPlugins={[remarkGfm]} components={components ?? defaultComponents}>
            {markdown}
        </ReactMarkdown>
    );
};

export default function ErrorBoundaryWrappedMarkdown({
    markdown,
    components,
}: {
    markdown: string;
    components?: Components;
}) {
    return (
        <ErrorBoundary fallbackRender={ErrorFallback}>
            <MarkdownContent markdown={markdown} components={components} />
        </ErrorBoundary>
    );
}
