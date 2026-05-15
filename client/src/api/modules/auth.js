export const authEndpoints = {
  register(payload) {
    return {
      service: "user",
      method: "post",
      url: "/users/register",
      data: payload,
    };
  },
  login(payload) {
    return {
      service: "user",
      method: "post",
      url: "/users/login",
      data: payload,
    };
  },
  refresh(payload) {
    return {
      service: "user",
      method: "post",
      url: "/users/refresh",
      data: payload,
    };
  },
  logout() {
    return {
      service: "user",
      method: "post",
      url: "/users/logout",
      data: {},
    };
  },
};
