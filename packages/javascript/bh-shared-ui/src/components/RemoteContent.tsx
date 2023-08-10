// Copyright 2023 Specter Ops, Inc.
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

import { Skeleton } from '@mui/material';
import React, { useEffect, useState } from 'react';
import DOMPurify from 'dompurify';
import { Divider, Link, Typography } from '@mui/material';
import ReactMarkdown from 'react-markdown';

const getComponents = (baseURL?: string) => {
    const COMPONENTS = {
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
                component='code'
                style={{
                    backgroundColor: '#e1e1e1',
                    borderRadius: '4px',
                    padding: inline ? '0 0.25em' : '0',
                }}
                {...props}
            />
        ),
        em: ({ node, ...props }: any) => <em {...props} />,
        hr: ({ node, ...props }: any) => <Divider {...props} />,
        img: ({ node, ...props }: any) => {
            const imgSrc = baseURL ? `${baseURL}${props.src}` : props.src;
            return <img {...props} alt={props.alt} src={imgSrc} style={{ maxWidth: '100%' }} />;
        },
        li: ({ node, ordered, ...props }: any) => <li {...props} />,
        ol: ({ node, ordered, ...props }: any) => (
            <ol
                style={{
                    paddingLeft: '1em',
                    marginBottom: '1em',
                }}
                {...props}
            />
        ),
        p: ({ node, ...props }: any) => <p style={{ margin: '1rem 0' }} {...props} />,
        pre: ({ node, ...props }: any) => (
            <pre
                style={{
                    fontSize: '0.875rem',
                    backgroundColor: '#e1e1e1',
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
                }}
                {...props}
            />
        ),
    };

    return COMPONENTS;
};

const RemoteContent: React.FC<{
    url: string;
    baseURL?: string;
    markdown?: boolean;
    fallback?: string | React.ReactNode;
}> = ({ url, baseURL, markdown = false, fallback = 'An error has occurred.' }) => {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<any>(undefined);
    const [data, setData] = useState<string | undefined>(undefined);

    useEffect(() => {
        setLoading(true);
        setError('');
        fetch(url)
            .then((res) => res.text())
            .then((data) => {
                setData(data);
                setError('');
            })
            .catch((reason) => {
                setData(undefined);
                setError(reason);
            })
            .finally(() => {
                setLoading(false);
            });
    }, [url]);

    if (loading) return <Skeleton />;

    if (error || data?.charAt(0) === '<') return <p style={{ margin: '1rem 0' }}>{fallback}</p>;

    if (markdown) return <ReactMarkdown components={getComponents(baseURL)}>{data!}</ReactMarkdown>;

    return <div>{DOMPurify.sanitize(data!)}</div>;
};

export default RemoteContent;
