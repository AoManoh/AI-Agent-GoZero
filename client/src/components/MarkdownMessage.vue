<template>
  <div class="markdown" v-html="sanitized" @click="handleMarkdownClick"></div>
</template>

<script setup>
import { computed } from "vue";
import { Marked, Renderer } from "marked";
import DOMPurify from "dompurify";
import hljs from "highlight.js/lib/core";
import bash from "highlight.js/lib/languages/bash";
import cpp from "highlight.js/lib/languages/cpp";
import css from "highlight.js/lib/languages/css";
import dockerfile from "highlight.js/lib/languages/dockerfile";
import go from "highlight.js/lib/languages/go";
import java from "highlight.js/lib/languages/java";
import javascript from "highlight.js/lib/languages/javascript";
import json from "highlight.js/lib/languages/json";
import markdown from "highlight.js/lib/languages/markdown";
import python from "highlight.js/lib/languages/python";
import sql from "highlight.js/lib/languages/sql";
import typescript from "highlight.js/lib/languages/typescript";
import xml from "highlight.js/lib/languages/xml";
import yaml from "highlight.js/lib/languages/yaml";

hljs.registerLanguage("bash", bash);
hljs.registerLanguage("cpp", cpp);
hljs.registerLanguage("css", css);
hljs.registerLanguage("dockerfile", dockerfile);
hljs.registerLanguage("go", go);
hljs.registerLanguage("java", java);
hljs.registerLanguage("javascript", javascript);
hljs.registerLanguage("json", json);
hljs.registerLanguage("markdown", markdown);
hljs.registerLanguage("python", python);
hljs.registerLanguage("sql", sql);
hljs.registerLanguage("typescript", typescript);
hljs.registerLanguage("xml", xml);
hljs.registerLanguage("yaml", yaml);

const props = defineProps({
  content: {
    type: String,
    default: "",
  },
});

const languageAliases = {
  c: "cpp",
  "c++": "cpp",
  js: "javascript",
  jsx: "javascript",
  sh: "bash",
  shell: "bash",
  ts: "typescript",
  tsx: "typescript",
  vue: "xml",
  yml: "yaml",
};

const escapeHtml = (value = "") =>
  String(value)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");

