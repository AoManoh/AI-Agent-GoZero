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
  font-size: 15px;
  line-height: 1.7;
  color: rgba(255, 255, 255, 0.85);
}

.markdown :deep(p) {
  margin: 0 0 1em;
}

.markdown :deep(p:last-child) {
  margin-bottom: 0;
}

.markdown :deep(pre) {
  background: #0d0d11;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 16px 0;
}

body.light-mode .markdown :deep(pre) {
  background: rgba(0, 0, 0, 0.05);
}

.markdown :deep(code) {
  font-family: var(--mono);
  background: rgba(255, 255, 255, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13.5px;
}

.markdown :deep(pre code) {
  background: transparent;
  padding: 0;
  border-radius: 0;
  color: #e2e2e2;
  font-size: 13.5px;
}

.markdown :deep(ul),
.markdown :deep(ol) {
  padding-left: 1.5rem;
  margin: 0.5rem 0 1rem;
}

.markdown :deep(li) {
  margin-bottom: 0.25rem;
}

.markdown :deep(a) {
  color: var(--color-glow-1);
  text-decoration: underline;
}

.markdown :deep(blockquote) {
  border-left: 3px solid rgba(255, 255, 255, 0.2);
  padding-left: 14px;
  margin: 1rem 0;
  color: var(--color-text-secondary);
}
</style>
