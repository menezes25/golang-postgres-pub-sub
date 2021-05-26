/*
 * * Eventos:
 * open – connection established
 * message – data received
 * error – websocket error
 * close – connection closed
*/

const BASE_URL = "ws://localhost:8880";

function handleWsOpen(openEvent) {
  console.debug('WebsocketOpen', openEvent);
  console.debug('Conexão estabilizada');
}

function handleWsMessage(messageEvent) {
  console.debug('WebsocketMessage', messageEvent);

  try {
    console.debug('Dados recebidos do servidor:', JSON.parse(messageEvent.data));  
  } catch (error) {
    console.debug(error);
    console.debug('Dados recebidos do servidor:', messageEvent.data);  
  }
}

async function handleWsClose(closeEvent) {
  console.debug('WebsocketClose', closeEvent);

  if (closeEvent.wasClean) {
    console.debug(`Conexão fechada corretamente, código: ${closeEvent.code}, motivo: ${closeEvent.reason}`);
  } else {
    console.debug(`Conexão fechada por erro de rede, código: ${closeEvent.code}, motivo: ${closeEvent.reason}`);
  }

  await sleep(10000)
  createNewSocket()
}

function handleWsError(errorEvent) {
  console.debug('WebsocketError', errorEvent);
}

const sleep = (delay) => new Promise((resolve) => setTimeout(resolve, delay))

function createNewSocket() {
  const socket = new WebSocket(`${BASE_URL}/ws/topic?listen=boleto,event`);
  socket.onopen = handleWsOpen;
  socket.onmessage = handleWsMessage;
  socket.onclose = handleWsClose;
  socket.onerror = handleWsError;
  return socket;
}

export default {
  createNewSocket,
};
