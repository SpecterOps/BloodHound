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

import React from 'react';

export const DisableQueryLimitContext = React.createContext<{
    setIsDisableQueryLimit: React.Dispatch<React.SetStateAction<boolean>>;
    isDisableQueryLimit: boolean;
}>({
    setIsDisableQueryLimit: () => {},
    isDisableQueryLimit: false,
});

export function useDisableQueryLimitContext() {
    const context = React.useContext(DisableQueryLimitContext);
    if (context === undefined) {
        throw new Error('useDisableQueryLimitContext must be used within a DisableQueryLimitContextProvider');
    }
    return context;
}
