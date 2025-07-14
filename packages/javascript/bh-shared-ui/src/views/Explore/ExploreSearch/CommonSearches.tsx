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

import { Skeleton } from '@bloodhoundenterprise/doodleui';
import { faChevronDown, faChevronUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box } from '@mui/material';
import fileDownload from 'js-file-download';
import { useEffect, useState } from 'react';
import { CommonSearches as prebuiltSearchListAGI } from '../../../commonSearchesAGI';
import { CommonSearches as prebuiltSearchListAGT } from '../../../commonSearchesAGT';
import FeatureFlag from '../../../components/FeatureFlag';
import PrebuiltSearchList from '../../../components/PrebuiltSearchList';
import {
    getExportQuery,
    useCypherSearch,
    useDeleteSavedQuery,
    usePrebuiltQueries,
    useSavedQueries,
} from '../../../hooks';
import { useNotifications } from '../../../providers';
import { QueryLineItem, QueryListSection, QuerySearchType } from '../../../types';
import { cn } from '../../../utils';
import QuerySearchFilter from './QuerySearchFilter';

type CommonSearchesProps = {
    onSetCypherQuery: (query: string) => void;
    onPerformCypherSearch: (query: string) => void;
    onSetSelected: (query: string) => void;
    onToggleCommonQueries: () => void;
    selected: string;
    showCommonQueries: boolean;
};

const InnerCommonSearches = ({
    onSetCypherQuery,
    onPerformCypherSearch,
    onSetSelected,
    onToggleCommonQueries,
    prebuiltSearchList,
    selected,
    showCommonQueries,
}: CommonSearchesProps & { prebuiltSearchList: QuerySearchType[] }) => {
    const userQueries = useSavedQueries();
    const deleteQueryMutation = useDeleteSavedQuery();
    const { addNotification } = useNotifications();
    const [searchTerm, setSearchTerm] = useState('');
    const [platform, setPlatform] = useState('');
    const [categoryFilter, setCategoryFilter] = useState<string[]>([]);
    const { cypherQuery } = useCypherSearch();

    //master list of pre-made queries
    const queryList = usePrebuiltQueries();
    const allCategories = queryList.map((item) => item.subheader);
    const uniqueCategoriesSet = new Set(allCategories);
    const categories = [...uniqueCategoriesSet].filter((category) => category !== '').sort();

    const [filteredList, setFilteredList] = useState<QueryListSection[]>([]);

    useEffect(() => {
        setFilteredList(queryList);
    }, [userQueries.data]);

    const handleClick = (query: string) => {
        if (selected === query) {
            //deselect
            onSetSelected('');
            // This first function is only necessary for the redux implementation and can be removed later, along with the associated prop
            onSetCypherQuery('');
            onPerformCypherSearch('');
        } else {
            onSetSelected(query);
            // This first function is only necessary for the redux implementation and can be removed later, along with the associated prop
            onSetCypherQuery(query);
            onPerformCypherSearch(query);
        }
    };

    const handleDeleteQuery = (id: number) =>
        deleteQueryMutation.mutate(id, {
            onSuccess: () => {
                addNotification(`Query deleted.`, 'userDeleteQuery');
            },
        });

    if (userQueries.isLoading) {
        return (
            <Box mt={2}>
                <Skeleton />
            </Box>
        );
    }

    const handleFilter = (searchTerm: string, platform: string, categories: string[]) => {
        setSearchTerm(searchTerm);
        setPlatform(platform);
        setCategoryFilter(categories);

        //local array variable
        let filteredData: QuerySearchType[] = queryList;

        if (searchTerm.length > 2) {
            filteredData = filteredData
                .map((obj) => ({
                    ...obj,
                    queries: obj.queries.filter((item: QueryLineItem) =>
                        item.name?.toLowerCase().includes(searchTerm.toLowerCase())
                    ),
                }))
                .filter((x) => x.queries.length);
        }
        if (platform) {
            filteredData = filteredData.filter((obj) => obj.category.toLowerCase() === platform.toLowerCase());
        }
        if (categories.length) {
            filteredData = filteredData
                .filter((item: QuerySearchType) => categories.includes(item.subheader))
                .filter((x) => x.queries.length);
        }
        setFilteredList(filteredData);
    };

    const handleClearFilters = () => {
        handleFilter('', '', []);
    };

    const handleExport = () => {
        if (!(selectedQuery && selectedQuery?.id)) return;
        getExportQuery(selectedQuery.id).then((res: any) => {
            const filename =
                res.headers['content-disposition']?.match(/^.*filename="(.*)"$/)?.[1] || `exported_queries.zip`;
            fileDownload(res.data, filename);
        });
    };

    function getSelectedQuery() {
        const comparator = selected ? selected : cypherQuery;
        for (const item of filteredList) {
            let result = null;
            result = item.queries.find((query) => {
                if (query.query === comparator) {
                    return query;
                }
            });
            if (result) {
                return result;
            }
        }
    }

    const selectedQuery = getSelectedQuery();
    return (
        <div className='flex flex-col h-full'>
            <div className='flex items-center'>
                <FontAwesomeIcon
                    className='px-2 mr-2'
                    icon={showCommonQueries ? faChevronDown : faChevronUp}
                    onClick={onToggleCommonQueries}
                />
                <h5 className='my-4 font-bold text-lg'>Pre-built Queries</h5>
            </div>

            <div className={cn({ hidden: !showCommonQueries })}>
                <QuerySearchFilter
                    queryFilterHandler={handleFilter}
                    exportHandler={handleExport}
                    categories={categories}
                    searchTerm={searchTerm}
                    platform={platform}
                    categoryFilter={categoryFilter}
                    selectedQuery={selectedQuery}></QuerySearchFilter>
            </div>

            <div className={cn('grow-1 min-h-0 overflow-auto', { hidden: !showCommonQueries })}>
                <PrebuiltSearchList
                    listSections={filteredList}
                    clickHandler={handleClick}
                    deleteHandler={handleDeleteQuery}
                    clearFiltersHandler={handleClearFilters}
                    selectedQuery={selectedQuery}
                />
            </div>
        </div>
    );
};

const CommonSearches = (props: CommonSearchesProps) => (
    <FeatureFlag
        flagKey='tier_management_engine'
        enabled={<InnerCommonSearches {...props} prebuiltSearchList={prebuiltSearchListAGT} />}
        disabled={<InnerCommonSearches {...props} prebuiltSearchList={prebuiltSearchListAGI} />}
    />
);

export default CommonSearches;
