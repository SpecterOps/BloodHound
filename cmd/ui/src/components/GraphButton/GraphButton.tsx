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

import { Button } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { FC } from 'react';
import { useCamera } from '@react-sigma/core';

const useStyles = makeStyles((theme) => ({
    button: {
        fontSize: '1rem',
        height: '1rem',
        lineHeight: '1rem',
        padding: theme.spacing(1.5),
        marginLeft: theme.spacing(2),
        marginTop: theme.spacing(2),
        border: 'none',
        boxSizing: 'initial',
        borderRadius: theme.shape.borderRadius,
        backgroundColor: theme.palette.background.paper,
        color: theme.palette.common.black,
        textTransform: 'capitalize',
        minWidth: 'initial',
        '&:hover': {
            backgroundColor: theme.palette.background.default,
            '@media (hover: none)': {
                backgroundColor: theme.palette.background.default,
            },
        },
    },
}));

export interface GraphButtonProps {
    onClick: () => void;
    displayText: string | JSX.Element;
}

const GraphButton: FC<GraphButtonProps> = ({ onClick, displayText }) => {
    const styles = useStyles();
    const { reset } = useCamera();

    return (
        <Button
            onClick={() => {
                onClick();
                reset();
            }}
            classes={{ root: styles.button }}>
            {displayText}
        </Button>
    );
};

export default GraphButton;
