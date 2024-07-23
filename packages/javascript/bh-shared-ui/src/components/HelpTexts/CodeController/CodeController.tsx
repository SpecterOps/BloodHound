// Copyright 2024 Specter Ops, Inc.
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

import { Theme, Typography } from '@mui/material';
import { makeStyles } from '@mui/styles';
import { PropsWithChildren, useMemo, useRef, useState } from 'react';
import clsx from 'clsx';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faAlignJustify, faCopy } from '@fortawesome/free-solid-svg-icons';
import { copyToClipboard } from '../../../utils/copyToClipboard';
import { Button } from '@bloodhoundenterprise/doodleui';

export const useStyles = makeStyles((theme: Theme) => ({
    codeController: {
        position: 'relative',
        '& .code': {
            'white-space': 'pre',
            'overflow-x': 'scroll',
        },
        '& .wrapped': {
            'white-space': 'pre-line',
        },
        '& .codeController': {
            display: 'flex',
            justifyContent: 'flex-end',
            position: 'absolute',
            right: '0',
            borderBottom: `1px solid ${theme.palette.color.primary}`,
            width: '100%',

            '& button': {
                color: theme.palette.color.primary,
                transition: 'opacity 100ms ease-in-out',
                boxShadow: 'none',
                fontSize: theme.typography.body1,
                padding: theme.spacing(0.5, 1),
                height: 'fit-content',

                '&:last-of-type': {
                    marginRight: '20px',
                },
            },
        },
    },
}));

interface Props {
    hideCopy?: boolean;
    hideWrap?: boolean;
}

/**
 * Wraps <Typography component="pre"> to add controls for copy/paste and code wrapping.
 * Implementation: please wrap code block in {`<coding>`}
 * @param hideCopy - default false - display copy text button
 * @param hideWrap - default false - display wrap text button
 * @returns
 */
function CodeController(props: PropsWithChildren<Props>) {
    const { hideCopy = false, hideWrap = false, children } = props;

    const [wrapped, setWrapped] = useState(true);
    const [copied, setCopied] = useState(false);
    const [scrollLeft, setScrollLeft] = useState(false);
    const [scrollRight, setScrollRight] = useState(false);

    const codeRef = useRef<HTMLPreElement>(null);

    const classes = useStyles();

    const handleScroll = (e: React.UIEvent<HTMLPreElement>) => {
        const { scrollLeft, scrollWidth, clientWidth } = e.currentTarget;
        setScrollLeft(scrollLeft + clientWidth < scrollWidth);
        setScrollRight(scrollLeft > 0);
    };

    const justifiedLeft = useMemo(() => {
        const perLine = (children?.toString() ?? '').split('\n');
        const nextNonBlankLine = perLine.find((x, i) => i !== 0 && !!x.trim());
        const leftIndentationIndex = nextNonBlankLine?.split('').findIndex((x) => !!x.trim());

        return perLine
            ?.map((line, i) => {
                if (i === 0) return line;
                const leftIndentation = line.slice(0, leftIndentationIndex).trim();
                const content = line.slice(leftIndentationIndex);
                return leftIndentation + content;
            })
            .join('\n');
    }, [children]);

    const handleCopy = async () => {
        setCopied(true);

        await copyToClipboard(justifiedLeft);

        setTimeout(() => {
            setCopied(false);
        }, 3000);
    };

    const handleWrap = () => {
        setWrapped((prev) => {
            if (prev) setScrollLeft(true);
            return !prev;
        });
    };

    return (
        <div className={classes.codeController}>
            <Typography
                component='pre'
                className={clsx('code', {
                    wrapped,
                    scrollLeft: !wrapped && scrollLeft,
                    scrollRight: !wrapped && scrollRight,
                })}
                ref={codeRef}
                onScroll={handleScroll}>
                {(!hideCopy || !hideWrap) && (
                    <>
                        <div className='codeController'>
                            {!hideCopy && (
                                <Button variant='text' onClick={handleCopy}>
                                    <FontAwesomeIcon icon={faCopy} />
                                    <Typography component='span' sx={{ marginLeft: '6px' }}>
                                        {copied ? 'Copied' : 'Copy'}
                                    </Typography>
                                </Button>
                            )}
                            {!hideWrap && (
                                <Button variant='text' onClick={handleWrap}>
                                    <FontAwesomeIcon icon={faAlignJustify} />
                                    <Typography component='span' sx={{ marginLeft: '6px' }}>
                                        {wrapped ? 'Unwrap' : 'Wrap'}
                                    </Typography>
                                </Button>
                            )}
                        </div>
                        <br />
                        <br />
                    </>
                )}
                {justifiedLeft}
            </Typography>
        </div>
    );
}

export default CodeController;
