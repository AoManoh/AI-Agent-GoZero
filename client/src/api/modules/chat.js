const cloneFormData = (source) => {
  const formData = new FormData();
  for (const [key, value] of source.entries()) {
    formData.append(key, value);
  }
  return formData;
};

const buildChatFormData = (payload, chatId) => {
  const formData = payload instanceof FormData ? cloneFormData(payload) : new FormData();

  if (!(payload instanceof FormData)) {
    if (typeof payload === "string") {
      formData.append("message", payload);
    } else if (payload && typeof payload === "object") {
      for (const [key, value] of Object.entries(payload)) {
        if (value !== undefined && value !== null && value !== "") {
          formData.append(key, value);
        }
      }
    }
  }

  if (chatId && !formData.has("chatId")) {
    formData.append("chatId", chatId);
  }

  return formData;
};

export const chatEndpoints = {
  interviewStream(payload, chatId) {
    const formData = buildChatFormData(payload, chatId);

    return {
      service: "chat",
      method: "post",
      url: "/ai/interview_app/chat/sse",
      data: formData,
      input: "form",
    };
  },
};
