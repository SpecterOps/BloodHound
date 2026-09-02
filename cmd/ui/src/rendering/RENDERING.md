# Custom Sigma Renderers

This directory holds all custom rendering programs we are using to override Sigmajs default functionality. Most overrides will need to be imported and instantiated in the Sigma settings located in our `SigmaChart` component.

## Node Programs

### Combined Nodes

The `combined` node type is from an [unmerged PR](https://github.com/jacomyal/sigma.js/pull/1206) in the Sigmajs github repository. It combines the `image` and `border` node types that already ship with Sigma.

**Example Usage:**

```
graph.addNode(key, {
    x: 10,
    y: 20,
    size: 30,
    type: 'combined',
    borderColor: '#000000',
    image: './icons/computer.svg'
});
```

### Glyph Nodes

The `glyphs` node type is further modified from the `combined` type, and allows us to render up to 4 glyphs around the edge of the node. Different glyph types are now defined in our `constants.ts` file along with our node definitions.

**Example Usage:**

```
graph.addNode(key, {
    x: 10,
    y: 20,
    size: 30,
    type: 'glyphs',
    borderColor: '#000000',
    image: './icons/computer.svg',
    glyphs: [
        {
            location: GlyphLocation.TOP_RIGHT,
            image: `./icons/${GLYPHS.tierZero.icon.iconName}.svg`,
            backgroundColor: GLYPHS.tierZero.background,
        },
    ]
});
```

**Notes:**

-   Only one glyph can be rendered for each value of the `GlyphLocation` enum; later elements in the `glyphs` array take precedence.
-   For performance concerns, we may want to modify this in the future to instantiate multiple different node types based on the number of glyphs we want to add. SigmaJS uses pre-allocated fixed length buffers to store node data, which means that we are hitting the shader 5 times as often vs. the `combined` type (even if we only want to render one glyph).

## Edge Programs

### Curved Edges

The `curved` edge type displays groups of edges as quadratic bezier curves radiating out from the line between the source and target node. These curves are calculated from `groupSize` and `groupPosition` values that need to be passed along with the edge.

A `direction` value can also be passed in to handle groups that have edges going in multiple directions between source and target (There is no implicit direction; passing in `EdgeDirection.BACKWARDS` will just flip the curve across the line of symmetry).

**Example Usage:**

```
graph.addEdgeWithKey(key, edge.source, edge.target, {
    size: 3,
    type: 'curved',
    label: edge.label,
    color: '#406F8E',
    groupPosition: 5,
    groupSize: 10,
    direction: EdgeDirection.BACKWARDS
});
```

### Arrow Edges

The `arrow` edge type is identical to the one that ships with Sigma except that we are using slightly larger arrowheads to match our curved edges.

## Hover Rendering

We are using a slightly modified version of Sigma's built-in `drawHover` function to add custom styles to hover/highlighted states. We can similarly overwrite the `drawLabel` and `drawEdgeLabel` functions as needed.

Sigma uses a 2D canvas element to display hover states and node labels that appears to be layered inbetween the webGL canvas used to render highlighted nodes and the canvas used to render all other nodes. This may make it difficult to render labels over the top of nodes, so for now they are rendered in the default sigma location to the right of the node.
