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
  // 当前后端接口对齐状态（2026-05-12 通过 curl probe 实跑 :8123 探活）：
  //
  //   ✅ 仓库代码已实现 + :8123 已加载：
  //     - GET    /api/ai/knowledge/documents              → 200
  //     - GET    /api/ai/knowledge/documents/:id/chunks   → 200
  //     - POST   /api/ai/knowledge/upload                 → 401（需登录 Bearer token）
  //     - POST   /api/ai/knowledge/test-query             → 400（路由存在；缺 body）
  //
  //   ⚠️  仓库代码已实现，但 :8123 进程是 v0.2 之前的旧 build，**重启 chat 后端即生效**：
  //     - GET    /api/ai/knowledge/folders                → 当前 404
  //     - POST   /api/ai/knowledge/folders                → 当前 404
  //     - PATCH  /api/ai/knowledge/folders/:id            → 当前 404
  //     - DELETE /api/ai/knowledge/folders/:id            → 当前 404
  //     - PATCH  /api/ai/knowledge/documents/:id/folder   → 当前 404
  //
  //   后端事实源（不要把前端备注当事实源，请以 chat.api 为准）：
  //     - 路由契约：api/chat/chat.api（type 定义 + 路由声明）
  //     - 路由表：  api/chat/internal/handler/routes.go（goctl 生成，自动同步）
  //     - handler： api/chat/internal/handler/knowledge_manager.go
  //     - logic：   api/chat/internal/logic/knowledge_*.go
  //     - 数据库：  db/user.sql 的 knowledge_folders / knowledge_base.folder_id
  //     - 开发记录：docs/development/2026-05-12-knowledge-v02-backend-interfaces.md
  //
  //   重启后未对齐时的兜底（前端已 try/catch 优雅降级，不阻塞主流程）：
  //     - knowledgeFolders 失败 → 视为 0 文件夹，sidebar 只显示 visibility 二分类
  //     - knowledgeMoveDocumentFolder 失败 → 抛错给 UI 层 alert
  //     - knowledgeCreateFolder/UpdateFolder/DeleteFolder 失败 → 同样 UI alert

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

  // ⚠️  仓库代码已实现，当前 :8123 运行旧 build 返回 404；重启后端即生效。
  //    期望响应（KnowledgeFoldersResp）：{ folders: KnowledgeFolderItem[]（树形 children 嵌套）, unfiledCount, total, totalCount, initialized, meta }
  knowledgeFolders() {
    return {
      service: "chat",
      method: "get",
      url: "/ai/knowledge/folders",
    };
  },

  // ⚠️  仓库代码已实现，当前 :8123 运行旧 build 返回 404；重启后端即生效。
  //    期望请求（KnowledgeCreateFolderReq）：{ name, parentId?, sortOrder? }
  //    期望响应（KnowledgeFolderMutationResp）：{ folder: KnowledgeFolderItem, meta }
  knowledgeCreateFolder(payload) {
    return {
      service: "chat",
      method: "post",
      url: "/ai/knowledge/folders",
      data: payload,
    };
  },

  // ⚠️  仓库代码已实现，当前 :8123 运行旧 build 返回 404；重启后端即生效。
  //    期望请求（KnowledgeUpdateFolderReq）：{ name?, parentId?, sortOrder?, setParent?, setSortOrder? }
  //    期望响应（KnowledgeFolderMutationResp）：{ folder: KnowledgeFolderItem, meta }
  knowledgeUpdateFolder(id, payload) {
    return {
      service: "chat",
      method: "patch",
      url: `/ai/knowledge/folders/${encodeURIComponent(id)}`,
      data: payload,
    };
  },

  // ⚠️  仓库代码已实现，当前 :8123 运行旧 build 返回 404；重启后端即生效。
  //    策略（refactor 9c38333）：仅允许删除空目录，非空返回 409 ErrKnowledgeFolderNotEmpty。
  //    期望响应（KnowledgeFolderDeleteResp）：{ deleted: bool, meta }
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

  // ⚠️  仓库代码已实现，当前 :8123 运行旧 build 返回 404；重启后端即生效。
  //    期望请求（KnowledgeMoveDocumentFolderReq）：{ folderId: number | 0=未归类 }
  //    期望响应（KnowledgeDocumentMutationResp）：{ document: KnowledgeDocumentItem, meta }
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
