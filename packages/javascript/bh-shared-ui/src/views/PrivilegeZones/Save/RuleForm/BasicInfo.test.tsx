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

import { AssetGroupTagSelectorAutoCertifyDisabled } from 'js-client-library';
import React from 'react';
import { FormProvider, useForm } from 'react-hook-form';
import { UseQueryResult } from 'react-query';
import { useAssetGroupTagInfo, usePZPathParams } from '../../../../hooks';
import { render, screen } from '../../../../test-utils';
import { PrivilegeZonesContext, defaultPrivilegeZoneCtxValue } from '../../PrivilegeZonesContext';
import BasicInfo from './BasicInfo';
import RuleFormContext, { initialValue } from './RuleFormContext';
import { RuleFormInputs } from './types';

vi.mock('../../../../hooks', async (importOriginal) => {
    const original: Record<string, any> = await importOriginal();
    return {
        ...original,
        useAssetGroupTagInfo: vi.fn(),
        usePZPathParams: vi.fn(),
    };
});

const mockZoneName = 'Tier Zero';

const mockTagQuery = (requireCertify: boolean | null) =>
    ({
        data: {
            id: 1,
            name: mockZoneName,
            require_certify: requireCertify,
        },
        isLoading: false,
        isError: false,
        isSuccess: true,
    }) as unknown as UseQueryResult<any>;

const mockRuleQuery = {
    data: undefined,
    isLoading: false,
    isError: false,
    isSuccess: false,
} as unknown as UseQueryResult<any>;

const mockPZPathParamsZone = {
    tagId: '1',
    tagType: 'zones' as const,
    tagTypeDisplay: 'Zone' as const,
    tagTypeDisplayPlural: 'Zones' as const,
    ruleId: '',
    zoneId: '1',
    labelId: undefined,
    memberId: undefined,
    hasLabelId: false,
    hasZoneId: true,
    isLabelPage: false,
    isZonePage: true,
    tagEditLink: vi.fn(),
    tagCreateLink: vi.fn(),
    ruleCreateLink: vi.fn(),
    ruleEditLink: vi.fn(),
    tagSummaryLink: vi.fn(),
    tagDetailsLink: vi.fn(),
    ruleDetailsLink: vi.fn(),
    objectDetailsLink: vi.fn(),
};

const CertificationComponent = React.lazy(() => Promise.resolve({ default: () => <></> }));

const BasicInfoWrapper: React.FC<{ certification?: React.LazyExoticComponent<React.FC> }> = ({ certification }) => {
    const form = useForm<RuleFormInputs>({
        defaultValues: {
            name: '',
            description: '',
            seeds: [],
            auto_certify: String(AssetGroupTagSelectorAutoCertifyDisabled),
        },
    });

    return (
        <FormProvider {...form}>
            <PrivilegeZonesContext.Provider value={{ ...defaultPrivilegeZoneCtxValue, Certification: certification }}>
                <RuleFormContext.Provider value={{ ...initialValue, ruleQuery: mockRuleQuery }}>
                    <BasicInfo control={form.control} />
                </RuleFormContext.Provider>
            </PrivilegeZonesContext.Provider>
        </FormProvider>
    );
};

describe('BasicInfo', () => {
    beforeEach(() => {
        vi.mocked(usePZPathParams).mockReturnValue(mockPZPathParamsZone as any);
    });

    describe('when isCertificationDisabledOnZoneLevel is false', () => {
        beforeEach(() => {
            vi.mocked(useAssetGroupTagInfo).mockReturnValue(mockTagQuery(false));
        });

        it('shows the Automatic Certification label and the zone-level disabled message', async () => {
            render(<BasicInfoWrapper certification={CertificationComponent} />);

            expect(await screen.findByText('Automatic Certification')).toBeInTheDocument();

            const disabledMessage = screen.getByText(/Certification disabled by the Zone's settings\. Please edit/);
            expect(disabledMessage).toHaveTextContent(
                `Certification disabled by the Zone's settings. Please edit ${mockZoneName} settings to manage rule-specific certification settings.`
            );

            expect(screen.queryByText(/Choose how new objects are certified\./)).not.toBeInTheDocument();
        });
    });

    describe('when isCertificationDisabledOnZoneLevel is true', () => {
        beforeEach(() => {
            vi.mocked(useAssetGroupTagInfo).mockReturnValue(mockTagQuery(true));
        });

        it('shows the Automatic Certification label and the certification dropdown with its full description', async () => {
            render(<BasicInfoWrapper certification={CertificationComponent} />);

            expect(await screen.findByText('Automatic Certification')).toBeInTheDocument();
            expect(screen.getByText(/Choose how new objects are certified\./)).toBeInTheDocument();

            expect(screen.getByText('Direct Objects')).toBeInTheDocument();
            expect(
                screen.getByText(
                    /Only the object explicitly selected either by object ID or cypher query are certified automatically\./
                )
            ).toBeInTheDocument();

            expect(screen.getByText('All Objects')).toBeInTheDocument();
            expect(
                screen.getByText(
                    /Means every object, including those tied to direct objects, is certified automatically\./
                )
            ).toBeInTheDocument();

            expect(screen.getByText(/Means all certification is manual\./)).toBeInTheDocument();

            expect(screen.queryByText(/Certification disabled by the Zone's settings/)).not.toBeInTheDocument();
        });
    });
});
