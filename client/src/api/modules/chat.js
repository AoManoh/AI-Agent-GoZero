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

  // ============ Knowledge Manager（WorkbenchKnowledge.vue 用） ============
  // 知识库文档列表：匿名只读公共知识，登录后包含当前用户私有
  knowledgeDocuments(params = {}) {
    return {
      service: "chat",
      method: "get",
      url: "/ai/knowledge/documents",
      params,
    };
  },

  knowledgeFolders() {
    return {
      service: "chat",
      method: "get",
      url: "/ai/knowledge/folders",
    };
  },

  knowledgeCreateFolder(payload) {
    return {
      service: "chat",
      method: "post",
      url: "/ai/knowledge/folders",
      data: payload,
    };
  },

  knowledgeUpdateFolder(id, payload) {
    return {
      service: "chat",
      method: "patch",
      url: `/ai/knowledge/folders/${encodeURIComponent(id)}`,
      data: payload,
    };
  },

  knowledgeDeleteFolder(id) {
    return {
      service: "chat",
      method: "delete",
      url: `/ai/knowledge/folders/${encodeURIComponent(id)}`,
    };
  },

  // 单个文档的分块预览
  knowledgeDocumentChunks(id, params = {}) {
    return {
      service: "chat",
      method: "get",
      url: `/ai/knowledge/documents/${encodeURIComponent(id)}/chunks`,
      params,
    };
  },

  // 知识库 PDF 上传（multipart/form-data，需登录 Bearer token；普通用户写入私人知识）
  knowledgeUpload(formData) {
    return {
      service: "chat",
      method: "post",
      url: "/ai/knowledge/upload",
      data: formData,
      headers: {
        "Content-Type": "multipart/form-data",
      },
      timeout: 120000,
    };
  },

  knowledgeMoveDocumentFolder(id, payload) {
    return {
      service: "chat",
      method: "patch",
      url: `/ai/knowledge/documents/${encodeURIComponent(id)}/folder`,
      data: payload,
    };
  },

  // 测试召回（管理页验证 TopK 检索）
  knowledgeTestQuery(payload) {
    return {
      service: "chat",
      method: "post",
      url: "/ai/knowledge/test-query",
      data: payload,
    };
  },
};
