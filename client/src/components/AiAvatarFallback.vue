<template>
  <div class="ai-avatar-fallback" :class="type">
    <span class="ai-avatar-glyph">{{ glyph }}</span>
  </div>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  type: {
    type: String,
    default: "default",
  },
});

const glyph = computed(() => {
  if (props.type === "interview") return "AI";
  if (props.type === "super") return "S";
  return "AI";
});
</script>

<style scoped>
.ai-avatar-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font: 700 12px var(--mono);
  letter-spacing: 0;
}

.ai-avatar-glyph {
  user-select: none;
}

.interview {
  background: var(--t);
  color: var(--bg);
}

.default,
.super {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--b);
  color: var(--t2);
}
</style> 