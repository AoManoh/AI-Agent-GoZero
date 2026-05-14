import { createRouter, createWebHistory } from "vue-router";
import { authStorage } from "../api/core";

const routes = [
  {
    path: "/",
    name: "Home",
    component: () => import("../views/Home.vue"),
    meta: {
      title: "AI 面试官 - 技术面试新体验",
      description:
        "AI面试官提供沉浸式编程面试模拟，深度追问与实时反馈助力全面提升技术能力。",
    },
  },
  {
    path: "/chat",
    name: "Chat",
    component: () => import("../views/Chat.vue"),
    meta: {
      title: "AI 模拟面试",
      description: "与AI面试官展开深入对话，获得实时评估与建议。",
      requiresAuth: true,
    },
  },
  {
    path: "/login",
    name: "Login",
    component: () => import("../views/Login.vue"),
    meta: {
      title: "登录 AI 面试官",
      description: "登录以继续你的模拟面试练习。",
    },
  },
  {
    path: "/register",
    name: "Register",
    component: () => import("../views/Register.vue"),
    meta: {
      title: "注册 AI 面试官账号",
      description: "创建账户，开启沉浸式技术面试体验。",
    },
  },

  // ============ Workbench routes ============
  // 工作台主页 + 4 个 section 子页。所有 workbench 路由都需要登录。
  // 未登录访问会被 router.beforeEach 拦截到 /login?redirect=<原 fullPath>。
  {
    path: "/workbench",
    name: "Workbench",
    component: () => import("../views/Workbench.vue"),
    meta: {
      title: "工作台 · AI 面试官",
      description: "查看练习进度、最近面试与能力雷达，快速进入下一场面试。",
      requiresAuth: true,
    },
  },
  {
    path: "/workbench/new",
    name: "WorkbenchNew",
    component: () => import("../views/WorkbenchNew.vue"),
    meta: {
      title: "新建面试 · AI 面试官",
      description: "选择岗位、难度和重点方向，配置一场专属面试。",
      requiresAuth: true,
    },
  },
  {
    path: "/workbench/resume",
    name: "WorkbenchResume",
    component: () => import("../views/WorkbenchResume.vue"),
    meta: {
      title: "简历管理 · AI 面试官",
      description: "上传与管理简历，让 AI 基于项目内容做深度追问。",
      requiresAuth: true,
    },
  },
  // 简历完整详情钻深页（设计图 241 / 261）。
  // 路由策略：独立 SFC + 独立路由（D-Q3 决策），保留浏览器前进后退能力，
  // 与主面板列表 ?artifact=:id query 命名分离，避免冲突。
  // 入口：/workbench/resume 主面板右栏 [看完整详情 →]，使用 router.push 保留历史（D-U3）。
  // 详见 docs/requirements/2026-05-12-workbench-resume-redesign.md §6.3 + §7.3。
  {
    path: "/workbench/resume/:id",
    name: "WorkbenchResumeDetail",
    component: () => import("../views/WorkbenchResumeDetail.vue"),
    meta: {
      title: "简历完整详情 · AI 面试官",
      description: "查看完整 AI 简历画像：18+ 原文 chunks、5 大评估面板与详细追问列表。",
      requiresAuth: true,
    },
  },
  {
    path: "/workbench/bank",
    name: "WorkbenchBank",
    component: () => import("../views/WorkbenchBank.vue"),
    meta: {
      title: "题库浏览 · AI 面试官",
      description: "按方向、难度、能力筛选真实面试题，看 AI 如何追问。",
      requiresAuth: true,
    },
  },
  {
    path: "/workbench/knowledge",
    name: "WorkbenchKnowledge",
    component: () => import("../views/WorkbenchKnowledge.vue"),
    meta: {
      title: "知识库 · AI 面试官",
      description: "上传文档，AI 会基于检索增强 (RAG) 在面试时引用你提供的资料。",
      requiresAuth: true,
    },
  },
  // 报告中心：v1 聚合页，入口由工作台报告卡承接，顶部导航暂不单列。
  {
    path: "/workbench/reports",
    name: "WorkbenchReports",
    component: () => import("../views/WorkbenchReports.vue"),
    meta: {
      title: "报告中心 · AI 面试官",
      description: "查看跨会话能力分析、单场报告与练习模式表现。",
      requiresAuth: true,
    },
  },

  // 兜底重定向：未匹配路由回到首页
  {
    path: "/:pathMatch(.*)*",
    redirect: "/",
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// === Auth guard ===
// 设计：仅检查 localStorage 里是否有 accessToken；不发起 /users/profile 校验
// 因为：(1) 路由切换需要瞬时响应，不能等 HTTP；
//      (2) 真实失效会被业务接口的 401 拦截器统一处理（后续接入）。
// 未登录时携带 redirect query 跳到 /login，便于 Login 完成后回流。
router.beforeEach((to, from, next) => {
  if (!to.meta?.requiresAuth) {
    next();
    return;
  }
  const session = authStorage.getSession();
  if (session?.accessToken) {
    next();
    return;
  }
  next({
    path: "/login",
    query: { redirect: to.fullPath },
  });
});

router.afterEach((to) => {
  if (to.meta?.title) {
    document.title = to.meta.title;
  }
});

export default router;
