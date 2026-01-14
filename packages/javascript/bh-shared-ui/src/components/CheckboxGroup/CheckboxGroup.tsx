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

import { Checkbox, FormControlLabel, FormGroup } from '@mui/material';

export interface CheckboxGroupProps {
    groupTitle: string;
    handleCheckboxFilter: any;
    options: {
        name: string;
        label: string;
    }[];
}

const CheckboxGroup: React.FC<CheckboxGroupProps> = ({ groupTitle, handleCheckboxFilter, options }) => {
    return (
        <section>
            <h3>{groupTitle}</h3>
            <FormGroup>
                {options.map((option: any, index: number) => {
                    return (
                        <FormControlLabel
                            control={
                                <Checkbox
                                    className='ml-2 p-1'
                                    onChange={handleCheckboxFilter}
                                    name={option.name}
                                    color='primary'
                                />
                            }
                            label={option.label}
                            key={index}
                        />
                    );
                })}
            </FormGroup>
        </section>
    );
};

export default CheckboxGroup;
