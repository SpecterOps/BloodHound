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

import userEvent from '@testing-library/user-event';
import { render, screen, within } from 'src/test-utils';
import Marketplace from './Marketplace';
import { COMMUNITY_EXTENSIONS_DISCLAIMER } from './marketplaceCatalog';

describe('Marketplace', () => {
    it('renders the CE catalog, outbound links, and complete community disclaimer', () => {
        render(<Marketplace />);

        expect(screen.getByRole('heading', { name: 'Marketplace' })).toBeInTheDocument();
        const enterpriseExtensionsRegion = screen.getByRole('region', { name: 'OpenGraph Enterprise Extensions' });
        const extensionsRegion = screen.getByRole('region', { name: 'Community Extensions' });
        const enterpriseIntegrationsRegion = screen.getByRole('region', { name: 'Enterprise Integrations' });
        const communityIntegrationsRegion = screen.getByRole('region', { name: 'Community Integrations' });
        expect(enterpriseExtensionsRegion).toHaveTextContent('· 5 items');
        expect(extensionsRegion).toHaveTextContent('· 26 items');
        expect(enterpriseIntegrationsRegion).toHaveTextContent('· 11 items');
        expect(communityIntegrationsRegion).toHaveTextContent('· 3 items');
        expect(screen.getAllByRole('article')).toHaveLength(45);
        expect(screen.getAllByText('Enterprise Extension')).toHaveLength(5);
        expect(screen.getAllByText('Community Extension')).toHaveLength(26);
        expect(screen.getAllByText('Enterprise Integration')).toHaveLength(11);
        expect(screen.getAllByText('Community Integration')).toHaveLength(3);
        expect(screen.getByLabelText('Community Extensions disclaimer')).toHaveTextContent(
            COMMUNITY_EXTENSIONS_DISCLAIMER
        );

        const onePasswordCard = screen.getByRole('article', { name: '1Password' });
        expect(within(onePasswordCard).getByText('SpecterOps')).toBeInTheDocument();
        expect(within(onePasswordCard).getByTestId('product-logo')).toHaveAttribute(
            'src',
            '/img/product-logos/community-1password.png'
        );
        const onePasswordLink = within(onePasswordCard).getByRole('link', { name: /View on GitHub 1Password/ });
        expect(onePasswordLink).toHaveAttribute('href', 'https://github.com/SpecterOps/1PassHound');
        expect(onePasswordLink).toHaveAttribute('target', '_blank');
        expect(onePasswordLink).toHaveAttribute('rel', 'noopener noreferrer');

        const awsCard = screen.getByRole('article', { name: 'AWS' });
        expect(within(awsCard).getByText('Beta')).toBeInTheDocument();
        expect(within(awsCard).getByTestId('product-logo')).toHaveAttribute('src', '/img/product-logos/aws.svg');

        const enterpriseNote = screen.getByRole('note', { name: 'BloodHound Enterprise extension capabilities' });
        expect(enterpriseNote).toHaveTextContent('Unlock automated analysis with BloodHound Enterprise');
        expect(enterpriseNote).toHaveTextContent(
            'In BloodHound Community Edition, these extensions do not include the continuous collection, automated analysis, findings, prioritization, trends, and reporting available with BloodHound Enterprise.'
        );
        const enterpriseLink = within(enterpriseNote).getByRole('link', {
            name: /Learn More About BloodHound Enterprise/,
        });
        expect(enterpriseLink).toHaveAttribute('href', 'https://specterops.io/get-a-demo/');
        expect(enterpriseLink).toHaveAttribute('target', '_blank');
        expect(enterpriseLink).toHaveAttribute('rel', 'noopener noreferrer');

        const integrationNote = screen.getByRole('note', {
            name: 'BloodHound Enterprise integration capabilities',
        });
        expect(integrationNote).toHaveTextContent('Connect your security ecosystem with BloodHound Enterprise');
        expect(integrationNote).toHaveTextContent(
            'Enterprise Integrations are listed here for discovery. They connect BloodHound Enterprise with supported security, automation, and IT operations platforms and are not available in BloodHound Community Edition.'
        );
        expect(
            within(integrationNote).getByRole('link', { name: /Learn More About BloodHound Enterprise/ })
        ).toHaveAttribute('href', 'https://specterops.io/get-a-demo/');

        const mcpCard = screen.getByRole('article', { name: 'BloodHound MCP' });
        expect(within(mcpCard).getByText('mwnickerson')).toBeInTheDocument();
        expect(within(mcpCard).getByRole('link', { name: /Learn More BloodHound MCP/ })).toHaveAttribute(
            'href',
            'https://github.com/mwnickerson/bloodhound_mcp'
        );
        expect(screen.queryByText(/upgrade your license/i)).not.toBeInTheDocument();
    });

    it('searches by item content and filters by type, publisher, and availability', async () => {
        const user = userEvent.setup();
        render(<Marketplace />);

        const search = screen.getByRole('searchbox', { name: 'Search Marketplace' });
        const typeFilter = screen.getByRole('combobox', { name: 'Filter Marketplace items by type' });
        const publisherFilter = screen.getByRole('combobox', { name: 'Filter Marketplace items by publisher' });
        const availabilityFilter = screen.getByRole('combobox', {
            name: 'Filter Marketplace items by availability',
        });
        expect(typeFilter).toHaveTextContent('All types');
        expect(publisherFilter).toHaveTextContent('All publishers');
        expect(availabilityFilter).toHaveTextContent('All availability');

        await user.type(search, 'terminal companion');
        expect(screen.getByRole('article', { name: 'CypherHound' })).toBeInTheDocument();
        expect(screen.getAllByRole('article')).toHaveLength(1);
        expect(screen.queryByRole('region', { name: 'Community Extensions' })).not.toBeInTheDocument();

        await user.clear(search);
        await user.click(typeFilter);
        await user.click(screen.getByRole('option', { name: 'OG Extensions' }));
        expect(screen.getByRole('region', { name: 'Community Extensions' })).toBeInTheDocument();
        expect(screen.queryByRole('region', { name: 'Enterprise Integrations' })).not.toBeInTheDocument();
        expect(screen.queryByRole('region', { name: 'Community Integrations' })).not.toBeInTheDocument();

        await user.click(publisherFilter);
        await user.click(screen.getByRole('option', { name: 'SpecterOps' }));
        expect(screen.getAllByRole('article')).toHaveLength(8);
        expect(screen.getByRole('article', { name: 'MSSQL' })).toBeInTheDocument();

        await user.click(availabilityFilter);
        await user.click(screen.getByRole('option', { name: 'Early Access' }));
        expect(screen.getAllByRole('article')).toHaveLength(2);
        expect(screen.getByRole('article', { name: 'AWS' })).toBeInTheDocument();
        expect(screen.getByRole('article', { name: 'Microsoft Entra Agent ID' })).toBeInTheDocument();

        await user.click(typeFilter);
        await user.click(screen.getByRole('option', { name: 'Integrations' }));
        expect(screen.getByRole('status')).toHaveTextContent('No Marketplace items match your search and filters.');
        expect(screen.queryByLabelText('Community Extensions disclaimer')).not.toBeInTheDocument();

        await user.click(availabilityFilter);
        await user.click(screen.getByRole('option', { name: 'All availability' }));
        await user.click(publisherFilter);
        await user.click(screen.getByRole('option', { name: 'All publishers' }));
        expect(screen.getByRole('region', { name: 'Enterprise Integrations' })).toHaveTextContent('· 11 items');
        expect(screen.getByRole('region', { name: 'Community Integrations' })).toHaveTextContent('· 3 items');
        expect(screen.getAllByRole('article')).toHaveLength(14);

        await user.click(publisherFilter);
        await user.click(screen.getByRole('option', { name: 'Integration Partners' }));
        expect(screen.getAllByRole('article')).toHaveLength(4);
        expect(screen.getByRole('article', { name: 'Tines' })).toBeInTheDocument();
        expect(screen.queryByRole('region', { name: 'Community Integrations' })).not.toBeInTheDocument();
    });
});
