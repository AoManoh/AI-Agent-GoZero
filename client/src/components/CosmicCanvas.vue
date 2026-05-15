<template>
  <div class="cosmic-canvas-container" :class="{ 'chat-mode': isChatMode }">
    <canvas ref="sfCanvas" id="sf"></canvas>
    <canvas ref="ocCanvas" id="oc"></canvas>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();
const isChatMode = ref(false);

// Template refs
const sfCanvas = ref(null);
const ocCanvas = ref(null);

// State & animation ids
let sc, ox;
let SW, SH;
let stars = [];
let ripples = [];
let _lx = 0, _ly = 0, _rt = 0;

// Config state
const SYS = { energy: 0, t: 0 };
const sVel = { x: 0, y: 0, tx: 0, ty: 0, z: 0.00065 };

// Animation loops
let sDrawReqId = null;
let loopReqId = null;

// Handle route change for Dynamic Degradation
watch(() => route.name, (newRouteName) => {
  if (newRouteName === 'Chat') {
    isChatMode.value = true;
    sVel.z = 0.0001; // 降级: 减慢星空飞行速度
  } else {
    isChatMode.value = false;
    sVel.z = 0.00065; // 恢复: 高能星空
  }
}, { immediate: true });

function sGen() {
  if (!sfCanvas.value || !ocCanvas.value) return;
  SW = window.innerWidth;
  SH = window.innerHeight;
  sfCanvas.value.width = SW;
  sfCanvas.value.height = SH;
  ocCanvas.value.width = SW;
  ocCanvas.value.height = SH;

  const numStars = Math.floor((SW + SH) / 9);
  stars = Array.from({ length: numStars }, () => ({
    x: Math.random() * SW,
    y: Math.random() * SH,
    z: 0.14 + Math.random() * 0.86,
    a: Math.random() * Math.PI * 2,
    pulse: Math.random() * 0.004 + 0.001,
    size: 0.35 + Math.random() * 1.1,
    brightness: 0.22 + Math.random() * 0.5
  }));
}

function sDraw() {
  if (!sc) return;
  sc.clearRect(0, 0, SW, SH);
  sc.fillStyle = 'rgba(4,4,6,1)';
  sc.fillRect(0, 0, SW, SH);

  sVel.tx *= 0.965;
  sVel.ty *= 0.965;
  sVel.x += (sVel.tx - sVel.x) * 0.82;
  sVel.y += (sVel.ty - sVel.y) * 0.82;

  const isSt = Math.abs(sVel.x) + Math.abs(sVel.y) < 0.04;
  stars.forEach(s => {
    s.x += sVel.x * s.z * 0.5;
    s.y += sVel.y * s.z * 0.5;
    s.x += (s.x - SW / 2) * sVel.z * s.z;
    s.y += (s.y - SH / 2) * sVel.z * s.z;
    s.z += sVel.z;
    s.a += s.pulse * (isSt ? 1 : 0.75);

    if (s.x < 0 || s.x > SW || s.y < 0 || s.y > SH) {
      s.z = 0.1;
      s.x = Math.random() * SW;
      s.y = Math.random() * SH;
    }

    const sz = 2.6 * s.z * s.size;
    const twink = Math.sin(s.a) * 0.44 + 0.56;

    sc.beginPath();
    sc.globalAlpha = s.brightness * (0.32 + 0.68 * twink);
    sc.strokeStyle = 'rgba(255,255,255,1)';
    sc.lineWidth = sz;
    sc.lineCap = 'round';
    sc.moveTo(s.x, s.y);
    sc.lineTo(s.x + (sVel.x || 0.1), s.y + (sVel.y || 0.1));
    sc.stroke();
    sc.globalAlpha = 1;
  });

  sDrawReqId = requestAnimationFrame(sDraw);
}

function overlayLoop(t) {
  if (!ox) return;
  SYS.t = t;
  SYS.energy *= 0.965;
  ox.clearRect(0, 0, SW, SH);

  for (let i = ripples.length - 1; i >= 0; i--) {
    const rp = ripples[i];
    rp.r += (44 - rp.r) * 0.1;
    rp.life -= 0.024 + SYS.energy * 0.01;

    if (rp.life <= 0) {
      ripples.splice(i, 1);
      continue;
    }

    ox.strokeStyle = `rgba(255,255,255,${rp.life * 0.17 * (0.4 + SYS.energy * 0.6)})`;
    ox.lineWidth = 0.7 + rp.life * 0.4;
    ox.beginPath();
    ox.arc(rp.x, rp.y, rp.r, 0, Math.PI * 2);
    ox.stroke();
  }

  loopReqId = requestAnimationFrame(overlayLoop);
}

function handleMouseMove(e) {
  const dx = e.clientX - _lx;
  const dy = e.clientY - _ly;
  const v = Math.sqrt(dx * dx + dy * dy);

  SYS.energy = Math.min(1, SYS.energy + v * 0.006);
  _lx = e.clientX;
  _ly = e.clientY;
  _rt += v;

  if (_rt > 26) {
    ripples.push({ x: e.clientX, y: e.clientY, r: 0, life: 1 });
    _rt = 0;
  }

  sVel.tx += (dx / 8) * -0.1;
  sVel.ty += (dy / 8) * -0.1;
}

function handleResize() {
  sGen();
}

onMounted(() => {
  if (sfCanvas.value && ocCanvas.value) {
    sc = sfCanvas.value.getContext('2d');
    ox = ocCanvas.value.getContext('2d');
    sGen();
    sDrawReqId = requestAnimationFrame(sDraw);
    loopReqId = requestAnimationFrame(overlayLoop);

    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('resize', handleResize);
  }
});

onBeforeUnmount(() => {
  if (sDrawReqId) cancelAnimationFrame(sDrawReqId);
  if (loopReqId) cancelAnimationFrame(loopReqId);
  window.removeEventListener('mousemove', handleMouseMove);
  window.removeEventListener('resize', handleResize);
});
</script>

<style scoped>
.cosmic-canvas-container {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background-color: var(--bg, #020204);
  transition: background-color 1s ease;
}

.cosmic-canvas-container.chat-mode {
  background-color: #060608;
}

#sf {
  position: absolute;
  inset: 0;
  z-index: 0;
  opacity: 1;
  transition: opacity 1.5s ease;
}

#oc {
  position: absolute;
  inset: 0;
  z-index: 20;
  opacity: 1;
  transition: opacity 1.5s ease;
}

/* 动态降级视觉控制 */
.cosmic-canvas-container.chat-mode #sf {
  opacity: 0.12;
}

.cosmic-canvas-container.chat-mode #oc {
  opacity: 0.3;
}
</style>
