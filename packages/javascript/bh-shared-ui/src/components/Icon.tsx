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

import { IconButton, Theme, Tooltip } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import withStyles from '@mui/styles/withStyles';
import { FC, MouseEvent, PropsWithChildren } from 'react';

interface Props {
    tip?: string;
    tipPlacment?: 'left' | 'bottom';
    click?: (event: MouseEvent) => void;
    size?: 'small' | 'medium';
    badge?: number;
    className?: string;
}

const useStyles = makeStyles<Theme, { overflow: boolean }>({
    icon: {
        position: 'relative',
    },
    iconMedium: {
        borderRadius: 0,
        height: '40px',
    },
    iconSmall: {
        borderRadius: 0,
        padding: 0,
    },
    badge: (props) => ({
        position: 'absolute',
        bottom: '3px',
        right: '3px',
        background: '#ae1dff',
        color: '#fff',
        width: '20px',
        height: '20px',
        lineHeight: props.overflow ? '21px' : '20px',
        fontSize: props.overflow ? '10px' : '12px',
        borderRadius: '10px',
    }),
});

const LightTooltip = withStyles((theme) => ({
    tooltip: {
        backgroundColor: theme.palette.common.white,
        color: 'rgba(0, 0, 0, 0.87)',
        boxShadow: theme.shadows[1],
        fontSize: 12,
        fontWeight: 'normal',
    },
}))(Tooltip);

const Icon: FC<PropsWithChildren<Props>> = ({
    tip,
    tipPlacment = 'bottom',
    click,
    children,
    size,
    badge = 0,
    className,
}): JSX.Element => {
    const overflow: boolean = badge > 99;
    const badgeText: string | null = overflow ? '99+' : badge > 0 ? badge.toString() : null;
    const styles = useStyles({ overflow });

    let iconClass = size ? (size === 'small' ? styles.iconSmall : styles.iconMedium) : styles.iconMedium;
    iconClass += ` ${styles.icon} ${className}`;

    const icon = (
        <IconButton className={`${iconClass} icon`} onClick={click} size='large'>
            {children}
            {badgeText && <Badge text={badgeText} overflow={overflow} />}
        </IconButton>
    );

    return tip ? (
        <LightTooltip title={tip} placement={tipPlacment}>
            {icon}
        </LightTooltip>
    ) : (
        icon
    );
};

const Badge: FC<{ text: string; overflow?: boolean }> = ({ text, overflow = false }): JSX.Element => {
    const styles = useStyles({ overflow });
    return <span className={styles.badge}>{text}</span>;
};

export default Icon;
