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

import { AppIcon, PageWithTitle } from 'bh-shared-ui';
import {
    ButtonVariants,
    Input,
    Link,
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
    Typography,
} from 'doodle-ui';
import { useState } from 'react';
import ProductCard from './ProductCard';
import {
    COMMUNITY_EXTENSIONS_DISCLAIMER,
    communityExtensions,
    communityIntegrations,
    enterpriseExtensions,
    enterpriseIntegrations,
    type CatalogItem,
    type MarketplaceAvailability,
    type MarketplacePublisher,
} from './marketplaceCatalog';

type MarketplaceTypeFilter = 'all' | 'extensions' | 'integrations';
type MarketplacePublisherFilter = 'all' | MarketplacePublisher;
type MarketplaceAvailabilityFilter = 'all' | MarketplaceAvailability;

const typeFilters: { label: string; value: MarketplaceTypeFilter }[] = [
    { label: 'All types', value: 'all' },
    { label: 'OG Extensions', value: 'extensions' },
    { label: 'Integrations', value: 'integrations' },
];

const publisherFilters: { label: string; value: MarketplacePublisherFilter }[] = [
    { label: 'All publishers', value: 'all' },
    { label: 'SpecterOps', value: 'specterops' },
    { label: 'Community', value: 'community' },
    { label: 'Integration Partners', value: 'partner' },
];

const availabilityFilters: { label: string; value: MarketplaceAvailabilityFilter }[] = [
    { label: 'All availability', value: 'all' },
    { label: 'Generally Available', value: 'general' },
    { label: 'Early Access', value: 'early-access' },
];

const matchesSearch = (item: CatalogItem, search: string) =>
    `${item.name} ${item.author} ${item.description}`.toLowerCase().includes(search);

const matchesPublisher = (item: CatalogItem, publisher: MarketplacePublisherFilter) =>
    publisher === 'all' || item.publisher === publisher;

const matchesAvailability = (item: CatalogItem, availability: MarketplaceAvailabilityFilter) =>
    availability === 'all' || item.availability === availability;

const formatItemCount = (count: number) => `${count} item${count === 1 ? '' : 's'}`;
const sectionDividerClasses = 'mt-8 border-t border-neutral-light-4 pt-8 dark:border-neutral-dark-4';

