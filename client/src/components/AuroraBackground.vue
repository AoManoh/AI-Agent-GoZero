<template>
  <div class="aurora-background">
    <div class="aurora-shape shape1"></div>
    <div class="aurora-shape shape2"></div>
  </div>
</template>

<style scoped>
.aurora-background {
  position: fixed;
  inset: 0;
  z-index: -1;
  overflow: hidden;
}

.aurora-shape {
  position: absolute;
  border-radius: 50%;
  filter: blur(120px);
  opacity: var(--color-aurora-opacity);
  transition: opacity 0.3s ease;
}

.shape1 {
  width: 500px;
  height: 500px;
  background: var(--color-glow-1);
  top: -150px;
  left: -150px;
  animation: move1 15s infinite alternate;
}

.shape2 {
  width: 400px;
  height: 400px;
  background: var(--color-glow-2);
  bottom: -100px;
  right: -100px;
  animation: move2 18s infinite alternate;
}

@keyframes move1 {
  to {
    transform: translate(100px, 150px) scale(1.2);
  }
}

@keyframes move2 {
  to {
    transform: translate(-150px, -100px) scale(0.8);
  }
}
</style>
