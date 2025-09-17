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

//list of ids to be excluded from Quick Ingest -- useExecuteOnFileDrag
export enum QuickUploadExclusionIds {
    ImportQueryDialog = 'import-query-dialog',
    DefaultNoDataDialog = 'default-no-data-import-query-dialog',
}

export const getExcludedIds = () => {
    const ids = Object.values(QuickUploadExclusionIds);

    for (const id of ids) {
        const element = document.getElementById(id);
        if (element) {
            return true;
        }
    }
    return false;
};
