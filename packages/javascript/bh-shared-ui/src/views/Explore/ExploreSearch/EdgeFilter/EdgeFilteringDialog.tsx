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

import { faAnglesDown, faAnglesUp, faChevronDown, faChevronUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, Checkbox, Dialog, DialogActions, DialogContent, DialogDescription, DialogTitle } from 'doodle-ui';
import { useEffect, useMemo, useRef, useState } from 'react';
import { SearchInput } from '../../../../components/SearchInput';
import { BUILTIN_EDGE_CATEGORIES, Category, EdgeCheckboxType, Subcategory } from './edgeCategories';
import { useEdgeCategories } from './useEdgeCategories';

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
    const title = 'Path Edge Filtering';
    const description = 'Select the edge types to include in your pathfinding search.';
    const [searchQuery, setSearchQuery] = useState('');
    const [collapseSignal, setCollapseSignal] = useState(0);
    const [expandSignal, setExpandSignal] = useState(0);
    const isSearching = searchQuery.trim().length > 0;

    return (
        <Dialog open={isOpen}>
            <DialogContent maxWidth='md' className='flex flex-col h-[80vh]'>
                <div className='flex items-start justify-between'>
                    <div>
                        <DialogTitle>{title}</DialogTitle>
                        <DialogDescription className='ml-1 mt-1'>{description}</DialogDescription>
                    </div>
                    <div className='flex gap-1'>
                        <Button
                            variant='text'
                            size='small'
                            disabled={isSearching}
                            onClick={() => setExpandSignal((s) => s + 1)}>
                            <FontAwesomeIcon icon={faAnglesDown} className='mr-1' />
                            Expand All
                        </Button>
                        <Button
                            variant='text'
                            size='small'
                            disabled={isSearching}
                            onClick={() => setCollapseSignal((s) => s + 1)}>
                            <FontAwesomeIcon icon={faAnglesUp} className='mr-1' />
                            Collapse All
                        </Button>
                    </div>
                </div>

                <div className='flex gap-2 mt-2'>
                    <SearchInput
                        id='edge-filter-search'
                        placeholder='Search edges...'
                        value={searchQuery}
                        onInputChange={setSearchQuery}
                        className='flex-1'
                    />
                </div>

                <div className='overflow-y-auto flex-1 mt-2'>
                    <CategoryList
                        selectedFilters={selectedFilters}
                        handleUpdate={handleUpdate}
                        searchQuery={searchQuery}
                        collapseSignal={collapseSignal}
                        expandSignal={expandSignal}
                    />
                </div>

                <DialogActions className='flex justify-end gap-2 pt-4'>
                    <Button variant='tertiary' onClick={handleCancel}>
                        Cancel
                    </Button>
                    <Button onClick={handleApply}>Apply</Button>
                </DialogActions>
            </DialogContent>
        </Dialog>
    );
};

interface CategoryListProps {
    selectedFilters: Array<EdgeCheckboxType>;
    handleUpdate: (checked: EdgeCheckboxType[]) => void;
    searchQuery: string;
    collapseSignal: number;
    expandSignal: number;
}

const CategoryList = ({
    selectedFilters,
    handleUpdate,
    searchQuery,
    collapseSignal,
    expandSignal,
}: CategoryListProps) => {
    const { isLoading, edgeCategories } = useEdgeCategories();

    const categories = isLoading ? BUILTIN_EDGE_CATEGORIES : edgeCategories;

    const filteredCategories = useMemo(() => {
        if (!searchQuery.trim()) return categories;

        const lowerQuery = searchQuery.toLowerCase();
        return categories
            .map((category) => ({
                ...category,
                subcategories: category.subcategories
                    .map((subcategory) => ({
                        ...subcategory,
                        edgeTypes: subcategory.edgeTypes.filter((edgeType) =>
                            edgeType.toLowerCase().includes(lowerQuery)
                        ),
                    }))
                    .filter((subcategory) => subcategory.edgeTypes.length > 0),
            }))
            .filter((category) => category.subcategories.length > 0);
    }, [categories, searchQuery]);

    return (
        <ul role='list'>
            {filteredCategories.map((category: Category) => {
                const { categoryName } = category;
                return (
                    <CategoryListItem
                        key={categoryName}
                        category={category}
                        checked={selectedFilters}
                        setChecked={handleUpdate}
                        defaultExpanded
                        collapseSignal={collapseSignal}
                        expandSignal={expandSignal}
                        searchQuery={searchQuery}
                    />
                );
            })}
        </ul>
    );
};

