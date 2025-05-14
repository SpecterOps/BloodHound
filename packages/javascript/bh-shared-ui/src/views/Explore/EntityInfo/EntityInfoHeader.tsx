import { faAngleDoubleUp, faMinus, faPlus } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Typography } from '@mui/material';
import React from 'react';
import Icon from '../../../components/Icon';
import NodeIcon from '../../../components/NodeIcon/NodeIcon';
import { useExploreParams } from '../../../hooks';
import { EntityKinds } from '../../../utils';
import { useHeaderStyles } from '../InfoStyles';
import { useObjectInfoPanelContext } from '../providers/ObjectInfoPanelProvider';

export interface HeaderProps {
    expanded: boolean;
    name: string;
    onToggleExpanded: (expanded: boolean) => void;
    nodeType?: EntityKinds;
}

const Header: React.FC<HeaderProps> = ({ name, nodeType, onToggleExpanded, expanded }) => {
    const styles = useHeaderStyles();
    const { setIsObjectInfoPanelOpen } = useObjectInfoPanelContext();
    const { setExploreParams, expandedPanelSections } = useExploreParams();

    const handleCollapseAll = () => {
        setIsObjectInfoPanelOpen(false);
        if (expandedPanelSections?.length) {
            setExploreParams({
                expandedPanelSections: [],
            });
        }
    };

    return (
        <Box className={styles.header}>
            <Icon
                className={styles.icon}
                click={() => {
                    onToggleExpanded(!expanded);
                }}>
                <FontAwesomeIcon icon={expanded ? faMinus : faPlus} />
            </Icon>

            {nodeType && <NodeIcon nodeType={nodeType} />}

            <Typography
                data-testid='explore_entity-information-panel_header-text'
                variant='h6'
                noWrap
                className={styles.headerText}>
                {name}
            </Typography>

            <Icon
                tip='Collapse All'
                click={handleCollapseAll}
                className={styles.icon}
                data-testid='explore_entity-information-panel_button-collapse-all'>
                <FontAwesomeIcon icon={faAngleDoubleUp} />
            </Icon>
        </Box>
    );
};

export default Header;
