import { runEndpoint, runStreamEndpoint } from "../api/core.js";
import { authEndpoints } from "../api/modules/auth.js";
import { chatEndpoints } from "../api/modules/chat.js";
import { userEndpoints } from "../api/modules/user.js";

const wrapEndpoints = (endpoints, runner = runEndpoint) => {
  const wrapped = {};
  for (const [name, factory] of Object.entries(endpoints)) {
    wrapped[name] = (...args) => runner(factory(...args));
  }
  return wrapped;
};

// chat 域里只有 interviewStream 走 SSE 流式，其他都是普通 HTTP。
const { interviewStream: _interviewStreamFactory, ...chatHttpEndpoints } = chatEndpoints;

export const apiService = {
  auth: wrapEndpoints(authEndpoints),
  user: wrapEndpoints(userEndpoints),
  chat: {
    interviewStream: (payload, chatId) =>
      runStreamEndpoint(chatEndpoints.interviewStream(payload, chatId)),
    ...wrapEndpoints(chatHttpEndpoints),
  },
};

export const useApi = () => apiService;
