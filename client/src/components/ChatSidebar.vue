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

    <div class="sb-scroll">
      <template v-for="group in groupedSessions" :key="group.label">
        <div v-if="group.items.length > 0" class="sb-section">{{ group.label }}</div>
        <div v-if="group.items.length > 0" class="sb-list">
          <div
            v-for="session in group.items"
            :key="session.id"
            class="s-item"
            :class="{ active: session.id === activeSessionId }"
            :title="session.title"
            @click="$emit('select-session', session.id)"
          >
            <span class="dot"></span>{{ session.title || '未命名会话' }}
          </div>
        </div>
      </template>
      <div v-if="!sessions.length" class="sb-empty">暂无会话，点击 + 新建</div>
    </div>

    <div class="sb-footer">
      <button class="btn-ghost user-btn" type="button">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20.94c1.5 0 2.75 1.06 4 1.06 3 0 6-8 6-12.22A4.91 4.91 0 0 0 17 5c-2.22 0-4 1.44-5 2-1-.56-2.78-2-5-2a4.9 4.9 0 0 0-5 4.78C2 14 5 22 8 22c1.25 0 2.5-1.06 4-1.06Z"/></svg>
        {{ username || '候选人' }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  sessions: {
    type: Array,
    default: () => [],
  },
  activeSessionId: {
    type: String,
    default: "",
  },
  username: {
    type: String,
    default: "",
  },
});

defineEmits(["new-chat", "select-session"]);

const DAY_MS = 24 * 60 * 60 * 1000;

const groupedSessions = computed(() => {
  const now = Date.now();
  const todayStart = new Date(now);
  todayStart.setHours(0, 0, 0, 0);
  const todayStartTs = todayStart.getTime();
  const sevenDaysAgoTs = todayStartTs - 7 * DAY_MS;

  const today = [];
  const week = [];
  const older = [];

  const sorted = [...props.sessions].sort(
    (a, b) => (b.updatedAt || 0) - (a.updatedAt || 0)
  );

  sorted.forEach((session) => {
    const ts = session.updatedAt || session.createdAt || 0;
    if (ts >= todayStartTs) {
      today.push(session);
    } else if (ts >= sevenDaysAgoTs) {
      week.push(session);
    } else {
      older.push(session);
    }
  });

  return [
    { label: "Today", items: today },
    { label: "Previous 7 Days", items: week },
    { label: "Older", items: older },
  ];
});
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
  font-size: var(--fs-md);
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-new {
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
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

.sb-scroll {
  flex: 1;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.08) transparent;
}

.sb-scroll::-webkit-scrollbar {
  width: 6px;
}

.sb-scroll::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.08);
  border-radius: 3px;
}

.sb-section {
  padding: 16px 20px 6px;
  font-size: var(--fs-2xs);
  font-family: var(--mono);
  color: rgba(255, 255, 255, 0.4);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.sb-list {
  padding: 0 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.sb-empty {
  padding: 32px 20px;
  text-align: center;
  font-size: var(--fs-xs);
  color: rgba(255, 255, 255, 0.35);
  font-family: var(--mono);
}

.s-item {
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  font-size: var(--fs-sm);
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
  font-size: var(--fs-sm);
  padding: 6px 12px;
  border-radius: 6px;
  transition: background 0.2s;
}

.btn-ghost:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--t);
}

.user-btn {
  width: 100%;
  text-align: left;
  padding: 8px 12px;
  color: var(--t);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}

@media (max-width: 768px) {
  .sidebar {
    display: none; 
  }
}
</style>
