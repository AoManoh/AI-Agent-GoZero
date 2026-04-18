<template>
  <aside class="workspace-rail surface-panel">
    <section class="rail-block">
      <div class="rail-head">
        <div>
          <h2>会话</h2>
        </div>
        <button class="primary-button sidebar-create" type="button" @click="$emit('create-session')">
          新建会话
        </button>
      </div>
    </section>

    <section class="rail-block">
      <div class="session-list">
        <div v-if="!workbenchRailItems.length" class="empty-state-card compact-empty">
          <strong>暂无会话</strong>
        </div>
        <button
          v-for="session in workbenchRailItems"
          :key="session.id"
          class="session-card"
          :class="{ active: session.id === activeSessionId }"
          type="button"
          @click="$emit('activate-session', session)"
        >
          <div class="session-card-top">
            <span>{{ session.displayTime }}</span>
          </div>
          <strong>{{ session.title }}</strong>
        </button>
      </div>
    </section>
  </aside>
</template>

<script setup>
defineProps({
  workbenchRailItems: {
    type: Array,
    default: () => [],
  },
  activeSessionId: {
    type: String,
    default: "",
  },
});

defineEmits(["create-session", "activate-session"]);
</script>

<style scoped>
.workspace-rail {
  display: grid;
  align-content: start;
  gap: 10px;
  padding: 14px;
}

.rail-block {
  display: grid;
  gap: 10px;
}

.rail-head {
  display: grid;
  gap: 12px;
}

.rail-head h2 {
  margin: 0;
  font-family: var(--font-body);
  font-size: 1.05rem;
  font-weight: 600;
}

.sidebar-create {
  width: fit-content;
  min-height: 36px;
  padding: 0 12px;
  justify-self: start;
}

.session-list {
  display: grid;
  gap: 8px;
  max-height: 380px;
  overflow-y: auto;
}

.empty-state-card,
.session-card {
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.05);
  background: rgba(255, 255, 255, 0.025);
}

.compact-empty,
.session-card {
  padding: 10px 12px;
}

.empty-state-card {
  display: grid;
  gap: 4px;
}

.session-card {
  width: 100%;
  display: grid;
  gap: 8px;
  text-align: left;
  cursor: pointer;
  transition:
    border-color var(--transition-base),
    background-color var(--transition-base);
}

.session-card:hover,
.session-card.active {
  border-color: var(--color-border-strong);
  background: rgba(255, 255, 255, 0.04);
}

.session-card-top {
  display: flex;
  justify-content: flex-start;
  gap: 12px;
  color: var(--color-text-muted);
  font-size: 0.78rem;
}

.session-card strong {
  font-size: 0.94rem;
  font-weight: 600;
}
</style>