const Marketplace = () => {
    const [search, setSearch] = useState('');
    const [typeFilter, setTypeFilter] = useState<MarketplaceTypeFilter>('all');
    const [publisherFilter, setPublisherFilter] = useState<MarketplacePublisherFilter>('all');
    const [availabilityFilter, setAvailabilityFilter] = useState<MarketplaceAvailabilityFilter>('all');
    const normalizedSearch = search.trim().toLowerCase();
    const filterCatalog = (item: CatalogItem) =>
        matchesSearch(item, normalizedSearch) &&
        matchesPublisher(item, publisherFilter) &&
        matchesAvailability(item, availabilityFilter);
    const filteredEnterpriseExtensions = enterpriseExtensions.filter(filterCatalog);
    const filteredCommunityExtensions = communityExtensions.filter(filterCatalog);
    const filteredEnterpriseIntegrations = enterpriseIntegrations.filter(filterCatalog);
    const filteredCommunityIntegrations = communityIntegrations.filter(filterCatalog);
    const showEnterpriseExtensions = typeFilter !== 'integrations' && filteredEnterpriseExtensions.length > 0;
    const showCommunityExtensions = typeFilter !== 'integrations' && filteredCommunityExtensions.length > 0;
    const showEnterpriseIntegrations = typeFilter !== 'extensions' && filteredEnterpriseIntegrations.length > 0;
    const showCommunityIntegrations = typeFilter !== 'extensions' && filteredCommunityIntegrations.length > 0;

    return (
        <PageWithTitle className='pb-5' data-testid='marketplace' title='Marketplace'>
            <div className='mb-8 flex flex-col gap-4 lg:flex-row lg:flex-wrap lg:items-center'>
                <div className='relative w-full lg:w-96 xl:w-[30rem]'>
                    <AppIcon.MagnifyingGlass
                        aria-hidden='true'
                        className='pointer-events-none absolute left-3 top-1/2 z-10 -translate-y-1/2 text-muted-foreground'
                        size={20}
                    />
                    <Input
                        aria-label='Search Marketplace'
                        className='bg-white pl-10 dark:bg-neutral-dark-2'
                        onChange={(event) => setSearch(event.target.value)}
                        placeholder='Search extensions and integrations...'
                        type='search'
                        value={search}
                        variant='outlined'
                    />
                </div>
                <div className='flex items-center gap-2'>
                    <label className='font-medium text-muted-foreground' htmlFor='marketplace-type-filter'>
                        Type
                    </label>
                    <Select onValueChange={(value: MarketplaceTypeFilter) => setTypeFilter(value)} value={typeFilter}>
                        <SelectTrigger
                            aria-label='Filter Marketplace items by type'
                            className='w-40'
                            id='marketplace-type-filter'>
                            <SelectValue />
                        </SelectTrigger>
                        <SelectPortal>
                            <SelectContent>
                                {typeFilters.map(({ label, value }) => (
                                    <SelectItem key={value} value={value}>
                                        {label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </SelectPortal>
                    </Select>
                </div>
                <div className='flex items-center gap-2'>
                    <label className='font-medium text-muted-foreground' htmlFor='marketplace-publisher-filter'>
                        Publisher
                    </label>
                    <Select
                        onValueChange={(value: MarketplacePublisherFilter) => setPublisherFilter(value)}
                        value={publisherFilter}>
                        <SelectTrigger
                            aria-label='Filter Marketplace items by publisher'
                            className='w-48'
                            id='marketplace-publisher-filter'>
                            <SelectValue />
                        </SelectTrigger>
                        <SelectPortal>
                            <SelectContent>
                                {publisherFilters.map(({ label, value }) => (
                                    <SelectItem key={value} value={value}>
                                        {label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </SelectPortal>
                    </Select>
                </div>
                <div className='flex items-center gap-2'>
                    <label className='font-medium text-muted-foreground' htmlFor='marketplace-availability-filter'>
                        Availability
                    </label>
                    <Select
                        onValueChange={(value: MarketplaceAvailabilityFilter) => setAvailabilityFilter(value)}
                        value={availabilityFilter}>
                        <SelectTrigger
                            aria-label='Filter Marketplace items by availability'
                            className='w-48'
                            id='marketplace-availability-filter'>
                            <SelectValue />
                        </SelectTrigger>
                        <SelectPortal>
                            <SelectContent>
                                {availabilityFilters.map(({ label, value }) => (
                                    <SelectItem key={value} value={value}>
                                        {label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </SelectPortal>
                    </Select>
                </div>
            </div>

            <div>
                {showEnterpriseExtensions && (
                    <section aria-label='OpenGraph Enterprise Extensions'>
                        <Typography className='mb-2' component='h2' variant='h2'>
                            OpenGraph Enterprise Extensions{' '}
                            <span className='ml-1 text-base font-normal text-muted-foreground'>
                                · {formatItemCount(filteredEnterpriseExtensions.length)}
                            </span>
                        </Typography>
                        <Typography className='mb-5 text-muted-foreground' variant='body2'>
                            SpecterOps-authored extensions for additional identity and infrastructure platforms.
                            Automated analysis, findings, trends, and reporting for these extensions require BloodHound
                            Enterprise.
                        </Typography>
                        <div
                            aria-label='BloodHound Enterprise extension capabilities'
                            className='mb-5 flex flex-col gap-4 rounded-md bg-status-info-fill p-4 text-status-info-text lg:flex-row lg:items-center lg:justify-between'
                            role='note'>
                            <div className='flex items-start gap-3'>
                                <AppIcon.Info
                                    aria-hidden='true'
                                    className='mt-0.5 shrink-0 text-status-info-main'
                                    size={24}
                                />
                                <div>
                                    <Typography component='h3' variant='h3'>
                                        Unlock automated analysis with BloodHound Enterprise
                                    </Typography>
                                    <Typography className='mt-1' variant='body2'>
                                        Enterprise Extensions are listed here for discovery. In BloodHound Community
                                        Edition, these extensions do not include the continuous collection, automated
                                        analysis, findings, prioritization, trends, and reporting available with
                                        BloodHound Enterprise.
                                    </Typography>
                                </div>
                            </div>
                            <Link
                                className={`${ButtonVariants({ variant: 'secondary' })} shrink-0 self-start rounded-md font-bold lg:self-auto`}
                                href='https://specterops.io/get-a-demo/'
                                variant='unstyled'>
                                Learn More About BloodHound Enterprise
                            </Link>
                        </div>
                        <div className='grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-4'>
                            {filteredEnterpriseExtensions.map((extension) => (
                                <ProductCard
                                    key={extension.name}
                                    {...extension}
                                    categoryLabel='Enterprise Extension'
                                    linkLabel='Learn More'
                                />
                            ))}
                        </div>
                    </section>
                )}

                {showCommunityExtensions && (
                    <section
                        aria-label='Community Extensions'
                        className={showEnterpriseExtensions ? sectionDividerClasses : undefined}>
                        <Typography className='mb-2' component='h2' variant='h2'>
                            Community Extensions{' '}
                            <span className='ml-1 text-base font-normal text-muted-foreground'>
                                · {formatItemCount(filteredCommunityExtensions.length)}
                            </span>
                        </Typography>
                        <Typography className='mb-4 text-muted-foreground' variant='body2'>
                            Open-source OpenGraph extensions authored and maintained by community members.
                        </Typography>
                        <div
                            aria-label='Community Extensions disclaimer'
                            className='mb-5 rounded-md border border-neutral-light-4 bg-neutral-light-1 p-4 dark:border-neutral-dark-4 dark:bg-neutral-dark-2'>
                            <Typography className='text-muted-foreground' variant='body2'>
                                {COMMUNITY_EXTENSIONS_DISCLAIMER}
                            </Typography>
                        </div>
                        <div className='grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-4'>
                            {filteredCommunityExtensions.map((extension) => (
                                <ProductCard
                                    key={extension.name}
                                    {...extension}
                                    categoryLabel='Community Extension'
                                    linkLabel='View on GitHub'
                                />
                            ))}
                        </div>
                    </section>
                )}

                {showEnterpriseIntegrations && (
                    <section
                        aria-label='Enterprise Integrations'
                        className={
                            showEnterpriseExtensions || showCommunityExtensions ? sectionDividerClasses : undefined
                        }>
                        <Typography className='mb-2' component='h2' variant='h2'>
                            Enterprise Integrations{' '}
                            <span className='ml-1 text-base font-normal text-muted-foreground'>
                                · {formatItemCount(filteredEnterpriseIntegrations.length)}
                            </span>
                        </Typography>
                        <Typography className='mb-5 text-muted-foreground' variant='body2'>
                            Supported integrations that connect BloodHound Enterprise with security, automation, and IT
                            operations platforms.
                        </Typography>
                        <div
                            aria-label='BloodHound Enterprise integration capabilities'
                            className='mb-5 flex flex-col gap-4 rounded-md bg-status-info-fill p-4 text-status-info-text lg:flex-row lg:items-center lg:justify-between'
                            role='note'>
                            <div className='flex items-start gap-3'>
                                <AppIcon.Info
                                    aria-hidden='true'
                                    className='mt-0.5 shrink-0 text-status-info-main'
                                    size={24}
                                />
                                <div>
                                    <Typography component='h3' variant='h3'>
                                        Connect your security ecosystem with BloodHound Enterprise
                                    </Typography>
                                    <Typography className='mt-1' variant='body2'>
                                        Enterprise Integrations are listed here for discovery. They connect BloodHound
                                        Enterprise with supported security, automation, and IT operations platforms and
                                        are not available in BloodHound Community Edition.
                                    </Typography>
                                </div>
                            </div>
                            <Link
                                className={`${ButtonVariants({ variant: 'secondary' })} shrink-0 self-start rounded-md font-bold lg:self-auto`}
                                href='https://specterops.io/get-a-demo/'
                                variant='unstyled'>
                                Learn More About BloodHound Enterprise
                            </Link>
                        </div>
                        <div className='grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-4'>
                            {filteredEnterpriseIntegrations.map((integration) => (
                                <ProductCard
                                    key={integration.name}
                                    {...integration}
                                    categoryLabel='Enterprise Integration'
                                    linkLabel='Learn More'
                                />
                            ))}
                        </div>
                    </section>
                )}

                {showCommunityIntegrations && (
                    <section
                        aria-label='Community Integrations'
                        className={
                            showEnterpriseExtensions || showCommunityExtensions || showEnterpriseIntegrations
                                ? sectionDividerClasses
                                : undefined
                        }>
                        <Typography className='mb-2' component='h2' variant='h2'>
                            Community Integrations{' '}
                            <span className='ml-1 text-base font-normal text-muted-foreground'>
                                · {formatItemCount(filteredCommunityIntegrations.length)}
                            </span>
                        </Typography>
                        <Typography className='mb-5 text-muted-foreground' variant='body2'>
                            Community tools that connect to BloodHound Community Edition for analysis, automation, and
                            local workflows.
                        </Typography>
                        <div className='grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-4'>
                            {filteredCommunityIntegrations.map((integration) => (
                                <ProductCard
                                    key={integration.name}
                                    {...integration}
                                    categoryLabel='Community Integration'
                                    linkLabel='Learn More'
                                />
                            ))}
                        </div>
                    </section>
                )}

                {!showEnterpriseExtensions &&
                    !showCommunityExtensions &&
                    !showEnterpriseIntegrations &&
                    !showCommunityIntegrations && (
                        <Typography className='py-12 text-center text-muted-foreground' role='status' variant='body1'>
                            No Marketplace items match your search and filters.
                        </Typography>
                    )}
            </div>
        </PageWithTitle>
    );
};

export default Marketplace;
