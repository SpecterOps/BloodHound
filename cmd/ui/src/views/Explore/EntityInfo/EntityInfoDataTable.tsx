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

import { EntityInfoDataTableProps, InfiniteScrollingTable, abortEntitySectionRequest } from 'bh-shared-ui';
import { useQuery } from 'react-query';
import { useDispatch } from 'react-redux';
import { NODE_GRAPH_RENDER_LIMIT } from 'src/constants';
import { putGraphData, putGraphError, saveResponseForExport, setGraphLoading } from 'src/ducks/explore/actions';
import { addSnackbar } from 'src/ducks/global/actions';
import { sourceNodeSelected } from 'src/ducks/searchbar/actions';
import { transformFlatGraphResponse } from 'src/utils';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';

const EntityInfoDataTable: React.FC<EntityInfoDataTableProps> = ({ id, label, endpoint, sections }) => {
    const dispatch = useDispatch();

    const countQuery = useQuery(
        ['relatedCount', label, id],
        () => {
            if (endpoint) {
                return endpoint({ skip: 0, limit: 128 });
            }
            if (sections) return Promise.all(sections.map((section) => section.endpoint?.({ skip: 0, limit: 128 })));
            return Promise.reject('Invalid call data provided for relationship list query');
        },
        { refetchOnWindowFocus: false, retry: false }
    );

    const handleOnChange = async (label: string, isOpen: boolean) => {
        if (!endpoint) return;

        if (isOpen && countQuery.data?.count < NODE_GRAPH_RENDER_LIMIT) {
            abortEntitySectionRequest();

            dispatch(setGraphLoading(true));

            await endpoint({ type: 'graph' })
                .then((result) => {
                    const formattedData = transformFlatGraphResponse(result);

                    dispatch(saveResponseForExport(formattedData));
                    dispatch(putGraphData(result));
                })
                .catch((err) => {
                    if (err?.code === 'ERR_CANCELED') {
                        return;
                    }
                    dispatch(putGraphError(err));
                    dispatch(addSnackbar('Query failed. Please try again.', 'nodeRelationshipGraphQuery', {}));
                })
                .finally(() => {
                    dispatch(setGraphLoading(false));
                });
        }
    };

    const handleOnClick = (item: any) => {
        dispatch(
            sourceNodeSelected({
                objectid: item.id,
                type: item.type,
                name: item.name,
            })
        );
    };

    let count: number | undefined;
    if (Array.isArray(countQuery.data)) {
        count = countQuery.data.reduce((acc, val) => {
            const count = val.count ?? 0;
            return acc + count;
        }, 0);
    } else if (countQuery.data) {
        count = countQuery.data.count ?? 0;
    }

    return (
        <EntityInfoCollapsibleSection
            label={label}
            count={count}
            isLoading={countQuery.isLoading}
            isError={countQuery.isError}
            error={countQuery.error}
            onChange={handleOnChange}>
            {endpoint && (
                <InfiniteScrollingTable itemCount={count} fetchDataCallback={endpoint} onClick={handleOnClick} />
            )}
            {sections && sections.map((nestedSection, index) => <EntityInfoDataTable key={index} {...nestedSection} />)}
        </EntityInfoCollapsibleSection>
    );
};

export default EntityInfoDataTable;
