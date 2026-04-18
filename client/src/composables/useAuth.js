import { computed, ref } from "vue";
import { authStorage } from "../api/index.js";
import { useApi } from "./useApi";

const authState = ref(authStorage.getSession());
let initialized = false;

const syncAuthState = () => {
  authState.value = authStorage.getSession();
};

const applySession = ({ accessToken = "", refreshToken = "", username = "" }) => {
  authStorage.setSession({
    accessToken,
    refreshToken,
    username,
  });
  syncAuthState();
};

const clearSession = () => {
  authStorage.clearSession();
  syncAuthState();
};

export function useAuth() {
  const api = useApi();

  if (!initialized) {
    syncAuthState();
    initialized = true;
  }

  const isAuthenticated = computed(() => Boolean(authState.value.accessToken));
  const username = computed(() => authState.value.username || "");

  const register = async ({ username: nextUsername, password, confirmPassword }) => {
    return api.auth.register({
      username: nextUsername,
      password,
      confirmPassword,
    });
  };

  const login = async ({ username: nextUsername, password }) => {
    const response = await api.auth.login({ username: nextUsername, password });
    applySession({
      accessToken: response.accessToken || "",
      refreshToken: response.refreshToken || "",
      username: nextUsername,
    });
    return response;
  };

  const renewSession = async () => {
    if (!authState.value.refreshToken) {
      clearSession();
      return null;
    }

    const response = await api.auth.refresh({
      refreshToken: authState.value.refreshToken,
    });

    applySession({
      accessToken: response.accessToken || "",
      refreshToken: response.refreshToken || "",
      username: authState.value.username,
    });

    return response;
  };

  const logout = async () => {
    try {
      if (authState.value.accessToken) {
        await api.auth.logout();
      }
    } finally {
      clearSession();
    }
  };

  return {
    authState,
    username,
    isAuthenticated,
    setSession: applySession,
    applySession,
    clearSession,
    register,
    login,
    renewSession,
    logout,
  };
}
