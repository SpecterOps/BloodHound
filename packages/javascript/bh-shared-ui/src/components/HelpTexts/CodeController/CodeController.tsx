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

import { Button, Typography } from '@mui/material';
import { makeStyles } from '@mui/styles';
import { PropsWithChildren, useMemo, useRef, useState } from 'react';
import clsx from 'clsx';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faAlignJustify, faCopy } from '@fortawesome/free-solid-svg-icons';
import { copyToClipboard } from '../../../utils/copyToClipboard';

export const useStyles = makeStyles(() => ({
    codeController: {
        position: 'relative',
        '& .code': {
            'white-space': 'pre',
            overflow: 'scroll',
        },
        '& .wrapped': {
            'white-space': 'pre-line',
        },
        '& .scrollLeft': {
            'box-shadow': 'inset -5px 0px 5px black;',
        },
        '& .scrollRight': {
            'box-shadow': 'inset 5px 0px 5px black;',
        },
        '& .scrollRight.scrollLeft': {
            'box-shadow': 'inset 5px 0px 5px 0px black, inset -5px 0px 5px 0px black',
        },
        '& .codeController': {
            display: 'flex',
            justifyContent: 'flex-end',
            position: 'absolute',
            right: '0',
            borderBottom: '1px solid white',
            width: '100%',

            '& button': {
                color: 'white',
                transition: 'opacity 100ms ease-in-out',
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

    // Trims off tab spacing at the beginning and end of new lines
    const justifiedLeft = useMemo(() => {
        const perLine = (children?.toString() ?? '').split('\n');
        const nextNonBlankLine = perLine.find((x, i) => i !== 0 && !!x.trim());

        const startingIndex = nextNonBlankLine?.split('').findIndex((x) => !!x.trim());
        return perLine?.map((x) => x.slice(startingIndex)).join('\n');
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
                                <Button sx={{ p: 0.5, m: 0, fontSize: '12px' }} onClick={handleCopy}>
                                    <FontAwesomeIcon icon={faCopy} />
                                    <Typography component='span' sx={{ marginLeft: '6px' }}>
                                        {copied ? 'Copied' : 'Copy'}
                                    </Typography>
                                </Button>
                            )}
                            {!hideWrap && (
                                <Button
                                    sx={{ p: 0.5, m: 0, marginRight: '20px', fontSize: '12px' }}
                                    onClick={handleWrap}>
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
