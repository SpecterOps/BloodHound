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

export const FileUploadDialogContext = createContext<{
    setShowFileIngestDialog: React.Dispatch<React.SetStateAction<boolean>>;
    showFileIngestDialog: boolean;
}>({
    setShowFileIngestDialog: () => {},
    showFileIngestDialog: false,
});

export const FileUploadDialogProvider = ({ children }: { children: ReactNode }) => {
    const [showFileIngestDialog, setShowFileIngestDialog] = useState(false);

    const value = {
        showFileIngestDialog,
        setShowFileIngestDialog,
    };

    return <FileUploadDialogContext.Provider value={value}>{children}</FileUploadDialogContext.Provider>;
};

export const useFileUploadDialogContext = () => useContext(FileUploadDialogContext);
