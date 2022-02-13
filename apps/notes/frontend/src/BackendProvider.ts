import { Observable } from 'lib0/observable'

import * as Y from 'yjs';

export default class BackendProvider {
    docName: string
    yDoc: Y.Doc
    ws: WebSocket
    backendUrl: string
    
    constructor(docName, yDoc, backendUrl) {
        this.docName = docName
        this.yDoc = yDoc
        this.backendUrl = backendUrl
    }

    async connect() {

        const response = await fetch('http://' + this.backendUrl + '/doc?name=' + this.docName)
        const update = new Uint8Array(await response.arrayBuffer())
        console.log(update)
        if(update.length > 0) {
            Y.applyUpdate(this.yDoc, update)
        }

        this.ws = new WebSocket('ws://' + this.backendUrl + '?name=' + this.docName)

        await new Promise((resolve, reject) => {
            console.log('hello world')
            this.ws.addEventListener('open', event => {
                console.log('hello')
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
}