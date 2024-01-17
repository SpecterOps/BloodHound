// Copyright 2023 Specter Ops, Inc.
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

import DownloadCollectors from './DownloadCollectors';
import { screen, render, waitFor } from 'src/test-utils';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import userEvent from '@testing-library/user-event';
import fileDownload from 'js-file-download';
import { Mock } from 'vitest';

vi.mock('js-file-download');

const server = setupServer();
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('DownloadCollectors', () => {
    it('displays a list of available sharphound and azurehound files to download', async () => {
        const testSharpHoundCollectors = {
            data: {
                latest: 'v2.0.0',
                versions: [
                    {
                        version: 'v2.0.0',
                        sha256sum: '4D9579899C56B69C78352B5F85A86',
                        deprecated: false,
                    },
                    {
                        version: 'v1.0.0',
                        sha256sum: '92BD9C32CB12F396F364BF1A786ED',
                        deprecated: true,
                    },
                ],
            },
        };
        const testAzureHoundCollectors = {
            data: {
                latest: 'v2.0.0',
                versions: [
                    {
                        version: 'v2.0.0',
                        sha256sum: 'EF48BD61478B766828CCD1C325F86',
                    },
                    {
                        version: 'v1.0.0',
                        sha256sum: '24BFBD626755CC84F298F9632B3D1',
                    },
                ],
            },
        };
        server.use(
            rest.get('/api/v2/collectors/sharphound', (req, res, ctx) => {
                return res(ctx.json(testSharpHoundCollectors));
            }),
            rest.get('/api/v2/collectors/azurehound', (req, res, ctx) => {
                return res(ctx.json(testAzureHoundCollectors));
            })
        );

        render(<DownloadCollectors />);
        expect(screen.getByText('Download Collectors')).toBeInTheDocument();
        expect(screen.getByText('SharpHound')).toBeInTheDocument();
        expect(screen.getByText('AzureHound')).toBeInTheDocument();

        // wait for network calls to finish loading
        await waitFor(() => screen.getByRole('heading', { name: 'SharpHound v2.0.0 (Latest)' }));

        // expect a heading, download button and checksum download button for each reported sharphound and azurehound version
        expect(screen.getByRole('heading', { name: 'SharpHound v2.0.0 (Latest)' }));
        expect(screen.getByRole('button', { name: 'Download SharpHound v2.0.0 (.zip)' })).toBeInTheDocument();
        expect(
            screen.getByRole('button', {
                name: (accessibleName, element) => {
                    return element.textContent === '4D9579899C56B69C78352B5F85A86';
                },
            })
        );

        expect(screen.getByRole('heading', { name: 'SharpHound v1.0.0 (Deprecated)' }));
        expect(screen.getByRole('button', { name: 'Download SharpHound v1.0.0 (.zip)' })).toBeInTheDocument();
        expect(
            screen.getByRole('button', {
                name: (accessibleName, element) => {
                    return element.textContent === '92BD9C32CB12F396F364BF1A786ED';
                },
            })
        );

        expect(screen.getByRole('heading', { name: 'AzureHound v2.0.0 (Latest)' }));
        expect(screen.getByRole('button', { name: 'Download AzureHound v2.0.0 (.zip)' })).toBeInTheDocument();
        expect(
            screen.getByRole('button', {
                name: (accessibleName, element) => {
                    return element.textContent === 'EF48BD61478B766828CCD1C325F86';
                },
            })
        );

        expect(screen.getByRole('heading', { name: 'AzureHound v1.0.0' }));
        expect(screen.getByRole('button', { name: 'Download AzureHound v1.0.0 (.zip)' })).toBeInTheDocument();
        expect(
            screen.getByRole('button', {
                name: (accessibleName, element) => {
                    return element.textContent === '24BFBD626755CC84F298F9632B3D1';
                },
            })
        );
    });

    it('triggers a file download when a download button is clicked', async () => {
        const user = userEvent.setup();
        const testSharpHoundCollectors = {
            data: {
                latest: 'v2.0.0',
                versions: [
                    {
                        version: 'v2.0.0',
                        sha256sum: '4D9579899C56B69C78352B5F85A86',
                    },
                    {
                        version: 'v1.0.0',
                        sha256sum: '92BD9C32CB12F396F364BF1A786ED',
                    },
                ],
            },
        };
        const testAzureHoundCollectors = {
            data: {
                latest: 'v2.0.0',
                versions: [
                    {
                        version: 'v2.0.0',
                        sha256sum: 'EF48BD61478B766828CCD1C325F86',
                    },
                    {
                        version: 'v1.0.0',
                        sha256sum: '24BFBD626755CC84F298F9632B3D1',
                    },
                ],
            },
        };
        const testFilename = 'sharphound-v2.0.0.zip';
        const testChecksumFilename = 'sharphound-v2.0.0.zip.sha256';
        server.use(
            rest.get('/api/v2/collectors/sharphound', (req, res, ctx) => {
                return res(ctx.json(testSharpHoundCollectors));
            }),
            rest.get('/api/v2/collectors/azurehound', (req, res, ctx) => {
                return res(ctx.json(testAzureHoundCollectors));
            }),
            rest.get('/api/v2/collectors/sharphound/v2.0.0', (req, res, ctx) => {
                return res(
                    ctx.set({
                        'Content-Type': 'application/octet-stream',
                        'Content-Disposition': `attachment; filename="${testFilename}`,
                    })
                );
            }),
            rest.get('/api/v2/collectors/sharphound/v2.0.0/checksum', (req, res, ctx) => {
                return res(
                    ctx.set({
                        'Content-Type': 'application/octet-stream',
                        'Content-Disposition': `attachment; filename="${testChecksumFilename}`,
                    })
                );
            })
        );

        render(<DownloadCollectors />);

        // wait for network calls to finish loading
        await waitFor(() => screen.getByRole('button', { name: 'Download SharpHound v2.0.0 (.zip)' }));

        // mocked fileDownload will be called with matching filenames when download button is clicked
        await user.click(screen.getByRole('button', { name: 'Download SharpHound v2.0.0 (.zip)' }));
        await waitFor(() => expect((fileDownload as Mock).mock.calls[0][1]).toBe(testFilename));

        await user.click(
            screen.getByRole('button', {
                name: (accessibleName, element) => {
                    return element.textContent === '4D9579899C56B69C78352B5F85A86';
                },
            })
        );
        await waitFor(() => expect((fileDownload as Mock).mock.calls[1][1]).toBe(testChecksumFilename));
    });

    it('displays a warning when there is a problem fetching the list of collectors', async () => {
        console.error = vi.fn();
        server.use(
            rest.get('/api/v2/collectors/sharphound', (req, res, ctx) => {
                return res(ctx.status(500));
            }),
            rest.get('/api/v2/collectors/azurehound', (req, res, ctx) => {
                return res(ctx.status(500));
            })
        );

        render(<DownloadCollectors />);

        // wait for network calls to finish loading
        await waitFor(() => screen.findByText('There are currently no versions of SharpHound available for download'));
        expect(
            screen.getByText('There are currently no versions of AzureHound available for download')
        ).toBeInTheDocument();
        expect(
            screen.getByText(
                'A browser extension (such as an ad blocker or other privacy extension) may prevent download links on this page from being displayed. Pause or disable your browser extensions and then refresh this page.'
            )
        ).toBeInTheDocument();
    });
});
