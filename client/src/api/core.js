import axios from "axios";

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api";
export const USER_API_BASE_URL = import.meta.env.VITE_USER_API_BASE_URL || API_BASE_URL;
export const CHAT_API_BASE_URL = import.meta.env.VITE_CHAT_API_BASE_URL || API_BASE_URL;

export const ACCESS_TOKEN_KEY = "gozero-ai-access-token";
export const REFRESH_TOKEN_KEY = "gozero-ai-refresh-token";
export const AUTH_USER_KEY = "gozero-ai-auth-user";
export const AUTH_CHANGED_EVENT = "gozero-ai-auth-changed";

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
    window.dispatchEvent(new CustomEvent(AUTH_CHANGED_EVENT));
  },
  clearSession() {
    if (typeof window === "undefined") {
      return;
    }
    window.localStorage.removeItem(ACCESS_TOKEN_KEY);
    window.localStorage.removeItem(REFRESH_TOKEN_KEY);
    window.localStorage.removeItem(AUTH_USER_KEY);
    window.dispatchEvent(new CustomEvent(AUTH_CHANGED_EVENT));
  },
};

export const getRequestErrorMessage = (error, fallback = "请求失败，请稍后再试") =>
  error?.response?.data?.message ||
  error?.response?.data?.msg ||
  error?.message ||
  fallback;

let refreshSessionPromise = null;

const isAuthEndpoint = (url = "") =>
  ["/users/login", "/users/register", "/users/refresh"].some((path) =>
    String(url || "").includes(path)
  );

const parseErrorResponse = async (response) => {
  let message = `HTTP error! status: ${response.status}`;
  try {
    const raw = await response.text();
    if (raw) {
      const parsed = JSON.parse(raw);
      message = parsed?.message || parsed?.msg || message;
    }
  } catch {
    // 保留默认 HTTP 状态错误。
  }
  const error = new Error(message);
  error.response = { status: response.status };
  return error;
};

export const refreshAccessToken = async () => {
  if (refreshSessionPromise) {
    return refreshSessionPromise;
  }

  const session = authStorage.getSession();
  if (!session.refreshToken) {
    authStorage.clearSession();
    return Promise.reject(new Error("登录已过期，请重新登录"));
  }

  refreshSessionPromise = axios
    .post(
      `${USER_API_BASE_URL}/users/refresh`,
      { refreshToken: session.refreshToken },
      { timeout: 60000 }
    )
    .then((response) => {
      const data = response.data || {};
      if (!data.accessToken || !data.refreshToken) {
        throw new Error("刷新登录态失败");
      }
      authStorage.setSession({
        accessToken: data.accessToken,
        refreshToken: data.refreshToken,
        username: session.username,
      });
      return data.accessToken;
    })
    .catch((error) => {
      authStorage.clearSession();
      throw error;
    })
    .finally(() => {
      refreshSessionPromise = null;
    });

  return refreshSessionPromise;
};

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
    async (error) => {
      const originalConfig = error?.config || {};
      if (
        error?.response?.status === 401 &&
        !originalConfig.__authRetry &&
        !isAuthEndpoint(originalConfig.url)
      ) {
        try {
          const accessToken = await refreshAccessToken();
          originalConfig.__authRetry = true;
          originalConfig.headers = originalConfig.headers || {};
          originalConfig.headers.Authorization = `Bearer ${accessToken}`;
          return instance.request(originalConfig);
        } catch {
          // 继续走统一错误转换，由页面决定是否跳转登录。
        }
      }

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
  let abortController = null;
  let messageCallback = null;
  let errorCallback = null;
  let closed = false;

  const fetchSSE = async () => {
    try {
      let response = null;
      for (let attempt = 0; attempt < 2; attempt += 1) {
        abortController = new AbortController();
        const accessToken = readStorage(ACCESS_TOKEN_KEY);
        response = await fetch(`${baseURL}${endpoint.url}`, {
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

        if (response.status === 401 && attempt === 0 && !isAuthEndpoint(endpoint.url)) {
          await refreshAccessToken();
          continue;
        }
        break;
      }

      if (!response?.ok || !response.body) {
        throw await parseErrorResponse(response);
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";
      let eventName = "message";

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
          if (line.startsWith("event: ")) {
            eventName = line.slice(7).trim() || "message";
            continue;
          }
          if (line.startsWith("data: ")) {
            const data = line.slice(6).replace(/\r$/, "");
            const dataTrimmed = data.trim();
            if (dataTrimmed === "[DONE]") {
              messageCallback?.("[DONE]", eventName);
              return;
            }
            if (data) {
              messageCallback?.(data, eventName);
            }
            eventName = "message";
          }
        }
      }
    } catch (error) {
      if (closed || error?.name === "AbortError") {
        return;
      }
      errorCallback?.(error);
    }
  };

  fetchSSE();

  const stream = {
    close() {
      closed = true;
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
