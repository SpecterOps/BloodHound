// Copyright 2025 Specter Ops, Inc.
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
const ShortcutSection = ({ heading, bindings }: { heading: string; bindings: string[][] }) => (
    <div className='mb-5' key={heading}>
        <div className='font-bold flex sm:justify-center p-2'>{heading}</div>
        <div className='flex flex-col gap-2 text-sm'>
            {bindings.map((binding: string[]) => (
                <div key={`${binding[0]}-${heading}`} className='flex gap-2'>
                    <div className='w-1/2 text-right p-2 flex md:justify-end sm:justify-center xs:justify-center items-center'>
                        {' '}
                        {binding[1]}
                    </div>
                    <div className='w-1/2 border-2 rounded-md p-2 text-center flex justify-center items-center'>
                        Alt/Option + {binding[0]}
                    </div>
                </div>
            ))}
        </div>
    </div>
);

export default ShortcutSection;
