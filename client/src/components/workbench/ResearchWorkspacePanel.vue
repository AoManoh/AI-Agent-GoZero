<template>
  <section class="research-workspace">
    <aside class="research-column">
      <span class="panel-label">研究对象</span>
      <strong>{{ primaryCard?.title }}</strong>
      <p>{{ primaryCard?.copy }}</p>
      <p v-if="focusLine" class="research-focus-line">重点：{{ focusLine }}</p>
    </aside>

    <div class="research-surface">
      <article class="research-surface-card featured">
        <span class="panel-label">来源</span>
        <strong>{{ supportingCards[0]?.title || "现有来源线索" }}</strong>
        <p>{{ supportingCards[0]?.copy || "继续补充资料来源，让研究对象更完整。" }}</p>
      </article>

      <article class="research-surface-card">
        <span class="panel-label">阶段判断</span>
        <strong>{{ supportingCards[1]?.title || "阶段性发现" }}</strong>
        <p>{{ supportingCards[1]?.copy || "把阶段性观察沉淀成可继续引用的研究结论。" }}</p>
      </article>

      <article class="research-surface-card action-card">
        <span class="panel-label">下一步</span>
        <strong>继续验证</strong>
        <ul class="research-action-list">
          <li v-for="action in actions" :key="action.title">
            <strong>{{ action.title }}</strong>
            <span>{{ action.copy }}</span>
          </li>
        </ul>
      </article>
    </div>
  </section>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  artifacts: {
    type: Array,
    default: () => [],
  },
  actions: {
    type: Array,
    default: () => [],
  },
  focusList: {
    type: Array,
    default: () => [],
  },
});

const primaryCard = computed(() => props.artifacts[0] || null);
const supportingCards = computed(() => props.artifacts.slice(1, 3));
const focusLine = computed(() => props.focusList.slice(0, 3).join(" · "));
</script>

<style scoped>
.research-workspace {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
  padding: 20px;
  border-radius: 24px;
  border: 1px solid rgba(255, 255, 255, 0.05);
  background: rgba(255, 255, 255, 0.02);
}

.research-column,
.research-surface-card {
  display: grid;
  gap: 10px;
  padding: 18px;
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.05);
  background: rgba(255, 255, 255, 0.02);
}

.research-column strong,
.research-surface-card strong {
  font-size: 0.98rem;
  font-weight: 600;
}

.research-column p,
.research-surface-card p,
.research-action-list span {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.66;
}

.research-focus-line {
  color: var(--color-text-muted);
  font-size: 0.84rem;
}

.research-surface {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 16px;
}

.featured {
  min-height: 220px;
}

.action-card {
  grid-column: 1 / -1;
}

.research-action-list {
  list-style: none;
  display: grid;
  gap: 10px;
}

.research-action-list li {
  display: grid;
  gap: 4px;
}

.research-action-list li strong {
  font-size: 0.92rem;
}

@media (max-width: 1220px) {
  .research-workspace,
  .research-surface {
    grid-template-columns: 1fr;
  }
}
</style>
