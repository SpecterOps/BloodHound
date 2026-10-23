// Copyright 2026 Specter Ops, Inc.
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

import { cn } from '../../utils';
import { FieldDisplayProps } from './Field.types';
import FieldRows from './FieldRows';

const SpreadField = ({ label, values, isList }: FieldDisplayProps & { isList: boolean }) => (
    <FieldRows
        label={<div className='mr-2 grow shrink-0 font-bold'>{label}</div>}
        values={values}
        rowClassName={cn('flex flex-row flex-wrap', isList && 'justify-end')}
        renderValue={(value, content) => (
            <div className='overflow-hidden text-ellipsis break-all' title={value}>
                {content}
            </div>
        )}
    />
);

export default SpreadField;
