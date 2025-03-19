import { Popover, PopoverContent, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { useState } from 'react';
import { AppIcon } from '../../../../components/AppIcon';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';

const selectors = ['Selector A', 'Selector B', 'Selector C'];

const SelectorList = () => {
    const [menuOpen, setMenuOpen] = useState<{ [key: number]: boolean }>({});

    const handleMenuClick = (index: number) => {
        setMenuOpen((prev) => ({
            ...prev,
            [index]: !prev[index], //Toggle only the clicked popover
        }));
    };

    const handleViewClick = () => {};
    const handleEditClick = () => {};
    const handleDeleteClick = () => {};

    return (
        <EntityInfoCollapsibleSection label='Selectors' count={selectors.length}>
            {selectors.map((selector, index) => {
                return (
                    <div className={`flex items-center gap-2 p-2 ${index % 2 === 0 ? 'bg-[#E3E7EA]' : ''}`} key={index}>
                        <Popover open={!!menuOpen[index]}>
                            <PopoverTrigger asChild>
                                <button onClick={() => handleMenuClick(index)}>
                                    <AppIcon.VerticalEllipsis />
                                </button>
                            </PopoverTrigger>
                            <PopoverContent
                                className='w-80 px-4 py-2 flex flex-col gap-2'
                                onInteractOutside={() => setMenuOpen({})}
                                onEscapeKeyDown={() => setMenuOpen({})}>
                                <div className='cursor-pointer p-2 hover:bg-[#E3E7EA]' onClick={handleViewClick}>
                                    View
                                </div>
                                <div className='cursor-pointer p-2 hover:bg-[#E3E7EA]' onClick={handleEditClick}>
                                    Edit
                                </div>
                                <div className='cursor-pointer p-2 hover:bg-[#E3E7EA]' onClick={handleDeleteClick}>
                                    Delete
                                </div>
                            </PopoverContent>
                        </Popover>
                        {selector}
                    </div>
                );
            })}
        </EntityInfoCollapsibleSection>
    );
};

export default SelectorList;
