<template>
  <div class="page auth-page">
    <div class="auth-aurora" aria-hidden="true">
      <div class="auth-blob auth-blob-a"></div>
      <div class="auth-blob auth-blob-b"></div>
    </div>
    <SiteHeader>
      <template #actions>
        <ThemeToggle :theme="theme" @toggle="toggleTheme" />
      </template>
    </SiteHeader>

    <div class="auth-wrapper">
      <div class="auth-card">
        <!-- 叙事动画：左上接触点（16,2）出发，两条独立路径同时延伸到右下（384,468）汇合。
             ring-tr: 顺顶边 → 右边 路径
             ring-lb: 顺左边 → 底边 路径
             同起点 同终点 同时延伸 → 在右下汇合 → 整体消散。 -->
        <svg class="auth-card-ring" viewBox="0 0 400 470" preserveAspectRatio="none" aria-hidden="true">
          <!-- 两条 path 都从卡片左上圆角中心 (8,8) 出发，先 L 到圆角末端，再沿两条边延伸到右下。
               stroke-width 2 让流光更细。 -->
          <path class="ring-tr"
                d="M 8 8 L 16 2 H 384 A 14 14 0 0 1 398 16 V 454 A 14 14 0 0 1 384 468"
                fill="none" stroke="rgba(220,155,90,1)" stroke-width="2"
                vector-effect="non-scaling-stroke"
                pathLength="100" />
          <path class="ring-lb"
                d="M 8 8 L 2 16 V 454 A 14 14 0 0 0 16 468 H 384"
                fill="none" stroke="rgba(220,155,90,1)" stroke-width="2"
                vector-effect="non-scaling-stroke"
                pathLength="100" />
        </svg>
        <h2 class="auth-title">创建账户</h2>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="register-username">用户名</label>
            <input
              id="register-username"
              v-model="username"
              type="text"
              class="form-input"
              autocomplete="username"
              minlength="6"
              maxlength="50"
              required
              :disabled="loading"
            />
          </div>
          <div class="form-group">
            <label for="register-password">密码</label>
            <input
              id="register-password"
              v-model="password"
              type="password"
              class="form-input"
              autocomplete="new-password"
              minlength="6"
              maxlength="50"
              required
              :disabled="loading"
            />
          </div>
          <div class="form-group">
            <label for="register-confirm">确认密码</label>
            <input
              id="register-confirm"
              v-model="confirmPassword"
              type="password"
              class="form-input"
              autocomplete="new-password"
              minlength="6"
              maxlength="50"
              required
              :disabled="loading"
            />
            <p class="error-message" v-if="showError">两次密码不一致</p>
          </div>
          <button type="submit" class="auth-button" :disabled="loading">
            {{ loading ? "注册中…" : "注册" }}
          </button>
        </form>
        <p class="switch-auth">
          已有账户？
          <router-link to="/login" class="switch-auth-link">立即登录</router-link>
        </p>
      </div>
    </div>

    <AppFooter />
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import SiteHeader from "../components/SiteHeader.vue";
import ThemeToggle from "../components/ThemeToggle.vue";
import AppFooter from "../components/AppFooter.vue";
import { useTheme } from "../composables/useTheme";
import { useAuth } from "../composables/useAuth";

const router = useRouter();
const { theme, toggleTheme } = useTheme();
const { register } = useAuth();

const username = ref("");
const password = ref("");
const confirmPassword = ref("");
const loading = ref(false);

const showError = computed(
  () =>
    Boolean(password.value) &&
    Boolean(confirmPassword.value) &&
    password.value !== confirmPassword.value
);

