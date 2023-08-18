import { Box, Paper } from "@mui/material";
import { FC } from "react";

const SearchCurrentResults: FC<{
    data: string;
}> = ({ data }) => {
    
    return (
        <Box borderRadius={1} bgcolor={'white'} height={200} width={200} marginLeft={2}>Search {data}</Box>
    );
}

export default SearchCurrentResults;