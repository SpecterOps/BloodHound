import { faCancel, faCheck } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { EntityField, format } from '../../utils';
import NodeIcon from '../NodeIcon';
import { isIconField } from './explore-table-utils';

const FALLBACK_STRING = '--';

const ExploreTableDataCell = ({ value, columnKey }: { value: EntityField['value']; columnKey: string }) => {
    if (columnKey === 'nodetype') {
        return (
            <div className='flex justify-center'>
                <NodeIcon nodeType={value?.toString() || ''} />
            </div>
        );
    }
    if (isIconField(value)) {
        return (
            <div className='flex justify-center items-center pb-1 pt-1'>
                <FontAwesomeIcon
                    icon={value ? faCheck : faCancel}
                    color={value ? 'green' : 'lightgray'}
                    className='scale-125'
                />
            </div>
        );
    }

    const stringyKey = columnKey?.toString();

    return format({ keyprop: stringyKey, value, label: stringyKey }) || FALLBACK_STRING;
};

export default ExploreTableDataCell;
