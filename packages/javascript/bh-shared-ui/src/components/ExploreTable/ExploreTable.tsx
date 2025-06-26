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

import { Button } from '@bloodhoundenterprise/doodleui';
import React from 'react';
import useToggle from '../../hooks/useToggle';
import { cn } from '../../utils/theme';
import ManageColumns from './ManageColumns';

interface ExploreTableProps {
    open: boolean;
    onClose: () => void;
}

const ExploreTable: React.FC<ExploreTableProps> = (props) => {
    const { open, onClose } = props;

    const [openManageColumns, toggleOpenManageColumns] = useToggle(false);

    if (!open) return null;

    return (
        <div className='absolute bottom-4 left-4 right-4 h-1/2 bg-pink-300 flex justify-center items-center'>
            <div className={cn({ hidden: openManageColumns })}>
                <Button onClick={toggleOpenManageColumns}>Manage Columns</Button>
                <Button onClick={onClose}>CLOSE</Button>
            </div>
            <ManageColumns open={openManageColumns} onClose={toggleOpenManageColumns} />
        </div>
    );
};

export default ExploreTable;
