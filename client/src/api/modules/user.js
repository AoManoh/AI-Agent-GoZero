export const userEndpoints = {
  profile() {
    return {
      service: "user",
      method: "get",
      url: "/users/profile",
    };
  },
  sessions() {
    return {
      service: "user",
      method: "get",
      url: "/users/sessions",
    };
  },
  sessionDetail(id) {
    return {
      service: "user",
      method: "get",
      url: `/users/sessions/${encodeURIComponent(id)}`,
    };
  },
  demoInterviewSceneRandom(params = {}) {
    return {
      service: "user",
      method: "get",
      url: "/users/demo/interview-scenes/random",
      params,
      timeout: 3000,
    };
  },
  resumeUpload(formData) {
    return {
      service: "user",
      method: "post",
      url: "/users/resume/upload",
      data: formData,
      headers: {
        "Content-Type": "multipart/form-data",
      },
      timeout: 120000,
    };
  },
};
