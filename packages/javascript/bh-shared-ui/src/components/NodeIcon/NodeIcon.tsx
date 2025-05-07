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

import { faQuestion, IconName } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Tooltip } from '@mui/material';
import { Theme } from '@mui/material/styles';
import makeStyles from '@mui/styles/makeStyles';
import { lazy } from 'react';
import { EntityKinds } from '../..';
import { CUSTOM_ICONS, NODE_ICON } from '../../utils/icons';

const LazyIcon = lazy(() => import('../CustomIcon'));

interface NodeIconProps {
    nodeType: EntityKinds | string;
}

const useStyles = makeStyles<Theme, NodeIconProps, string>({
    root: {
        display: 'inline-block',
        marginRight: '4px',
        position: 'relative',
    },
    container: {
        backgroundColor: (props) => NODE_ICON[props.nodeType]?.color || '#FFFFFF',
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
    let iconElement: JSX.Element = <FontAwesomeIcon icon={faQuestion} transform='shrink-2' />;

    if (nodeType in CUSTOM_ICONS) {
        try {
            let iconName = CUSTOM_ICONS[nodeType] as IconName;
            iconElement = <LazyIcon icon={iconName} transform='shrink-2' />;
        } catch (e) {
            console.log(e);
        }
    } else if (nodeType in NODE_ICON) {
        iconElement = <FontAwesomeIcon icon={NODE_ICON[nodeType].icon} transform='shrink-2' />;
    }

    return (
        <Tooltip title={nodeType || ''} describeChild={true}>
            <Box className={classes.root}>
                <Box className={classes.container}>{iconElement}</Box>
            </Box>
        </Tooltip>
    );
};

export default NodeIcon;
