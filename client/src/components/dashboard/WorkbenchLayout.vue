<template>
  <!--
    WorkbenchLayout：5 个 /workbench 页面共享的视觉外壳。
    - 与 Home 同源的 aurora 背景层（冷月光 + 暖琥珀双 blob，mix-blend-mode: screen）
    - SiteHeader fixed 80px + 5 个 section nav + 主 CTA "+ 新建面试" + 头像
    - main 留 80px 顶部空间避让 fixed header；业务页自己控制首屏节奏
    - AppFooter 复用，与 SiteHeader 镜像（80px / max-width 1320 / 1px 顶分隔）
    使用方式：
      <WorkbenchLayout>
        <YourSectionContent />
      </WorkbenchLayout>
    section 高亮由 useRoute().path 自动判断；不需要 props 传 current。
  -->
  <div class="workbench-page">
    <div class="workbench-aurora" aria-hidden="true">
      <div class="wb-blob wb-blob-a"></div>
      <div class="wb-blob wb-blob-b"></div>
    </div>

    <SiteHeader :fixed="true">
      <template #actions>
        <nav class="wb-nav">
          <router-link
            to="/workbench"
            class="wb-nav-link"
            :class="{ 'wb-nav-active': isHome }"
          >工作台</router-link>
          <router-link
            to="/workbench/new"
            class="wb-nav-link"
            :class="{ 'wb-nav-active': isNew }"
          >面试</router-link>
          <router-link
            to="/workbench/bank"
            class="wb-nav-link"
            :class="{ 'wb-nav-active': isBank }"
          >题库</router-link>
          <router-link
            to="/workbench/knowledge"
            class="wb-nav-link"
            :class="{ 'wb-nav-active': isKnowledge }"
          >知识库</router-link>
          <router-link
            to="/workbench/resume"
            class="wb-nav-link"
            :class="{ 'wb-nav-active': isResume }"
          >简历</router-link>
        </nav>
        <router-link to="/workbench/new" class="wb-cta">
          <span class="wb-cta-plus" aria-hidden="true">+</span>
          <span>新建面试</span>
        </router-link>
        <button
          type="button"
          class="wb-avatar"
          :title="username ? `${username} · 退出登录` : '退出登录'"
          @click="handleLogout"
        >
          {{ userInitial }}
        </button>
      </template>
    </SiteHeader>

    <main class="wb-main">
      <slot />
    </main>

    <AppFooter />
  </div>
</template>

<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import SiteHeader from "../SiteHeader.vue";
import AppFooter from "../AppFooter.vue";
import { useAuth } from "../../composables/useAuth";

const route = useRoute();
const router = useRouter();
const { username, logout } = useAuth();

// section 高亮：精确匹配 /workbench；前缀匹配子路由。
// 使用 startsWith 而非 includes 避免 /workbench/bank 误命中 /workbench/banking 类
// 未来路由（当前不存在但保留语义安全）。
const isHome = computed(() => route.path === "/workbench");
const isNew = computed(() => route.path.startsWith("/workbench/new"));
const isBank = computed(() => route.path.startsWith("/workbench/bank"));
const isKnowledge = computed(() => route.path.startsWith("/workbench/knowledge"));
const isResume = computed(() => route.path.startsWith("/workbench/resume"));

const userInitial = computed(() => {
  const name = username.value.trim();
  if (!name) return "U";
  // 中英混合：取首个非空字符。中文姓名取姓氏字符即可。
  return name.charAt(0).toUpperCase();
});

const handleLogout = async () => {
  await logout();
  router.push({ path: "/login" });
};
</script>

<style scoped>
/* 关键：position: relative + isolation: isolate 创建独立 stacking context，
   让 .workbench-aurora 的 z-index: -1 仅在 .workbench-page 内部可见，
   不会被 body 背景或其他 fixed 元素吞没。
   这条规则与 Login/Register 的 .auth-page stacking 修复同源。 */
.workbench-page {
  position: relative;
  isolation: isolate;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* aurora 背景层：与 Home.vue .aurora-bg 同构。
   - position: fixed + inset: 0 让 aurora 永远贴满视口（不随滚动滑出）
   - z-index: -1 在 .workbench-page 自身 stacking 内垫底，但仍处于
     CosmicCanvas (z-index: 0) 之上的合成栈中
   - pointer-events: none 不拦截任何鼠标交互
   - overflow: hidden 切住 -25vw / -15vw 大 blob 越界
   - mix-blend-mode: screen 让 blob 容器的"黑"被透明化，
     只把彩色叠加到星空和页面背景上 */
.workbench-aurora {
  position: fixed;
  inset: 0;
  z-index: -1;
  pointer-events: none;
  overflow: hidden;
  mix-blend-mode: screen;
}

.wb-blob {
  position: absolute;
  border-radius: 50%;
  /* 重度 blur 让 blob 边缘极柔，肉眼看不到圆形边界。 */
  filter: blur(90px);
  will-change: transform;
}

/* 左上：冷月光近白，alpha 压到 0.07。screen 模式下混合后微微发亮，
   不会让整体"发灰"。与 Home aurora-blob.aurora-a 同色同尺寸。 */
.wb-blob-a {
  top: -25vw;
  left: -15vw;
  width: 80vw;
  height: 80vw;
  background: rgba(225, 230, 245, 0.07);
  animation: wb-aurora-drift-a 80s ease-in-out infinite;
}

/* 右下：暖琥珀近白，alpha 压到 0.05。与 Home aurora-blob.aurora-b 同色同尺寸。
   两个 blob 永不同步漂移，振幅 4-5vw/vh，周期 80s/95s 错开。 */
.wb-blob-b {
  bottom: -20vw;
  right: -15vw;
  width: 70vw;
  height: 70vw;
  background: rgba(245, 230, 210, 0.05);
  animation: wb-aurora-drift-b 95s ease-in-out infinite;
}

@keyframes wb-aurora-drift-a {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(5vw, 4vh); }
}

