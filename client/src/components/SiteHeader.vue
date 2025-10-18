<template>
  <header class="site-header">
    <slot name="left">
      <router-link class="logo-link" to="/">
        <img :src="logo" alt="AI 面试官 Logo" class="logo" />
      </router-link>
    </slot>

    <slot name="center"></slot>

    <nav class="main-nav">
      <slot name="actions"></slot>
    </nav>
  </header>
</template>

<script setup>
defineProps({
  logo: {
    type: String,
    default: "http://pic.aomanoh.com/note/20250903035824302.jpg",
  },
});
</script>

<style scoped>
.site-header {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.logo-link {
  cursor: pointer;
  display: inline-flex;
}

.logo {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 0 10px rgba(255, 255, 255, 0.2);
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
}

body.light-mode .logo {
  border-color: rgba(0, 0, 0, 0.1);
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
}

.main-nav {
  display: flex;
  align-items: center;
  gap: 16px;
}
</style>
