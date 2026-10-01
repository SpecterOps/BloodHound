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

const colors = presets.theme.extend.colors;

interface SwatchObject {
    objKey: string;
}

const Swatch = (props: SwatchObject) => {
    const { objKey } = props;
    return (
        <div className='border mb-4 flex flex-col items-center p-4'>
            <div className={`w-16 h-16 mb-1 border bg-${objKey}`}></div>
            <p>{objKey}</p>
        </div>
    );
};

const ColorPalette = () => {
    return (
        <div className='mb-8'>
            <div className='grid w-full grid-cols-3 gap-4'>
                {Object.keys(colors).map((key) => {
                    return <Swatch key={key} objKey={key} />;
                })}
            </div>
        </div>
    );
};

export { ColorPalette };
