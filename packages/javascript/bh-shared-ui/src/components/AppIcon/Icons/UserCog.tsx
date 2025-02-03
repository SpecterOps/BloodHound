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

export const UserCog: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG name='user-cog' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' {...props}>
            <BasePath d='M4.08496 17.5V16.85C4.08496 16.51 4.24496 16.19 4.49496 16.04C6.18496 15.03 8.11496 14.5 10.085 14.5C10.115 14.5 10.135 14.5 10.165 14.51C10.265 13.81 10.465 13.14 10.755 12.53C10.535 12.51 10.315 12.5 10.085 12.5C7.66496 12.5 5.40496 13.17 3.47496 14.32C2.59496 14.84 2.08496 15.82 2.08496 16.85V19.5H11.345C10.925 18.9 10.595 18.22 10.375 17.5H4.08496Z' />
            <BasePath d='M10.085 11.5C12.295 11.5 14.085 9.71 14.085 7.5C14.085 5.29 12.295 3.5 10.085 3.5C7.87496 3.5 6.08496 5.29 6.08496 7.5C6.08496 9.71 7.87496 11.5 10.085 11.5ZM10.085 5.5C11.185 5.5 12.085 6.4 12.085 7.5C12.085 8.6 11.185 9.5 10.085 9.5C8.98496 9.5 8.08496 8.6 8.08496 7.5C8.08496 6.4 8.98496 5.5 10.085 5.5Z' />
            <BasePath d='M20.835 15.5C20.835 15.28 20.805 15.08 20.775 14.87L21.915 13.86L20.915 12.13L19.465 12.62C19.145 12.35 18.785 12.14 18.385 11.99L18.085 10.5H16.085L15.785 11.99C15.385 12.14 15.025 12.35 14.705 12.62L13.255 12.13L12.255 13.86L13.395 14.87C13.365 15.08 13.335 15.28 13.335 15.5C13.335 15.72 13.365 15.92 13.395 16.13L12.255 17.14L13.255 18.87L14.705 18.38C15.025 18.65 15.385 18.86 15.785 19.01L16.085 20.5H18.085L18.385 19.01C18.785 18.86 19.145 18.65 19.465 18.38L20.915 18.87L21.915 17.14L20.775 16.13C20.805 15.92 20.835 15.72 20.835 15.5ZM17.085 17.5C15.985 17.5 15.085 16.6 15.085 15.5C15.085 14.4 15.985 13.5 17.085 13.5C18.185 13.5 19.085 14.4 19.085 15.5C19.085 16.6 18.185 17.5 17.085 17.5Z' />
        </BaseSVG>
    );
};

export default UserCog;
