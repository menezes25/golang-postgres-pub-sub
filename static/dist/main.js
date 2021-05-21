import api from "./api.js";
import websocket from "./websocket.js";

console.debug(api, websocket);

const dom = {
  btnCreate: document.querySelector('#btn-create'),
  inputEventName: document.querySelector('#input-event-name'),
}

dom.btnCreate.addEventListener('click', async () => {
  try {
    const res = await api.saveNewEvent({ event: dom.inputEventName.value });
    console.debug(res);
  } catch (error) {
    console.debug(error);
    alert(error.message);
    return;
  }

});

//TODO: rotina para reconexão
const socket = websocket.createNewSocket();
console.debug(socket);
