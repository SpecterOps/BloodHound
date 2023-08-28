import { Box, List, ListItem, Paper, SxProps, TextField } from "@mui/material";
import { useCombobox } from "downshift";
import { FC, useEffect, useRef, useState } from "react";
import SearchResultItem from "../SearchResultItem";
import { FlatNode, GraphNodes } from "./types";
import { useOnClickOutside } from "../../hooks";

export const PLACEHOLDER_TEXT = "Search Current Results";
export const NO_RESULTS_TEXT = "No result found in current results";

const SearchCurrentNodes: FC<{
    sx?: SxProps;
    currentNodes: GraphNodes;
    onSelect: (node: FlatNode) => void;
    onClose?: () => void;
}> = ({ sx, currentNodes, onSelect, onClose }) => {
    
    const containerRef = useRef<HTMLDivElement>(null);
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

    useOnClickOutside(containerRef, () => onClose && onClose());

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
        <div ref={containerRef}>
            <Box component={Paper} {...sx} {...getComboboxProps()}>
                <Box overflow={"auto"} maxHeight={350} marginBottom={items.length === 0 ? 0 : 1}>
                    <List data-testid={"current-results-list"} dense {...getMenuProps({
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
                        {items.length === 0 && inputValue && <ListItem disabled sx={{ fontSize: 14 }}>{NO_RESULTS_TEXT}</ListItem>}
                    </List>
                </Box>
                <TextField
                    inputRef={inputRef}
                    placeholder={PLACEHOLDER_TEXT}
                    variant="outlined"
                    size="small"
                    fullWidth
                    {...getInputProps()}
                    InputProps={{ sx: { fontSize: 14 } }}
                />
            </Box>
        </div>
    );
}

export default SearchCurrentNodes;