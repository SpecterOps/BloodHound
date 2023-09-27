import { Box, Fade, ListItem, Tooltip, Typography } from "@mui/material";
import { FC, HTMLAttributes } from "react";
import NodeIcon from "../NodeIcon";

const AutocompleteOption: FC<{
    props: HTMLAttributes<HTMLLIElement>,
    id: string,
    name?: string,
    type: string,
    actionLabel?: string,
}> = ({ props, id, name, type, actionLabel }) => {
    return (
        <ListItem
            {...props}
            key={id}
            role='option'
            style={{
                display: 'block',
                maxWidth: '100%',
            }}>
            <Typography variant='body2'> {actionLabel}</Typography>
            <Box style={{ display: 'flex', justifyContent: 'flex-start' }}>
                <NodeIcon nodeType={type}></NodeIcon>
                <Tooltip
                    title={name || id}
                    placement='top-start'
                    TransitionComponent={Fade}>
                    <Typography
                        variant='body1'
                        style={{
                            display: 'block',
                            maxWidth: '100%',
                            whiteSpace: 'nowrap',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                        }}>
                        {name || id}
                    </Typography>
                </Tooltip>
            </Box>
        </ListItem>
    )
}

export default AutocompleteOption;