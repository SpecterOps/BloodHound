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
import * as ScrollAreaPrimitive from '@radix-ui/react-scroll-area';
import * as React from 'react';
import { cn } from '../utils';

interface ScrollbarProps extends React.ComponentPropsWithoutRef<typeof ScrollAreaPrimitive.Root> {
    scrollbarWidth?: number;
    thumbHeight?: number;
    scrollbarColor?: string;
    thumbColor?: string;
}

const ScrollArea = React.forwardRef<React.ElementRef<typeof ScrollAreaPrimitive.Root>, ScrollbarProps>(
    ({ className, scrollbarWidth, thumbHeight, scrollbarColor, thumbColor, children, ...props }, ref) => (
        <>
            <ScrollAreaPrimitive.Root ref={ref} type='auto' {...props} className={cn('overflow-hidden', className)}>
                <ScrollAreaViewport
                    scrollbarWidth={scrollbarWidth}
                    thumbHeight={thumbHeight}
                    scrollbarColor={scrollbarColor}
                    thumbColor={thumbColor}
                    {...props}
                    ref={ref}>
                    {children}
                </ScrollAreaViewport>
            </ScrollAreaPrimitive.Root>
        </>
    )
);
ScrollArea.displayName = 'ScrollArea';

const ScrollAreaViewport = React.forwardRef<React.ElementRef<typeof ScrollAreaPrimitive.Viewport>, ScrollbarProps>(
    ({ children, scrollbarWidth, thumbHeight, scrollbarColor, thumbColor, ...props }, ref) => {
        const baseThumbStyle = {
            backgroundColor: thumbColor ? thumbColor : '',
        };
        const thumbStyle = thumbHeight
            ? {
                  ...baseThumbStyle,
                  height: `${thumbHeight}px`,
              }
            : {
                  ...baseThumbStyle,
              };

        const thumbStyleHorizontal = thumbHeight
            ? {
                  ...baseThumbStyle,
                  width: `${thumbHeight}px`,
              }
            : {
                  ...baseThumbStyle,
              };

        return (
            <>
                <ScrollAreaPrimitive.Viewport ref={ref} className={cn('w-full h-full')} {...props}>
                    {children}
                </ScrollAreaPrimitive.Viewport>
                <ScrollAreaScrollbar
                    style={{
                        width: scrollbarWidth ? `${scrollbarWidth}px` : '6px',
                        backgroundColor: scrollbarColor ? scrollbarColor : 'transparent',
                    }}
                    className='z-10'
                    orientation='vertical'>
                    <ScrollAreaThumb
                        style={{
                            ...thumbStyle,
                        }}
                    />
                </ScrollAreaScrollbar>
                <ScrollAreaScrollbar
                    style={{
                        height: scrollbarWidth ? `${scrollbarWidth}px` : '6px',
                        backgroundColor: scrollbarColor ? scrollbarColor : 'transparent',
                    }}
                    orientation='horizontal'>
                    <ScrollAreaThumb
                        style={{
                            height: '100%',
                            ...thumbStyleHorizontal,
                        }}
                    />
                </ScrollAreaScrollbar>
                <ScrollAreaCorner />
            </>
        );
    }
);
ScrollAreaViewport.displayName = 'ScrollAreaViewport';

const ScrollAreaScrollbar = React.forwardRef<
    React.ElementRef<typeof ScrollAreaPrimitive.Scrollbar>,
    React.ComponentPropsWithoutRef<typeof ScrollAreaPrimitive.Scrollbar>
>(({ className, children, ...props }, ref) => (
    <ScrollAreaPrimitive.Scrollbar
        ref={ref}
        className={cn(
            'transition-all data-[state=hidden]:animate-fade-out data-[state=visible]:animate-fade-in',
            className
        )}
        orientation='vertical'
        {...props}>
        {children}
    </ScrollAreaPrimitive.Scrollbar>
));
ScrollAreaScrollbar.displayName = 'ScrollAreaScrollbar';

const ScrollAreaThumb = React.forwardRef<
    React.ElementRef<typeof ScrollAreaPrimitive.Thumb>,
    React.ComponentPropsWithoutRef<typeof ScrollAreaPrimitive.Thumb>
>(({ className, ...props }, ref) => (
    <ScrollAreaPrimitive.Thumb ref={ref} className={cn('flex-1 bg-neutral-400 rounded-[10px]', className)} {...props} />
));
ScrollAreaThumb.displayName = 'ScrollAreaThumb';

const ScrollAreaCorner = ScrollAreaPrimitive.Corner;

export { ScrollArea, ScrollAreaCorner, ScrollAreaScrollbar, ScrollAreaThumb, ScrollAreaViewport };
