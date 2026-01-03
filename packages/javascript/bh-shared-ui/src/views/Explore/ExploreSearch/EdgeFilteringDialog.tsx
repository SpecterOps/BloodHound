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
import { faChevronDown, faChevronUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    Box,
    Checkbox,
    Collapse,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Divider,
    FormControlLabel,
    Grid,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    SvgIcon,
    TextField,
    Typography,
    useTheme,
} from '@mui/material';
import { useState } from 'react';
import { AllEdgeTypes, Category, EdgeCheckboxType, Subcategory } from '../../../edgeTypes';

interface EdgeFilteringDialogProps {
    selectedFilters: EdgeCheckboxType[];
    isOpen: boolean;
    handleApply: () => void;
    handleUpdate: (checked: EdgeCheckboxType[]) => void;
    handleCancel: () => void;
}

const EdgeFilteringDialog = ({
    selectedFilters,
    isOpen,
    handleApply,
    handleUpdate,
    handleCancel,
}: EdgeFilteringDialogProps) => {
    const [searchQuery, setSearchQuery] = useState('');
    const title = 'Path Edge Filtering';
    const description = 'Select the edge types to include in your pathfinding search.';

    const handleClose = () => {
        handleCancel();
    };

    const handleApplyClick = () => {
        handleApply();
    };

    const handleExited = () => {
        setSearchQuery('');
    };

    return (
        <Dialog open={isOpen} fullWidth maxWidth={'md'} onClose={handleClose} TransitionProps={{ onExited: handleExited }}>
            <DialogTitle>{title}</DialogTitle>
            <Divider sx={{ ml: 1, mr: 1 }} />
            <Typography variant='subtitle1' ml={3} mt={1}>
                {description}
            </Typography>

            <DialogContent>
                <TextField
                    fullWidth
                    placeholder='Search edges...'
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    sx={{ mb: 2 }}
                />
                <CategoryList
                    selectedFilters={selectedFilters}
                    handleUpdate={handleUpdate}
                    searchQuery={searchQuery}
                />
            </DialogContent>

            <DialogActions>
                <Button variant='tertiary' onClick={handleClose}>
                    Cancel
                </Button>
                <Button onClick={handleApplyClick}>Apply</Button>
            </DialogActions>
        </Dialog>
    );
};

interface CategoryListProps {
    selectedFilters: Array<EdgeCheckboxType>;
    handleUpdate: (checked: EdgeCheckboxType[]) => void;
    searchQuery: string;
}

const CategoryList = ({ selectedFilters, handleUpdate, searchQuery }: CategoryListProps) => {
    const filterCategory = (category: Category): boolean => {
        if (!searchQuery) return true;

        const query = searchQuery.toLowerCase();

        // Check if any edge type matches
        const hasMatchingEdge = category.subcategories.some((subcategory) => {
            return subcategory.edgeTypes.some((edgeType) =>
                edgeType.toLowerCase().includes(query)
            );
        });

        return hasMatchingEdge;
    };

    return (
        <List>
            {AllEdgeTypes.filter(filterCategory).map((category: Category) => {
                const { categoryName } = category;
                return (
                    <CategoryListItem
                        key={categoryName}
                        category={category}
                        checked={selectedFilters}
                        setChecked={handleUpdate}
                        searchQuery={searchQuery}
                    />
                );
            })}
        </List>
    );
};

interface CategoryListItemProps {
    category: Category;
    checked: EdgeCheckboxType[];
    setChecked: (checked: EdgeCheckboxType[]) => void;
    searchQuery: string;
}

const CategoryListItem = ({ category, checked, setChecked, searchQuery }: CategoryListItemProps) => {
    const { categoryName, subcategories } = category;

    const categoryFilter = (element: EdgeCheckboxType) => element.category === categoryName;

    const filterSubcategory = (subcategory: Subcategory): boolean => {
        if (!searchQuery) return true;

        const query = searchQuery.toLowerCase();
        const hasMatchingEdge = subcategory.edgeTypes.some((edgeType) => edgeType.toLowerCase().includes(query));

        return hasMatchingEdge;
    };

    return (
        <IndeterminateListItem
            name={categoryName}
            checkboxFilter={categoryFilter}
            checked={checked}
            setChecked={setChecked}
            forceExpand={!!searchQuery}
            collapsibleContent={
                <List sx={{ pl: 2 }}>
                    {subcategories.filter(filterSubcategory).map((subcategory) => {
                        return (
                            <SubcategoryListItem
                                key={subcategory.name}
                                checked={checked}
                                setChecked={setChecked}
                                subcategory={subcategory}
                                searchQuery={searchQuery}
                            />
                        );
                    })}
                </List>
            }
        />
    );
};

interface SubcategoryListItemProps {
    subcategory: Subcategory;
    checked: EdgeCheckboxType[];
    setChecked: (checked: EdgeCheckboxType[]) => void;
    searchQuery: string;
}

