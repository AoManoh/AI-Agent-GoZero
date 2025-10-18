<script setup>
import { RouterView } from "vue-router";
import AuroraBackground from "./components/AuroraBackground.vue";
import { useTheme } from "./composables/useTheme";

useTheme();
</script>

<template>
  <div class="app-root">
    <AuroraBackground />
    <RouterView />
  </div>
</template>

<style scoped>
.app-root {
  min-height: 100vh;
  position: relative;
  display: flex;
  flex-direction: column;
}
</style>