async function handleSubmit() {
  if (loading.value) return;
  if (showError.value) return;
  if (!username.value || !password.value) {
    ElMessage.warning("请输入用户名和密码");
    return;
  }

  loading.value = true;
  try {
    await register({
      username: username.value,
      password: password.value,
      confirmPassword: confirmPassword.value,
    });
    ElMessage.success("注册成功，请登录");
    router.push({ name: "Login" });
  } catch (error) {
    ElMessage.error(error?.message || "注册失败，请稍后再试");
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.auth-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  /* 修复 aurora z-index: -1 被 body 黑色实心 bg 吞掉的问题：
     让 .auth-page 创建局部 stacking context，这样 aurora 在 .auth-page 内部
     的 -1 层 仍在 .auth-page 自身 0 层 + body bg 之上。 */
  position: relative;
  isolation: isolate;
}

.auth-wrapper {
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 5vh 20px;
  flex-grow: 1;
}

.auth-card {
  width: 100%;
  max-width: 400px;
  /* 与 Home .step-card 同款双层渐变面板（padding-box 哑光金属底 + border-box
     高光描边），加三层 inset/外阴影构成亚克力质感。 */
  background:
    linear-gradient(180deg,
      rgba(18, 19, 24, 1) 0%,
      rgba(11, 12, 16, 1) 50%,
      rgba(7, 8, 11, 1) 100%
    ) padding-box,
    linear-gradient(160deg,
      rgba(255, 255, 255, 0.14) 0%,
      rgba(255, 255, 255, 0.03) 30%,
      rgba(255, 255, 255, 0.02) 70%,
      rgba(255, 255, 255, 0.07) 100%
    ) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  /* 克制档：外发光 alpha 0.40→0.25，warm-pulse 起伏 0.18→0.32 更柔和不抢戏 form。 */
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.05),
    inset 0 50px 60px -50px rgba(255, 255, 255, 0.025),
    0 16px 40px rgba(0, 0, 0, 0.35),
    0 0 100px -30px rgba(220, 155, 90, 0.25);
  backdrop-filter: blur(10px);
  padding: 40px;
  animation: auth-card-warm-pulse 4s ease-in-out infinite;
}

/* 暖琥珀呼吸脉冲（4s），克制起伏 alpha 0.18→0.32 + spread 轻微。 */
@keyframes auth-card-warm-pulse {
  0%, 100% {
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.05),
      inset 0 50px 60px -50px rgba(255, 255, 255, 0.025),
      0 16px 40px rgba(0, 0, 0, 0.35),
      0 0 90px -30px rgba(220, 155, 90, 0.18);
  }
  50% {
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.05),
      inset 0 50px 60px -50px rgba(255, 255, 255, 0.025),
      0 16px 40px rgba(0, 0, 0, 0.35),
      0 0 120px -25px rgba(220, 155, 90, 0.32);
  }
}

/* 双向扩散闭环叙事动画 SVG 宽度：与 Login.vue 同一设计。
   - drop-shadow 让 stroke 自带柔和光晕，取代 blur（避免与 box-shadow 外发光混色）。 */
.auth-card-ring {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  filter: drop-shadow(0 0 6px rgba(220, 155, 90, 0.7));
  overflow: visible;
}

/* ring-tr 和 ring-lb 同起点、同终点、同动画节奏。 */
.ring-tr,
.ring-lb {
  stroke-dasharray: 0 100;
  stroke-linecap: round;
  animation: ring-edge-sweep 6s ease-in-out infinite;
}

/* 叙事动画（两条 path 同步延伸）：
   - 0–10%: opacity 0 → 1，起点闪现
   - 10–55%: stroke-dasharray 0 → 100，两条 path 同时延伸到右下汇合
   - 55–75%: 汇合完整保持 opacity 1
   - 75–95%: opacity 1 → 0，整体消散
   - 95–100%: 重置、准备下一轮 */
@keyframes ring-edge-sweep {
  0%   { stroke-dasharray: 0 100;  opacity: 0; }
  10%  { stroke-dasharray: 0 100;  opacity: 1; }
  55%  { stroke-dasharray: 100 0; opacity: 1; }
  75%  { stroke-dasharray: 100 0; opacity: 1; }
  95%  { stroke-dasharray: 100 0; opacity: 0; }
  100% { stroke-dasharray: 0 100;  opacity: 0; }
}

.auth-title {
  text-align: center;
  font-family: var(--display);
  font-size: 2rem;
  font-weight: 600;
  letter-spacing: -0.01em;
  margin-bottom: 30px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-weight: 500;
  margin-bottom: 8px;
}

.form-input {
  width: 100%;
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  /* HTML input 默认不继承父字体，显式声明与 .auth-button / 全局 body 同 sans 字族，
     避免 fallback 到 Arial 与 .auth-title (var(--display)) 形成字族断裂 */
  font: 1rem var(--sans);
  transition: border-color 0.2s, box-shadow 0.2s;
}

:global(body.light-mode) .form-input {
  background: rgba(255, 255, 255, 0.5);
}

