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
import { ReactNode, useCallback, useEffect, useRef } from 'react';
import { SNACKBAR_DURATION } from '../../../../constants';
import { cn } from '../../../../utils';
export type CypherSearchMessageProps = {
    messageState: {
        showMessage: boolean;
        message?: ReactNode;
    };
    clearMessage: () => void;
};

const CypherSearchMessage = (props: CypherSearchMessageProps) => {
    const { clearMessage, messageState } = props;
    const { showMessage, message } = messageState;
    const timeoutRef = useRef<number | undefined>(undefined);

    const startTimer = useCallback(() => {
        if (timeoutRef.current) {
            clearTimeout(timeoutRef.current);
        }
        timeoutRef.current = window.setTimeout(() => {
            clearMessage();
        }, SNACKBAR_DURATION);
    }, [clearMessage]);

    const clearTimer = useCallback(() => {
        if (timeoutRef.current) {
            clearTimeout(timeoutRef.current);
        }
    }, []);

    useEffect(() => {
        if (showMessage) {
            startTimer();
        } else {
            clearTimer();
        }

        return () => {
            if (timeoutRef.current) {
                clearTimeout(timeoutRef.current);
            }
        };
    }, [clearMessage, showMessage, startTimer, clearTimer]);

    return (
        <div className='w-full pr-1'>
            <div
                role='status'
                aria-live='polite'
                className={cn('leading-none opacity-0 scale-90 transition-all duration-300 ease-in-out', {
                    'opacity-100 scale-100 transition-all duration-300 ease-in-out': showMessage,
                })}>
                {message}
            </div>
        </div>
    );
};

export default CypherSearchMessage;
