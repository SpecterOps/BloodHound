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

import { cva } from 'class-variance-authority';

export const LabelVariants = cva('leading-none peer-disabled:cursor-not-allowed peer-disabled:text-disabled', {
    variants: {
        size: {
            small: 'text-sm font-medium',
            medium: 'text-base font-bold',
            large: 'text-lg font-bold',
        },
    },
    defaultVariants: {
        size: 'medium',
    },
});
