import { FC } from 'react';
import { Typography } from '@mui/material';

const Abuse: FC = () => {
    return (
        <Typography variant='body2'>
            An attacker who is a member of "Authenticated Users" triggers a traditional SMB based coercion from the target computer to their attacker host.
            The attacker relays this authentication attempt to the target system the inbound account has admin access to.
        </Typography>
    );
};

export default Abuse;
