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

import { fromJS, List } from 'immutable';
import { render, screen } from '../../../test-utils';
import { OperationSummaryWithEdition } from './OperationsEditionPlugin';

const getComponent = (name: string) => {
    const components: Record<string, React.FC<any>> = {
        authorizeOperationBtn: () => null,
        OperationSummaryMethod: ({ method }) => <span>{method}</span>,
        OperationSummaryPath: () => <span>/api/v2/test</span>,
        JumpToPath: () => null,
    };

    return components[name];
};

const operationProps = fromJS({
    summary: 'Test endpoint',
    isAuthorized: false,
    method: 'get',
    op: { summary: 'Test endpoint' },
    showSummary: true,
    path: '/api/v2/test',
    displayOperationId: false,
});

const renderOperationSummary = ({
    isCommunity = true,
    isEnterprise = true,
}: {
    isCommunity?: boolean;
    isEnterprise?: boolean;
} = {}) =>
    render(
        <OperationSummaryWithEdition
            isShown={false}
            toggleShown={vi.fn()}
            getComponent={getComponent}
            authActions={{}}
            authSelectors={{}}
            operationProps={operationProps}
            specPath={List()}
            isCommunity={isCommunity}
            isEnterprise={isEnterprise}
        />
    );

describe('OperationSummaryWithEdition', () => {
    it('shows the Enterprise BloodHound dog logo in the Enterprise brand color when the endpoint is available', () => {
        renderOperationSummary({ isEnterprise: true });

        const [, enterpriseLogoLabel] = screen.getAllByText('app-icon-bh-logo');
        const enterpriseLogo = enterpriseLogoLabel.closest('svg');

        expect(enterpriseLogo).toHaveAttribute('aria-hidden', 'true');
        expect(enterpriseLogo).toHaveStyle({ color: 'var(--bhe-main)' });
        expect(enterpriseLogo).toHaveAttribute('width', '50px');
        expect(enterpriseLogo).toHaveAttribute('height', '33px');
        expect(enterpriseLogo).toHaveAttribute('viewBox', '0 2.75 24 18.5');
    });

    it('shows the Community BloodHound dog logo in the Community brand color when the endpoint is available', () => {
        renderOperationSummary({ isCommunity: true });
        const [communityLogoLabel] = screen.getAllByText('app-icon-bh-logo');
        const communityLogo = communityLogoLabel.closest('svg');

        expect(communityLogo).toHaveAttribute('aria-hidden', 'true');
        expect(communityLogo).toHaveStyle({ color: 'var(--bhce-main)' });
        expect(communityLogo).toHaveAttribute('width', '50px');
        expect(communityLogo).toHaveAttribute('height', '33px');
        expect(communityLogo).toHaveAttribute('viewBox', '0 2.75 24 18.5');
    });

    it('shows the Enterprise BloodHound dog logo in grey when the endpoint is unavailable', () => {
        renderOperationSummary({ isEnterprise: false });

        const [, enterpriseLogoLabel] = screen.getAllByText('app-icon-bh-logo');
        const enterpriseLogo = enterpriseLogoLabel.closest('svg');

        expect(enterpriseLogo?.style.color).toBe('grey');
    });

    it('includes edition availability in the operation button name', () => {
        renderOperationSummary({ isCommunity: true, isEnterprise: false });

        expect(screen.getByRole('button')).toHaveAccessibleName(
            'get /api/v2/test. Available in BloodHound Community Edition. Not available in BloodHound Enterprise'
        );
    });

    it('shows the Community BloodHound dog logo in grey when the endpoint is unavailable', () => {
        renderOperationSummary({ isCommunity: false });

        const [communityLogoLabel] = screen.getAllByText('app-icon-bh-logo');
        const communityLogo = communityLogoLabel.closest('svg');

        expect(communityLogo?.style.color).toBe('grey');
    });
});
