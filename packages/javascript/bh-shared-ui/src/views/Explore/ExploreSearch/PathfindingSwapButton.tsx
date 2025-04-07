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
import { faExchangeAlt } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

type PathfindingSwapButtonProps = {
    disabled?: boolean;
    onSwapPathfindingInputs: () => void;
};
const PathfindingSwapButton = ({ disabled, onSwapPathfindingInputs }: PathfindingSwapButtonProps) => {
    return (
        <Button
            className='h-7 w-7 min-w-7 p-0 rounded-[4px] border-black/25 text-white'
            disabled={disabled}
            onClick={onSwapPathfindingInputs}>
            <FontAwesomeIcon icon={faExchangeAlt} className='fa-rotate-90' />
        </Button>
    );
};

export default PathfindingSwapButton;
