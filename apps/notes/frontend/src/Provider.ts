import * as random from 'lib0/random'
import * as encoding from 'lib0/encoding'
import * as decoding from 'lib0/decoding'
import { Observable } from 'lib0/observable'


import * as Y from 'yjs' // eslint-disable-line
import Peer from 'simple-peer'

import * as syncProtocol from 'y-protocols/sync'
import * as awarenessProtocol from 'y-protocols/awareness'

import * as ipfsHttpClient from 'ipfs-http-client'

type ProviderEvents = 'ready' 
                    | 'error'
                    | 'newPeer'
                    | 'synced'

enum SignallingEvents {
  Announce = 0,
  Acknowledge = 1,
  Signal = 2,
}

enum MessageTypes {
  Sync = 0,
  Awareness = 1,
}


type ProviderPeer = Peer.Instance & { ready?: boolean }

export default class YjsProvider extends Observable<ProviderEvents> {

  docName: string
  yDoc: Y.Doc
  ipfs: ipfsHttpClient.IPFSHTTPClient
  peerId: string
  peers: Map<string, ProviderPeer>
  awareness: awarenessProtocol.Awareness

  constructor(docName: string, yDoc: Y.Doc, ipfs: ipfsHttpClient.IPFSHTTPClient) {
    super()
    this.docName = docName
    this.yDoc = yDoc
    this.ipfs = ipfs
    this.peerId = random.uuidv4()   
    this.peers = new Map()
    this.awareness = new awarenessProtocol.Awareness(yDoc)
  }

  _broadcastMessage(encoder: encoding.Encoder) {
    this.peers.forEach(peer => {
      if (peer.ready) {
        peer.send(encoding.toUint8Array(encoder))
      }
    })
  }

  async connect() {
    try {

      this.yDoc.on('update', update => {
        const encoder = encoding.createEncoder()
        encoding.writeVarString(encoder, this.peerId)
        encoding.writeUint8(encoder, MessageTypes.Sync) 
        syncProtocol.writeUpdate(encoder, update)
        this._broadcastMessage(encoder)
      })
  
      this.awareness.on('update', ({ added, updated, removed }) => {
        const changedClients = added.concat(updated).concat(removed)
        const encoder = encoding.createEncoder()
        encoding.writeVarString(encoder, this.peerId)
        encoding.writeVarUint(encoder, MessageTypes.Awareness)
        encoding.writeVarUint8Array(encoder, awarenessProtocol.encodeAwarenessUpdate(this.awareness, changedClients))
        this._broadcastMessage(encoder)
      })

      await this.ipfs.pubsub.subscribe(this.docName, msg => {
        const decoder = decoding.createDecoder(msg.data)
        const sender = decoding.readVarString(decoder)
        if(sender === this.peerId) {
          return
        }
        const event: SignallingEvents = decoding.readUint8(decoder)
        switch(event) {
          case SignallingEvents.Signal: {
            this._handleSignal(sender, decoder)
            break
          }
          case SignallingEvents.Announce: {
            this._handleAnnounce(sender, decoder)
            break
          }
          case SignallingEvents.Acknowledge: {
            this._handleAcknowledgement(sender, decoder)
          }
        }
      })
  
      await this._announce()
      this.emit('ready', [])
    }
    catch(error) {
      console.error(error)
      this.emit('error', [error])
    }

  }

  disconnect() {
    this.peers.forEach((peer, id) => {
      peer.destroy()
    })
  }

  async _announce() {
    const encoder = encoding.createEncoder()
    encoding.writeVarString(encoder, this.peerId)
    encoding.writeUint8(encoder, SignallingEvents.Announce)
    await this.ipfs.pubsub.publish(this.docName, encoding.toUint8Array(encoder))
  }

  async _acknowledge(recipient: string) {
    const encoder = encoding.createEncoder()
    encoding.writeVarString(encoder, this.peerId)
    encoding.writeUint8(encoder, SignallingEvents.Acknowledge)
    encoding.writeVarString(encoder, recipient)
    await this.ipfs.pubsub.publish(this.docName, encoding.toUint8Array(encoder))
  }

