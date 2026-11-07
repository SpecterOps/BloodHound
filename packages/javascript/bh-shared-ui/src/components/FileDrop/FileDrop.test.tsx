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

import type { ComponentProps } from 'react';
import { fireEvent, render, screen } from '../../test-utils';
import FileDrop from './FileDrop';

describe('FileDrop', () => {
    const setup = (props: Partial<ComponentProps<typeof FileDrop>> = {}) => {
        const onDrop = vi.fn();
        render(<FileDrop onDrop={onDrop} disabled={false} {...props} />);

        return {
            onDrop,
            dropZone: screen.getByRole('button'),
            fileInput: screen.getByTestId('ingest-file-upload'),
            instructions: screen.getByText(/Click here or drag and drop/),
            icon: screen.getByText('inbox', { exact: true }),
        };
    };

    it('opens the file chooser when the text or icon is clicked', () => {
        const { dropZone, fileInput, instructions, icon } = setup();
        const clickSpy = vi.spyOn(fileInput, 'click').mockImplementation(() => undefined);

        fireEvent.click(instructions);
        fireEvent.click(icon);

        expect(clickSpy).toHaveBeenCalledTimes(2);
        expect(dropZone).toHaveAttribute('aria-label', 'Choose JSON or zip/compressed JSON files to upload');
    });

    it('keeps the drop zone as the hit target for decorative children', () => {
        const { dropZone, instructions, icon } = setup();

        expect(dropZone).toContainElement(instructions);
        expect(instructions).toHaveClass('pointer-events-none');
        expect(icon).toBeInTheDocument();
    });

    it('shows hover and drag-active states', () => {
        const { dropZone } = setup();

        fireEvent.mouseEnter(dropZone);
        expect(dropZone).toHaveClass('bg-neutral-3');

        fireEvent.mouseLeave(dropZone);
        expect(dropZone).not.toHaveClass('bg-neutral-3');

        fireEvent.dragEnter(dropZone);
        expect(dropZone).toHaveClass('bg-neutral-3');

        fireEvent.dragLeave(dropZone);
        expect(dropZone).not.toHaveClass('bg-neutral-3');
    });

    it('forwards selected and dropped files', () => {
        const { dropZone, fileInput, onDrop } = setup();
        const file = new File(['contents'], 'test.json', { type: 'application/json' });

        fireEvent.change(fileInput, { target: { files: [file] } });
        expect(onDrop).toHaveBeenNthCalledWith(1, [file]);

        fireEvent.drop(dropZone, { dataTransfer: { files: [file] } });
        expect(onDrop).toHaveBeenNthCalledWith(2, [file]);
    });

    it('supports a disabled drop zone', () => {
        const { dropZone, fileInput } = setup({ disabled: true });

        expect(dropZone).toBeDisabled();
        expect(fileInput).toBeDisabled();
    });

    it('uses single-file copy and labeling when multiple is disabled', () => {
        const { dropZone, instructions } = setup({ multiple: false });

        expect(dropZone).toHaveAccessibleName('Choose a JSON file to upload');
        expect(instructions).toHaveTextContent('Click here or drag and drop to upload a JSON file');
    });

    it('renders the default multiple-file copy', () => {
        render(<FileDrop onDrop={vi.fn()} disabled={false} />);

        expect(
            screen.getByText('Click here or drag and drop to upload JSON or zip/compressed JSON files')
        ).toBeVisible();
    });
});
