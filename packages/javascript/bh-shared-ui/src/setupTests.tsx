// Copyright 2023 Specter Ops, Inc.
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

/* eslint-disable @typescript-eslint/no-unused-vars */
import matchers from '@testing-library/jest-dom/matchers';
import { expect } from 'vitest';
//@ts-ignore
import React from 'react';
//@ts-ignore
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import 'whatwg-fetch';

// jest-dom extensions
expect.extend(matchers);

// mocks
beforeEach(() => {
    vi.clearAllMocks();
});

// See https://fontawesome.com/v5.15/how-to-use/on-the-web/using-with/react#unit-testing for more information
vi.mock('@fortawesome/react-fontawesome', () => ({
    FontAwesomeIcon: vi.fn((props) => {
        if (typeof props.icon === 'string') return <span>{props.icon}</span>;

        return <span>{props.icon.iconName}</span>;
    }),
}));