  async _signal(recipient: string, data: any) {
    const encoder = encoding.createEncoder()
    encoding.writeVarString(encoder, this.peerId)
    encoding.writeUint8(encoder, SignallingEvents.Signal)
    encoding.writeVarString(encoder, recipient)
    encoding.writeAny(encoder, data)
    await this.ipfs.pubsub.publish(this.docName, encoding.toUint8Array(encoder))
  }

  _createPeer(peerId: string, opts?: Peer.Options) {
    const peer: ProviderPeer = new Peer(opts)
    peer.ready = false

    peer.on('signal', data => {
      this._signal(peerId, data)
    })
    peer.on('data', data => {
      const decoder = decoding.createDecoder(data)
      const sender = decoding.readVarString(decoder)
      const messageType: MessageTypes = decoding.readUint8(decoder)
      switch(messageType) {
        case MessageTypes.Sync: {
          const encoder = encoding.createEncoder()
          encoding.writeVarString(encoder, this.peerId)
          encoding.writeUint8(encoder, MessageTypes.Sync)
          const syncMessageType = syncProtocol.readSyncMessage(decoder, encoder, this.yDoc, sender)
          if(syncMessageType === syncProtocol.messageYjsSyncStep1) {
            peer.send(encoding.toUint8Array(encoder))
          }
          else {
            this.emit('synced', [peerId])
          }
          break
        }
        case MessageTypes.Awareness: {
          awarenessProtocol.applyAwarenessUpdate(this.awareness, decoding.readVarUint8Array(decoder), sender)
          break
        }
      }
    })
    peer.on('close', () => {
      peer.destroy()
      this.peers.delete(peerId)
    })
    peer.on('connect', () => {
      peer.ready = true
      const syncEncoder = encoding.createEncoder()
      encoding.writeVarString(syncEncoder, this.peerId)
      encoding.writeUint8(syncEncoder, MessageTypes.Sync)
      syncProtocol.writeSyncStep1(syncEncoder, this.yDoc)
      peer.send(encoding.toUint8Array(syncEncoder))

      const awarenessStates = this.awareness.getStates()
      if (awarenessStates.size > 0) {
        const awarenessEncoder = encoding.createEncoder()
        encoding.writeVarString(awarenessEncoder, this.peerId)
        encoding.writeVarUint(awarenessEncoder, MessageTypes.Awareness)
        const awarenessUpdate = awarenessProtocol.encodeAwarenessUpdate(this.awareness, Array.from(awarenessStates.keys()))
        encoding.writeVarUint8Array(awarenessEncoder, awarenessUpdate)
        peer.send(encoding.toUint8Array(awarenessEncoder))
      }

      this.emit('newPeer', [peerId])
    })
    peer.on('error', (error) => {
      peer.destroy()
      this.peers.delete(peerId)
    })

    this.peers.set(peerId, peer)

    return peer
  }


  async _handleAnnounce(sender: string, decoder: decoding.Decoder) {
    this._acknowledge(sender)
    if(this.peerId < sender) {
      this._createPeer(sender, { initiator: true })
    }
  }

  async _handleAcknowledgement(sender: string, decoder: decoding.Decoder) {
    const recipient = decoding.readVarString(decoder)
    if(recipient !== this.peerId) {
      return
    }
    if(this.peerId < sender) {
      this._createPeer(sender, { initiator: true })
    }
  }

  async _handleSignal(sender: string, decoder: decoding.Decoder) {

    const recipient = decoding.readVarString(decoder)
    if(recipient !== this.peerId) {
      return
    }
    const data = decoding.readAny(decoder)
    if(!this.peers.has(sender)) {
      this._createPeer(sender)
    }
    const peer = this.peers.get(sender)
    peer.signal(data)
  }
}



