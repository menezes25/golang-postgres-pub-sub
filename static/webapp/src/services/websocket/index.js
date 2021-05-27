import Swal from "sweetalert2";
import store from "@/store";

const BASE_URL = process.env.VUE_APP_WEBSOCKET_BASE_URL;

const sleep = (delay) => new Promise((resolve) => setTimeout(resolve, delay));

function handleWsOpen(/*openEvent*/) {
  console.info('[WEBSOCKET] Conexão estabilizada');

  Swal.fire({
    icon: 'success',
    titleText: 'WS Conexão estabilizada',
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    showCloseButton: true,
  });
}

function handleWsMessage(messageEvent) {
  let payload;
  try {
    payload = JSON.parse(messageEvent.data);
  } catch (error) {
    console.error('[WEBSOCKET] Dados não JSON recebidos do servidor:', messageEvent.data, error);
    return;
  }

  console.info('[WEBSOCKET] Dados JSON recebidos do servidor:', JSON.parse(messageEvent.data));
  store.addEventsInProgress(payload);
}

async function handleWsClose(closeEvent) {
  if (closeEvent.wasClean) {
    console.debug(`[WEBSOCKET] Conexão fechada corretamente, código: ${closeEvent.code}, motivo: ${closeEvent.reason}`);
  } else {
    console.debug(`[WEBSOCKET] Conexão fechada por erro de rede, código: ${closeEvent.code}, motivo: ${closeEvent.reason}`);
  }

  await sleep(10000)
  startWebSocket()
}

function handleWsError(errorEvent) {
  console.debug('[WEBSOCKET] Error:', errorEvent);

  const readyStateText = new Map()
    .set(WebSocket.CONNECTING, 'CONNECTING')
    .set(WebSocket.OPEN, 'OPEN')
    .set(WebSocket.CLOSING, 'CLOSING')
    .set(WebSocket.CLOSED, 'CLOSED')

  Swal.fire({
    icon: 'error',
    titleText: `WS Error: state ${readyStateText.get(errorEvent.srcElement.readyState)}`,
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    showCloseButton: true,
  });
}

function startWebSocket() {
  const socket = new WebSocket(`${BASE_URL}/ws/topic?listen=boleto,event`);
  socket.onopen = handleWsOpen;
  socket.onmessage = handleWsMessage;
  socket.onclose = handleWsClose;
  socket.onerror = handleWsError;

  Swal.fire({
    title: 'WS Conectando...',
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    showCloseButton: true,
  });
  Swal.showLoading()

  return socket;
}

export default {
  startWebSocket,
};
