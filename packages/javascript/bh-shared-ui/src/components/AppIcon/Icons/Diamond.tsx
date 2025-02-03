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

export const Diamond: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG name='diamond' version='1.1' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 768 768' {...props}>
            <g id='diamond'></g>
            <BasePath d='M271.525 156.745l112.502 117.448 112.501-117.448h-224.998zM548.467 190.97l-93.086 97.113h165.857l-72.772-97.113zM612.755 348.703h-457.457l228.729 247.905 228.726-247.905zM146.812 288.083h165.856l-93.088-97.113-72.772 97.113zM705.199 338.726l-298.288 323.298c-5.785 6.314-14.141 9.977-22.887 9.977s-16.972-3.663-22.887-9.977l-298.289-323.298c-9.902-10.735-10.672-26.774-1.928-38.391l143.999-191.958c5.785-7.702 15.045-12.375 24.816-12.375h308.576c9.77 0 19.028 4.545 24.814 12.375l143.998 191.958c8.741 11.618 7.841 27.656-1.929 38.391z' />
        </BaseSVG>
    );
};

export default Diamond;
