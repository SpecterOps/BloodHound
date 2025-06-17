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

import { useMemo, useState } from 'react';
import { SelectorsInfoPanelContext } from './SelectorsInfoPanelContext';

type SelectorsInfoPanelContextProviderProps = { children: React.ReactNode };

export function SelectorsInfoPanelContextProvider({ children }: SelectorsInfoPanelContextProviderProps) {
    const [isSelectorsInfoPanelOpen, setIsSelectorsInfoPanelOpen] = useState<boolean>(true);
    const value = useMemo(
        () => ({
            isSelectorsInfoPanelOpen,
            setIsSelectorsInfoPanelOpen,
        }),
        [isSelectorsInfoPanelOpen]
    );
    return <SelectorsInfoPanelContext.Provider value={value}>{children}</SelectorsInfoPanelContext.Provider>;
}
