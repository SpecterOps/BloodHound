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

import { Button } from '@bloodhoundenterprise/doodleui';
import { faFilter } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useState } from 'react';
import EdgeFilteringDialog from './EdgeFilteringDialog';
import { usePathfindingFilterSwitch } from './switches';

const EdgeFilter = () => {
    const [isOpenDialog, setIsOpenDialog] = useState(false);

    const { selectedFilters, initialize, handleApplyFilters, handleUpdateFilters, handleCancelFilters } =
        usePathfindingFilterSwitch();

    return (
        <>
            <Button
                className={'h-7 w-7 min-w-7 p-0 rounded-[4px] border-black/25 text-white'}
                onClick={() => {
                    setIsOpenDialog(true);
                    // what is the initial state of edge filters?  save it
                    initialize();
                }}>
                <FontAwesomeIcon icon={faFilter} />
            </Button>
            <EdgeFilteringDialog
                isOpen={isOpenDialog}
                selectedFilters={selectedFilters}
                handleApply={() => {
                    setIsOpenDialog(false);
                    handleApplyFilters();
                }}
                handleUpdate={handleUpdateFilters}
                handleCancel={() => {
                    setIsOpenDialog(false);
                    handleCancelFilters();
                }}
            />
        </>
    );
};

export default EdgeFilter;
