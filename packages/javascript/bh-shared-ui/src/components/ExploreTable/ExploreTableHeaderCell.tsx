import { faCaretDown, faCaretUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Tooltip } from '@mui/material';
import { cn, formatPotentiallyUnknownLabel } from '../../utils';
import { MungedTableRowWithId } from './explore-table-utils';

const KEYS_TO_RENDER_AS_ICON = ['nodetype'];

const ExploreTableHeaderCell = ({
    headerKey,
    sortBy,
    sortOrder,
    onClick,
    dataType,
}: {
    headerKey: keyof MungedTableRowWithId;
    sortBy?: keyof MungedTableRowWithId;
    dataType: string;
    sortOrder?: string;
    onClick: () => void;
}) => {
    const label = formatPotentiallyUnknownLabel(String(headerKey));
    return (
        <Tooltip title={<p>{label}</p>}>
            <div
                className={cn(
                    'flex items-center m-0 cursor-pointer h-full w-full hover:bg-neutral-100 dark:hover:bg-neutral-dark-4',
                    {
                        'justify-center':
                            dataType === 'boolean' || KEYS_TO_RENDER_AS_ICON.includes(headerKey.toString()),
                    }
                )}
                onClick={onClick}>
                <div className='truncate'>{label}</div>
                <div className={cn('pl-2', { ['opacity-0']: sortBy !== headerKey })}>
                    {!sortOrder && <FontAwesomeIcon icon={faCaretDown} />}
                    {sortOrder === 'asc' && <FontAwesomeIcon icon={faCaretUp} />}
                    {sortOrder === 'desc' && <FontAwesomeIcon icon={faCaretDown} />}
                </div>
            </div>
        </Tooltip>
    );
};

export default ExploreTableHeaderCell;
