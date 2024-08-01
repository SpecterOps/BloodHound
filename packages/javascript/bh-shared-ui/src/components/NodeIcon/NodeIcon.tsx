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
import { NODE_ICON } from '../../utils/icons';
import { EntityKinds } from '../..';

interface NodeIconProps {
    nodeType: EntityKinds | string;
}

// hashes a string to a lightly shaded hex color
export const stringToLightHexColor = (str: string) => {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }

    let color = '#';
    for (let i = 0; i < 3; i++) {
        let value = (hash >> (i * 8)) & 0xff;
        value = Math.floor((value + 255) / 2);
        color += ('00' + value.toString(16)).substr(-2);
    }

    return color;
};

const useStyles = makeStyles<Theme, NodeIconProps, string>({
    root: {
        display: 'inline-block',
        marginRight: '4px',
        position: 'relative',
    },
    container: {
        backgroundColor: (props) => NODE_ICON[props.nodeType]?.color || stringToLightHexColor(props.nodeType),
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
    const classes = useStyles({ nodeType });

    return (
        <Tooltip title={nodeType || ''} describeChild={true}>
            <Box className={classes.root}>
                <Box className={classes.container}>
                    {NODE_ICON[nodeType]?.icon ? (
                        <FontAwesomeIcon icon={NODE_ICON[nodeType].icon} transform='shrink-2' />
                    ) : (
                        nodeType.at(0)?.toUpperCase()
                    )}
                </Box>
            </Box>
        </Tooltip>
    );
};

export default NodeIcon;
