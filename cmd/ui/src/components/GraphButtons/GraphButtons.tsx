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

import { faCropAlt } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Menu, MenuItem } from '@mui/material';
import { useSigma } from '@react-sigma/core';
import { random } from 'graphology-layout';
import forceAtlas2 from 'graphology-layout-forceatlas2';
import isEmpty from 'lodash/isEmpty';
import { Children, FC, ReactNode, useState } from 'react';
import GraphButton from 'src/components/GraphButton';
import { GraphButtonProps } from 'src/components/GraphButton/GraphButton';
import { resetCamera } from 'src/ducks/graph/utils';
import { RankDirection, layoutDagre } from 'src/hooks/useLayoutDagre/useLayoutDagre';

interface GraphButtonsProps {
    rankDirection?: RankDirection;
    options?: GraphButtonOptions;
    nonLayoutButtons?: GraphButtonProps[];
}

export type GraphButtonOptions = {
    standard: boolean;
    sequential: boolean;
};

const GraphButtons: FC<GraphButtonsProps> = ({ rankDirection, options, nonLayoutButtons }) => {
    if (isEmpty(options)) options = { standard: false, sequential: false };
    const { standard, sequential } = options;

    const sigma = useSigma();
    const graph = sigma.getGraph();
    const { assign: assignDagre } = layoutDagre(
        {
            graph: {
                rankdir: rankDirection || RankDirection.LEFT_RIGHT,
                ranksep: rankDirection === RankDirection.LEFT_RIGHT ? 500 : 50,
            },
        },
        graph
    );

    const runSequentialLayout = (): void => {
        assignDagre();
        resetCamera(sigma);
    };

    const runStandardLayout = (): void => {
        random.assign(graph, { scale: 1000 });
        forceAtlas2.assign(graph, {
            iterations: 128,
            settings: {
                scalingRatio: 1000,
                barnesHutOptimize: true,
            },
        });
        resetCamera(sigma);
    };

    const reset = (): void => {
        resetCamera(sigma);
    };

    return (
        <Box position={'absolute'} bottom={16} display={'flex'}>
            <GraphButton onClick={reset} displayText={<FontAwesomeIcon icon={faCropAlt} />} />

            <MenuButton label='Layout'>
                {sequential && <MenuItem onClick={runSequentialLayout}>Sequential</MenuItem>}
                {standard && <MenuItem onClick={runStandardLayout}>Standard</MenuItem>}
            </MenuButton>

            <MenuButton label='Export'>
                <MenuItem onClick={() => {}}>JSON</MenuItem>
            </MenuButton>

            {nonLayoutButtons?.length && (
                <>
                    {nonLayoutButtons.map((props, index) => (
                        <GraphButton key={index} onClick={props.onClick} displayText={props.displayText} />
                    ))}
                </>
            )}
        </Box>
    );
};

const MenuButton: FC<{ label: string; children: ReactNode }> = (props) => {
    const { label, children } = props;
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

    const open = Boolean(anchorEl);

    const handleClose = () => setAnchorEl(null);

    return (
        <>
            <GraphButton
                onClick={(event: React.MouseEvent<HTMLButtonElement>) => {
                    setAnchorEl(event.currentTarget);
                }}
                aria-controls={open ? `${label}-menu` : undefined}
                aria-haspopup='true'
                aria-expanded={open ? 'true' : undefined}
                displayText={label}></GraphButton>
            <Menu
                id={`${label}-menu`}
                open={open}
                anchorEl={anchorEl}
                onClose={handleClose}
                MenuListProps={{
                    'aria-labelledby': `${label}-button`,
                }}>
                {Children.map(children, (child) => {
                    return <div onClick={handleClose}>{child}</div>;
                })}
            </Menu>
        </>
    );
};

export default GraphButtons;
