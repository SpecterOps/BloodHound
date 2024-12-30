// Copyright 2024 Specter Ops, Inc.
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

import { setWorldConstructor, World } from '@cucumber/cucumber';
import { IFixture } from '../../support/FixtureManager.js';

// Extend World base class properties and methods with PlayWrightWorld Class
export default class PlaywrightWorld extends World {
    fixture: IFixture;
}

setWorldConstructor(PlaywrightWorld);
