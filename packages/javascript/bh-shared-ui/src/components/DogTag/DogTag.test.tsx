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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen, waitForElementToBeRemoved } from '../../test-utils';
import DogTag from './DogTag';

const server = setupServer(
    rest.get('/api/v2/dog-tags', async (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    'example.feature.bool.value': true,
                    'example.feature.int.value': 67,
                    'example.feature.string.value': 'test',
                },
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('DogTag', () => {
    it('renders enabled prop when DogTag bool expected value equals the DogTag value', async () => {
        render(
            <DogTag
                dogTagKey='example.feature.bool.value'
                expectedValue={true}
                disabled={<div>This feature is disabled!</div>}
                enabled={<div>This feature is enabled!</div>}
            />
        );
        expect(await screen.findByText('This feature is enabled!')).toBeInTheDocument();
    });

    it('renders disabled prop when DogTag bool expected value does not equals the DogTag value', async () => {
        render(
            <DogTag
                dogTagKey='example.feature.bool.value'
                expectedValue={false}
                disabled={<div>This feature is disabled!</div>}
                enabled={<div>This feature is enabled!</div>}
            />
        );
        expect(await screen.findByText('This feature is disabled!')).toBeInTheDocument();
    });

    it('renders enabled prop when DogTag int expected value equals the DogTag value', async () => {
        render(
            <DogTag
                dogTagKey='example.feature.int.value'
                expectedValue={67}
                disabled={<div>This feature is disabled!</div>}
                enabled={<div>This feature is enabled!</div>}
            />
        );
        expect(await screen.findByText('This feature is enabled!')).toBeInTheDocument();
    });

    it('renders disabled prop when DogTag int expected value does not equal the DogTag value', async () => {
        render(
            <DogTag
                dogTagKey='example.feature.int.value'
                expectedValue={0}
                disabled={<div>This feature is disabled!</div>}
                enabled={<div>This feature is enabled!</div>}
            />
        );
        expect(await screen.findByText('This feature is disabled!')).toBeInTheDocument();
    });

    it('renders enabled prop when DogTag string expected value equals the DogTag value', async () => {
        render(
            <DogTag
                dogTagKey='example.feature.string.value'
                expectedValue={'test'}
                disabled={<div>This feature is disabled!</div>}
                enabled={<div>This feature is enabled!</div>}
            />
        );
        expect(await screen.findByText('This feature is enabled!')).toBeInTheDocument();
    });

    it('renders disabled prop when DogTag string expected value does not equal the DogTag value', async () => {
        render(
            <DogTag
                dogTagKey='example.feature.string.value'
                expectedValue={'wrong-answer'}
                disabled={<div>This feature is disabled!</div>}
                enabled={<div>This feature is enabled!</div>}
            />
        );
        expect(await screen.findByText('This feature is disabled!')).toBeInTheDocument();
    });

    it('renders nothing if DogTag is not equal to expected value', async () => {
        render(<DogTag dogTagKey='example.feature.bool.value' expectedValue={false} />);
        await waitForElementToBeRemoved(() => screen.queryByText('Loading...'));
        expect(document.querySelector('div')?.innerHTML).toEqual('');
    });

    it('renders nothing if DogTag is not equal to expected value and disabled prop is not provided', async () => {
        render(<DogTag dogTagKey='example.feature.bool.value' expectedValue={false} />);
        await waitForElementToBeRemoved(() => screen.queryByText('Loading...'));
        expect(document.querySelector('div')?.innerHTML).toEqual('');
    });

    it('renders an error fallback when DogTag is not found', async () => {
        console.error = vi.fn();
        render(<DogTag dogTagKey='not-found' errorFallback={<span>Error</span>} />);
        await waitForElementToBeRemoved(() => screen.queryByText('Loading...'));
        expect(screen.getByText('Error')).toBeInTheDocument();
        expect(console.error).toHaveBeenCalledWith('DogTag "not-found" not found');
    });
});
