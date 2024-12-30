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

import { PrismaClient } from '@prisma/client';
// Database connection will increase once we create more BDD scenarios.
// Per Prisma best practice: create a single prisma instance for each worker to perform database operations
// context: https://www.prisma.io/docs/orm/prisma-client/setup-and-configuration/databases-connections#re-using-a-single-prismaclient-instance
let prisma = new PrismaClient();
export default prisma;
