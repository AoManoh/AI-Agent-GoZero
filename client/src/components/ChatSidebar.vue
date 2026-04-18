<template>
  <div class="sidebar">
    <div class="sb-header">
      <div class="sb-brand">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><rect width="24" height="24" rx="5" fill="rgba(255,255,255,.08)" stroke="rgba(255,255,255,.12)" stroke-width="1"/><circle cx="12" cy="12" r="7" stroke="rgba(255,255,255,.45)" stroke-width="1"/><line x1="7" y1="10" x2="13" y2="10" stroke="rgba(255,255,255,.9)" stroke-width="1.5" stroke-linecap="round"/><line x1="11" y1="14" x2="17" y2="14" stroke="rgba(255,255,255,.9)" stroke-width="1.5" stroke-linecap="round"/></svg>
        面试记录
      </div>
      <button class="btn-new" title="新建面试" @click="$emit('new-chat')">
        <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><path d="M6 2v8M2 6h8"/></svg>
      </button>
    </div>
    <div class="sb-section">Today</div>
    <div class="sb-list">
      <div class="s-item active"><span class="dot"></span>当前模拟面试</div>
    </div>
    <div class="sb-footer">
      <button class="btn-ghost" style="width:100%; text-align:left; padding:8px 12px; display:flex; align-items:center; gap:8px;">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20.94c1.5 0 2.75 1.06 4 1.06 3 0 6-8 6-12.22A4.91 4.91 0 0 0 17 5c-2.22 0-4 1.44-5 2-1-.56-2.78-2-5-2a4.9 4.9 0 0 0-5 4.78C2 14 5 22 8 22c1.25 0 2.5-1.06 4-1.06Z"/></svg>
        候选人
      </button>
    </div>
  </div>
</template>

<script setup>
defineEmits(["new-chat"]);
</script>

<style scoped>
.sidebar {
  width: 280px;
  background: rgba(6, 6, 8, 0.7);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border-right: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sb-header {
  padding: 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.sb-brand {
  font-size: 14px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-new {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: transparent;
  color: var(--t2);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all .2s;
}

.btn-new:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--t);
}

.sb-section {
  padding: 20px 20px 8px;
  font-size: 11px;
  font-family: var(--mono);
  color: rgba(255, 255, 255, 0.4);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.sb-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.s-item {
  padding: 10px 12px;
  border-radius: 8px;
  font-size: 13.5px;
  color: var(--t2);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 10px;
  transition: background .2s;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.s-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.s-item.active {
  background: rgba(255, 255, 255, 0.08);
  color: var(--t);
  font-weight: 500;
}

.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--t3);
  flex-shrink: 0;
}

.s-item.active .dot {
  background: #fff;
}

.sb-footer {
  padding: 16px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.04);
}

.btn-ghost {
  background: transparent;
  border: none;
  color: var(--t2);
  cursor: pointer;
  font-family: var(--sans);
  font-size: 13px;
  padding: 6px 12px;
  border-radius: 6px;
  transition: background 0.2s;
}

.btn-ghost:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--t);
}

@media (max-width: 768px) {
  .sidebar {
    display: none; 
  }
}
</style>
