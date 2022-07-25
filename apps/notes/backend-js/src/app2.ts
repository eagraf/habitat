import express from 'express';
import { initLibp2p, getLibp2p } from './libp2p.service';
import cors from 'cors'
import router from './routes'
import { pipe } from 'it-pipe'
import { multiaddr } from 'ipfs-http-client';

const PORT = process.argv[2];

async function main() {
    
    const libp2p = await initLibp2p()

    await libp2p.start()

    libp2p.multiaddrs.forEach(ma => {
        console.log(ma.toString())
    })

    console.log(libp2p.peerId.toB58String())

    if (process.argv.length >= 5) {
        const ma = new multiaddr(process.argv[4])
        const latency = await libp2p.ping(ma)
        console.log(`pinged ${ma} in ${latency} ms`)
    }
    // libp2p.handle('/notes/1.0.0', async ({ stream }) => {
    //     for await (const chunk of stream.source) {
    //         console.log(chunk)
    //     }
    // })

    const app = express()
    app.use(cors())
  

    const server = app.listen(PORT, () => {
        return console.log(`server is listening on ${PORT}`);
    });

    const stop = async () => {
        // stop libp2p
        await libp2p.stop()
        console.log('libp2p has stopped')
        process.exit(0)
    }
      
    process.on('SIGTERM', stop)
    process.on('SIGINT', stop)
}

main()