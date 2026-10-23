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

import { Typography, VisuallyHidden } from 'doodle-ui';
import { FieldDisplayProps } from './Field.types';

const InlineField = ({ label, values }: FieldDisplayProps) => (
    <>
        {values.map(({ value, accessibleLabel }, index) => (
            <div className='p-2' key={index}>
                {index === 0 && (
                    <Typography component='span' variant='body2' className='mr-2 font-semibold text-text-main'>
                        {typeof label === 'string' ? label.replace(/:$/, '') : label}
                    </Typography>
                )}
                <Typography component='span' variant='body2' className='break-words' title={value}>
                    {accessibleLabel ? (
                        <>
                            <span aria-hidden='true'>{value}</span>
                            <VisuallyHidden>{accessibleLabel}</VisuallyHidden>
                        </>
                    ) : (
                        value
                    )}
                </Typography>
            </div>
        ))}
    </>
);

export default InlineField;
