import http from 'http';
import WebSocket from 'ws'


export default async function websocket(server: http.Server) {
    const wsServer = new WebSocket.Server({
        noServer: true
    })

    wsServer.on('connection', async (ws, req) => {
        ws.binaryType = 'arraybuffer'
        //TODO: make this better
        const noteName = req.url!.substring(7)
    })

    server.on('upgrade', (request, socket, head) => {
        wsServer.handleUpgrade(request, socket, head, ws => {
            wsServer.emit('connection', ws, request)
        })
    })
}