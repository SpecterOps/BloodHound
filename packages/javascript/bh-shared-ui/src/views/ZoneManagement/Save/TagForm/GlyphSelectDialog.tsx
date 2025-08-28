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

import {
    Button,
    Card,
    CardContent,
    Dialog,
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
    Input,
} from '@bloodhoundenterprise/doodleui';
import { IconName } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import React, { useState } from 'react';
import { FixedSizeList, ListChildComponentProps } from 'react-window';
import { AppIcon } from '../../../../components';
import { useZonePathParams } from '../../../../hooks';
import { freeIconsList } from '../../../../utils';

const InnerElement = ({ style, ...rest }: any) => (
    <ul style={{ ...style, overflowX: 'hidden', marginTop: 0, overflowY: 'auto' }} {...rest}></ul>
);

const Row = ({ data: filteredList, index, style }: ListChildComponentProps<IconName[]>) => {
    filteredList;

    return (
        <li className={'flex justify-evenly items-center'} style={{ ...style }}>
            <Button variant={'text'} className='relative'>
                <Card className='flex items-center justify-center h-24 w-24'>
                    <CardContent className='first:pt-0'>
                        <FontAwesomeIcon icon={filteredList[index * 5]} size='2xl' />
                    </CardContent>
                </Card>
                <p className='absolute -bottom-16'>{filteredList[index * 5]}</p>
            </Button>
            <Button variant={'text'} className='relative'>
                <Card className='flex items-center justify-center h-24 w-24'>
                    <CardContent className='first:pt-0'>
                        <FontAwesomeIcon icon={filteredList[index * 5 + 1]} size='2xl' />
                    </CardContent>
                </Card>
                <p className='absolute -bottom-16'>{filteredList[index * 5 + 1]}</p>
            </Button>
            <Button variant={'text'} className='relative'>
                <Card className='flex items-center justify-center h-24 w-24'>
                    <CardContent className='first:pt-0'>
                        <FontAwesomeIcon icon={filteredList[index * 5 + 2]} size='2xl' />
                    </CardContent>
                </Card>
                <p className='absolute -bottom-16'>{filteredList[index * 5 + 2]}</p>
            </Button>
            <Button variant={'text'} className='relative'>
                <Card className='flex items-center justify-center h-24 w-24'>
                    <CardContent className='first:pt-0'>
                        <FontAwesomeIcon icon={filteredList[index * 5 + 3]} size='2xl' />
                    </CardContent>
                </Card>
                <p className='absolute -bottom-16'>{filteredList[index * 5 + 3]}</p>
            </Button>
            <Button variant={'text'} className='relative'>
                <Card className='flex items-center justify-center h-24 w-24'>
                    <CardContent className='first:pt-0'>
                        <FontAwesomeIcon icon={filteredList[index * 5 + 4]} size='2xl' />
                    </CardContent>
                </Card>
                <p className='absolute -bottom-16'>{filteredList[index * 5 + 4]}</p>
            </Button>
        </li>
    );
};

export const VirtualizedIconList = ({ filteredList }: { filteredList: IconName[] }) => {
    return (
        <FixedSizeList
            height={64 * 10}
            itemCount={Math.ceil(filteredList.length / 5)}
            itemData={filteredList}
            itemSize={128 + 64}
            innerElementType={InnerElement}
            width={'100%'}
            initialScrollOffset={0}
            style={{ borderRadius: 4 }}>
            {Row}
        </FixedSizeList>
    );
};

const GlyphSelectDialog: React.FC<{
    open: boolean;
    handleCancel: () => void;
    handleSelect: (selectedGlyph?: IconName) => void;
    tagName?: string;
    selected?: IconName;
}> = ({ open, handleCancel, handleSelect, selected }) => {
    const { tagKindDisplay } = useZonePathParams();
    const [query, setQuery] = useState('');
    const [selectedIcon, setSelectedIcon] = useState<IconName | undefined>(selected);

    const handleConfirm = () => {
        handleSelect(selectedIcon);
    };

    const handleClear = () => setSelectedIcon(undefined);
    handleClear;

    return (
        <Dialog open={open} data-testid='confirmation-dialog'>
            <DialogPortal>
                <DialogContent className='z-[1400]' maxWidth='lg'>
                    <DialogTitle className='text-lg'>Select a Glyph</DialogTitle>
                    <DialogDescription className='text-lg'>
                        The selected glyph will apply to all nodes with tagged in this {tagKindDisplay} for displaying
                        in the Explore graph.
                    </DialogDescription>
                    <div className='flex flex-col gap-6 justify-between items-center'>
                        <span className='flex items-center w-64 self-start'>
                            <AppIcon.MagnifyingGlass className='-mr-4' />
                            <Input placeholder='Search' onChange={(e) => setQuery(e.target.value)} className='pl-8' />
                        </span>
                        <VirtualizedIconList
                            filteredList={freeIconsList.filter((iconName) => iconName.includes(query))}
                        />
                    </div>
                    <DialogActions>
                        <Button variant='tertiary' onClick={handleCancel} data-testid='confirmation-dialog_button-no'>
                            Cancel
                        </Button>
                        <Button onClick={handleConfirm} data-testid='confirmation-dialog_button-yes'>
                            Confirm
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default GlyphSelectDialog;
