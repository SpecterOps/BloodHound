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

import { makeStyles } from '@mui/styles';

export const groupSpecialFormat = (sourceType: string | undefined, sourceName: string | undefined) => {
    if (!sourceType || !sourceName) return 'This entity has';
    if (sourceType === 'Group') {
        return `The members of the ${typeFormat(sourceType)} ${sourceName} have`;
    } else {
        return `The ${typeFormat(sourceType)} ${sourceName} has`;
    }
};

export const typeFormat = (type: string | undefined): string => {
    if (!type) return '';
    if (type === 'GPO' || type === 'OU') {
        return type;
    } else if (type === 'CertTemplate') {
        return 'certificate template';
    } else if (type === 'EnterpriseCA') {
        return 'enterprise CA';
    } else if (type === 'RootCA') {
        return 'root CA';
    } else if (type === 'NTAuthStore') {
        return 'NTAuth store';
    } else if (type === 'AIACA') {
        return 'AIA CA';
    } else {
        return type.toLowerCase();
    }
};

export const useHelpTextStyles = makeStyles((theme) => ({
    containsCodeEl: {
        '& code': {
            backgroundColor: 'darkgrey',
            padding: '2px .5ch',
            fontWeight: 'normal',
            fontSize: '.875em',
            borderRadius: '3px',
            display: 'inline',

            overflowWrap: 'break-word',
            whiteSpace: 'pre-wrap',
        },
    },
}));
