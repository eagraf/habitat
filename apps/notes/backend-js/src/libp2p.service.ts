import Libp2p from "libp2p"
import TCP from 'libp2p-tcp'
import { Noise } from '@chainsafe/libp2p-noise'
import PeerId from 'peer-id';
//@ts-ignore
import  MPLEX  from 'libp2p-mplex'

let libp2p: Libp2p

export async function initLibp2p(peerId?: PeerId, options?: Libp2p.Libp2pOptions & Libp2p.CreateOptions) {
    libp2p = await Libp2p.create({
        ...options,
        addresses: {
            listen: ['/ip4/127.0.0.1/tcp/0']
        },
        modules: {
            transport: [TCP],
            streamMuxer: [MPLEX],
            connEncryption: [new Noise()],
            
            ...options?.modules
        },
    })

    await libp2p.handle('/echo/1.0.0', ({ stream }) => console.log(stream))

    return libp2p
}

export function getLibp2p() {
    return libp2p;
}