interface CategoryListItemProps {
    category: Category;
    checked: EdgeCheckboxType[];
    setChecked: (checked: EdgeCheckboxType[]) => void;
    defaultExpanded?: boolean;
    collapseSignal?: number;
    expandSignal?: number;
    searchQuery?: string;
}

const CategoryListItem = ({
    category,
    checked,
    setChecked,
    defaultExpanded = false,
    collapseSignal = 0,
    expandSignal = 0,
    searchQuery = '',
}: CategoryListItemProps) => {
    const { categoryName, subcategories } = category;

    const categoryFilter = (element: EdgeCheckboxType) => element.category === categoryName;

    const isSearching = searchQuery.trim().length > 0;
    const childDefaultExpanded = expandSignal > collapseSignal || (expandSignal === collapseSignal && defaultExpanded);

    return (
        <IndeterminateListItem
            name={categoryName}
            checkboxFilter={categoryFilter}
            checked={checked}
            setChecked={setChecked}
            defaultExpanded={defaultExpanded}
            collapseSignal={collapseSignal}
            expandSignal={expandSignal}
            forceExpanded={isSearching}
            collapsibleContent={
                <ul className='pl-4'>
                    {subcategories.map((subcategory) => {
                        return (
                            <SubcategoryListItem
                                key={subcategory.name}
                                categoryName={categoryName}
                                checked={checked}
                                setChecked={setChecked}
                                subcategory={subcategory}
                                defaultExpanded={childDefaultExpanded}
                                collapseSignal={collapseSignal}
                                expandSignal={expandSignal}
                                searchQuery={searchQuery}
                            />
                        );
                    })}
                </ul>
            }
        />
    );
};

interface SubcategoryListItemProps {
    categoryName: string;
    subcategory: Subcategory;
    checked: EdgeCheckboxType[];
    setChecked: (checked: EdgeCheckboxType[]) => void;
    defaultExpanded?: boolean;
    collapseSignal?: number;
    expandSignal?: number;
    searchQuery?: string;
}

const SubcategoryListItem = ({
    categoryName,
    subcategory,
    checked,
    setChecked,
    defaultExpanded = false,
    collapseSignal = 0,
    expandSignal = 0,
    searchQuery = '',
}: SubcategoryListItemProps) => {
    const { name, edgeTypes } = subcategory;

    const subcategoryFilter = (element: EdgeCheckboxType) =>
        element.category === categoryName && element.subcategory === name;

    const isSearching = searchQuery.trim().length > 0;

    return (
        <IndeterminateListItem
            name={name}
            checkboxFilter={subcategoryFilter}
            checked={checked}
            setChecked={setChecked}
            defaultExpanded={defaultExpanded}
            collapseSignal={collapseSignal}
            expandSignal={expandSignal}
            forceExpanded={isSearching}
            collapsibleContent={
                <div className='pl-8'>
                    <EdgesView
                        edgeTypes={edgeTypes}
                        checked={checked}
                        setChecked={setChecked}
                        searchQuery={searchQuery}
                    />
                </div>
            }
        />
    );
};

const HighlightMatch = ({ text, query }: { text: string; query: string }) => {
    if (!query.trim()) return <>{text}</>;

    const lowerText = text.toLowerCase();
    const lowerQuery = query.toLowerCase();
    const matchIndex = lowerText.indexOf(lowerQuery);

    if (matchIndex === -1) return <>{text}</>;

    const before = text.slice(0, matchIndex);
    const match = text.slice(matchIndex, matchIndex + query.length);
    const after = text.slice(matchIndex + query.length);

    return (
        <>
            {before}
            <span className='text-link'>{match}</span>
            {after}
        </>
    );
};

interface EdgesViewProps {
    edgeTypes: string[];
    checked: EdgeCheckboxType[];
    setChecked: (checked: EdgeCheckboxType[]) => void;
    searchQuery?: string;
}

