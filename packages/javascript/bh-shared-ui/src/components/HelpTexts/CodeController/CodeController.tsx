import { Box, Button, Typography } from '@mui/material';
import { makeStyles } from '@mui/styles';
import { PropsWithChildren, useRef, useState } from 'react';
import clsx from 'clsx';

export const useStyles = makeStyles((theme) => ({
    codeController: {
        position: 'relative',
        'text-wrap': 'nowrap',
        '& .code': {
            'text-wrap': 'nowrap',
            overflow: 'scroll',
        },
        '& .wrapped': {
            'text-wrap': 'wrap',
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
        },
        '& button': {
            color: 'white',
            transition: 'opacity 100ms ease-in-out',
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
 * @param hideCopy - default true - display copy text button
 * @param hideWrap - default true - display wrap text button
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

    const handleCopy = async () => {
        setCopied(true);
        // Trims off the white space at the beginning and end of new lines
        const justifiedLeft = (children?.toString() ?? '')
            .split('\n')
            .map((s) => s.trim())
            .join('\n');

        await navigator.clipboard.writeText(justifiedLeft);

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
                                    {copied ? 'Copied!' : 'Copy'}
                                </Button>
                            )}
                            {!hideWrap && (
                                <Button
                                    sx={{ p: 0.5, m: 0, marginRight: '20px', fontSize: '12px' }}
                                    onClick={handleWrap}>
                                    {wrapped ? 'Unwrap' : 'Wrap'}
                                </Button>
                            )}
                        </div>
                        <br />
                        <br />
                    </>
                )}
                {children}
            </Typography>
        </div>
    );
}

export default CodeController;
