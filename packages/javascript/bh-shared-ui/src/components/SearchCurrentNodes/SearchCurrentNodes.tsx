import { Box, List, ListItem, Paper, TextField } from "@mui/material";
import { useCombobox } from "downshift";
import { FC, useEffect, useRef, useState } from "react";
import SearchResultItem from "../SearchResultItem";
import { FlatNode, GraphNodes } from "./types";

const SearchCurrentNodes: FC<{
    currentNodes: GraphNodes;
    onSelect: (node: FlatNode) => void;
}> = ({ currentNodes, onSelect }) => {
    
    const inputRef = useRef<HTMLInputElement>(null);
    const [items, setItems] = useState<FlatNode[]>([]);
    const [selectedNode, setSelectedNode] = useState<FlatNode | null | undefined>(null)

    // Node data is a lot easier to work with in the combobox if we transform to an array of flat objects
    const flatNodeList: FlatNode[] = Object.entries(currentNodes).map(([key, value]) => {
        return { id: key, ...value }
    });

    useEffect(() => inputRef.current?.focus(), []);

    useEffect(() => {
        if (selectedNode) onSelect(selectedNode);
    }, [selectedNode, onSelect]);

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
        },
        stateReducer: (_state, actionAndChanges) => {
            const { changes, type } = actionAndChanges;
            switch (type) {
                case useCombobox.stateChangeTypes.ItemClick:
                    if (changes.selectedItem) setSelectedNode(changes.selectedItem);
                    return { ...changes, inputValue: '' }
                default:
                    return changes
            }
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
            <Box overflow={"auto"} maxHeight={350} marginBottom={items.length === 0 ? 0 : 1}>
                <List dense {...getMenuProps({
                    hidden: items.length === 0 && !inputValue,
                    style: { paddingTop: 0 }
                })}>
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
            <TextField
                inputRef={inputRef}
                placeholder={"Search Current Results"}
                variant="outlined"
                size="small"
                fullWidth
                {...getInputProps()}
            />
        </Box>
    );
}

export default SearchCurrentNodes;