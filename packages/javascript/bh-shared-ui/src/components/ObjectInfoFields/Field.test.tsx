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
import { Field } from './Field';

describe.each(['spread', 'inline'] as const)('Field (%s layout)', (layout) => {
    it.each([
        { description: 'single-item', values: ['HOST/one'] },
        { description: 'multiple-item', values: ['HOST/one', 'HOST/two'] },
    ])('aligns $description lists for the selected layout with one label', ({ values }) => {
        render(<Field layout={layout} label='Service Principal Names:' value={values} />);

        expect(screen.getAllByText(/^Service Principal Names:?$/)).toHaveLength(1);
        values.forEach((value) => {
            const element = screen.getByText(value);
            expect(element).toHaveAttribute('title', value);
            if (layout === 'spread') {
                expect(element.parentElement).toHaveClass('justify-end');
            } else {
                expect(element.parentElement).not.toHaveClass('justify-end');
            }
        });
    });

    it('preserves rich labels', () => {
        render(<Field layout={layout} label={<span>Rich label</span>} value='value' />);
        expect(screen.getByText('Rich label')).toBeInTheDocument();
        expect(screen.getByText('value')).toBeInTheDocument();
    });

    it('should render a Field when the provided value is false', () => {
        render(<Field layout={layout} label='Test Field (Boolean)' value={false} keyprop='test-field-boolean' />);
        expect(screen.getByText('FALSE')).toBeInTheDocument();
    });

    it('should render a Field when the provided value is 0', () => {
        render(<Field layout={layout} label='Test Field (Number)' value={0} keyprop='test-field-number' />);
        expect(screen.getByText('0')).toBeInTheDocument();
    });

    it('should render a Field when the provided value is -0', () => {
        render(<Field layout={layout} label='Test Field (Number)' value={-0} keyprop='test-field-number' />);
        expect(screen.getByText('0')).toBeInTheDocument();
        expect(screen.queryByText('-0')).not.toBeInTheDocument();
    });

    it('should render an accessible placeholder when the provided value is an empty string', () => {
        render(<Field layout={layout} label='Test Field (String)' value={''} keyprop='test-field-string' />);
        expect(screen.getByText('Test Field (String)')).toBeInTheDocument();
        expect(screen.getByText('—')).toHaveAttribute('aria-hidden', 'true');
        expect(screen.getByText('Empty string')).toBeInTheDocument();
    });

    it('should render an accessible placeholder when the provided value is an empty array', () => {
        render(<Field layout={layout} label='Test Field (Array)' value={[]} keyprop='test-field-array' />);
        expect(screen.getByText('Test Field (Array)')).toBeInTheDocument();
        expect(screen.getByText('NONE')).toHaveAttribute('aria-hidden', 'true');
        expect(screen.getByText('Empty array, zero values')).toBeInTheDocument();
    });

    it('should render accessible placeholders for empty values in a list', () => {
        render(
            <Field layout={layout} label='Test Field (Array)' value={['value', '', null]} keyprop='test-field-array' />
        );
        expect(screen.getByText('value')).not.toHaveAttribute('aria-hidden');
        screen.getAllByText('—').forEach((placeholder) => {
            expect(placeholder).toHaveAttribute('aria-hidden', 'true');
        });
        expect(screen.getByText('Empty string')).toBeInTheDocument();
        expect(screen.getByText('Null value')).toBeInTheDocument();
    });

    it('should render an accessible placeholder when the provided value is null', () => {
        render(<Field layout={layout} label='Test Field (Null)' value={null} keyprop='test-field-null' />);
        expect(screen.getByText('Test Field (Null)')).toBeInTheDocument();
        expect(screen.getByText('—')).toHaveAttribute('aria-hidden', 'true');
        expect(screen.getByText('Null value')).toBeInTheDocument();
    });

    it('should render a real value matching an empty state placeholder as ordinary text', () => {
        render(<Field layout={layout} label='Test Field (String)' value='NONE' keyprop='test-field-string' />);
        expect(screen.getByText('NONE')).not.toHaveAttribute('aria-hidden');
        expect(screen.queryByText('Empty array, zero values')).not.toBeInTheDocument();
    });
});
