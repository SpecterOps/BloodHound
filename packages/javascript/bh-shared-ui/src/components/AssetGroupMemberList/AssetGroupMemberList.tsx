import { Box } from "@mui/material";
import { FC } from "react"
import NodeIcon from "../NodeIcon";

const AssetGroupMemberList: FC<{
    assetGroupMembers: any[],
    onSelectMember: (member: any) => void,
}> = ({ assetGroupMembers, onSelectMember }) => {
    return (
        <Box>
            {assetGroupMembers?.map(member => {
                return (
                    <Box
                        onClick={() => onSelectMember(member )}
                        padding={1}
                        border={1}
                        key={member.object_id}
                    >
                        <NodeIcon nodeType={member.primary_kind} />{member.name}
                    </Box>
                )
            })}
        </Box>
    );
}

export default AssetGroupMemberList;