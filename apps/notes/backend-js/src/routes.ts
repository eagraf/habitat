import { Router } from "express"
import { getLibp2p } from "./libp2p.service"
import { createFromB58String } from 'peer-id'
import { multiaddr } from "multiaddr";

const router = Router();

router.post('/connect', async (req, res) => {
    const libp2p = getLibp2p();

    try {
        const addrStr = req.query.addr as string
        const addr = multiaddr(addrStr)
    
        const peerStr = req.query.peer as string
    
        const peerId = createFromB58String(peerStr)
    
        await libp2p.peerStore.addressBook.add(peerId, [addr])
        res.sendStatus(200)

    }
    catch {
        res.sendStatus(400)
    }
})

export default router