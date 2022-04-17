import { Observable } from 'lib0/observable'

import * as Y from 'yjs';

import Provider from './Provider'

export default class BackendProvider implements Provider {
    docName: string
    yDoc: Y.Doc
    ws: WebSocket
    backendUrl: string
    
    constructor(docName: string, yDoc: Y.Doc, backendUrl: string) {
        this.docName = docName
        this.yDoc = yDoc
        this.backendUrl = backendUrl
    }

    async connect() {
        this.ws = new WebSocket('ws://' + this.backendUrl + '?name=' + this.docName)
        this.ws.binaryType = 'arraybuffer'

        this.ws.addEventListener('message', event => {
            const change = new Uint8Array(event.data as ArrayBuffer)
            Y.applyUpdate(this.yDoc, change)
        })

        await new Promise((resolve, reject) => {
            this.ws.addEventListener('open', event => {
                resolve(event)
            })
            this.ws.addEventListener('error', error => {
                console.error(error)
                reject(error)
            })
        })

        this.yDoc.on('update', (update: Uint8Array, origin) => {
            if(origin?.key) {
                this.ws.send(update.buffer)
                console.log('sent update')
            }
        })
    }

    async disconnect() {
        if(this.ws) {
            this.ws.close()
        }
    }
}