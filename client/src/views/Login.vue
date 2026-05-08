<template>
  <div class="page auth-page">
    <SiteHeader>
      <template #actions>
        <ThemeToggle :theme="theme" @toggle="toggleTheme" />
      </template>
    </SiteHeader>

    <div class="auth-wrapper">
      <div class="auth-card">
        <h2 class="auth-title">登录</h2>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="login-username">用户名</label>
            <input
              id="login-username"
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
            <label for="login-password">密码</label>
            <input
              id="login-password"
              v-model="password"
              type="password"
              class="form-input"
              autocomplete="current-password"
              minlength="6"
              maxlength="50"
              required
              :disabled="loading"
            />
          </div>
          <button type="submit" class="cta-button auth-button" :disabled="loading">
            {{ loading ? "登录中…" : "登录" }}
          </button>
        </form>
        <p class="switch-auth">
          还没有账户？
          <router-link to="/register" class="switch-auth-link">立即注册</router-link>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import SiteHeader from "../components/SiteHeader.vue";
import ThemeToggle from "../components/ThemeToggle.vue";
import { useTheme } from "../composables/useTheme";
import { useAuth } from "../composables/useAuth";

const router = useRouter();
const { theme, toggleTheme } = useTheme();
const { login } = useAuth();

const username = ref("");
const password = ref("");
const loading = ref(false);

async function handleSubmit() {
  if (loading.value) return;
  if (!username.value || !password.value) {
    ElMessage.warning("请输入用户名和密码");
    return;
  }

  loading.value = true;
  try {
    await login({ username: username.value, password: password.value });
    ElMessage.success("登录成功");
    router.push({ name: "Chat" });
  } catch (error) {
    ElMessage.error(error?.message || "登录失败，请稍后再试");
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
  border-radius: var(--radius-lg);
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
  border-radius: var(--radius-sm);
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
