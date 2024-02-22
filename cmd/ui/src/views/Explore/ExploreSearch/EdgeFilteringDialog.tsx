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

import { useTheme } from '@mui/material/styles';
import {
    Box,
    Button,
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
    Typography,
} from '@mui/material';
import { useState } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faChevronDown, faChevronUp } from '@fortawesome/free-solid-svg-icons';
import { pathFiltersSaved } from 'src/ducks/searchbar/actions';
import { AllEdgeTypes, Category, Subcategory } from 'bh-shared-ui';
import { useAppDispatch, useAppSelector } from 'src/store';

interface EdgeFilteringDialogProps {
    isOpen: boolean;
    handleCancel: () => void;
    handleApply: () => void;
}

const EdgeFilteringDialog = ({ isOpen, handleCancel, handleApply }: EdgeFilteringDialogProps) => {
    const selectedFilters: EdgeCheckboxType[] = useAppSelector((state) => state.search.pathFilters);

    const onCancel = () => {
        handleCancel();
    };

    const onApply = () => {
        handleApply();
    };

    const title = 'Path Edge Filtering';
    const description = 'Select the edge types to include in your pathfinding search.';

    return (
        <Dialog open={isOpen} fullWidth maxWidth={'md'}>
            <DialogTitle>{title}</DialogTitle>
            <Divider sx={{ ml: 1, mr: 1 }} />
            <Typography variant='subtitle1' ml={3} mt={1}>
                {description}
            </Typography>

            <DialogContent>
                <CategoryList selectedFilters={selectedFilters} />
            </DialogContent>

            <DialogActions>
                <Button onClick={onCancel}>Cancel</Button>
                <Button onClick={onApply}>Apply</Button>
            </DialogActions>
        </Dialog>
    );
};

export type EdgeCheckboxType = {
    category: string;
    subcategory: string;
    edgeType: string;
    checked: boolean;
};

interface CategoryListProps {
    selectedFilters: Array<EdgeCheckboxType>;
}

const CategoryList = ({ selectedFilters }: CategoryListProps) => {
    const dispatch = useAppDispatch();

    return (
        <List>
            {AllEdgeTypes.map((category: Category) => {
                const { categoryName } = category;
                return (
                    <CategoryListItem
                        key={categoryName}
                        category={category}
                        checked={selectedFilters}
                        setChecked={(checked: EdgeCheckboxType[]) => dispatch(pathFiltersSaved(checked))}
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
}

const CategoryListItem = ({ category, checked, setChecked }: CategoryListItemProps) => {
    const { categoryName, subcategories } = category;

    const categoryFilter = (element: EdgeCheckboxType) => element.category === categoryName;

    return (
        <IndeterminateListItem
            name={categoryName}
            checkboxFilter={categoryFilter}
            checked={checked}
            setChecked={setChecked}
            collapsibleContent={
                <List sx={{ pl: 2 }}>
                    {subcategories.map((subcategory) => {
                        return (
                            <SubcategoryListItem
                                key={subcategory.name}
                                checked={checked}
                                setChecked={setChecked}
                                subcategory={subcategory}
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
}

const SubcategoryListItem = ({ subcategory, checked, setChecked }: SubcategoryListItemProps) => {
    const { name, edgeTypes } = subcategory;

    const subcategoryFilter = (element: EdgeCheckboxType) => element.subcategory === name;

    return (
        <IndeterminateListItem
            name={name}
            checkboxFilter={subcategoryFilter}
            checked={checked}
            setChecked={setChecked}
            collapsibleContent={
                <List sx={{ pl: 4 }}>
                    <ListItem sx={{ display: 'block' }}>
                        <EdgesView edgeTypes={edgeTypes} checked={checked} setChecked={setChecked} />
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
        <Box bgcolor={theme.palette.grey[200]} p={1} borderRadius={1}>
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
}

const IndeterminateListItem = ({
    name,
    checkboxFilter,
    checked,
    setChecked,
    collapsibleContent,
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

    return (
        <>
            <ListItem
                disablePadding
                dense
                secondaryAction={
                    <IconButton
                        title={showCollapsibleContent ? `minimize-${name}` : `expand-${name}`}
                        onClick={toggleCollapsibleContent}>
                        <SvgIcon>
                            <FontAwesomeIcon icon={showCollapsibleContent ? faChevronUp : faChevronDown} />
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
            <Collapse in={showCollapsibleContent}>{collapsibleContent}</Collapse>
        </>
    );
};
export default EdgeFilteringDialog;
