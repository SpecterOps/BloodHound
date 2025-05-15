import { Alert, Skeleton } from '@mui/material';
import React, { useEffect } from 'react';
import { useExploreParams, useFetchEntityProperties, usePreviousValue } from '../../../hooks';
import { SearchValue } from '../../../store';
import { EntityField, formatObjectInfoFields } from '../../../utils';
import { BasicObjectInfoFields } from '../BasicObjectInfoFields';
import { FieldsContainer, ObjectInfoFields } from '../fragments';
import { useObjectInfoPanelContext } from '../providers/ObjectInfoPanelProvider';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';
import { EntityInfoContentProps } from './EntityInfoContent';

const EntityObjectInformation: React.FC<EntityInfoContentProps> = ({ id, nodeType, databaseId }) => {
    const { setExploreParams } = useExploreParams();
    const { isObjectInfoPanelOpen, setIsObjectInfoPanelOpen } = useObjectInfoPanelContext();
    const { entityProperties, informationAvailable, isLoading, isError } = useFetchEntityProperties({
        objectId: id,
        nodeType,
        databaseId,
    });

    const previousId = usePreviousValue(id);

    useEffect(() => {
        if (previousId !== id) {
            setIsObjectInfoPanelOpen(true);
        }
    }, [previousId, id, setIsObjectInfoPanelOpen]);

    const sectionLabel = 'Object Information';

    const handleOnChange = () => {
        setIsObjectInfoPanelOpen(!isObjectInfoPanelOpen);
    };

    if (isLoading) return <Skeleton data-testid='entity-object-information-skeleton' variant='text' />;

    if (isError || !informationAvailable)
        return (
            <EntityInfoCollapsibleSection
                onChange={handleOnChange}
                isExpanded={isObjectInfoPanelOpen}
                label={sectionLabel}>
                <FieldsContainer>
                    <Alert severity='error'>Unable to load object information for this node.</Alert>
                </FieldsContainer>
            </EntityInfoCollapsibleSection>
        );

    const formattedObjectFields: EntityField[] = formatObjectInfoFields(entityProperties);

    const handleSourceNodeSelected = (sourceNode: SearchValue) => {
        setExploreParams({ primarySearch: sourceNode.objectid, searchType: 'node' });
    };

    return (
        <EntityInfoCollapsibleSection onChange={handleOnChange} isExpanded={isObjectInfoPanelOpen} label={sectionLabel}>
            <FieldsContainer>
                <BasicObjectInfoFields
                    nodeType={nodeType}
                    handleSourceNodeSelected={handleSourceNodeSelected}
                    {...entityProperties}
                />
                <ObjectInfoFields fields={formattedObjectFields} />
            </FieldsContainer>
        </EntityInfoCollapsibleSection>
    );
};

export default EntityObjectInformation;
