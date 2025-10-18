<template>
  <div class="markdown" v-html="sanitized"></div>
</template>

<script setup>
import { computed } from "vue";
import { marked } from "marked";
import DOMPurify from "dompurify";

const props = defineProps({
  content: {
    type: String,
    default: "",
  },
});

const renderer = new marked.Renderer();
renderer.link = (href, title, text) => {
  const target = '_blank';
  const rel = 'noopener noreferrer';
  const safeHref = href ?? '#';
  const safeTitle = title ? ` title="${title}"` : "";
  return `<a href="${safeHref}" target="${target}" rel="${rel}"${safeTitle}>${text}</a>`;
};

marked.setOptions({
  breaks: true,
  gfm: true,
  renderer,
});

const sanitized = computed(() => {
  const html = marked.parse(props.content || "");
  return DOMPurify.sanitize(html);
});
</script>

<style scoped>
.markdown {
  font-size: 1rem;
  line-height: 1.7;
  color: var(--color-text-primary);
}

.markdown :deep(p) {
  margin: 0 0 0.8em;
}

.markdown :deep(pre) {
  background: rgba(0, 0, 0, 0.35);
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
}

body.light-mode .markdown :deep(pre) {
  background: rgba(0, 0, 0, 0.05);
}

.markdown :deep(code) {
  font-family: "Fira Code", "SFMono-Regular", monospace;
}

.markdown :deep(ul),
.markdown :deep(ol) {
  padding-left: 1.5rem;
  margin: 0.5rem 0;
}

.markdown :deep(a) {
  color: var(--color-glow-2);
  text-decoration: underline;
}

.markdown :deep(blockquote) {
  border-left: 4px solid rgba(255, 255, 255, 0.2);
  padding-left: 12px;
  margin: 1rem 0;
  color: var(--color-text-secondary);
}
</style>