const EdgesView = ({ edgeTypes, checked, setChecked, searchQuery = '' }: EdgesViewProps) => {
    const changeCheckbox = (isChecked: boolean | 'indeterminate', edgeType: string) => {
        const newChecked = [...checked];
        const indexToUpdate = newChecked.findIndex((element) => element.edgeType === edgeType);
        newChecked[indexToUpdate] = { ...newChecked[indexToUpdate], checked: isChecked === true };

        setChecked(newChecked);
    };

    return (
        <ul className='p-2 rounded-lg list-none flex flex-wrap'>
            {[...edgeTypes]
                .sort((a, b) => a.localeCompare(b))
                .map((edgeType, index) => {
                    const edgeIsChecked = checked.find((element) => element.edgeType === edgeType)?.checked ?? false;
                    return (
                        <li key={index} className='w-1/2 min-w-0 pr-4'>
                            <label className='flex items-start gap-2 cursor-pointer py-0.5 hover:underline'>
                                <Checkbox
                                    aria-label={edgeType}
                                    name={edgeType}
                                    checked={edgeIsChecked}
                                    onCheckedChange={(value) => changeCheckbox(value, edgeType)}
                                    className='mt-0.5 shrink-0'
                                />
                                <span className='text-sm break-words min-w-0'>
                                    <HighlightMatch text={edgeType} query={searchQuery} />
                                </span>
                            </label>
                        </li>
                    );
                })}
        </ul>
    );
};

interface IndeterminateListItemProps {
    name: string;
    checkboxFilter: (element: EdgeCheckboxType) => boolean;
    checked: EdgeCheckboxType[];
    setChecked: (checked: EdgeCheckboxType[]) => void;
    defaultExpanded?: boolean;
    collapseSignal?: number;
    expandSignal?: number;
    forceExpanded?: boolean;
    collapsibleContent: React.ReactNode;
}

const IndeterminateListItem = ({
    name,
    checkboxFilter,
    checked,
    setChecked,
    defaultExpanded = false,
    collapseSignal = 0,
    expandSignal = 0,
    forceExpanded = false,
    collapsibleContent,
}: IndeterminateListItemProps) => {
    const [showCollapsibleContent, setShowCollapsibleContent] = useState(defaultExpanded);
    const initialCollapseSignal = useRef(collapseSignal);
    const initialExpandSignal = useRef(expandSignal);

    useEffect(() => {
        if (collapseSignal !== initialCollapseSignal.current) {
            setShowCollapsibleContent(false);
        }
    }, [collapseSignal]);

    useEffect(() => {
        if (expandSignal !== initialExpandSignal.current) {
            setShowCollapsibleContent(true);
        }
    }, [expandSignal]);

    const checkboxes = checked.filter(checkboxFilter);
    const elementIsChecked = (element: EdgeCheckboxType) => element.checked;
    const elementIsEqual = (element: EdgeCheckboxType, _index: number, arr: EdgeCheckboxType[]) =>
        element.checked === arr[0].checked;

    const allChecked = checkboxes.every(elementIsChecked);
    const isIndeterminate = !checkboxes.every(elementIsEqual);
    const checkedState: boolean | 'indeterminate' = isIndeterminate ? 'indeterminate' : allChecked;

    const changeAllChildren = (value: boolean) => {
        const newChecked = [...checked];
        newChecked.forEach((checkbox, index, arr) => {
            if (checkboxFilter(checkbox)) {
                arr[index] = { ...checkbox, checked: value };
            }
        });

        setChecked(newChecked);
    };

    const handleCheck = () => {
        // if item is checked, uncheck all children
        if (allChecked) {
            changeAllChildren(false);
        } else {
            // if item is not checked, check all children
            changeAllChildren(true);
        }
    };

    const isExpanded = forceExpanded || showCollapsibleContent;
    const toggleCollapsibleContent = () => setShowCollapsibleContent((v) => !v);

    return (
        <li className='list-none'>
            <div className='flex items-center py-1 group'>
                <div
                    role='button'
                    tabIndex={0}
                    className='flex items-center gap-2 flex-1 cursor-pointer p-1 text-left rounded-l peer hover:bg-neutral-3'
                    onClick={toggleCollapsibleContent}
                    onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault();
                            toggleCollapsibleContent();
                        }
                    }}>
                    <Checkbox
                        aria-label={name}
                        checked={checkedState}
                        onCheckedChange={() => handleCheck()}
                        onClick={(e) => e.stopPropagation()}
                        icon={isIndeterminate ? <span className='block w-2/3 mx-auto h-0.5 bg-current' /> : undefined}
                    />
                    <span className='text-sm font-medium'>{name}</span>
                </div>
                <button
                    type='button'
                    aria-label={isExpanded ? `minimize-${name}` : `expand-${name}`}
                    className='px-2 self-stretch flex items-center cursor-pointer bg-transparent border-none rounded-r peer-hover:bg-neutral-3'
                    onClick={toggleCollapsibleContent}>
                    <FontAwesomeIcon icon={isExpanded ? faChevronUp : faChevronDown} />
                </button>
            </div>
            {isExpanded && collapsibleContent}
        </li>
    );
};

export default EdgeFilteringDialog;
