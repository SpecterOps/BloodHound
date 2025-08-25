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
import { useRef } from 'react';
import { Transition } from 'react-transition-group';
import { SNACKBAR_DURATION } from '../../../../constants';
type CypherSearchMessageProps = {
    // message?: string;
    // showMessage: boolean;
    messageState: {
        showMessage: boolean;
        message?: string;
    };
    clearMessage: () => void;
};

const CypherSearchMessage = (props: CypherSearchMessageProps) => {
    const { clearMessage, messageState } = props;
    const { showMessage, message } = messageState;

    const nodeRef = useRef(null);

    const duration = 300;

    const defaultStyle = {
        transition: `opacity ${duration}ms ease-in-out, transform ${duration}ms`,
        opacity: 0,
        transform: 'scale(1)',
    };

    const transitionStyles: { [key: string]: React.CSSProperties } = {
        entering: { opacity: 1, transform: 'translateX(0) scale(0.96)' },
        entered: { opacity: 1 },
        exiting: { opacity: 0, transform: 'scale(0.9)' },
        exited: { opacity: 0 },
    };

    const handleEntered = () => {
        setTimeout(() => {
            clearMessage();
        }, SNACKBAR_DURATION);
    };

    return (
        <div className='w-full'>
            <Transition nodeRef={nodeRef} in={showMessage} timeout={duration} onEntered={handleEntered}>
                {(state) => (
                    <div
                        ref={nodeRef}
                        style={{
                            ...defaultStyle,
                            ...transitionStyles[state],
                        }}
                        className='leading-none'>
                        {message}
                    </div>
                )}
            </Transition>
        </div>
    );
};

export default CypherSearchMessage;
