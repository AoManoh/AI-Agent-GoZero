<script setup>
import { RouterView } from "vue-router";
import CosmicCanvas from "./components/CosmicCanvas.vue";
import { useTheme } from "./composables/useTheme";

useTheme();
</script>

<template>
  <div class="app-root">
    <CosmicCanvas />
    <RouterView />
  </div>
</template>

<style scoped>
.app-root {
  min-height: 100vh;
  width: 100vw;
  position: relative;
  display: block; /* Removed flex-column which breaks 100vh layout when nesting .workbench */
}
</style>
