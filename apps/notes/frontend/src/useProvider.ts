import { useState, useEffect } from 'react';
import Provider from './Provider'
import * as Y from 'yjs';
import { IPFSHTTPClient } from 'ipfs-http-client';

export default function useProvider(docName: string, yDoc: Y.Doc, ipfs: IPFSHTTPClient) {
    const [provider, setProvider] = useState<Provider>(() => new Provider(docName, yDoc, ipfs))
    useEffect(() => {
        if(docName !== provider.docName) {
            console.log(docName, provider.docName)
            provider.disconnect()
            const newProvider = new Provider(docName, yDoc, ipfs)
            setProvider(newProvider)
        }
    }, [docName, yDoc, ipfs])

    return provider
}