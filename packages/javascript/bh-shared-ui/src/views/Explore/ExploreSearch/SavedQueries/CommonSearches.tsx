import { Button } from '@bloodhoundenterprise/doodleui';
import { faChevronDown, faChevronUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import fileDownload from 'js-file-download';
import { useCallback, useEffect, useState } from 'react';
import PrebuiltSearchList from '../../../../components/PrebuiltSearchList';
import {
    getExportQuery,
    useCypherSearch,
    useDeleteSavedQuery,
    usePrebuiltQueries,
    useSavedQueries,
} from '../../../../hooks';
import { useSelf } from '../../../../hooks/useSelf';
import { useNotifications } from '../../../../providers';
import { QueryLineItem, QueryListSection } from '../../../../types';
import { cn } from '../../../../utils';
import { useSavedQueriesContext } from '../../providers';
import ConfirmDeleteQueryDialog from './ConfirmDeleteQueryDialog';
import QuerySearchFilter from './QuerySearchFilter';

type CommonSearchesProps = {
    onSetCypherQuery: (query: string) => void;
    onPerformCypherSearch: (query: string) => void;
    onToggleCommonQueries: () => void;
    showCommonQueries: boolean;
};

const CommonSearches = ({
    onSetCypherQuery,
    onPerformCypherSearch,
    onToggleCommonQueries,
    showCommonQueries,
}: CommonSearchesProps) => {
    const { selectedId, selectedQuery, setSelectedId } = useSavedQueriesContext();
    const { cypherQuery } = useCypherSearch();

    const userQueries = useSavedQueries();
    const deleteQueryMutation = useDeleteSavedQuery();
    const { addNotification } = useNotifications();

    const [searchTerm, setSearchTerm] = useState('');
    const [platform, setPlatform] = useState('');
    const [source, setSource] = useState('');
    const [open, setOpen] = useState(false);
    const [queryId, setQueryId] = useState<number>();

    const [categoryFilter, setCategoryFilter] = useState<string[]>([]);

    //master list of pre-made queries
    const queryList: QueryListSection[] = usePrebuiltQueries();
    const allCategories = queryList.map((item) => item.subheader);
    const uniqueCategoriesSet = new Set(allCategories);
    const categories = [...uniqueCategoriesSet].filter((category) => category !== '').sort();

    const [filteredList, setFilteredList] = useState<QueryListSection[]>([]);

    const { getSelfId } = useSelf();
    const { data: selfId } = getSelfId;

    const handleFilter = useCallback(
        (searchTerm: string, platform: string, categories: string[], source: string) => {
            setSearchTerm(searchTerm);
            setPlatform(platform);
            setCategoryFilter(categories);
            setSource(source);
            //local array variable
            let filteredData: QueryListSection[] = queryList;
            const hasSelf = typeof selfId === 'string' && selfId.length > 0;

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
                filteredData = filteredData.filter((obj) => obj.category?.toLowerCase() === platform.toLowerCase());
            }
            if (categories.length) {
                filteredData = filteredData
                    .filter((item: QueryListSection) => categories.includes(item.subheader))
                    .filter((x) => x.queries.length);
            }
            if (source && source === 'prebuilt') {
                filteredData = filteredData
                    .map((obj) => ({
                        ...obj,
                        queries: obj.queries.filter((item: QueryLineItem) => !item.id),
                    }))
                    .filter((x) => x.queries.length);
            } else if (source && source === 'personal') {
                if (!hasSelf) {
                    filteredData = [];
                } else {
                    filteredData = filteredData
                        .map((obj) => ({
                            ...obj,
                            queries: obj.queries.filter((item: QueryLineItem) => item.user_id === selfId),
                        }))
                        .filter((x) => x.queries.length);
                }
            } else if (source && source === 'shared') {
                if (!hasSelf) {
                    filteredData = [];
                } else {
                    filteredData = filteredData
                        .map((obj) => ({
                            ...obj,
                            queries: obj.queries.filter((item: QueryLineItem) => item.id && item.user_id !== selfId),
                        }))
                        .filter((x) => x.queries.length);
                }
            }
            setFilteredList(filteredData);
        },
        [queryList, selfId]
    );

    useEffect(() => {
        setFilteredList(queryList);
        handleFilter(searchTerm, platform, categoryFilter, source);
    }, [userQueries.data, categoryFilter, platform, searchTerm, source, queryList, handleFilter]);

    const handleClick = (query: string, id: number | undefined) => {
        if (query === cypherQuery && selectedId === id) {
            //deselect
            setSelectedId(undefined);
            onSetCypherQuery('');
            onPerformCypherSearch('');
        } else {
            setSelectedId(id);
            onSetCypherQuery(query);
            onPerformCypherSearch(query);
        }
    };

    const handleDeleteQuery = (id: number) => {
        setQueryId(id);
        setOpen(true);
    };

    const handleClose = () => {
        setOpen(false);
        setQueryId(undefined);
    };

    const confirmDeleteQuery = (id: number) => {
        deleteQueryMutation.mutate(id, {
            onSuccess: () => {
                addNotification(`Query deleted.`, 'userDeleteQuery');
                setOpen(false);
                setQueryId(undefined);
            },
        });
    };

    const handleClearFilters = () => {
        handleFilter('', '', [], '');
    };

    const handleExport = () => {
        if (!(selectedQuery && selectedQuery?.id)) return;
        getExportQuery(selectedQuery.id).then((res) => {
            const filename =
                res.headers['content-disposition']?.match(/^.*filename="(.*)"$/)?.[1] || `exported_queries.zip`;
            fileDownload(res.data, filename);
        });
    };

    return (
        <div className='flex flex-col h-full'>
            <div className='flex items-center'>
                <Button
                    onClick={onToggleCommonQueries}
                    className='flex justify-start items-center w-full pl-0'
                    data-testid='common-queries-toggle'
                    variant={'text'}>
                    <FontAwesomeIcon className='px-2 mr-2' icon={showCommonQueries ? faChevronDown : faChevronUp} />
                    <span className='my-4 font-semibold text-lg'>Saved Queries</span>
                </Button>
            </div>

            <div className={cn({ hidden: !showCommonQueries })}>
                <QuerySearchFilter
                    queryFilterHandler={handleFilter}
                    exportHandler={handleExport}
                    deleteHandler={handleDeleteQuery}
                    categories={categories}
                    searchTerm={searchTerm}
                    platform={platform}
                    categoryFilter={categoryFilter}
                    source={source}></QuerySearchFilter>
            </div>

            <div className={cn('grow-1 min-h-0 overflow-auto', { hidden: !showCommonQueries })}>
                <PrebuiltSearchList
                    listSections={filteredList}
                    clickHandler={handleClick}
                    deleteHandler={handleDeleteQuery}
                    clearFiltersHandler={handleClearFilters}
                    showCommonQueries={showCommonQueries}
                />
            </div>

            <ConfirmDeleteQueryDialog
                open={open}
                queryId={queryId}
                deleteHandler={confirmDeleteQuery}
                handleClose={handleClose}
            />
        </div>
    );
};

export default CommonSearches;
