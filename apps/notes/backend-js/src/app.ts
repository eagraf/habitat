import express from 'express'
import { Server as WsServer } from 'ws'
import * as Y from 'yjs'
import * as ipfsHttpClient from 'ipfs-http-client'
import OrbitDb from 'orbit-db'
import stream from 'stream'
import cors from 'cors'
import EventStore from 'orbit-db-eventstore'
import { URL } from 'url'

const PORT = 4000;


async function main() {
  const ipfs = ipfsHttpClient.create({
    host: 'localhost',
    port: 5001,
  })

  const orbitdb = await OrbitDb.createInstance(ipfs)
  console.log('orbitdb instance created')

  const addr = await orbitdb.determineAddress('docName', 'eventlog', {
    accessController: {
      write: ['*']
    }
  })

  console.log(addr)


  const app = express();
  app.use(cors())

  app.get("/doc", async (req, res) => {
    const docName = req.query['name'] as string
    const db = (await orbitdb.open(docName)) as EventStore<Uint8Array>
    await db.load()
    let state = null
    const iterator = db.iterator({ limit: -1 })
    let { value, done } = iterator.next()
    while(!done) {
      const update = value.payload.value
      if(state) {
        state = Y.mergeUpdates([state, update])
      }
      else {
        state = update
      }
      const next = iterator.next()
      value = next.value
      done = next.done
    } 

    console.log(state)

    const readStream = new stream.PassThrough();

    readStream.end(state)
    readStream.pipe(res)

    await db.close()

  })

  const wsServer = new WsServer({ noServer: true });
  wsServer.on('connection', async (ws, request) => {
    console.log('connection')
    ws.binaryType = 'arraybuffer'
    let changes: Uint8Array
    ws.on('message', data => {
      const update = new Uint8Array(data as ArrayBuffer)
      if(changes) {
        changes = Y.mergeUpdates([changes, update])
      }
      else {
        changes = update
      }
    })

    //TODO: make this better
    console.log(request.url)
    const docName = request.url!.substring(7)
    console.log(docName)

    const db = (await orbitdb.open(docName)) as EventStore<Uint8Array>
    ws.on('close', async (code, reason) => {
      if(changes) {
        await db.load()
        await db.add(changes)
        console.log('changes added')
      }
      await db.close()
    })
    db.events.on('replicated', () => {
      
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