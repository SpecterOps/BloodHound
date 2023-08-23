import { Box, List, ListItem, Paper, TextField } from "@mui/material";
import { useCombobox } from "downshift";
import { FC, useEffect, useRef, useState } from "react";
import SearchResultItem from "../SearchResultItem";

export type FlatNode = GraphNode & { id: string; };

export type GraphNode = {
    label: string;
    kind: string;
    objectId: string;
    lastSeen: string;
    isTierZero: boolean;
    descendent_count?: number | null;
};

type GraphNodes = Record<string, GraphNode>;

const SearchCurrentNodes: FC<{
    currentNodes: GraphNodes;
    onBlur?: () => void;
}> = ({ currentNodes, onBlur }) => {
    
    const inputRef = useRef<HTMLInputElement>(null);
    const [items, setItems] = useState<FlatNode[]>([]);

    const flatNodeList: FlatNode[] = Object.entries(currentNodes).map(([key, value]) => {
        return { id: key, ...value }
    });

    useEffect(() => {
        inputRef.current?.focus();
    }, []);

    const { getInputProps, getMenuProps, getComboboxProps, getItemProps, inputValue } = useCombobox({
        items,
        onInputValueChange: ({ inputValue }) => {
            const filteredNodes = flatNodeList.filter(node => {
                const label = node.label.toLowerCase();
                const objectId = node.objectId.toLowerCase();
                const lowercaseInputValue = inputValue?.toLowerCase() || '';

                if (inputValue === '') return false;

                return label.includes(lowercaseInputValue) || objectId.includes(lowercaseInputValue);
            });
            setItems(filteredNodes);
        }
    });

    return (
        <Box
            component={Paper}
            borderRadius={1}
            bgcolor={'white'}
            width={600}
            marginLeft={2}
            padding={1}
            {...getComboboxProps()}
        >
            <TextField
                inputRef={inputRef}
                InputProps={{ onBlur }}
                variant="outlined"
                size="small"
                fullWidth
                {...getInputProps()}
            />
            <Box overflow={"auto"} height={350}>
                <List dense {...getMenuProps()}>
                    {items.map((node, index) => {
                        return (
                            <SearchResultItem
                                item={node}
                                index={index}
                                key={index}
                                highlightedIndex={0}
                                keyword={inputValue}
                                getItemProps={getItemProps}
                            />
                        );
                    })}
                    {items.length === 0 && inputValue && <ListItem>No Results Found</ListItem>}
                </List>
            </Box>
        </Box>
    );
}

export default SearchCurrentNodes;