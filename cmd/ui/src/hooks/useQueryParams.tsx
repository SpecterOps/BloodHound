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

import { useMemo } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

const useQueryParams = () => {
    const { search } = useLocation();
    const navigate = useNavigate();
    const queryParams = useMemo(() => new URLSearchParams(search), [search]);

    const _push = (queryParams: URLSearchParams) => {
        navigate({
            search: queryParams.toString(),
        });
    };

    /**
     *
     * @param name The name of the parameter to set.
     * @param value The value of the parameter to set.
     *
     * Sets the query paramter and updates the current URL.
     */
    const setQueryParam = (name: string, value?: string) => {
        if (queryParams.get(name) === value) return;

        if (value !== undefined) queryParams.set(name, value);
        else queryParams.delete(name);

        _push(queryParams);
    };

    /**
     *
     * @param name The name of the parameter to delete.
     *
     * Deletes the query paramter and updates the current URL.
     */
    const deleteQueryParam = (name: string) => {
        if (queryParams.get(name) === null) return;

        queryParams.delete(name);

        _push(queryParams);
    };

    return { queryParams, setQueryParam, deleteQueryParam };
};

export default useQueryParams;
