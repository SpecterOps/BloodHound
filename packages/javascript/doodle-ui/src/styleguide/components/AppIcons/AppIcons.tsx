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
import { Card, CardContent } from '../../../components/Card';
import * as IconOptions from './components';

const componentNames = Object.keys(IconOptions);

const nonLogos = componentNames.filter((x) => !x.includes('Full'));

export type AppIconOptions = keyof typeof IconOptions;

export const AppIcon = IconOptions;

const AppIcons = () => {
    return (
        <div className='mb-8'>
            <div className='mb-4'>
                <div className='grid grid-cols-1 md:grid-cols-2 gap-x-4 gap-y-8 mb-8'>
                    <div className='flex flex-col items-center justify-center'>
                        <Card className='flex items-center justify-center h-32  mb-2'>
                            <CardContent className='px-6'>
                                <AppIcon.BHCELogoFull size={300} className='text-base' />
                            </CardContent>
                        </Card>
                        <p>BHCELogoFull</p>
                    </div>
                    <div className='flex flex-col items-center justify-center'>
                        <Card className='flex items-center justify-center h-32  mb-2'>
                            <CardContent className='px-6'>
                                <AppIcon.BHELogoFull size={300} className='text-primary' />
                            </CardContent>
                        </Card>
                        <p>BHELogoFull</p>
                    </div>
                </div>
            </div>
            <div className='grid grid-cols-3 md:grid-cols-4 gap-x-4 gap-y-8 mb-4'>
                {nonLogos.map((x: string) => {
                    const MyComponent = AppIcon[x as keyof typeof AppIcon];
                    return (
                        <div key={x} className='flex flex-col items-center justify-center'>
                            <Card className='flex items-center justify-center h-24 w-24 mb-2'>
                                <CardContent>
                                    <MyComponent key={x} size={32} />
                                </CardContent>
                            </Card>
                            <p>{MyComponent.name}</p>
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

export { AppIcons };
