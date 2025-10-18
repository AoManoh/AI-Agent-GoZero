<template>
  <section class="page-section" :class="variant">
    <div class="section-container">
      <slot />
    </div>
  </section>
</template>

<script setup>
defineProps({
  variant: {
    type: String,
    default: "default",
  },
});
</script>

<style scoped>
.page-section {
  width: 100%;
  display: flex;
  justify-content: center;
  padding: 60px 20px;
}

.section-container {
  width: 100%;
  max-width: 900px;
}

.page-section.tight {
  padding: 40px 20px;
}

.page-section.hero {
  padding-top: 5vh;
}
</style>
