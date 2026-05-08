import axios from "axios";

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api";
export const USER_API_BASE_URL = import.meta.env.VITE_USER_API_BASE_URL || API_BASE_URL;
export const CHAT_API_BASE_URL = import.meta.env.VITE_CHAT_API_BASE_URL || API_BASE_URL;

export const ACCESS_TOKEN_KEY = "gozero-ai-access-token";
export const REFRESH_TOKEN_KEY = "gozero-ai-refresh-token";
export const AUTH_USER_KEY = "gozero-ai-auth-user";

const readStorage = (key) => {
  if (typeof window === "undefined") {
    return "";
  }
  return window.localStorage.getItem(key) || "";
};

export const authStorage = {
  getSession() {
    return {
      accessToken: readStorage(ACCESS_TOKEN_KEY),
      refreshToken: readStorage(REFRESH_TOKEN_KEY),
      username: readStorage(AUTH_USER_KEY),
    };
  },
  setSession({ accessToken = "", refreshToken = "", username = "" }) {
    if (typeof window === "undefined") {
      return;
    }
    window.localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
    window.localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
    window.localStorage.setItem(AUTH_USER_KEY, username);
  },
  clearSession() {
    if (typeof window === "undefined") {
      return;
    }
    window.localStorage.removeItem(ACCESS_TOKEN_KEY);
    window.localStorage.removeItem(REFRESH_TOKEN_KEY);
    window.localStorage.removeItem(AUTH_USER_KEY);
  },
};

export const getRequestErrorMessage = (error, fallback = "请求失败，请稍后再试") =>
  error?.response?.data?.message ||
  error?.response?.data?.msg ||
  error?.message ||
  fallback;

const attachAuthInterceptors = (instance) => {
  instance.interceptors.request.use((config) => {
    const accessToken = readStorage(ACCESS_TOKEN_KEY);
    if (accessToken) {
      config.headers = config.headers || {};
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    return config;
  });

  instance.interceptors.response.use(
    (response) => response,
    (error) => {
      const nextError = new Error(
        error?.response?.data?.message || error?.response?.data?.msg || error.message || "请求失败"
      );
      nextError.response = error?.response;
      return Promise.reject(nextError);
    }
  );
};

const createHttpClient = (baseURL, timeout = 60000) => {
  const instance = axios.create({
    baseURL,
    timeout,
  });

  attachAuthInterceptors(instance);
  return instance;
};

export const requestClients = {
  user: createHttpClient(USER_API_BASE_URL),
  chat: createHttpClient(CHAT_API_BASE_URL),
};

export const runEndpoint = async (endpoint) => {
  const {
    service = "user",
    method = "get",
    url,
    data,
    params,
    headers,
    timeout,
    responseType,
  } = endpoint;

  const client = requestClients[service];

  const response = await client.request({
    method,
    url,
    data,
    params,
    headers,
    timeout,
    responseType,
  });

  return response.data;
};

const buildQueryString = (params = {}) =>
  Object.entries(params)
    .filter(([, value]) => value !== undefined && value !== null && value !== "")
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
    .join("&");

const connectSSEByQuery = (baseURL, endpoint, onMessage, onError) => {
  const queryString = buildQueryString(endpoint.params);
  const fullUrl = `${baseURL}${endpoint.url}${queryString ? `?${queryString}` : ""}`;
  const eventSource = new EventSource(fullUrl);

  eventSource.onmessage = (event) => {
    onMessage?.(event.data);
  };

  eventSource.onerror = (error) => {
    onError?.(error);
    eventSource.close();
  };

  return eventSource;
};

const connectSSEByFetch = (baseURL, endpoint) => {
  const accessToken = readStorage(ACCESS_TOKEN_KEY);
  let abortController = null;
  let messageCallback = null;
  let errorCallback = null;

  const fetchSSE = async () => {
    try {
      abortController = new AbortController();

      const response = await fetch(`${baseURL}${endpoint.url}`, {
        method: endpoint.method || "POST",
        body: endpoint.data,
        signal: abortController.signal,
        headers: accessToken
          ? {
              Authorization: `Bearer ${accessToken}`,
              ...(endpoint.headers || {}),
            }
          : endpoint.headers,
      });

      if (!response.ok || !response.body) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();

        if (done) {
          messageCallback?.("[DONE]");
          break;
        }

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() || "";

        for (const line of lines) {
          if (line.startsWith("data: ")) {
            const data = line.slice(6).replace(/\r$/, "");
            const dataTrimmed = data.trim();
            if (dataTrimmed === "[DONE]") {
              messageCallback?.("[DONE]");
              return;
            }
            if (data) {
              messageCallback?.(data);
            }
          }
        }
      }
    } catch (error) {
      errorCallback?.(error);
    }
  };

  fetchSSE();

  const stream = {
    close() {
      abortController?.abort();
    },
  };

  Object.defineProperty(stream, "onmessage", {
    get: () => messageCallback,
    set: (callback) => {
      messageCallback = callback;
    },
  });

  Object.defineProperty(stream, "onerror", {
    get: () => errorCallback,
    set: (callback) => {
      errorCallback = callback;
    },
  });

  return stream;
};

export const runStreamEndpoint = (endpoint, onMessage, onError) => {
  const baseURL = endpoint.service === "chat" ? CHAT_API_BASE_URL : USER_API_BASE_URL;

  if (endpoint.input === "query") {
    return connectSSEByQuery(baseURL, endpoint, onMessage, onError);
  }

  return connectSSEByFetch(baseURL, endpoint);
};