.form-input:focus {
  outline: none;
  /* Home .step-num 同款暖琥珀，与 .auth-blob-a 形成色彩呼应 */
  border-color: rgba(220, 155, 90, 0.9);
  box-shadow: 0 0 12px rgba(220, 155, 90, 0.35);
}

.error-message {
  color: var(--color-error);
  font-size: 0.9rem;
  margin-top: 8px;
}

.auth-button {
  /* 自洽样式：白底黑字 + 大尺寸（auth 主动作权重高于 Home btn-ns）。
     与 Home .btn-ns 同语言（var(--t) 白底 + var(--bg) 黑字 + opacity hover），
     但 padding/font-size 保持原 cta-button 的大版本，不依赖外部全局类。 */
  display: inline-block;
  width: 100%;
  padding: 14px;
  border: none;
  border-radius: var(--radius-md);
  background: var(--t);
  color: var(--bg, #020204);
  font: 600 1.1rem var(--sans);
  cursor: pointer;
  margin-top: 10px;
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.auth-button:hover {
  opacity: 0.88;
}

.auth-button:active {
  transform: translateY(1px);
}

.auth-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.switch-auth {
  text-align: center;
  margin-top: 20px;
  color: var(--color-text-secondary);
}

.switch-auth-link {
  /* 次级暖白：不抢 form 主操作的注意力，hover 才有强反馈 */
  color: var(--t2);
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s ease;
}

.switch-auth-link:hover {
  color: var(--t);
  text-decoration: underline;
}

/* 极光装饰层（A+ 档）：与 Home aurora 同语言但为 auth 页专门调明。
   关键参数反直觉点：blur 不是越大越柔和，过大 blur 会把 alpha “稀释”到边缘、
   失去光团轮廓感。本档采用 blur 70px + alpha 0.5-0.6 让暖琥珀 / 冷月白两团
   在 1920 大屏上有明显“光团中心”而非「均匀薄雾」。drift 8vw/6vh 位移住
   在 40-50s 周期内可见运动。卡片额外加一层暖琥珀 outer glow，让它看起来
   「浮在暖光之上」（此点与 Home .step-card 重要差异：step-card 在拥挤的步骤
   区不需要漂浮感，auth-card 在空旷背景上需要）。 */
.auth-aurora {
  position: fixed;
  inset: 0;
  z-index: -1;
  pointer-events: none;
  overflow: hidden;
}

.auth-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(70px);
  will-change: transform;
}

/* 左上：暖琥珀（与 .form-input:focus 暖色 + Home step-num 同色系）
   超克制档：左上仕然「一大坠」增强割裂、中心退出视口更多 + alpha 再降 + blur 加大让光「漫」。 */
.auth-blob-a {
  top: -10vw;
  left: -10vw;
  width: 36vw;
  height: 36vw;
  max-width: 580px;
  max-height: 580px;
  background: rgba(220, 155, 90, 0.40);
  filter: blur(100px);
  animation:
    auth-aurora-drift-a 40s ease-in-out infinite,
    auth-blob-pulse 4s ease-in-out infinite;
}

/* 右下：冷月白（与 Home aurora-a 同色系，形成冷暖呼应）
   克制档： alpha 0.65→0.45 size 580→480 保持冷暖对冲但不轻重不平衡。 */
.auth-blob-b {
  right: -10vw;
  bottom: -10vw;
  width: 30vw;
  height: 30vw;
  max-width: 480px;
  max-height: 480px;
  background: rgba(225, 230, 245, 0.45);
  animation: auth-aurora-drift-b 50s ease-in-out infinite;
}

@keyframes auth-aurora-drift-a {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(8vw, 6vh); }
}

@keyframes auth-aurora-drift-b {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(-8vw, -6vh); }
}

/* B-soft 最终档：blob-a 与 .auth-card warm-pulse 同周期同相位（4s），
   0%/100% 弱、 50% 强；成「全页齐步呼吸」。 */
@keyframes auth-blob-pulse {
  0%, 100% { opacity: 0.55; }
  50%      { opacity: 1.00; }
}

/* 移动端隐藏极光，避免性能压力 + 让窄屏更聚焦 form 操作 */
@media (max-width: 768px) {
  .auth-aurora {
    display: none;
  }
}
</style>
