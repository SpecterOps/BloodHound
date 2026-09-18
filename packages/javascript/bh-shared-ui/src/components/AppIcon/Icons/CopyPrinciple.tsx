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

import React from 'react';
import { BasePath, BaseSVG, BaseSVGProps } from './utils';

export const CopyPrinciple: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='copy-principle'
            width='16'
            height='16'
            viewBox='0 0 16 16'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
            {...props}>
            <BasePath d='M11.2588 4.81168V0.888856C11.2588 0.653117 11.1652 0.427033 10.9985 0.26034C10.8318 0.0936471 10.6057 0 10.37 0H0.888856C0.653117 0 0.427033 0.0936471 0.26034 0.26034C0.0936471 0.427033 0 0.653117 0 0.888856V10.37C0 10.6057 0.0936471 10.8318 0.26034 10.9985C0.427033 11.1652 0.653117 11.2588 0.888856 11.2588H4.81168C4.97539 12.2834 5.41906 13.2427 6.0937 14.031C6.76833 14.8193 7.64769 15.4057 8.63466 15.7257C9.62164 16.0456 10.6778 16.0866 11.6866 15.844C12.6954 15.6015 13.6175 15.0848 14.3512 14.3512C15.0848 13.6175 15.6015 12.6954 15.844 11.6866C16.0866 10.6778 16.0456 9.62164 15.7257 8.63466C15.4057 7.64769 14.8193 6.76833 14.031 6.0937C13.2427 5.41906 12.2834 4.97539 11.2588 4.81168ZM1.77771 1.77771H9.48113V9.48113H1.77771V1.77771ZM10.37 14.2217C9.50401 14.219 8.66405 13.9254 7.98491 13.3881C7.30577 12.8508 6.82682 12.101 6.62494 11.2588H10.37C10.6057 11.2588 10.8318 11.1652 10.9985 10.9985C11.1652 10.8318 11.2588 10.6057 11.2588 10.37V6.61902C12.1806 6.83404 12.9909 7.3809 13.5351 8.15525C14.0793 8.9296 14.3194 9.87722 14.2095 10.8173C14.0996 11.7573 13.6474 12.624 12.9392 13.252C12.231 13.8799 11.3165 14.2251 10.37 14.2217Z' />
        </BaseSVG>
    );
};

export default CopyPrinciple;
