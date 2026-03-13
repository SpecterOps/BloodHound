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
import presets from '../../../tailwind/preset';

const fonts = presets.theme.extend.fontSize;

const FontSizes = () => {
    return (
        <div className='mb-8'>
            {Object.entries(fonts).map(([key, val]) => (
                <div key={key} className={`mb-5`}>
                    <div className={`grid grid-cols-3 leading-none text-${key}`}>
                        <p>{key}</p>
                        <p className='font-bold'>{key}</p>
                        {key !== 'headline-1' && key !== 'headline-2' && <p className='underline'>{key}</p>}
                    </div>
                    <p>
                        <span className='text-eyeline'>{val}</span>
                    </p>
                </div>
            ))}
        </div>
    );
};

export { FontSizes };
