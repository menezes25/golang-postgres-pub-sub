import axios from "axios";

class ApiError extends Error {
  constructor(message, response=undefined, originalError=undefined) {
    super(message);
    this.name = 'ApiError';
    this.message = message;
    this.response = response;
    this.originalError = originalError;
  }
}

const apiInstance = axios.create({
  baseURL: process.env.VUE_APP_API_BASE_URL,
});

// {"event": "custom-event"}
async function saveNewEvent(event) {
  try {
    const res = await apiInstance.post(`/api/event`, event);
    return res.data;
  } catch (error) {
    throw new ApiError(error.message, error?.response?.data, error);
  }
}

export default {
  saveNewEvent,
}
