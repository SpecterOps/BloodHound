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

import { faEyeSlash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Tooltip } from '@mui/material';

const HiddenEntityIconRootStyles = 'inline-block relative';

const HiddenEntityIconContainerStyles =
    'flex border border-neutral-dark-1 rounded-full p-1 w-[22px] h-[22px] items-center justify-center text-sm text-neutral-dark-1 bg-neutral-light-1 ';

const HiddenEntityIcon: React.FC = () => {
    return (
        <Tooltip title='Hidden' describeChild={true}>
            <Box className={HiddenEntityIconRootStyles}>
                <Box className={HiddenEntityIconContainerStyles}>
                    <FontAwesomeIcon icon={faEyeSlash} transform='shrink-2' />
                </Box>
            </Box>
        </Tooltip>
    );
};

export default HiddenEntityIcon;
