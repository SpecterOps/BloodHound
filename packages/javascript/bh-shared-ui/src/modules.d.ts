declare module '@neo4j-cypher/react-codemirror' {
    type CypherEditorProps = import('@neo4j-cypher/react-codemirror/src/react-codemirror.d.ts').CypherEditorProps;
    export class CypherEditor extends React.Component<CypherEditorProps, any> {}
}
