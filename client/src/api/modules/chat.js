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
  //
  // 当前后端接口对齐状态（2026-05-12 通过 curl probe 实跑 :8123 探活，参见 docs/work-logs/）：
  //
  //   ✅ 已实现：
  //     - GET    /api/ai/knowledge/documents              → 200
  //     - GET    /api/ai/knowledge/documents/:id/chunks   → 200
  //     - POST   /api/ai/knowledge/upload                 → 401（需登录 Bearer token；存在）
  //     - POST   /api/ai/knowledge/test-query             → 400（路由存在；缺 body）
  //
  //   ❌ 当前后端尚未实现（仓库 routes.go 已注册，但实际运行的旧版本未包含）：
  //     - GET    /api/ai/knowledge/folders                → 404
  //     - POST   /api/ai/knowledge/folders                → 404
  //     - PATCH  /api/ai/knowledge/folders/:id            → 404
  //     - DELETE /api/ai/knowledge/folders/:id            → 404
  //     - PATCH  /api/ai/knowledge/documents/:id/folder   → 404
  //
  // 后端补齐之前，前端的 KnowledgeSidebar 目录树 / 目录 CRUD / 「移动到目录」select 都会 fallback：
  //   - knowledgeFolders 失败 → 视为 0 文件夹，sidebar 只显示 visibility 二分类
  //   - knowledgeMoveDocumentFolder 失败 → 抛错给 UI 层 alert
  //   - knowledgeCreateFolder/UpdateFolder/DeleteFolder 失败 → 同样 UI alert
  //
  // 后续待办：后端补齐 5 个 folder 接口后，请同步 review 此处注释 + 字段对齐
  // （KnowledgeFolderItem.path/depth/documentCount/sortOrder/children；KnowledgeFolderDeleteResp.deleted）。

  // 知识库文档列表：匿名只读公共知识，登录后包含当前用户私有
  // ✅ 后端已实现
  knowledgeDocuments(params = {}) {
    return {
      service: "chat",
      method: "get",
      url: "/ai/knowledge/documents",
      params,
    };
  },

  // ❌ TODO(backend-align): 当前 :8123 后端 404，前端 fallback 为 0 文件夹。
  //    期望响应：{ folders: KnowledgeFolderItem[]（树形 children 嵌套）, unfiledCount, totalCount, initialized, meta }
  knowledgeFolders() {
    return {
      service: "chat",
      method: "get",
      url: "/ai/knowledge/folders",
    };
  },

  // ❌ TODO(backend-align): 当前 :8123 后端 404。
  //    期望请求：{ name, parentId? } | 期望响应：{ folder: KnowledgeFolderItem, meta }
  knowledgeCreateFolder(payload) {
    return {
      service: "chat",
      method: "post",
      url: "/ai/knowledge/folders",
      data: payload,
    };
  },

  // ❌ TODO(backend-align): 当前 :8123 后端 404。
  //    期望请求：{ name?, parentId?, sortOrder? } | 期望响应：{ folder: KnowledgeFolderItem, meta }
  knowledgeUpdateFolder(id, payload) {
    return {
      service: "chat",
      method: "patch",
      url: `/ai/knowledge/folders/${encodeURIComponent(id)}`,
      data: payload,
    };
  },

  // ❌ TODO(backend-align): 当前 :8123 后端 404。
  //    新策略（refactor 9c38333）：仅允许删除空目录，非空返回 409 ErrKnowledgeFolderNotEmpty。
  //    期望响应：{ deleted: bool, meta }
  knowledgeDeleteFolder(id) {
    return {
      service: "chat",
      method: "delete",
      url: `/ai/knowledge/folders/${encodeURIComponent(id)}`,
    };
  },

  // 单个文档的分块预览
  // ✅ 后端已实现（reader 模式 limit=500 拉全量；list 模式 limit=6 拉预览）
  knowledgeDocumentChunks(id, params = {}) {
    return {
      service: "chat",
      method: "get",
      url: `/ai/knowledge/documents/${encodeURIComponent(id)}/chunks`,
      params,
    };
  },

  // 知识库 PDF 上传（multipart/form-data，需登录 Bearer token；普通用户写入私人知识）
  // ✅ 后端已实现（依赖 MCP :8080 做 PDF 解析；MCP 未启动时上传会失败）
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

  // ❌ TODO(backend-align): 当前 :8123 后端 404。
  //    期望请求：{ folderId: number | 0=未归类 } | 期望响应：{ document: KnowledgeDocumentItem, meta }
  //    用于卡片右上 select「移动到目录」交互；对应 svc.MoveKnowledgeDocumentFolder。
  knowledgeMoveDocumentFolder(id, payload) {
    return {
      service: "chat",
      method: "patch",
      url: `/ai/knowledge/documents/${encodeURIComponent(id)}/folder`,
      data: payload,
    };
  },

  // 测试召回（管理页验证 TopK 检索）
  // ✅ 后端已实现（路由存在；缺 body 时 400）
  knowledgeTestQuery(payload) {
    return {
      service: "chat",
      method: "post",
      url: "/ai/knowledge/test-query",
      data: payload,
    };
  },
};
