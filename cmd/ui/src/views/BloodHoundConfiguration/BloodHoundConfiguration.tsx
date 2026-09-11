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

import { AnalyzeNowConfiguration, CitrixRDPConfiguration, PageWithTitle } from 'bh-shared-ui';

const BloodHoundConfiguration = () => {
    return (
        <PageWithTitle
            title='BloodHound Configuration'
            pageDescription={
                <p className='text-sm'>
                    Modify the configuration of your BloodHound tenant. See our{' '}
                    <a
                        className='text-link underline'
                        href='https://bloodhound.specterops.io/analyze-data/bloodhound-gui/configuration'>
                        documentation
                    </a>{' '}
                    for more details on each option.
                </p>
            }>
            <div className='flex flex-col gap-6 mt-4'>
                <AnalyzeNowConfiguration description='This will re-run analysis in the BloodHound environment, recreating all Attack Paths that exist as a result of complex configurations.' />
                <CitrixRDPConfiguration />
            </div>
        </PageWithTitle>
    );
};

export default BloodHoundConfiguration;
