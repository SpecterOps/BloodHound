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

import { SelectedEnvironment } from '../../../components/SimpleEnvironmentSelector/types';
import { render, screen } from '../../../test-utils';
import { getStatsComponent } from './getStatsComponent';

vi.mock('../DomainInfo', () => ({
    DomainInfo: ({ contextId }: { contextId: string }) => <div data-testid='domain-info'>{contextId}</div>,
    ActiveDirectoryPlatformInfo: () => <div data-testid='ad-platform-info' />,
}));

vi.mock('../TenantInfo', () => ({
    default: ({ contextId }: { contextId: string }) => <div data-testid='tenant-info'>{contextId}</div>,
    AzurePlatformInfo: () => <div data-testid='azure-platform-info' />,
}));

vi.mock('../OpenGraphInfo', () => ({
    default: ({ contextId }: { contextId: string }) => <div data-testid='open-graph-info'>{contextId}</div>,
    OpenGraphPlatformInfo: ({ contextKindId }: { contextKindId: number }) => (
        <div data-testid='open-graph-platform-info'>{contextKindId}</div>
    ),
}));

const dataErrorHandler = vi.fn();

const renderStats = (selectedEnvironment: SelectedEnvironment | null) => {
    const result = getStatsComponent(selectedEnvironment, dataErrorHandler);
    if (result) render(result);
    return result;
};

describe('getStatsComponent', () => {
    it('renders DomainInfo for a collected active-directory environment', () => {
        renderStats({ type: 'active-directory', id: 'ad-id' });

        expect(screen.getByTestId('domain-info')).toHaveTextContent('ad-id');
    });

    it('returns null for active-directory without an id', () => {
        expect(renderStats({ type: 'active-directory', id: null })).toBeNull();
    });

    it('renders ActiveDirectoryPlatformInfo for the active-directory-platform aggregate', () => {
        renderStats({ type: 'active-directory-platform', id: null });

        expect(screen.getByTestId('ad-platform-info')).toBeInTheDocument();
    });

    it('renders TenantInfo for a collected azure environment', () => {
        renderStats({ type: 'azure', id: 'az-id' });

        expect(screen.getByTestId('tenant-info')).toHaveTextContent('az-id');
    });

    it('returns null for azure without an id', () => {
        expect(renderStats({ type: 'azure', id: null })).toBeNull();
    });

    it('renders AzurePlatformInfo for the azure-platform aggregate', () => {
        renderStats({ type: 'azure-platform', id: null });

        expect(screen.getByTestId('azure-platform-info')).toBeInTheDocument();
    });

    it('renders OpenGraphPlatformInfo for an open graph platform aggregate with a kind id', () => {
        renderStats({ type: 'aws-platform', id: null, environment_kind_id: 101 });

        expect(screen.getByTestId('open-graph-platform-info')).toHaveTextContent('101');
    });

    it('returns null for an open graph platform aggregate without a kind id', () => {
        expect(renderStats({ type: 'aws-platform', id: null })).toBeNull();
    });

    it('renders OpenGraphInfo for a collected open graph environment with an id', () => {
        renderStats({ type: 'aws', id: 'og-id' });

        expect(screen.getByTestId('open-graph-info')).toHaveTextContent('og-id');
    });

    it('returns null for an open graph environment without an id', () => {
        expect(renderStats({ type: 'aws', id: null })).toBeNull();
    });

    it('returns null when no environment is selected', () => {
        expect(renderStats(null)).toBeNull();
    });
});
