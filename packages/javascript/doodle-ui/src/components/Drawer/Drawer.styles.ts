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

import { cn } from '../utils';

export const drawerOverlayClasses = cn(
    'cn-drawer-overlay fixed inset-0 z-[1410] min-h-dvh bg-[color:var(--drawer-overlay-fill)] opacity-[max(var(--drawer-overlay-min-opacity,0),calc(1-var(--drawer-swipe-progress)))] transition-opacity',
    'motion-reduce:transition-none',
    '[transition-duration:450ms] [transition-timing-function:cubic-bezier(0.32,0.72,0,1)] select-none',
    'data-[ending-style]:pointer-events-none data-[ending-style]:opacity-0 data-[ending-style]:[transition-duration:calc(var(--drawer-swipe-strength)*400ms)]',
    'data-[snap-points]:[--drawer-overlay-min-opacity:0.5]',
    'data-[starting-style]:opacity-0',
    'data-[swiping]:[transition-duration:0ms]'
);

export const drawerSwipeHandleClasses = cn(
    'cn-drawer-swipe-handle relative z-10 mx-auto my-2 h-1.5 w-12 shrink-0 cursor-grab rounded-full bg-neutral-4 transition-opacity duration-200',
    'group-data-[swipe-axis=x]/drawer-popup:mx-2 group-data-[swipe-axis=x]/drawer-popup:my-auto group-data-[swipe-axis=x]/drawer-popup:h-12 group-data-[swipe-axis=x]/drawer-popup:w-1.5',
    'group-data-[nested-drawer-open]/drawer-popup:opacity-0',
    'group-data-[nested-drawer-swiping]/drawer-popup:opacity-100',
    'group-data-[swipe-direction=left]/drawer-popup:order-last',
    'group-data-[swipe-direction=up]/drawer-popup:order-last',
    'active:cursor-grabbing',
    'motion-reduce:transition-none'
);

export const drawerViewportClasses =
    'pointer-events-none fixed inset-0 z-[1500] select-none data-[modal=true]:pointer-events-auto';

