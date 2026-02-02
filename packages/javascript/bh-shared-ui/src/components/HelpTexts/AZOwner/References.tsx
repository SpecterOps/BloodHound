import {FC} from "react";
import {Box, Link} from "@mui/material";

const References: FC = () => {
    return (
        <Box className='overflow-x-auto'>
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://blog.netspi.com/attacking-azure-with-custom-script-extensions/'>
                https://blog.netspi.com/attacking-azure-with-custom-script-extensions/
            </Link>
            <br/>
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://docs.microsoft.com/en-us/azure/role-based-access-control/built-in-roles#owner'>
                https://docs.microsoft.com/en-us/azure/role-based-access-control/built-in-roles#owner
            </Link>
        </Box>
    )
};

export default References;
