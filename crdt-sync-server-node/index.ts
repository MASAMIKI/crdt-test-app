import express, { Application } from 'express';
import * as http from "http";
import * as WebSocket from "ws";

const app: Application = express();
const host = 'localhost';
const port = process.env.PORT || 8080;
const server = http.createServer(app);
const wss = new WebSocket.Server({ noServer: true });
const { setupWSConnection } = require('y-websocket/bin/utils');

// setup websocket connectionをカスタマイズする必要がある。
wss.on('connection', setupWSConnection);
server.on('upgrade', (request, socket, head) => {
  wss.handleUpgrade(request, socket, head, (ws) => {
    wss.emit('connection', ws, request);
  });
});

server.listen(Number(port), host, () => {
  console.log(`Server is Fire at http://${host}:${port}`);
})
