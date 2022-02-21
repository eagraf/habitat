import express, { raw } from 'express'
import { Server as WsServer } from 'ws'
import * as Y from 'yjs'
import * as ipfsHttpClient from 'ipfs-http-client'
import OrbitDb from 'orbit-db'
import stream from 'stream'
import cors from 'cors'
import EventStore from 'orbit-db-eventstore'
import { URL } from 'url'

import decoding from 'lib0/decoding'
import encoding from 'lib0/encoding'
import syncProtocol from 'y-protocols/sync'


const PORT = process.argv[2];

async function main() {
  const ipfs = ipfsHttpClient.create({
    host: 'localhost',
    port: parseInt(process.argv[3]),
  })

  const orbitdb = await OrbitDb.createInstance(ipfs)
  console.log('orbitdb instance created')

  const app = express()
  app.use(cors())

  const wsServer = new WsServer({ noServer: true })
  wsServer.on('connection', async (ws, request) => {
    console.log('connection')
    ws.binaryType = 'arraybuffer'

    let newChanges: Uint8Array
    ws.on('message', rawData => {
      const data = new Uint8Array(rawData as ArrayBuffer)
      if(newChanges) {
        newChanges = Y.mergeUpdates([data, newChanges])
      }
      else {
        newChanges = data
      }
    })

    //TODO: make this better
    const docName = request.url!.substring(7)
    console.log(docName)

    const db = (await orbitdb.eventlog<Uint8Array>(docName, {
      accessController: {
        write: ["*"]
      },
      create: true,
    }))

    db.events.on('replicate', () => {
      console.log('replicate')
    })

    db.events.on('peer', () => {
      console.log('peer')
    })

    db.events.on('peer.exchanged', () => {
      console.log('exchanged')
    })

    db.events.on('replicate.progress', (address, hash, entry, progress, have) => {
      console.log("Replicating", address, entry, progress, have)
      ws.send(entry.payload.value)
      //ws.send(entry.payload.value)
    })

    let oldChanges: Uint8Array
    db.events.on('load.progress', (address, hash, entry, progress, total) => {
      if(oldChanges) {
        oldChanges = Y.mergeUpdates([oldChanges, entry.payload.value])
      }
      else {
        oldChanges = entry.payload.value
      }
    })

    await db.load()
    console.log("loaded")
    
    if(oldChanges!) {
      ws.send(oldChanges!.buffer)
    }

    ws.on('close', async (code, reason) => {
      if(newChanges) {
        await db.add(newChanges)
        console.log('changes added')
      }
    })
  })
  
  const server = app.listen(PORT, () => {
    return console.log(`server is listening on ${PORT}`);
  });

  server.on('upgrade', (request, socket, head) => {
    wsServer.handleUpgrade(request, socket, head, ws => {
      wsServer.emit('connection', ws, request)
    })
  })
}

main()