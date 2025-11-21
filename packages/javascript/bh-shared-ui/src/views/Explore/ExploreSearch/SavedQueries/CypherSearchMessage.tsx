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
import { ReactNode } from 'react';
import { cn } from '../../../../utils';

export type MessageState = {
    showMessage: boolean;
    message?: ReactNode;
};

export type CypherSearchMessageProps = {
    messageState: {
        showMessage: boolean;
        message?: ReactNode;
    };
    setMessageState: React.Dispatch<React.SetStateAction<MessageState>>;
};

const CypherSearchMessage = (props: CypherSearchMessageProps) => {
    const { messageState, setMessageState } = props;
    const { message } = messageState;

    return (
        <div
            onAnimationEnd={() => {
                setMessageState((prev) => ({
                    ...prev,
                    showMessage: false,
                }));
            }}
            onTransitionEnd={(animationEvent) => {
                const element = animationEvent.target as HTMLElement;
                if (!element.className.includes('__message-still-visible')) {
                    setMessageState(() => ({
                        message: '',
                        showMessage: false,
                    }));
                }
            }}
            className={cn('w-full pr-1', {
                'animate-[null-animation_4s]': messageState.showMessage,
            })}>
            <div
                role='status'
                aria-live='polite'
                onAnimationEnd={(e) => e.stopPropagation()}
                className={cn('leading-none animate-in fade-in duration-300 scale-90 opacity-0', {
                    'opacity-100 scale-100 __message-still-visible': messageState.showMessage,
                })}>
                {message}
            </div>
        </div>
    );
};

export default CypherSearchMessage;
