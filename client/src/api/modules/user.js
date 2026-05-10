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
  createSession(payload = {}) {
    return {
      service: "user",
      method: "post",
      url: "/users/sessions",
      data: payload,
    };
  },
  sessionBootstrap(id) {
    return {
      service: "user",
      method: "get",
      url: `/users/sessions/${encodeURIComponent(id)}/bootstrap`,
    };
  },
  sessionFlowState(id) {
    return {
      service: "user",
      method: "get",
      url: `/users/sessions/${encodeURIComponent(id)}/flow-state`,
    };
  },
  finishSession(id) {
    return {
      service: "user",
      method: "post",
      url: `/users/sessions/${encodeURIComponent(id)}/finish`,
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

  // ============ Workbench / Interview / Resume Artifacts ============
  // 工作台首屏聚合数据（Workbench.vue 主页一次拿全部）
  workbenchBootstrap() {
    return {
      service: "user",
      method: "get",
      url: "/users/workbench/bootstrap",
    };
  },

  // 新建面试配置页：方向 / 难度 / focus / 面试官风格 / 默认配置
  interviewPresets() {
    return {
      service: "user",
      method: "get",
      url: "/users/interview/presets",
    };
  },

  // 预览新建面试的计划题目；也供题库页按方向+难度筛选拉题
  // params: { directionKey?, difficulty?, focusKeys?, interviewerStyle?, limit? }
  interviewPlanPreview(params = {}) {
    return {
      service: "user",
      method: "get",
      url: "/users/interview/plan/preview",
      params,
    };
  },

  // 结构化面试题库列表；前端题库页优先消费该接口，静态 JSON 仅作离线降级。
  interviewQuestions(params = {}) {
    return {
      service: "user",
      method: "get",
      url: "/users/interview/questions",
      params,
    };
  },

  interviewQuestionDetail(id) {
    return {
      service: "user",
      method: "get",
      url: `/users/interview/questions/${encodeURIComponent(id)}`,
    };
  },

  interviewQuestionStats() {
    return {
      service: "user",
      method: "get",
      url: "/users/interview/question-stats",
    };
  },

  // 当前用户简历列表（按绑定会话聚合）
  resumeArtifacts() {
    return {
      service: "user",
      method: "get",
      url: "/users/resume/artifacts",
    };
  },

  // 单份简历详情和分块预览
  resumeArtifactDetail(id) {
    return {
      service: "user",
      method: "get",
      url: `/users/resume/artifacts/${encodeURIComponent(id)}`,
    };
  },

  // 单份简历的 AI 分析（技能/亮点/风险/建议追问题）
  resumeArtifactAnalysis(id, params = {}) {
    return {
      service: "user",
      method: "get",
      url: `/users/resume/artifacts/${encodeURIComponent(id)}/analysis`,
      params,
    };
  },

  // ============ Report Center（报告中心，对接 /api/users/report-center/*） ============
  // 后端 5 个端点都已交付（详见 api/user/user.api 的 service user-api 第 2 段）。
  // 报告中心完整 SFC 由独立需求阶段实现；当前 WorkbenchReports.vue 占位 SFC 也消费 bootstrap，
  // 让占位页对接真实接口契约，未来升级到完整 SFC 时数据来源不需要换。

  // 工作台单入口：一次拿到 overview + modes 列表 + 当前模式详情，首屏少抖动
  reportCenterBootstrap(params = {}) {
    return {
      service: "user",
      method: "get",
      url: "/users/report-center/bootstrap",
      params,
    };
  },

  // 跨会话能力概览（总分 / 维度强弱 / 完成数 / 平均分）
  reportCenterOverview() {
    return {
      service: "user",
      method: "get",
      url: "/users/report-center/overview",
    };
  },

  // 报告会话列表（可按 mode / modeKey / status / limit 过滤）
  reportCenterSessions(params = {}) {
    return {
      service: "user",
      method: "get",
      url: "/users/report-center/sessions",
      params,
    };
  },

  // 模式卡片列表（练习模式分布：方向 / 难度 / focus 维度）
  reportCenterModes() {
    return {
      service: "user",
      method: "get",
      url: "/users/report-center/modes",
    };
  },

  // 单个模式详情（卡片 + 该模式下报告列表 + 过滤器）
  reportCenterModeDetail(modeKey, params = {}) {
    return {
      service: "user",
      method: "get",
      url: `/users/report-center/modes/${encodeURIComponent(modeKey)}`,
      params,
    };
  },
};
