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

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Tooltip } from '@mui/material';
import { Theme } from '@mui/material/styles';
import makeStyles from '@mui/styles/makeStyles';
import { EntityKinds, IconDictionary, IconInfo, NODE_ICON, UNKNOWN_ICON } from 'bh-shared-ui';
import { useAppSelector } from 'src/store';

interface NodeIconProps {
    nodeType: EntityKinds | string;
}

interface IconInfoProp {
    icon: IconInfo;
}

const useStyles = makeStyles<Theme, IconInfoProp, string>({
    root: {
        display: 'inline-block',
        marginRight: '4px',
        position: 'relative',
    },
    container: {
        backgroundColor: (props) => props.icon?.color || '#FFFFFF',
        border: '1px solid #000000',
        padding: '2px',
        borderRadius: '50%',
        height: '22px',
        width: '22px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontSize: '14px',
        color: '#000000DD',
    },
});

const NodeIcon: React.FC<NodeIconProps> = ({ nodeType }) => {
    const customIcons = useAppSelector((state) => state.global.customNodeInformation.customIcons);
    const icon = GetIconInfo(nodeType, customIcons);

    const classes = useStyles({ icon });

    return (
        <Tooltip title={nodeType || ''} describeChild={true}>
            <Box className={classes.root}>
                <Box className={classes.container}>
                    <FontAwesomeIcon icon={icon.icon} transform='shrink-2' />
                </Box>
            </Box>
        </Tooltip>
    );
};

export const GetIconInfo = (iconName: string, customIcons: IconDictionary): IconInfo => {
    if (iconName in customIcons) {
        return customIcons[iconName];
    }

    if (iconName in NODE_ICON) {
        return NODE_ICON[iconName];
    }

    return UNKNOWN_ICON;
};

export default NodeIcon;
