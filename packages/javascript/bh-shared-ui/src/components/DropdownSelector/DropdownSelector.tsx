import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { Box, Button, MenuItem, Popover, Tooltip, Typography } from "@mui/material";
import { FC, useState } from "react";
import { DropdownOption } from "./types";

const DropdownSelector: FC<{
    options: DropdownOption[],
    selectedText: string,
    fullWidth?: boolean,
    onChange: (selection: DropdownOption) => void,
}> = ({ options, selectedText, fullWidth, onChange }) => {

    const [anchorEl, setAnchorEl] = useState(null);
    const open = Boolean(anchorEl);
    
    const handleClick = (e: any) => {
        setAnchorEl(e.currentTarget);
    }

    const handleClose = () => {
        setAnchorEl(null);
    }

    return (
        <Box p={1}>
            <Button
                sx={{
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    display: 'block',
                }}
                fullWidth={fullWidth}
                variant='contained'
                disableElevation
                color='primary'
                onClick={handleClick}
                data-testid='data-quality_context-selector'>
                {selectedText}
            </Button>
            <Popover
                open={open}
                anchorEl={anchorEl}
                onClose={handleClose}
                anchorOrigin={{
                    vertical: 'bottom',
                    horizontal: 'center',
                }}
                transformOrigin={{
                    vertical: 'top',
                    horizontal: 'center',
                }}
                data-testid='data-quality_context-selector-popover'>
                {options.map(option => {
                    return (
                        <MenuItem
                            style={{
                                display: 'flex',
                                justifyContent: 'space-between',
                                width: 450,
                                maxWidth: 450,
                            }}
                            key={option.key}
                            onClick={() => {
                                onChange(option);
                                handleClose();
                            }}>
                            <Tooltip title={option.value}>
                                <Typography
                                    style={{
                                        overflow: 'hidden',
                                        textTransform: 'uppercase',
                                        display: 'inline-block',
                                        textOverflow: 'ellipsis',
                                        maxWidth: 350,
                                    }}>
                                    {option.value}
                                </Typography>
                            </Tooltip>
                            {option.icon &&
                                <FontAwesomeIcon
                                    style={{ width: '10%', alignSelf: 'center' }}
                                    icon={option.icon}
                                    size='sm'
                                />
                            }
                        </MenuItem>
                    )})}
                </Popover>
            </Box>
    )
}

export default DropdownSelector;