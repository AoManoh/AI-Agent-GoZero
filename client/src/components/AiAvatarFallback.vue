<template>
  <div class="ai-avatar-fallback" :class="type">
    <!-- 插入 李÷ 图片url链接，保证图片大小为100% -->
    <span v-if="type === 'interview'" >
      <img src="http://pic.aomanoh.com/note/20250828030959474.jpeg" alt="AI Avatar" style="width: 100%; height: 100%;" />
    </span>
    <span v-else>🤖</span>
  </div>
</template>

<script setup>
defineProps({
  type: {
    type: String,
    default: 'default'
  }
})
</script>

<style scoped>
.ai-avatar-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  border-radius: 50%;
}

.interview {
  background: linear-gradient(45deg, #2196f3, #2196f3);
}

.default, .super {
  background: linear-gradient(45deg, #3f51b5, #5677fc);
}
</style> 