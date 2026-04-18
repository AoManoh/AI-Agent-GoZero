<template>
  <aside class="workspace-inspector surface-panel">
    <section class="inspector-block" v-if="currentModeShortLabel === 'Interview'">
      <div class="inspector-headline">
        <div>
          <h2>历史归档</h2>
          <p class="inspector-copy">只保留已归档会话。</p>
        </div>
        <button
          v-if="isAuthenticated"
          class="ghost-button inspector-refresh"
          type="button"
          :disabled="loadingRemoteContext"
          @click="$emit('refresh-remote-context')"
        >
          {{ loadingRemoteContext ? "刷新中..." : "刷新" }}
        </button>
      </div>

      <div v-if="!isAuthenticated" class="empty-state-card">
        <strong>访客模式</strong>
        <p>登录后即可查看真实归档记录。</p>
      </div>
      <div v-else-if="loadingRemoteContext && !archiveSessions.length" class="empty-state-card">
        <strong>归档同步中</strong>
        <p>正在同步历史会话。</p>
      </div>
      <div v-else-if="!archiveSessions.length" class="empty-state-card">
        <strong>暂无归档</strong>
        <p>完成一轮模拟后，这里会显示真实会话记录。</p>
      </div>
      <template v-else>
        <div class="archive-list">
          <button
            v-for="session in archiveSessions"
            :key="session.sessionId"
            class="archive-row"
            :class="{ active: session.sessionId === selectedArchiveId }"
            type="button"
            @click="$emit('select-archive', session.sessionId)"
          >
            <strong>{{ session.title || `云端会话 ${session.sessionId.slice(0, 8)}` }}</strong>
            <span>{{ formatArchiveTime(session.updatedAt || session.lastMessageAt || session.createdAt, true) }}</span>
          </button>
        </div>

        <div v-if="selectedArchiveDetail || loadingArchiveDetail" class="archive-detail-card">
          <div class="archive-detail-head">
            <strong>{{ selectedArchiveTitle }}</strong>
            <span>{{ selectedArchiveDetail?.session?.messageCount ?? archiveMessagesPreview.length }} 条记录</span>
          </div>

          <p v-if="loadingArchiveDetail" class="caption">正在加载会话详情...</p>
          <div v-else-if="archiveMessagesPreview.length" class="archive-message-list">
            <article
              v-for="message in archiveMessagesPreview"
              :key="`${message.role}-${message.createdAt}`"
              class="archive-message-item"
            >
              <strong>{{ message.role }}</strong>
              <p>{{ truncateText(message.content, 140) }}</p>
            </article>
          </div>
        </div>
      </template>
      <p v-if="remoteError" class="caption error-text">{{ remoteError }}</p>
    </section>

    <section class="inspector-block">
      <div class="inspector-headline">
        <div>
          <h2>复盘</h2>
          <p class="inspector-copy">只保留下一步最值得继续追问的点。</p>
        </div>
      </div>

      <div class="review-list">
        <article v-for="item in currentModeActions.slice(0, 3)" :key="item.title" class="review-item">
          <span>{{ item.index }}</span>
          <div>
            <strong>{{ item.title }}</strong>
            <p>{{ item.copy }}</p>
          </div>
        </article>
      </div>
    </section>
  </aside>
</template>

<script setup>
defineProps({
  isAuthenticated: {
    type: Boolean,
    default: false,
  },
  loadingRemoteContext: {
    type: Boolean,
    default: false,
  },
  loadingArchiveDetail: {
    type: Boolean,
    default: false,
  },
  remoteError: {
    type: String,
    default: "",
  },
  currentModeShortLabel: {
    type: String,
    default: "",
  },
  archiveSessions: {
    type: Array,
    default: () => [],
  },
  selectedArchiveId: {
    type: String,
    default: "",
  },
  selectedArchiveDetail: {
    type: Object,
    default: null,
  },
  archiveMessagesPreview: {
    type: Array,
    default: () => [],
  },
  selectedArchiveTitle: {
    type: String,
    default: "",
  },
  currentModeActions: {
    type: Array,
    default: () => [],
  },
  formatArchiveTime: {
    type: Function,
    required: true,
  },
  truncateText: {
    type: Function,
    required: true,
  },
});

defineEmits(["refresh-remote-context", "select-archive"]);
</script>

<style scoped>
.workspace-inspector {
  display: grid;
  align-content: start;
  gap: 18px;
  padding: 20px;
}

.inspector-block {
  display: grid;
  gap: 14px;
}

.inspector-headline {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
}

.inspector-headline h2 {
  margin: 0;
  font-family: var(--font-body);
  font-size: 1rem;
  font-weight: 600;
}

.inspector-copy {
  margin: 4px 0 0;
  color: var(--color-text-muted);
  font-size: 0.84rem;
  line-height: 1.6;
}

.inspector-refresh {
  min-height: 40px;
  padding: 0 12px;
}

.empty-state-card,
.archive-row,
.archive-detail-card,
.archive-message-item,
.review-item {
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.05);
  background: rgba(255, 255, 255, 0.025);
}

.empty-state-card,
.archive-detail-card,
.review-item {
  padding: 14px 16px;
}

.empty-state-card {
  display: grid;
  gap: 6px;
}

.empty-state-card p,
.archive-message-item p,
.review-item p {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.65;
}

.archive-list,
.archive-message-list,
.review-list {
  display: grid;
  gap: 10px;
}

.archive-row {
  width: 100%;
  display: grid;
  gap: 6px;
  padding: 14px 16px;
  text-align: left;
  cursor: pointer;
  transition:
    border-color var(--transition-base),
    background-color var(--transition-base),
    transform var(--transition-base);
}

.archive-row:hover,
.archive-row.active {
  border-color: var(--color-border-strong);
  background: rgba(255, 255, 255, 0.04);
  transform: translateY(-1px);
}

.archive-row span,
.archive-detail-head span,
.caption {
  color: var(--color-text-muted);
  font-size: 0.78rem;
}

.archive-row strong,
.archive-detail-head strong,
.review-item strong,
.archive-message-item strong {
  font-size: 0.94rem;
  font-weight: 600;
}

.archive-detail-card {
  display: grid;
  gap: 12px;
}

.archive-detail-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
}

.archive-message-item {
  display: grid;
  gap: 8px;
  padding: 12px 14px;
}

.review-item {
  display: grid;
  grid-template-columns: 36px minmax(0, 1fr);
  gap: 12px;
  align-items: start;
}

.review-item span {
  color: var(--color-text-muted);
  font-size: 0.78rem;
  line-height: 1.7;
}

.error-text {
  color: var(--color-error);
}
</style>
