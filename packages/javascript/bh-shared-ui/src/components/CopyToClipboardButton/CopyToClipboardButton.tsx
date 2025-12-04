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
import { faCheck, faCopy } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { AnimationEvent, KeyboardEvent, MouseEvent, useState } from 'react';
import { cn, copyToClipboard } from '../../utils';
import { adaptClickHandlerToKeyDown } from '../../utils/adaptClickHandlerToKeyDown';

type CopyToClipboardButtonProps = {
    onAnimationStart?: (e: AnimationEvent<HTMLDivElement>) => void;
    onAnimationEnd?: (e: AnimationEvent<HTMLDivElement>) => void;
    transitionDelay?: string;
    value: string | Array<any>;
};

const DEFAULT_TRANSITION_DELAY = 'delay-300';

export const CopyToClipboardButton = ({
    onAnimationEnd = () => {},
    onAnimationStart = () => {},
    transitionDelay = DEFAULT_TRANSITION_DELAY,
    value,
}: CopyToClipboardButtonProps) => {
    const [displayCopyCheckmark, setDisplayCopyCheckmark] = useState(false);
    const handleCopyToClipBoard = <T extends KeyboardEvent<HTMLElement> | MouseEvent<HTMLElement>>(e: T) => {
        e.stopPropagation(); // prevents the click event bubbling up the DOM and triggering the row click handler
        if (typeof value === 'string') {
            copyToClipboard(value);
        } else {
            copyToClipboard(value.join(', '));
        }
        setDisplayCopyCheckmark(true);
    };

    return (
        <>
            <div
                role='button'
                tabIndex={0}
                onClick={handleCopyToClipBoard}
                onKeyDown={adaptClickHandlerToKeyDown(handleCopyToClipBoard)}
                onAnimationStart={(animationEvent) => {
                    const element = animationEvent.target as HTMLElement;

                    if (element?.role === 'button') {
                        onAnimationStart(animationEvent);
                    }
                }}
                aria-label='Copy to clipboard'
                onAnimationEnd={(animationEvent) => {
                    const element = animationEvent.target as HTMLElement;

                    if (element?.role === 'button') {
                        setDisplayCopyCheckmark(false);
                        onAnimationEnd(animationEvent);
                    }
                }}
                className={cn(
                    'cursor-pointer absolute top-1/2 left-2 -translate-x-1/2 -translate-y-1/2 opacity-0 pr-1 group-hover:opacity-100 transition-opacity ease-in',
                    transitionDelay,
                    {
                        'animate-[null-animation_1s]': displayCopyCheckmark,
                    }
                )}>
                {displayCopyCheckmark && (
                    <FontAwesomeIcon
                        icon={faCheck}
                        onAnimationStart={(e) => e.stopPropagation()}
                        onAnimationEnd={(e) => e.stopPropagation()}
                        className='animate-in fade-in duration-300'
                    />
                )}
                {!displayCopyCheckmark && (
                    <FontAwesomeIcon
                        icon={faCopy}
                        onAnimationStart={(e) => e.stopPropagation()}
                        onAnimationEnd={(e) => e.stopPropagation()}
                        className='animate-in fade-in duration-300'
                    />
                )}
            </div>
            <span className={cn('group-hover:pl-5 transition-[padding-left] ease-in', transitionDelay)}>{value}</span>
        </>
    );
};
