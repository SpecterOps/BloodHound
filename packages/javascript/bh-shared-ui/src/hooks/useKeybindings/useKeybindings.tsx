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
import { useCallback, useEffect } from 'react';
import { NavigateFunction, useNavigate } from 'react-router-dom';

type KeyBindingCallbackOptions = {
    navigate: NavigateFunction;
};

interface KeyBindings extends Record<string, KeyBindings | ((options: KeyBindingCallbackOptions) => void)> {}

type KeyBindingsWithShift = { shift?: KeyBindings } & KeyBindings;

export const useKeybindings = (bindings: KeyBindingsWithShift = {}) => {
    const navigate = useNavigate();
    const handleKeyDown = useCallback(
        (event: KeyboardEvent) => {
            if (event.altKey && !event.metaKey) {
                event.preventDefault();

                if (event.shiftKey && !bindings.shift) {
                    return;
                }

                const bindingsMap: KeyBindingsWithShift = event.shiftKey && bindings.shift ? bindings.shift : bindings;

                const key = event.code;
                const func = bindingsMap[key] || bindingsMap[key?.toLowerCase()];

                if (typeof func === 'function') {
                    func({
                        navigate,
                    });
                }
            }
        },
        [bindings, navigate]
    );

    useEffect(() => {
        document.addEventListener('keydown', handleKeyDown);

        return () => {
            document.removeEventListener('keydown', handleKeyDown);
        };
    }, [handleKeyDown]);
};

export default useKeybindings;
