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
import { Card } from 'doodle-ui';
import { SchemaUploadDialog } from '../../components/SchemaUploadDialog/SchemaUploadDialog';

export const SchemaUploadCard = () => {
    return (
        <Card className='flex flex-col p-6 gap-4'>
            <h2 className='text-xl font-bold'>Extension Installation</h2>
            <div>
                <p>Install an OpenGraph Extension Bundle or JSON files to expand platform functionality, including:</p>
                <ul className='list-disc pl-6 mt-2'>
                    <li>Introducing new types of nodes and relationships</li>
                    <li>Adding rich contextual information to the attack graph</li>
                    <li>Modeling your specific technologies, environments, and workflows</li>
                </ul>
            </div>
            <SchemaUploadDialog />
        </Card>
    );
};
