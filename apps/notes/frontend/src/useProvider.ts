import { useState, useEffect } from 'react';
import Provider from './Provider'

export default function useProvider(docName, yDoc, ipfs) {
    const [provider, setProvider] = useState<Provider>(() => new Provider(docName, yDoc, ipfs))
    useEffect(() => {
        if(docName !== provider.docName) {
            console.log(docName, provider.docName)
            provider.disconnect()
            const newProvider = new Provider(docName, yDoc, ipfs)
            setProvider(newProvider)
        }
    }, [docName])

    return provider
}