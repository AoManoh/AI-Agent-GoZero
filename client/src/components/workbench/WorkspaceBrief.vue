<template>
  <section class="workspace-brief" :class="[modeSlug, { structured }]">
    <article class="brief-main">
      <span class="brief-label">当前产出</span>
      <strong>{{ artifactTitle }}</strong>
      <p>{{ latestOutputSnapshot }}</p>
    </article>

    <div class="brief-side">
      <article v-for="panel in panels" :key="panel.title" class="brief-point">
        <span class="brief-label">{{ panel.label }}</span>
        <strong>{{ panel.title }}</strong>
        <p>{{ panel.copy }}</p>
      </article>
    </div>

    <section class="artifact-board" v-if="structuredArtifacts.length">
      <article v-for="item in structuredArtifacts" :key="item.title" class="artifact-card">
        <span class="brief-label">{{ item.label }}</span>
        <strong>{{ item.title }}</strong>
        <p>{{ item.copy }}</p>
        <span class="artifact-meta">{{ item.meta }}</span>
      </article>
    </section>
  </section>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  modeName: {
    type: String,
    default: "",
  },
  artifactTitle: {
    type: String,
    default: "",
  },
  latestOutputSnapshot: {
    type: String,
    default: "",
  },
  panels: {
    type: Array,
    default: () => [],
  },
  structuredArtifacts: {
    type: Array,
    default: () => [],
  },
  structured: {
    type: Boolean,
    default: false,
  },
});

const modeSlug = computed(() => {
  if (props.modeName === "Research Desk") return "research";
  if (props.modeName === "Memory Atlas") return "memory";
  if (props.modeName === "Coach") return "coach";
  return "interview";
});
</script>

<style scoped>
.workspace-brief {
  display: grid;
  grid-template-columns: minmax(0, 1.12fr) minmax(320px, 0.88fr);
  gap: 14px;
  align-items: stretch;
}

.workspace-brief.structured {
  grid-template-columns: minmax(0, 1fr) minmax(320px, 0.82fr);
}

.brief-main,
.brief-point {
  display: grid;
  gap: 10px;
  padding: 18px;
  border-radius: 22px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.brief-main {
  align-content: start;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.04), transparent 30%),
    rgba(255, 255, 255, 0.03);
  transition:
    transform var(--transition-base),
    border-color var(--transition-base),
    box-shadow var(--transition-slow);
}

.brief-label {
  color: var(--color-text-muted);
  font-size: 0.8rem;
  line-height: 1.4;
}

.brief-main strong {
  font-family: var(--font-display);
  font-size: 1.28rem;
  letter-spacing: -0.03em;
}

.brief-main p {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.82;
}

.brief-main:hover,
.brief-point:hover,
.artifact-card:hover {
  transform: translateY(-4px);
  border-color: var(--color-border-strong);
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.18);
}

.brief-side {
  display: grid;
  gap: 12px;
}

.brief-point {
  background: rgba(255, 255, 255, 0.018);
  transition:
    transform var(--transition-base),
    border-color var(--transition-base),
    box-shadow var(--transition-slow);
}

.brief-point strong {
  font-size: 1rem;
}

.brief-point p {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.75;
}

.artifact-board {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.artifact-card {
  display: grid;
  gap: 10px;
  padding: 16px;
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.05);
  background: rgba(255, 255, 255, 0.018);
  transition:
    transform var(--transition-base),
    border-color var(--transition-base),
    box-shadow var(--transition-slow);
}

.artifact-card strong {
  font-size: 1rem;
}

.artifact-card p {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.72;
}

.artifact-meta {
  color: var(--color-text-muted);
  font-size: 0.8rem;
  line-height: 1.6;
}

.workspace-brief.research .artifact-board {
  grid-template-columns: minmax(0, 1.25fr) repeat(2, minmax(0, 1fr));
}

.workspace-brief.research .artifact-card:first-child {
  grid-column: span 1;
  grid-row: span 2;
  min-height: 240px;
  align-content: end;
  background:
    linear-gradient(180deg, rgba(154, 169, 189, 0.12), rgba(255, 255, 255, 0.02)),
    rgba(255, 255, 255, 0.02);
}

.workspace-brief.research .artifact-card:nth-child(2) {
  grid-column: span 2;
}

.workspace-brief.memory .artifact-board {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.workspace-brief.memory .artifact-card:first-child,
.workspace-brief.memory .artifact-card:nth-child(3) {
  grid-column: span 2;
}

.workspace-brief.coach .artifact-board {
  grid-template-columns: minmax(0, 1.15fr) repeat(2, minmax(0, 1fr));
}

.workspace-brief.coach .artifact-card:first-child {
  grid-row: span 2;
  min-height: 220px;
  align-content: end;
  background:
    linear-gradient(180deg, rgba(122, 163, 140, 0.12), rgba(255, 255, 255, 0.02)),
    rgba(255, 255, 255, 0.02);
}

:global(.workspace-page.theme-interview) .workspace-brief,
:global(.workspace-page.theme-interview) .brief-main {
  background:
    linear-gradient(180deg, rgba(126, 147, 186, 0.08), rgba(255, 255, 255, 0.02));
  border-color: rgba(126, 147, 186, 0.12);
}

:global(.workspace-page.theme-research) .workspace-brief,
:global(.workspace-page.theme-research) .brief-main {
  background:
    linear-gradient(180deg, rgba(154, 169, 189, 0.06), rgba(255, 255, 255, 0.02));
  border-color: rgba(154, 169, 189, 0.14);
}

:global(.workspace-page.theme-memory) .workspace-brief,
:global(.workspace-page.theme-memory) .brief-main {
  background:
    linear-gradient(180deg, rgba(204, 181, 134, 0.08), rgba(255, 255, 255, 0.02));
  border-color: rgba(204, 181, 134, 0.16);
}

:global(.workspace-page.theme-coach) .workspace-brief,
:global(.workspace-page.theme-coach) .brief-main {
  background:
    linear-gradient(180deg, rgba(122, 163, 140, 0.08), rgba(255, 255, 255, 0.02));
  border-color: rgba(122, 163, 140, 0.16);
}

@media (max-width: 1220px) {
  .workspace-brief {
    grid-template-columns: 1fr;
  }

  .artifact-board {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .artifact-board {
    grid-template-columns: 1fr;
  }
}
</style>
