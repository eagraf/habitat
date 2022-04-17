import express, { raw } from 'express'
import { Server as WsServer } from 'ws'
import * as Y from 'yjs'
import * as ipfsHttpClient from 'ipfs-http-client'
import OrbitDb from 'orbit-db'
import stream from 'stream'
import cors from 'cors'
import EventStore from 'orbit-db-eventstore'
import { URL } from 'url'

import * as decoding from 'lib0/decoding'
import * as encoding from 'lib0/encoding'
import syncProtocol from 'y-protocols/sync'

import storage from 'node-persist';

const PORT = process.argv[2];

enum DocState {
  Active,
  Inactive,
  Closed,
}

async function main() {
  const ipfs = ipfsHttpClient.create({
    host: 'localhost',
    port: parseInt(process.argv[3]),
  })

  await storage.init({
    dir: 'persist'
  })
  const docs = new Map<string, EventStore<Uint8Array> | null>()

  const stored = await storage.get('docs')
  console.log(stored)
  stored.forEach((x: string) => {
    docs.set(x, null)
  })

  const peerId = (await ipfs.id()).id
  console.log(peerId)

  ipfs.pubsub.subscribe('habitat_notes', async (msg) => {
    if(msg.from === peerId) {
      return
    }
    const decoder = decoding.createDecoder(msg.data)
    const docName = decoding.readVarString(decoder)
    console.log(docName)

    if(docs.has(docName)) {
      if(!docs.get(docName)) {
        console.log('waking up orbitdb')
        const db = await orbitdb.eventlog<Uint8Array>(docName, {
          accessController: {
            write: ["*"]
          },
          create: true,
        })
        docs.set(docName, db)
        db.events.on('peer.exchanged', async (peer) => {
          if(peer === msg.from) {
            console.log('exchanged and closing')
            db.close()
          }
          docs.set(docName, null)
        })
      }
    }
  })
  
  const orbitdb = await OrbitDb.createInstance(ipfs)
  console.log('orbitdb instance created')

  const app = express()
  app.use(cors())

  app.get('/docs', (req, res) => {
    console.log(Array.from(docs.keys()))
    res.send(Array.from(docs.keys()))
  })

  app.post('/newDoc', async (req, res) => {
    const docName = req.query.name as string
    if(OrbitDb.isValidAddress(docName)) {
      res.send(docName)
    }
    else {
      const addr = await orbitdb.determineAddress(docName, 'eventlog', {
        accessController: {
          write: ["*"]
        },
      })
      //@ts-ignore
      res.send('/orbitdb/' + addr.root + '/' + addr.path)
    }
  })

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


    const encoder = encoding.createEncoder()
    encoding.writeVarString(encoder, docName)
    ipfs.pubsub.publish('habitat_notes', encoding.toUint8Array(encoder))


    if(docs.get(docName)) {
      await docs.get(docName)!.close()
    }

    const db = (await orbitdb.eventlog<Uint8Array>(docName, {
      accessController: {
        write: ["*"]
      },
      create: true,
    }))
    
    await storage.setItem('docs', Array.from(docs.keys()))

    docs.set(docName, db)

    db.events.on('replicate.progress', (address, hash, entry, progress, have) => {
      //console.log("Replicating", address, entry, progress, have)
      ws.send(entry.payload.value)
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
      db.close()
      docs.set(docName, null)
      console.log("closed orbitdb after disconnect")

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