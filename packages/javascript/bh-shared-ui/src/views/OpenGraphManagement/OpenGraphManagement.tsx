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

import PageWithTitle from '../../components/PageWithTitle';
import { ActiveExtensionsCard } from './ActiveExtensionsCard';
import { SchemaUploadCard } from './SchemaUploadCard';

const OpenGraphManagement: React.FC = () => {
    return (
        <PageWithTitle
            title='OpenGraph Management'
            pageDescription={
                <p className='text-sm'>
                    OpenGraph Management provides a centralized space to define and maintain the structures that shape
                    how BloodHound understands relationships in an environment.
                </p>
            }>
            {/* Cards */}
            <section className='flex flex-col gap-4 mt-4 max-w-4xl'>
                <SchemaUploadCard />
                <ActiveExtensionsCard />
            </section>
        </PageWithTitle>
    );
};

export default OpenGraphManagement;
