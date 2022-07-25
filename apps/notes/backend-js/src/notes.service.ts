import { IPFSHTTPClient } from 'ipfs-http-client';
import storage from 'node-persist';
import OrbitDB from 'orbit-db';
import EventStore from 'orbit-db-eventstore'

const IPFS_PORT = parseInt(process.argv[3])

const STORAGE_KEY = 'notes'

const notes = new Map<string, EventStore<Uint8Array> | null>()

let orbitdb: OrbitDB

export async function initNotes(ipfs: IPFSHTTPClient) {
    await storage.init({
        dir: 'persist' + '_' + IPFS_PORT
    })

    orbitdb = await OrbitDB.createInstance(ipfs, {
        directory: './orbitdb_' + IPFS_PORT
    })

    const stored = await storage.get(STORAGE_KEY)
    if (stored) {
        stored.forEach((x: string) => {
            notes.set(x, null)
        })
    }
}

export function getNotes() {
    return notes
}

export async function addNote(name: string) {

    let addr: string
    if (OrbitDB.isValidAddress(name)) {
        addr = name
    }
    else {
        const odb = await orbitdb.determineAddress(name, 'eventlog', {
            accessController: {
                write: ["*"]
            },
        })
        //@ts-ignore
        addr = '/orbitdb/' + odb.root + '/' + odb.path
    }

    notes.set(addr, null)
    await storage.set(STORAGE_KEY, Array.from(notes.keys()))

    return addr
}

export async function openNote(addr: string): Promise<EventStore<Uint8Array>> {
    const db = await orbitdb.eventlog<Uint8Array>(addr, {
        accessController: {
            write: ["*"]
        },
        create: true,
    })
    notes.set(addr, db)
    return db
}

export async function closeNote(addr: string) {
    await notes.get(addr)!.close()
    notes.set(addr, null)
}