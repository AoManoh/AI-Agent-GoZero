import { createRouter, createWebHistory } from "vue-router";

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
  {
    path: "/:pathMatch(.*)*",
    redirect: "/",
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.afterEach((to) => {
  if (to.meta?.title) {
    document.title = to.meta.title;
  }
});

export default router;