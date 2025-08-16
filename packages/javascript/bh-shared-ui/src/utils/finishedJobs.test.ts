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

import { type ScheduledJobDisplay } from 'js-client-library';

import { toCollected, toFormatted, toMins } from './finishedJobs';

const MOCK_FINISHED_JOB: ScheduledJobDisplay = {
    id: 22,
    client_id: '718c9b04-9394-42c0-9d53-c87b689e2d92',
    client_name: 'GOAD',
    event_id: 123,
    execution_time: '2024-08-15T21:24:52.366579Z',
    start_time: '2024-08-15T21:25:21.990437Z',
    end_time: '2024-08-15T21:26:43.033448Z',
    status: 2,
    status_message: 'The service collected successfully',
    session_collection: true,
    local_group_collection: true,
    ad_structure_collection: true,
    cert_services_collection: true,
    ca_registry_collection: true,
    dc_registry_collection: true,
    all_trusted_domains: true,
    domain_controller: '',
    ous: [],
    domains: [],
    domain_results: [],
};

describe('toCollected', () => {
    it('shows the collection methods for the given job', () => {
        expect(toCollected(MOCK_FINISHED_JOB)).toBe(
            'Sessions, Local Groups, AD Structure, Certificate Services, CA Registry, DC Registry'
        );
    });

    it('shows some collection methods for the given job', () => {
        const NO_COLLECTIONS_JOB = {
            ...MOCK_FINISHED_JOB,
            session_collection: false,
            local_group_collection: false,
            ad_structure_collection: false,
            cert_services_collection: false,
            ca_registry_collection: false,
            dc_registry_collection: false,
        };
        expect(toCollected(NO_COLLECTIONS_JOB)).toBe('');
    });

    it('shows no collection methods for the given job', () => {
        const SOME_COLLECTIONS_JOB = {
            ...MOCK_FINISHED_JOB,
            session_collection: false,
            local_group_collection: false,
            ad_structure_collection: false,
            cert_services_collection: false,
        };
        expect(toCollected(SOME_COLLECTIONS_JOB)).toBe('CA Registry, DC Registry');
    });
});

describe('toFormatted', () => {
    it('formats the date string', () => {
        const result = toFormatted('2024-01-01T15:30:00.500Z');
        // Server TZ might not match local dev TZ
        // Match format like '2024-01-01 09:30 CST'
        expect(result).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2} [A-Z]{3,4}$/);
    });
});

describe('toMins', () => {
    it('shows an interval in mins', () => {
        expect(toMins('2024-01-01T15:30:00.500Z', '2024-01-02T03:00:00.000Z')).toBe('689 Min');
    });
});
