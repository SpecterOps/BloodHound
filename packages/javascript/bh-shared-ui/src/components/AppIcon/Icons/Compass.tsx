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

import React from 'react';
import { BasePath, BaseSVG, BaseSVGProps } from './utils';

export const Compass: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='compass'
            version='1.1'
            xmlns='http://www.w3.org/2000/svg'
            width='768'
            height='768'
            viewBox='0 0 768 768'
            {...props}>
            <g id='icomoon-ignore'></g>
            <BasePath d='M657.001 384c0-72.404-28.762-141.842-79.961-193.040-51.196-51.196-120.635-79.96-193.040-79.96s-141.842 28.762-193.040 79.96-79.96 120.637-79.96 193.040c0 72.405 28.763 141.842 79.96 193.039 51.196 51.199 120.637 79.961 193.040 79.961s141.842-28.762 193.040-79.961c51.199-51.196 79.961-120.635 79.961-193.040zM47.999 384c0-89.113 35.399-174.576 98.413-237.588s148.475-98.413 237.588-98.413c89.114 0 174.576 35.399 237.589 98.413 63.009 63.012 98.412 148.475 98.412 237.588 0 89.114-35.4 174.576-98.412 237.589-63.013 63.009-148.475 98.412-237.589 98.412s-174.576-35.4-237.588-98.412c-63.011-63.013-98.413-148.475-98.413-237.589zM450.545 474.694l-189.395 72.845c-25.462 9.845-50.532-15.224-40.688-40.691l72.844-189.392c4.331-11.157 12.994-19.82 24.151-24.151l189.392-72.844c25.465-9.844 50.534 15.224 40.69 40.688l-72.845 189.395c-4.199 11.156-12.993 19.818-24.148 24.149zM425.999 384c0-11.14-4.425-21.822-12.301-29.698s-18.562-12.301-29.698-12.301c-11.14 0-21.822 4.425-29.698 12.301s-12.301 18.56-12.301 29.698c0 11.139 4.425 21.822 12.301 29.698s18.56 12.301 29.698 12.301c11.139 0 21.822-4.425 29.698-12.301s12.301-18.562 12.301-29.699z' />
        </BaseSVG>
    );
};

export default Compass;