export const drawerPopupClasses = cn(
    // Base.
    'cn-drawer-popup group/drawer-popup pointer-events-auto fixed z-[1500] bg-neutral-1 text-main shadow-xl m-[var(--drawer-inset,0px)] flex h-[var(--drawer-content-height)] max-h-[var(--drawer-content-max-height,none)] min-h-0 w-[var(--drawer-content-width,auto)] [transform:translate3d(var(--translate-x,0px),var(--translate-y,0px),0)_scale(var(--stack-scale))] flex-col transition-[transform,height,opacity,filter]',
    'motion-reduce:transition-none',
    '[transition-duration:450ms] [transition-timing-function:cubic-bezier(0.22,1,0.36,1)] will-change-transform outline-none select-none [interpolate-size:allow-keywords]',
    // Nested.
    'data-[nested-drawer-open]:overflow-hidden data-[nested-drawer-open]:brightness-95',
    // Bleed.
    'after:pointer-events-none after:absolute after:[background-color:var(--drawer-bleed-background,var(--neutral-1))]',
    'data-[swipe-axis=x]:after:inset-y-0 data-[swipe-axis=x]:after:w-[var(--bleed)]',
    'data-[swipe-axis=y]:after:inset-x-0 data-[swipe-axis=y]:after:h-[var(--bleed)]',
    'data-[swipe-direction=down]:after:top-full',
    'data-[swipe-direction=left]:after:right-full',
    'data-[swipe-direction=right]:after:left-full',
    'data-[swipe-direction=up]:after:bottom-full',
    // Sizing.
    '[--drawer-content-height:var(--drawer-height,auto)]',
    'data-[swipe-axis=x]:[--drawer-content-width:75%]',
    'data-[swipe-axis=y]:[--drawer-content-max-height:calc(100dvh-6rem)]',
    'data-[swipe-axis=x]:sm:[--drawer-content-width:24rem]',
    // Snap points.
    'data-[swipe-axis=y]:data-[snap-points]:[--drawer-content-height:100dvh] data-[swipe-axis=y]:data-[snap-points]:[--drawer-content-max-height:100dvh]',
    // Stack.
    '[--bleed:3rem] [--peek:1rem] [--stack-height:var(--drawer-frontmost-height,var(--drawer-height,0px))] [--stack-peek-offset:max(0px,calc((var(--nested-drawers)-var(--stack-progress))*var(--peek)))] [--stack-progress:clamp(0,var(--drawer-swipe-progress),1)] [--stack-scale-base:max(0,calc(1-(var(--nested-drawers)*var(--stack-step))))] [--stack-scale:clamp(0,calc(var(--stack-scale-base)+(var(--stack-step)*var(--stack-progress))),1)] [--stack-shrink:calc(1-var(--stack-scale))] [--stack-step:0.05]',
    // Transitions.
    'data-[ending-style]:[transform:var(--closed-transform)] data-[ending-style]:opacity-[0.9999] data-[ending-style]:[transition-duration:calc(var(--drawer-swipe-strength)*400ms)]',
    'data-[nested-drawer-swiping]:[transition-duration:0ms]',
    'data-[ending-style]:data-[nested-drawer-swiping]:[transition-duration:calc(var(--drawer-swipe-strength)*400ms)]',
    'data-[starting-style]:[transform:var(--closed-transform)]',
    'data-[swiping]:[transition-duration:0ms]',
    'data-[ending-style]:data-[swiping]:[transition-duration:calc(var(--drawer-swipe-strength)*400ms)]',
    // Axis: y.
    'data-[swipe-axis=y]:inset-x-0 data-[swipe-axis=y]:data-[nested-drawer-open]:h-[var(--stack-height)]',
    // Axis: x.
    'data-[swipe-axis=x]:inset-y-0 data-[swipe-axis=x]:flex-row',
    // Direction: down.
    'data-[swipe-direction=down]:bottom-0 data-[swipe-direction=down]:origin-bottom data-[swipe-direction=down]:[--closed-transform:translate3d(0,calc(100%+var(--drawer-inset,0px)+2px),0)] data-[swipe-direction=down]:[--translate-y:calc(var(--drawer-snap-point-offset,0px)+var(--drawer-swipe-movement-y)-var(--stack-peek-offset)-(var(--stack-shrink)*var(--stack-height)))]',
    // Direction: up.
    'data-[swipe-direction=up]:top-0 data-[swipe-direction=up]:origin-top data-[swipe-direction=up]:[--closed-transform:translate3d(0,calc(-100%-var(--drawer-inset,0px)-2px),0)] data-[swipe-direction=up]:[--translate-y:calc(var(--drawer-snap-point-offset,0px)+var(--drawer-swipe-movement-y)+var(--stack-peek-offset)+(var(--stack-shrink)*var(--stack-height)))]',
    // Direction: left.
    'data-[swipe-direction=left]:left-0 data-[swipe-direction=left]:origin-left data-[swipe-direction=left]:[--closed-transform:translate3d(calc(-100%-var(--drawer-inset,0px)-2px),0,0)] data-[swipe-direction=left]:[--translate-x:calc(var(--drawer-swipe-movement-x)+var(--stack-peek-offset)+(var(--stack-shrink)*100%))]',
    // Direction: right.
    'data-[swipe-direction=right]:right-0 data-[swipe-direction=right]:origin-right data-[swipe-direction=right]:[--closed-transform:translate3d(calc(100%+var(--drawer-inset,0px)+2px),0,0)] data-[swipe-direction=right]:[--translate-x:calc(var(--drawer-swipe-movement-x)-var(--stack-peek-offset)-(var(--stack-shrink)*100%))]'
);

export const drawerInnerContentClasses = cn(
    'cn-drawer-content-base flex min-h-0 flex-1 flex-col gap-4 px-3 py-4 overflow-hidden overscroll-contain rounded-[inherit] transition-opacity duration-300 [transition-timing-function:cubic-bezier(0.45,1.005,0,1.005)]',
    'motion-reduce:transition-none',
    'select-text',
    'group-data-[nested-drawer-open]/drawer-popup:opacity-0',
    'group-data-[nested-drawer-swiping]/drawer-popup:opacity-100',
    'group-data-[swiping]/drawer-popup:select-none'
);

export const drawerHeaderClasses = 'cn-drawer-header-base flex shrink-0 items-start justify-between gap-4';

export const drawerFooterClasses = 'cn-drawer-footer-base mt-auto flex shrink-0';

export const drawerDescriptionClasses = 'cn-drawer-description text-balance text-muted';

export const drawerBodyClasses = 'min-h-0 flex-1 overflow-y-auto pr-3 -mr-3';
