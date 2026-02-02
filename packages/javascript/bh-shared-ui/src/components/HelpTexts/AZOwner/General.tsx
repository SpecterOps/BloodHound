import {FC} from "react";
import {Link, Typography} from "@mui/material";

const AZResourceGroupLink = (
    <Link
        target='_blank'
        rel='noopener noreferrer'
        href='https://bloodhound.specterops.io/resources/nodes/az-resource-group'>
        AZResourceGroup
    </Link>
);

const AZSubscriptionLink = (
    <Link
        target='_blank'
        rel='noopener noreferrer'
        href='https://bloodhound.specterops.io/resources/nodes/az-subscription'>
        AZSubscription
    </Link>
);

const AZVMLink = (
    <Link
        target='_blank'
        rel='noopener noreferrer'
        href='https://bloodhound.specterops.io/resources/nodes/az-vm'>
        AZVM
    </Link>
)

const General: FC = () => {
    return (
        <Typography variant='body2'>
            The principal is granted the Owner role on the resource.
            <br/><br/>
            AZOwner targets resources in AzureRM (for example {AZResourceGroupLink}, {AZSubscriptionLink} and {AZVMLink}) through role assignment called “Owner”.
        </Typography>
    )
};

export default General;
