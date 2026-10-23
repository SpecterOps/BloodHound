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

import { VisuallyHidden } from 'doodle-ui';
import { ReactNode } from 'react';
import { cn } from '../../utils';
import { FieldDisplayProps } from './Field.types';

type FieldRowsProps = FieldDisplayProps & {
    rowClassName?: string;
    renderValue: (value: string, content: ReactNode) => ReactNode;
};

const FieldRows = ({ label, values, rowClassName, renderValue }: FieldRowsProps) => (
    <>
        {values.map(({ value, accessibleLabel }, index) => (
            <div className={cn('p-2', rowClassName)} key={index}>
                {index === 0 && label}
                {renderValue(
                    value,
                    accessibleLabel ? (
                        <>
                            <span aria-hidden='true'>{value}</span>
                            <VisuallyHidden>{accessibleLabel}</VisuallyHidden>
                        </>
                    ) : (
                        value
                    )
                )}
            </div>
        ))}
    </>
);

export default FieldRows;