const escapeAttr = (value = "") => escapeHtml(value).replace(/`/g, "&#96;");

const normalizeLanguage = (lang = "") => {
  const first = String(lang || "").trim().split(/\s+/)[0].toLowerCase();
  return languageAliases[first] || first;
};

const renderer = new Renderer();

renderer.link = function link({ href, title, tokens }) {
  const text = this.parser.parseInline(tokens);
  const safeHref = escapeAttr(href || "#");
  const safeTitle = title ? ` title="${escapeAttr(title)}"` : "";
  return `<a href="${safeHref}" target="_blank" rel="noopener noreferrer"${safeTitle}>${text}</a>`;
};

renderer.code = ({ text, lang }) => {
  const normalizedLanguage = normalizeLanguage(lang);
  const displayLanguage = normalizedLanguage || "text";
  let highlighted = escapeHtml(text);

  if (normalizedLanguage && hljs.getLanguage(normalizedLanguage)) {
    try {
      highlighted = hljs.highlight(text, {
        language: normalizedLanguage,
        ignoreIllegals: true,
      }).value;
    } catch {
      highlighted = escapeHtml(text);
    }
  }

  return `<figure class="code-block" data-language="${escapeAttr(displayLanguage)}">
  <figcaption class="code-toolbar">
    <span class="code-language">${escapeHtml(displayLanguage)}</span>
    <button class="code-copy" type="button" data-copy-code aria-label="复制代码">复制</button>
  </figcaption>
  <pre><code class="hljs language-${escapeAttr(displayLanguage)}">${highlighted}</code></pre>
</figure>`;
};

const marked = new Marked({
  breaks: true,
  gfm: true,
  renderer,
});

const sanitized = computed(() => {
  const html = marked.parse(props.content || "");
  return DOMPurify.sanitize(html, {
    ADD_TAGS: ["button"],
    ADD_ATTR: ["target", "rel", "type", "data-copy-code", "aria-label", "data-language"],
  });
});

const writeClipboard = async (text) => {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }

  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand("copy");
  document.body.removeChild(textarea);
};

const handleMarkdownClick = async (event) => {
  const target = event.target;
  if (!(target instanceof HTMLElement)) return;

  const button = target.closest("[data-copy-code]");
  if (!button) return;

  const code = button.closest(".code-block")?.querySelector("pre code")?.innerText || "";
  if (!code) return;

  try {
    await writeClipboard(code);
    const originalText = button.textContent || "复制";
    button.textContent = "已复制";
    window.setTimeout(() => {
      button.textContent = originalText;
    }, 1200);
  } catch {
    button.textContent = "复制失败";
    window.setTimeout(() => {
      button.textContent = "复制";
    }, 1200);
  }
};
</script>

<style scoped>
.markdown {
  font-size: var(--fs-lg);
  line-height: 1.7;
  color: rgba(255, 255, 255, 0.85);
  overflow-wrap: anywhere;
}

.markdown :deep(p) {
  margin: 0 0 1em;
}

.markdown :deep(p:last-child) {
  margin-bottom: 0;
}

.markdown :deep(.code-block) {
  margin: 18px 0;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: var(--radius-md);
  background: #0b0c10;
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.34);
}

.markdown :deep(.code-toolbar) {
  min-height: 34px;
  padding: 0 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.04);
}

.markdown :deep(.code-language) {
  color: rgba(255, 255, 255, 0.58);
  font: 600 var(--fs-2xs) var(--mono);
  letter-spacing: 0;
  text-transform: uppercase;
}

.markdown :deep(.code-copy) {
  height: 24px;
  padding: 0 8px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: var(--radius-xs);
  background: rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.74);
  font: 500 var(--fs-xs) var(--sans);
  cursor: pointer;
  transition: background 0.16s ease, color 0.16s ease, border-color 0.16s ease;
}

.markdown :deep(.code-copy:hover) {
  border-color: rgba(255, 255, 255, 0.24);
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
}

.markdown :deep(pre) {
  margin: 0;
  padding: 18px 20px;
  overflow-x: auto;
  background: transparent;
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.22) transparent;
}

.markdown :deep(code) {
  font-family: var(--mono);
  background: rgba(255, 255, 255, 0.1);
  padding: 2px 6px;
  border-radius: var(--radius-xs);
  font-size: var(--fs-sm);
}

.markdown :deep(pre code) {
  display: block;
  min-width: max-content;
  background: transparent;
  padding: 0;
  border-radius: 0;
  color: #d7dbe7;
  font-size: var(--fs-sm);
  line-height: 1.72;
  white-space: pre;
  tab-size: 2;
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
  color: var(--t);
  text-decoration: underline;
  text-decoration-color: var(--bh);
  text-underline-offset: 3px;
  transition: text-decoration-color 0.2s;
}

.markdown :deep(a:hover) {
  text-decoration-color: var(--t);
}

.markdown :deep(blockquote) {
  border-left: 3px solid rgba(255, 255, 255, 0.2);
  padding-left: 14px;
  margin: 1rem 0;
  color: var(--color-text-secondary);
}

.markdown :deep(.hljs-keyword),
.markdown :deep(.hljs-selector-tag),
.markdown :deep(.hljs-literal),
.markdown :deep(.hljs-section),
.markdown :deep(.hljs-link) {
  color: #ffb86c;
}

.markdown :deep(.hljs-string),
.markdown :deep(.hljs-title),
.markdown :deep(.hljs-name),
.markdown :deep(.hljs-type),
.markdown :deep(.hljs-attribute),
.markdown :deep(.hljs-symbol),
.markdown :deep(.hljs-bullet),
.markdown :deep(.hljs-addition),
.markdown :deep(.hljs-variable),
.markdown :deep(.hljs-template-tag),
.markdown :deep(.hljs-template-variable) {
  color: #8bd49c;
}

.markdown :deep(.hljs-comment),
.markdown :deep(.hljs-quote),
.markdown :deep(.hljs-deletion),
.markdown :deep(.hljs-meta) {
  color: #7f8497;
}

.markdown :deep(.hljs-number),
.markdown :deep(.hljs-regexp),
.markdown :deep(.hljs-built_in),
.markdown :deep(.hljs-builtin-name),
.markdown :deep(.hljs-params) {
  color: #82aaff;
}

.markdown :deep(.hljs-function),
.markdown :deep(.hljs-class),
.markdown :deep(.hljs-property),
.markdown :deep(.hljs-operator),
.markdown :deep(.hljs-punctuation) {
  color: #cdd6f4;
}

body.light-mode .markdown :deep(.code-block) {
  background: #f8fafc;
  border-color: rgba(15, 23, 42, 0.12);
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.08);
}

body.light-mode .markdown :deep(.code-toolbar) {
  background: rgba(15, 23, 42, 0.04);
  border-bottom-color: rgba(15, 23, 42, 0.08);
}

body.light-mode .markdown :deep(.code-language),
body.light-mode .markdown :deep(.code-copy) {
  color: rgba(15, 23, 42, 0.68);
}
</style>
