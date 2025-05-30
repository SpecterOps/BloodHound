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
import { cn } from '../../utils/theme';

interface ManageColumnsProps {
    open: boolean;
    onClose: () => void;
}

const ManageColumns: React.FC<ManageColumnsProps> = (props) => {
    const { open, onClose } = props;
    return (
        <div className={cn({ hidden: !open })}>
            Manage Columns
            <Button onClick={onClose}>X</Button>
        </div>
    );
};

export default ManageColumns;
