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

import { faChevronDown, faChevronUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Skeleton } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { useState } from 'react';
import { CommonSearches as prebuiltSearchListAGI } from '../../../commonSearchesAGI';
import { CommonSearches as prebuiltSearchListAGT } from '../../../commonSearchesAGT';
import FeatureFlag from '../../../components/FeatureFlag';
import PrebuiltSearchList, { LineItem } from '../../../components/PrebuiltSearchList';
import { useDeleteSavedQuery, useSavedQueries } from '../../../hooks';
import { useNotifications } from '../../../providers';
import { QuerySearchType } from '../../../types';
import { cn } from '../../../utils';
import QuerySearchFilter from './QuerySearchFilter';
const AD_TAB = 'Active Directory';
const AZ_TAB = 'Azure';
const CUSTOM_TAB = 'Custom Searches';

const useStyles = makeStyles((theme) => ({
    tabs: {
        height: '35px',
        minHeight: '35px',
    },
    tab: {
        height: '35px',
        minHeight: '35px',
        color: theme.palette.color.primary,
    },
    list: {
        position: 'relative',
        overflow: 'hidden',
        '& ul': { padding: 0 },
    },
}));

type CommonSearchesProps = {
    onSetCypherQuery: (query: string) => void;
    onPerformCypherSearch: (query: string) => void;
};

const InnerCommonSearches = ({
    onSetCypherQuery,
    onPerformCypherSearch,
    prebuiltSearchList,
}: CommonSearchesProps & { prebuiltSearchList: QuerySearchType[] }) => {
    const userQueries = useSavedQueries();
    const deleteQueryMutation = useDeleteSavedQuery();
    const { addNotification } = useNotifications();

    const [showCommonQueries, setShowCommonQueries] = useState(false);
    const [searchTerm, setSearchTerm] = useState('');
    const [platform, setPlatform] = useState('');
    const [categoryFilter, setCategoryFilter] = useState<string[]>([]);

    const savedLineItems: LineItem[] =
        userQueries.data?.map((query) => ({
            description: query.name,
            cypher: query.query,
            canEdit: true,
            id: query.id,
        })) || [];

    const savedQueries = {
        category: 'Saved Queries',
        subheader: '',
        queries: savedLineItems,
    };

    //master list of pre-made queries
    const queryList = [...prebuiltSearchList, savedQueries];

    //list of categories for filter dropdown
    // const categories = queryList.
    const allCategories = queryList.map((item) => item.subheader);
    const uniqueCategoriesSet = new Set(allCategories);
    const categories = [...uniqueCategoriesSet].filter((category) => category !== '').sort();

    const [filteredList, setFilteredList] = useState<any[]>(queryList);

    const handleClick = (query: string) => {
        // This first function is only necessary for the redux implementation and can be removed later, along with the associated prop
        onSetCypherQuery(query);
        onPerformCypherSearch(query);
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
        let filteredData: any[] = queryList;

        if (searchTerm.length > 2) {
            filteredData = filteredData
                .map((obj) => ({
                    ...obj,
                    queries: obj.queries.filter((item: any) =>
                        item.description.toLowerCase().includes(searchTerm.toLowerCase())
                    ),
                }))
                .filter((x) => x.queries.length);
        }
        if (platform) {
            filteredData = filteredData.filter((obj) => obj.category.toLowerCase() === platform.toLowerCase());
        }
        if (categories.length) {
            filteredData = filteredData
                .filter((item: any) => categories.includes(item.subheader))
                .filter((x) => x.queries.length);
        }
        setFilteredList(filteredData);
    };

    const handleClearFilters = () => {
        handleFilter('', '', []);
    };

    return (
        <div className='flex flex-col h-full'>
            <div className='flex items-center'>
                <FontAwesomeIcon
                    className='px-2 mr-2'
                    icon={showCommonQueries ? faChevronDown : faChevronUp}
                    onClick={() => {
                        setShowCommonQueries((v) => !v);
                    }}
                />
                <h5 className='my-4 font-bold text-lg'>Pre-built Queries</h5>
            </div>

            <div className={cn('grow-1 min-h-0 overflow-auto', { hidden: !showCommonQueries })}>
                <QuerySearchFilter
                    queryFilterHandler={handleFilter}
                    categories={categories}
                    searchTerm={searchTerm}
                    platform={platform}
                    categoryFilter={categoryFilter}></QuerySearchFilter>
                <PrebuiltSearchList
                    listSections={filteredList}
                    clickHandler={handleClick}
                    deleteHandler={handleDeleteQuery}
                    clearFiltersHandler={handleClearFilters}
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
