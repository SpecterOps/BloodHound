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

import { render, screen } from '../../test-utils';
import HighlightedText from '.';

describe('HighlightedText', () => {
    it('should render when search string is at the beginning of the text', () => {
        render(<HighlightedText text='test string' search='test' />);
        expect(screen.getByText(/test/)).toBeInTheDocument();
        expect(screen.getByText(/test/)).toHaveAttribute('style', 'font-weight: bold;');
        expect(screen.getByText(/string/)).toBeInTheDocument();
    });

    it('should render when search string is in the middle of the text', () => {
        render(<HighlightedText text='test string' search='str' />);
        expect(screen.getByText(/test/)).toBeInTheDocument();
        expect(screen.getByText(/str/)).toBeInTheDocument();
        expect(screen.getByText(/str/)).toHaveAttribute('style', 'font-weight: bold;');
        expect(screen.getByText(/ing/)).toBeInTheDocument();
    });

    it('should render when search string at the end of the text', () => {
        render(<HighlightedText text='test string' search='string' />);
        expect(screen.getByText(/test/)).toBeInTheDocument();
        expect(screen.getByText(/string/)).toBeInTheDocument();
        expect(screen.getByText(/string/)).toHaveAttribute('style', 'font-weight: bold;');
    });

    it('should render when search string is not found in the text', () => {
        render(<HighlightedText text='test string' search='zzz' />);
        expect(screen.getByText(/test string/)).toBeInTheDocument();
    });

    it('should render when search string matches text entirely', () => {
        render(<HighlightedText text='test string' search='test string' />);
        expect(screen.getByText(/test string/)).toBeInTheDocument();
        expect(screen.getByText(/test string/)).toHaveAttribute('style', 'font-weight: bold;');
    });

    it('should handle special characters', () => {
        render(<HighlightedText text='(TESTLAB.LOCAL)+SOME@TEXT$![-[\]{}()*+?' search='CAL)+SOME@TEXT$!' />);
        expect(screen.getByText(/\(TESTLAB\.LO/)).toBeInTheDocument();
        expect(screen.getByText(/CAL\)\+SOME@TEXT\$!/)).toBeInTheDocument();
        expect(screen.getByText(/\[-\[\\\]\{\}\(\)\*\+\?/)).toBeInTheDocument();
        expect(screen.getByText(/CAL\)\+SOME@TEXT\$!/)).toHaveAttribute('style', 'font-weight: bold;');
    });
});