@keyframes wb-aurora-drift-b {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(-4vw, -5vh); }
}

/* === Header 内 nav 链接 === */
/* gap 4 而非 16：设计图 nav 链接是紧凑排列，靠 padding + active 下划线区分。 */
.wb-nav {
  display: flex;
  align-items: center;
  gap: 4px;
}

.wb-nav-link {
  font: 500 var(--fs-md) var(--sans);
  color: var(--t3);
  text-decoration: none;
  padding: 8px 14px;
  border-radius: var(--radius-sm);
  transition: color .2s ease, background-color .2s ease;
  /* line-height 固定 1.2 避免 active 状态下下划线位置因行高跳变。 */
  line-height: 1.2;
}

.wb-nav-link:hover {
  color: var(--t);
  background: rgba(255, 255, 255, 0.04);
}

/* active：白色文字 + 下方 2px 暖琥珀短下划线。下划线两侧各内缩 14px
   等于 padding-left/right，让线刚好覆盖文字宽度而非整个 link 宽度。 */
.wb-nav-active {
  color: var(--t);
  position: relative;
}

.wb-nav-active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 14px;
  right: 14px;
  height: 2px;
  background: rgba(220, 155, 90, 0.9);
  border-radius: 2px;
}

/* === 主 CTA："+ 新建面试" === */
/* 紧凑版：font 14 / padding 9 18，与 nav-link 同高 ~36px，
   不抢镜但仍是白底 + 黑字的主操作召唤。
   margin-left 12 与 nav 之间留出视觉分隔。 */
.wb-cta {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font: 600 var(--fs-md) var(--sans);
  color: var(--bg);
  background: var(--t);
  border: none;
  cursor: pointer;
  padding: 9px 18px;
  border-radius: var(--radius-sm);
  text-decoration: none;
  margin-left: 12px;
  transition: opacity .2s ease, transform .2s ease, box-shadow .2s ease;
  /* 静态阴影更轻，hover 略加深，避免 Home btn-p 那种主页 hero 级别的呼吸感。
     这里是 header 内嵌按钮，需要克制。 */
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.wb-cta:hover {
  opacity: .94;
  transform: translateY(-1px);
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.3);
}

.wb-cta-plus {
  font-weight: 700;
  font-size: var(--fs-md);
  line-height: 1;
}

/* === Avatar === */
/* 36×36 圆 + 字母首字母。与 nav-link / wb-cta 三件套同高 ~36-38px，
   视觉上 header 右侧形成"链接组 / CTA / 用户"三段式节奏。 */
.wb-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: var(--t);
  font: 600 var(--fs-sm) var(--sans);
  cursor: pointer;
  margin-left: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: border-color .2s ease, background-color .2s ease;
  /* 字母强制大写 + 居中 */
  text-transform: uppercase;
  letter-spacing: 0;
}

/* hover 边框暖琥珀，与 nav-active 下划线同色形成视觉关联：
   "你当前所在 section + 你是谁"两个用户感知都用同一个 accent 标记。 */
.wb-avatar:hover {
  border-color: rgba(220, 155, 90, 0.6);
  background: rgba(255, 255, 255, 0.08);
}

/* === main === */
/* 80px 与 SiteHeader fixed 高度配合避让。
   flex: 1 让 main 撑满剩余视口高度，footer 自然贴底（min-height: 100vh
   情况下短内容也能让 footer 在视口底而非内容紧下方）。 */
.wb-main {
  flex: 1;
  margin-top: 80px;
  padding-top: 0;
  position: relative;
  /* z-index: 1 保证 main 内容层在 aurora (z-index: -1) 之上、
     ripple 一类的浮层之下。 */
  z-index: 1;
}

/* === 响应式 === */
@media (max-width: 1024px) {
  .wb-nav-link {
    padding: 6px 10px;
    font-size: var(--fs-sm);
  }
  .wb-cta {
    padding: 7px 14px;
    margin-left: 8px;
  }
  .wb-cta span:not(.wb-cta-plus) {
    /* 中等屏幕仅留 + 图标，省空间。 */
    display: none;
  }
}

@media (max-width: 768px) {
  .wb-nav {
    /* 移动端隐藏 section nav，靠 CTA + avatar 维持核心入口。
       后续可加汉堡菜单 panel；本轮不做以避免增加状态管理。 */
    display: none;
  }
  .wb-cta {
    margin-left: 0;
  }
  .wb-main {
    padding-top: 0;
  }
}
</style>
