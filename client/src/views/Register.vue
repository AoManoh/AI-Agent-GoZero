<template>
  <div class="page auth-page">
    <SiteHeader>
      <template #actions>
        <ThemeToggle :theme="theme" @toggle="toggleTheme" />
      </template>
    </SiteHeader>

    <div class="auth-wrapper">
      <div class="auth-card">
        <h2 class="auth-title">创建账户</h2>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="register-email">邮箱</label>
            <input id="register-email" v-model="email" type="email" class="form-input" required />
          </div>
          <div class="form-group">
            <label for="register-password">密码</label>
            <input id="register-password" v-model="password" type="password" class="form-input" required />
          </div>
          <div class="form-group">
            <label for="register-confirm">确认密码</label>
            <input
              id="register-confirm"
              v-model="confirmPassword"
              type="password"
              class="form-input"
              required
            />
            <p class="error-message" v-if="showError">密码不匹配</p>
          </div>
          <button type="submit" class="cta-button auth-button">注册</button>
        </form>
        <p class="switch-auth">
          已有账户？
          <router-link to="/login" class="switch-auth-link">立即登录</router-link>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import SiteHeader from "../components/SiteHeader.vue";
import ThemeToggle from "../components/ThemeToggle.vue";
import { useTheme } from "../composables/useTheme";

const router = useRouter();
const { theme, toggleTheme } = useTheme();

const email = ref("");
const password = ref("");
const confirmPassword = ref("");

const showError = computed(() => password.value && confirmPassword.value && password.value !== confirmPassword.value);

function handleSubmit() {
  if (showError.value) return;
  router.push({ name: "Home" });
}
</script>

<style scoped>
.auth-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
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
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(10px);
  padding: 40px;
}

.auth-title {
  text-align: center;
  font-size: 2rem;
  font-weight: 600;
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
  border-radius: 6px;
  font-size: 1rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

:global(body.light-mode) .form-input {
  background: rgba(255, 255, 255, 0.5);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-glow-1);
  box-shadow: 0 0 10px rgba(59, 130, 246, 0.5);
}

.error-message {
  color: var(--color-error);
  font-size: 0.9rem;
  margin-top: 8px;
}

.auth-button {
  width: 100%;
  padding: 14px;
  font-size: 1.1rem;
  cursor: pointer;
  margin-top: 10px;
}

.switch-auth {
  text-align: center;
  margin-top: 20px;
  color: var(--color-text-secondary);
}

.switch-auth-link {
  color: var(--color-glow-1);
  text-decoration: none;
  font-weight: 500;
}

.switch-auth-link:hover {
  text-decoration: underline;
}
</style>
