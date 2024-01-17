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

import { Box, Divider, Drawer, List, ListItem, ListItemText, ListSubheader } from '@mui/material';
import { Theme } from '@mui/material/styles';
import createStyles from '@mui/styles/createStyles';
import makeStyles from '@mui/styles/makeStyles';
import find from 'lodash/find';
import findIndex from 'lodash/findIndex';
import React from 'react';
import { Link, useLocation } from 'react-router-dom';

const drawerWidth = 320;

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        drawer: {
            width: drawerWidth,
            flexShrink: 0,
        },
        drawerPaper: {
            paddingTop: '50px',
            width: drawerWidth,
            border: 'none',
        },
        drawerContainer: {
            overflow: 'auto',
        },
        list: {
            padding: theme.spacing(1),
        },
        listItem: {
            borderRadius: theme.shape.borderRadius,
            marginBottom: theme.spacing(0.5),
        },
        listItemTextPrimary: {
            fontSize: '0.875rem',
        },
        listSubheader: {
            fontSize: '1rem',
        },
        listSubheaderSelected: {
            color: theme.palette.primary.main,
        },
        listItemSelected: {
            color: theme.palette.primary.main,
        },
    })
);

const LeftNavWithContent: React.FC<{
    sections: {
        title: string;
        items: {
            path: string;
            label: string;
        }[];
    }[];
}> = ({ sections }) => {
    const classes = useStyles();
    const location = useLocation();
    const selectedSectionIndex = findIndex(
        sections,
        (section) => find(section.items, (item) => item.path === location.pathname) !== undefined
    );

    return (
        <Drawer
            className={classes.drawer}
            variant='permanent'
            classes={{
                paper: classes.drawerPaper,
            }}>
            <nav className={classes.drawerContainer}>
                <List className={classes.list}>
                    {sections.map((section, sectionIndex) => (
                        <React.Fragment key={sectionIndex}>
                            <Box mb={2}>
                                <ListSubheader
                                    className={`${
                                        selectedSectionIndex === sectionIndex ? classes.listSubheaderSelected : ''
                                    }`}
                                    classes={{ root: classes.listSubheader }}>
                                    {section.title}
                                </ListSubheader>
                                {section.items.map((item, itemIndex) => (
                                    <ListItem
                                        classes={{
                                            root: classes.listItem,
                                            selected: classes.listItemSelected,
                                        }}
                                        button
                                        key={itemIndex}
                                        component={Link}
                                        to={item.path}
                                        selected={location.pathname === item.path}>
                                        <ListItemText
                                            primary={item.label}
                                            classes={{
                                                primary: classes.listItemTextPrimary,
                                            }}
                                        />
                                    </ListItem>
                                ))}
                            </Box>

                            {sectionIndex !== sections.length - 1 && (
                                <Box mb={2}>
                                    <Divider />
                                </Box>
                            )}
                        </React.Fragment>
                    ))}
                </List>
            </nav>
        </Drawer>
    );
};

export default LeftNavWithContent;
