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
import { createContext, ReactNode, useContext, useState } from 'react';

export const KeyboardShortcutsDialogContext = createContext<{
    setShowKeyboardShortcutsDialog: React.Dispatch<React.SetStateAction<boolean>>;
    showKeyboardShortcutsDialog: boolean;
}>({
    setShowKeyboardShortcutsDialog: () => {},
    showKeyboardShortcutsDialog: false,
});

export const KeyboardShortcutsDialogProvider = ({ children }: { children: ReactNode }) => {
    const [showKeyboardShortcutsDialog, setShowKeyboardShortcutsDialog] = useState(false);

    const value = {
        showKeyboardShortcutsDialog,
        setShowKeyboardShortcutsDialog,
    };

    return <KeyboardShortcutsDialogContext.Provider value={value}>{children}</KeyboardShortcutsDialogContext.Provider>;
};

export const useKeyboardShortcutsDialogContext = () => useContext(KeyboardShortcutsDialogContext);