const SubcategoryListItem = ({ subcategory, checked, setChecked, searchQuery }: SubcategoryListItemProps) => {
    const { name, edgeTypes } = subcategory;

    const subcategoryFilter = (element: EdgeCheckboxType) => element.subcategory === name;

    const filteredEdgeTypes = searchQuery
        ? edgeTypes.filter((edgeType) => edgeType.toLowerCase().includes(searchQuery.toLowerCase()))
        : edgeTypes;

    return (
        <IndeterminateListItem
            name={name}
            checkboxFilter={subcategoryFilter}
            checked={checked}
            setChecked={setChecked}
            forceExpand={!!searchQuery}
            collapsibleContent={
                <List sx={{ pl: 4 }}>
                    <ListItem sx={{ display: 'block' }}>
                        <EdgesView edgeTypes={filteredEdgeTypes} checked={checked} setChecked={setChecked} />
                    </ListItem>
                </List>
            }
        />
    );
};

interface EdgesViewProps {
    edgeTypes: string[];

    checked: EdgeCheckboxType[];
    setChecked: (checked: EdgeCheckboxType[]) => void;
}

const EdgesView = ({ edgeTypes, checked, setChecked }: EdgesViewProps) => {
    const theme = useTheme();

    const changeCheckbox = (event: React.ChangeEvent<HTMLInputElement>, edgeType: string) => {
        const newChecked = [...checked];
        const indexToUpdate = newChecked.findIndex((element) => element.edgeType === edgeType);
        newChecked[indexToUpdate] = { ...newChecked[indexToUpdate], checked: event.target.checked };

        setChecked(newChecked);
    };

    return (
        <Box bgcolor={theme.palette.neutral.tertiary} p={1} borderRadius={1}>
            <Grid container spacing={2}>
                {edgeTypes.map((edgeType, index) => {
                    return (
                        <Grid item xs={6} sm={4} key={index}>
                            <FormControlLabel
                                label={edgeType}
                                control={
                                    <Checkbox
                                        inputProps={{ 'aria-label': edgeType }}
                                        name={edgeType}
                                        checked={checked.find((element) => element.edgeType === edgeType)?.checked}
                                        onChange={(e) => changeCheckbox(e, edgeType)}
                                    />
                                }
                            />
                        </Grid>
                    );
                })}
            </Grid>
        </Box>
    );
};

interface IndeterminateListItemProps {
    name: string;

    checkboxFilter: (element: EdgeCheckboxType) => boolean;

    checked: EdgeCheckboxType[];
    setChecked: (checked: EdgeCheckboxType[]) => void;

    collapsibleContent: React.ReactNode;
    forceExpand?: boolean;
}

const IndeterminateListItem = ({
    name,
    checkboxFilter,
    checked,
    setChecked,
    collapsibleContent,
    forceExpand = false,
}: IndeterminateListItemProps) => {
    const [showCollapsibleContent, setShowCollapsibleContent] = useState(false);

    const checkboxes = checked.filter(checkboxFilter);
    const elementIsChecked = (element: EdgeCheckboxType) => element.checked;
    const elementIsEqual = (element: EdgeCheckboxType, index: number, arr: EdgeCheckboxType[]) =>
        element.checked === arr[0].checked;

    const changeAllChildren = (value: boolean) => {
        const newChecked = [...checked];
        newChecked.forEach((checkbox, index, arr) => {
            if (checkboxFilter(checkbox)) {
                arr[index] = { ...checkbox, checked: value };
            }
        });

        setChecked(newChecked);
    };

    const onCheckboxClick = () => {
        const isChecked = checkboxes.every((element) => element.checked);
        // if item is checked, uncheck all children
        if (isChecked) {
            changeAllChildren(false);
        } else {
            // if item is not checked, check all children
            changeAllChildren(true);
        }
    };

    const toggleCollapsibleContent = () => setShowCollapsibleContent((v) => !v);

    const isExpanded = forceExpand || showCollapsibleContent;

    return (
        <>
            <ListItem
                disablePadding
                dense
                secondaryAction={
                    <IconButton
                        title={isExpanded ? `minimize-${name}` : `expand-${name}`}
                        onClick={toggleCollapsibleContent}>
                        <SvgIcon>
                            <FontAwesomeIcon icon={isExpanded ? faChevronUp : faChevronDown} />
                        </SvgIcon>
                    </IconButton>
                }>
                <ListItemButton onClick={toggleCollapsibleContent}>
                    <ListItemIcon>
                        <Checkbox
                            onClick={(e) => {
                                e.stopPropagation();
                                onCheckboxClick();
                            }}
                            inputProps={{
                                'aria-label': name,
                            }}
                            checked={checkboxes.every(elementIsChecked)}
                            indeterminate={!checkboxes.every(elementIsEqual)}
                        />
                    </ListItemIcon>
                    <ListItemText>{name}</ListItemText>
                </ListItemButton>
            </ListItem>
            <Collapse in={isExpanded}>{collapsibleContent}</Collapse>
        </>
    );
};
export default EdgeFilteringDialog;
