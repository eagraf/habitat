import * as Y from "yjs";

export default interface Provider {
    docName: string
    yDoc: Y.Doc
    
    connect(): void
    disconnect(): void
}