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
import { IconName, faClose } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import clsx from 'clsx';
import React, { FC, useState } from 'react';
import { FixedSizeList, ListChildComponentProps } from 'react-window';
import { AppIcon } from '../../../../components';
import { freeIconsList } from '../../../../utils';

const InnerElement = ({ style, ...rest }: any) => (
    <ul style={{ ...style, overflowX: 'hidden', marginTop: 0, overflowY: 'auto' }} {...rest}></ul>
);

const IconCard: FC<{ iconName: IconName; onClick: (iconName: IconName) => void }> = ({ iconName, onClick }) => {
    return (
        <Button
            variant={'text'}
            className={clsx(['relative', !iconName && 'invisible'])}
            onClick={() => {
                onClick(iconName);
            }}>
            <Card className='flex items-center justify-center h-24 w-24'>
                <CardContent className='first:pt-0'>
                    {iconName && <FontAwesomeIcon icon={iconName} size='2xl' />}
                </CardContent>
            </Card>
            <p className='absolute -bottom-16'>{iconName}</p>
        </Button>
    );
};

const Row = ({
    data,
    index: row,
    style,
}: ListChildComponentProps<{ filteredList: IconName[]; onClick: (iconName: IconName) => void }>) => {
    const { filteredList, onClick } = data;

    return (
        <li className={'flex justify-evenly items-center'} style={{ ...style }}>
            {Array.from({ length: 5 }, (_, index) => {
                return <IconCard key={row * 5 + index} iconName={filteredList[row * 5 + index]} onClick={onClick} />;
            })}
        </li>
    );
};

export const VirtualizedIconList = ({
    filteredList,
    onClick,
}: {
    filteredList: IconName[];
    onClick: (iconName: IconName) => void;
}) => {
    return (
        <FixedSizeList
            height={64 * 9}
            itemCount={Math.ceil(filteredList.length / 5)}
            itemData={{ filteredList, onClick }}
            itemSize={128 + 64}
            innerElementType={InnerElement}
            width={'100%'}
            className='rounded-md bg-neutral-3'
            initialScrollOffset={0}>
            {Row}
        </FixedSizeList>
    );
};

const GlyphSelectDialog: React.FC<{
    open: boolean;
    selected: IconName | undefined;
    onCancel: () => void;
    onSelect: (selectedGlyph?: IconName) => void;
    tagName?: string;
}> = ({ open, onCancel, onSelect, selected }) => {
    const [query, setQuery] = useState('');
    const [selectedIcon, setSelectedIcon] = useState<IconName | undefined>(selected);

    const handleConfirm = () => onSelect(selectedIcon);
    const handleClear = () => setSelectedIcon(undefined);
    const handleCancel = () => {
        setSelectedIcon(selected);
        onCancel();
    };

    return (
        <Dialog open={open} data-testid='confirmation-dialog'>
            <DialogPortal>
                <DialogContent className='z-[1400]' maxWidth='lg'>
                    <DialogTitle className='text-lg'>Select a Glyph</DialogTitle>
                    <DialogDescription className='text-lg'>
                        The selected glyph will apply to all nodes tagged in this Zone for displaying in the Explore
                        graph.
                    </DialogDescription>
                    <div className='flex flex-col gap-6 justify-between items-center'>
                        <div className='flex items-center w-full justify-between px-4'>
                            <div className='flex items-end gap-6 h-20'>
                                <div className='flex flex-col items-start justify-end w-44'>
                                    <span className='font-bold'>Current Selection:</span>
                                    <p>{selectedIcon || 'None Selected'}</p>
                                </div>
                                {selectedIcon && (
                                    <Button variant={'text'} onClick={handleClear}>
                                        <Card className='flex items-center justify-center size-16 relative dark:bg-neutral-4'>
                                            <FontAwesomeIcon icon={faClose} className='absolute top-1 right-1' />
                                            <CardContent className='first:pt-0 p-0'>
                                                <FontAwesomeIcon icon={selectedIcon} size='2xl' />
                                            </CardContent>
                                        </Card>
                                    </Button>
                                )}
                            </div>

                            <span className='flex items-center w-64 self-end'>
                                <AppIcon.MagnifyingGlass className='-mr-4' />
                                <Input
                                    placeholder='Search'
                                    onChange={(e) => setQuery(e.target.value)}
                                    className='pl-8'
                                />
                            </span>
                        </div>

                        <div className='p-2 size-full bg-neutral-3 rounded-md'>
                            <VirtualizedIconList
                                filteredList={freeIconsList.filter((iconName) => iconName.includes(query))}
                                onClick={(iconName) => {
                                    setSelectedIcon(iconName);
                                }}
                            />
                        </div>
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
