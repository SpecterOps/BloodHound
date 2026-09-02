// Copyright 2023 Specter Ops, Inc.
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

import { History, Transition } from 'history';
import { useCallback, useContext, useEffect } from 'react';
import { UNSAFE_NavigationContext as NavigationContext, Navigator } from 'react-router-dom';

// Navigator functionality has been removed in 6 so this adds some
// of the features we would use back in
// https://stackoverflow.com/a/71587163/8046487
// https://gist.github.com/rmorse/426ffcc579922a82749934826fa9f743

type ExtendNavigator = Navigator & Pick<History, 'block'>;

export function useBlocker(blocker: (tx: Transition) => void, when = true) {
    const { navigator } = useContext(NavigationContext);

    useEffect(() => {
        if (!when) return;

        const unblock = (navigator as ExtendNavigator).block((tx) => {
            const autoUnblockingTx = {
                ...tx,
                retry() {
                    unblock();
                    tx.retry();
                },
            };

            blocker(autoUnblockingTx);
        });

        return unblock;
    }, [navigator, blocker, when]);
}

export default function usePrompt(message: string, when = true) {
    const blocker = useCallback(
        (tx: Transition) => {
            if (window.confirm(message)) tx.retry();
        },
        [message]
    );

    useBlocker(blocker, when);
}
