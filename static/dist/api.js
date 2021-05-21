// const BASE_URL = "https://pocwebsocket.free.beeceptor.com";
const BASE_URL = "http://localhost:8880";

// lança um ApiResponseError se o status http for fora da faixa de sucesso (200-299)
function handleApiErrors(responseFetch, response) {
  if (!responseFetch.ok) {
    const err = new Error(response.message || `Request failed with status ${responseFetch.status} - ${responseFetch.statusText}`);
    err.name = 'ApiResponseError';
    err.response = response;
    throw err;
  }
}

// {"event": "custom-event"}
async function saveNewEvent(event) {
  const headers = new Headers();
  headers.append("Content-Type", "application/json");

  const requestOptions = {
    method: 'POST',
    headers: headers,
    body: JSON.stringify(event),
    redirect: 'follow'
  };

  const res = await fetch(`${BASE_URL}/api/event`, requestOptions);
  const resJson = await res.json();
  handleApiErrors(res, resJson);
  return resJson;
}

export default {
  saveNewEvent,
};
