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

import { Button, Card } from '@bloodhoundenterprise/doodleui';

export const SchemaUploadCard = () => {
    return (
        <Card className='flex flex-col p-6 gap-4'>
            <h2 className='text-xl font-bold'>Custom Schema Upload</h2>

            <p>
                Upload custom schema JSON files to introduce new node and edge types. Then apply and validate schema
                updates to tailor the attack graph model to specific environments, workflows, or needs.
            </p>

            <Button className='self-start' variant='secondary' disabled={true}>
                Upload File
            </Button>
        </Card>
    );
};
