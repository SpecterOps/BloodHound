import { Checkbox } from '@bloodhoundenterprise/doodleui';
import { faThumbTack } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { UseComboboxPropGetters, useMultipleSelection } from 'downshift';
import { cn } from '../../../utils';
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
        className='p-2 m-0 w-full hover:bg-gray-100 dark:hover:bg-neutral-dark-4 cursor-pointer'
        {...itemProps}
        onClick={(e) => {
            e.stopPropagation();
            onClick(item);
        }}>
        <button className="w-full text-left flex justify-between items-center'">
            <div>
                <Checkbox className={cn('mr-2', { '*:bg-blue-800': isSelected })} checked={isSelected} />
                <span>{item.value}</span>
            </div>
            {item.isPinned && (
                <FontAwesomeIcon
                    stroke=''
                    className='justify-self-end stroke-cyan-300 dark:fill-white'
                    icon={faThumbTack}
                />
            )}
        </button>
    </li>
);

export default ManageColumnsListItem;
