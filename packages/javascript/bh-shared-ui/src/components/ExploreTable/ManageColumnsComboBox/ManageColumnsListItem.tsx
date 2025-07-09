import { Checkbox } from '@bloodhoundenterprise/doodleui';
import { faThumbTack } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { UseComboboxPropGetters, useMultipleSelection } from 'downshift';
import { ManageColumnsComboBoxOption } from './ManageColumnsComboBox';

type ManageColumnsListItemProps = {
    isSelected?: boolean;
    item: ManageColumnsComboBoxOption;
    onClick:
        | ReturnType<typeof useMultipleSelection<ManageColumnsComboBoxOption>>['removeSelectedItem']
        | ReturnType<typeof useMultipleSelection<ManageColumnsComboBoxOption>>['addSelectedItem'];
    itemProps: ReturnType<UseComboboxPropGetters<ManageColumnsComboBoxOption>['getItemProps']>;
};

const ManageColumnsListItem = ({ isSelected, item, onClick, itemProps }: ManageColumnsListItemProps) => (
    <li
        className={`p-2 w-full hover:bg-gray-100 ${isSelected ? 'cursor-default' : 'cursor-pointer'} ${item.isPinned ? 'bg-gray-100' : ''}`}
        {...itemProps}
        aria-disabled={item?.isPinned}
        onClick={(e) => {
            e.stopPropagation();
            onClick(item);
        }}>
        <button
            className={`w-full text-left flex justify-between items-center ${isSelected ? 'cursor-default' : 'cursor-pointer'}`}>
            <div>
                <Checkbox className={`mr-2 ${isSelected ? '*:bg-blue-800' : ''}`} checked={isSelected} />
                <span>{item.value}</span>
            </div>
            {item.isPinned && (
                <FontAwesomeIcon
                    fill='white'
                    stroke=''
                    className='justify-self-end stroke-cyan-300'
                    icon={faThumbTack}
                />
            )}
        </button>
    </li>
);

export default ManageColumnsListItem;
