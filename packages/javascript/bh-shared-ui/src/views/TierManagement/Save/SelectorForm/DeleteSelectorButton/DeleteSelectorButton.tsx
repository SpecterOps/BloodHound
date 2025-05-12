import { Button } from '@bloodhoundenterprise/doodleui';
import { faTrashCan } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { AssetGroupTagSelector } from 'js-client-library';
import { FC } from 'react';

const DeleteSelectorButton: FC<{
    selectorId: string;
    selectorData: AssetGroupTagSelector | undefined;
    onClick: () => void;
}> = ({ selectorId, selectorData, onClick }) => {
    if (selectorId === '') return null;

    if (selectorData === undefined) return null;

    if (selectorData.is_default) return null;

    return (
        <Button variant={'text'} onClick={onClick}>
            <span>
                <FontAwesomeIcon icon={faTrashCan} /> Delete Selector
            </span>
        </Button>
    );
};

export default DeleteSelectorButton;
