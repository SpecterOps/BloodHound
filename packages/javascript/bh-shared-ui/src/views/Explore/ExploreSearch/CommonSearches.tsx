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
import makeStyles from '@mui/styles/makeStyles';
import { useState } from 'react';
import { CommonSearches as prebuiltSearchListAGI } from '../../../commonSearchesAGI';
import { CommonSearches as prebuiltSearchListAGT } from '../../../commonSearchesAGT';
import FeatureFlag from '../../../components/FeatureFlag';
import PrebuiltSearchList, { LineItem } from '../../../components/PrebuiltSearchList';
import { useDeleteSavedQuery, useSavedQueries } from '../../../hooks';
import { useNotifications } from '../../../providers';
import { CommonSearchType } from '../../../types';
import { cn } from '../../../utils';
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
}: CommonSearchesProps & { prebuiltSearchList: CommonSearchType[] }) => {
    // const classes = useStyles();

    // const [activeTab, setActiveTab] = useState(AD_TAB);

    // const handleTabChange = (event: React.SyntheticEvent, newValue: string) => {
    //     setActiveTab(newValue);
    // };
    const userQueries = useSavedQueries();
    const deleteQueryMutation = useDeleteSavedQuery();
    const { addNotification } = useNotifications();

    const [showCommonQueries, setShowCommonQueries] = useState(false);

    const savedLineItems: LineItem[] =
        userQueries.data?.map((query) => ({
            description: query.name,
            cypher: query.query,
            canEdit: true,
            id: query.id,
        })) || [];

    // console.log(savedLineItems);
    const savedQueries = {
        category: 'Saved Queries',
        subheader: '',
        lineItems: savedLineItems,
    };
    console.log('savedQueries');
    console.log(savedQueries);

    const adSections = prebuiltSearchList
        .filter(({ category }) => category === 'Active Directory')
        .map(({ category, subheader, queries }) => ({ category, subheader, lineItems: queries }));

    const azSections = prebuiltSearchList
        .filter(({ category }) => category === 'Azure')
        .map(({ category, subheader, queries }) => ({ category, subheader, lineItems: queries }));

    const queryList = [...prebuiltSearchList, savedQueries];
    console.log('queryList');
    console.log(queryList);

    const adAzSections = prebuiltSearchList
        .filter(({ category }) => category === 'Active Directory' || category === 'Azure')
        .map(({ category, subheader, queries }) => ({ category, subheader, lineItems: queries }));

    console.log('prebuiltSearchList');
    console.log(prebuiltSearchList);
    console.log('adAzSections');
    console.log(adAzSections);

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
                {/* <PrebuiltSearchList listSections={adSections} clickHandler={handleClick} />
                <PrebuiltSearchList listSections={azSections} clickHandler={handleClick} />
                <PersonalSearchList clickHandler={handleClick} /> */}
                <PrebuiltSearchList listSections={adAzSections} clickHandler={handleClick} />
                {/* <QuerySearchList clickHandler={handleClick} /> */}
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
