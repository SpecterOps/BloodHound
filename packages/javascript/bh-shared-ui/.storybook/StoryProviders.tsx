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
import { ReactNode, useState } from 'react';
import { HelmetProvider } from 'react-helmet-async';
import { QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router-dom';
import { AppNameProvider } from '../src';
import { AnnouncementProvider } from '../src/providers/AnnouncementProvider';

interface StoryProvidersProps {
    children: ReactNode;
    initialEntries?: string[];
}

export const StoryProviders = ({ children, initialEntries = ['/'] }: StoryProvidersProps) => {
    const [queryClient] = useState(
        () =>
            new QueryClient({
                defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false }, mutations: { retry: false } },
            })
    );

    return (
        <HelmetProvider>
            <QueryClientProvider client={queryClient}>
                <MemoryRouter
                    initialEntries={initialEntries}
                    future={{ v7_relativeSplatPath: true, v7_startTransition: true }}>
                    <AppNameProvider name='BloodHound Community Edition'>
                        <AnnouncementProvider>{children}</AnnouncementProvider>
                    </AppNameProvider>
                </MemoryRouter>
            </QueryClientProvider>
        </HelmetProvider>
    );
};
