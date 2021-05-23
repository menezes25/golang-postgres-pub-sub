/*
 * * Eventos:
 * open – connection established
 * message – data received
 * error – websocket error
 * close – connection closed
*/

const BASE_URL = "ws://localhost:8880/ws/new";

function handleWsOpen(openEvent) {
  console.debug('WebsocketOpen', openEvent);
  console.debug('Conexão estabilizada');
}

function handleWsMessage(messageEvent) {
  console.debug('WebsocketMessage', messageEvent);
  console.debug('Dados recebidos do servidor:', messageEvent.data);
}

function handleWsClose(closeEvent) {
  console.debug('WebsocketClose', closeEvent);

  if (closeEvent.wasClean) {
    console.debug(`Conexão fechada corretamente, código: ${closeEvent.code}, motivo: ${closeEvent.reason}`);
  } else {
    console.debug(`Conexão fechada por erro de rede, código: ${closeEvent.code}, motivo: ${closeEvent.reason}`);
  }
}

function handleWsError(errorEvent) {
  console.debug('WebsocketError', errorEvent);
}

function createNewSocket() {
  const socket = new WebSocket(BASE_URL+"/event");
  socket.onopen = handleWsOpen;
  socket.onmessage = handleWsMessage;
  socket.onclose = handleWsClose;
  socket.onerror = handleWsError;
  return socket;
}

export default {
  createNewSocket,
};
