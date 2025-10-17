import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue'),
    meta: {
      title: ' 首页- AI面试官智能体应用平台',
      description: 'AI面试官提供专业的编程面试模拟，智能算法面试、系统设计练习，实时反馈助你快速提升，成功拿到心仪offer'
    }
  },
  {
    path: '/interview',
    name: 'Interview',
    component: () => import('../views/Interview.vue'),
    meta: {
      title: '面试 - 助你拿到心仪offer',
      description: 'AI面试官提供专业的编程面试模拟，智能算法面试、系统设计练习，实时反馈助你快速提升，成功拿到心仪offer'
    }
  },
  {
    path: '/interview-master',
    name: 'InterviewMaster',
    component: () => import('../views/InterviewMaster.vue'),
    meta: {
      title: '模拟面试官 - 助你拿到心仪offer',
      description: 'AI面试官提供专业的编程面试模拟，智能算法面试、系统设计练习，实时反馈助你快速提升，成功拿到心仪offer'
    }
  }
  // {
  //   path: '/super-agent',
  //   name: 'SuperAgent',
  //   component: () => import('../views/SuperAgent.vue'),
  //   meta: {
  //     title: 'AI超级智能体 - 大鱼AI超级智能体应用平台',
  //     description: 'AI超级智能体是大鱼AI超级智能体应用平台的全能助手，能解答各类专业问题，提供精准建议和解决方案'
  //   }
  // }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router