<template>
  <!--
    KnowledgeSidebar：私人知识库子页左栏抽象组件。

    v0.1（mode='visibility'，当前实现）:
      - section1「知识范围」: 显示公共知识 / 我的私人资料 二分类，count 来自父组件传入的 documents 客户端聚合
      - section2「文件夹（v0.2 上线）」: 灰色 disabled 状态展示 5 个建议 folder 名（lazy-init 设计示意），不可点击
      - 顶部 stats 显示 documents 总数 / chunks 总数 / 向量维度

    v0.2（mode='folder-tree'，预留）:
      - 由父组件后续切换到该模式后内部渲染递归 folder 树（含「未归类」虚拟节点 + 拖拽排序）
      - 当前阶段仅占位 props/事件契约，避免 v0.2 切换时父组件变更

    Props/事件契约（保持稳定，v0.1 / v0.2 共用）:
      - mode: 'visibility' | 'folder-tree'
      - documents: KnowledgeDocumentItem[] - 后端返回的完整文档列表（v0.1 用于 visibility 聚合）
      - activeKey: 当前激活节点 key（'public' | 'private' | folderId 字符串）
      - @select="(key) => ..." - 用户点击节点时上抛 key

    决策来源:
      - docs/requirements/2026-05-12-workbench-knowledge-base-redesign.md Q5=B
      - 响应式约束 R: 组件宽度自适应父容器，section 字号 clamp(13px, 0.95vw, 15px)
  -->
  <aside class="kb-sidebar">
    <!-- section1: 知识范围（visibility 二分类） -->
    <section class="kb-sidebar-section">
      <header class="kb-sidebar-section-head">
        <span class="kb-sidebar-section-title">知识范围</span>
      </header>
      <ul class="kb-sidebar-list">
        <li
          class="kb-sidebar-item"
          :class="{ 'kb-sidebar-item-active': activeKey === 'public' }"
          @click="$emit('select', 'public')"
        >
          <span class="kb-sidebar-dot kb-dot-public" aria-hidden="true"></span>
          <span class="kb-sidebar-label">公共知识</span>
          <span class="kb-sidebar-count">{{ publicCount }}</span>
        </li>
        <li
          v-if="hasPrivateAccess"
          class="kb-sidebar-item"
          :class="{ 'kb-sidebar-item-active': activeKey === 'private' }"
          @click="$emit('select', 'private')"
        >
          <span class="kb-sidebar-dot kb-dot-private" aria-hidden="true"></span>
          <span class="kb-sidebar-label">我的私人资料</span>
          <span class="kb-sidebar-count">{{ privateCount }}</span>
        </li>
      </ul>
    </section>

    <!-- section2: 文件夹（v0.2 上线，灰色 disabled 占位） -->
    <section v-if="hasPrivateAccess" class="kb-sidebar-section kb-sidebar-section-disabled">
      <header class="kb-sidebar-section-head">
        <span class="kb-sidebar-section-title">文件夹</span>
        <span class="kb-sidebar-section-tag">v0.2 上线</span>
      </header>
      <ul class="kb-sidebar-list">
        <li
          v-for="placeholder in folderPlaceholders"
          :key="placeholder"
          class="kb-sidebar-item kb-sidebar-item-disabled"
          :title="`${placeholder} · 文件夹功能将在 v0.2 启用`"
        >
          <span class="kb-sidebar-folder-icon" aria-hidden="true">▸</span>
          <span class="kb-sidebar-label">{{ placeholder }}</span>
        </li>
      </ul>
    </section>

    <!-- 底部 stats 概览 -->
    <section class="kb-sidebar-stats">
      <div class="kb-sidebar-stat-line">
        <span class="kb-sidebar-stat-num">{{ documents.length }}</span>
        <span class="kb-sidebar-stat-lb">文档</span>
      </div>
      <div class="kb-sidebar-stat-line">
        <span class="kb-sidebar-stat-num">{{ totalChunks }}</span>
        <span class="kb-sidebar-stat-lb">片段</span>
      </div>
      <div v-if="embeddingLabel" class="kb-sidebar-stat-line">
        <span class="kb-sidebar-stat-num">{{ embeddingLabel }}</span>
        <span class="kb-sidebar-stat-lb">维度</span>
      </div>
    </section>
  </aside>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  mode: {
    type: String,
    default: "visibility",
    validator: (v) => ["visibility", "folder-tree"].includes(v),
  },
  documents: {
    type: Array,
    default: () => [],
  },
  activeKey: {
    type: String,
    default: "public",
  },
  hasPrivateAccess: {
    type: Boolean,
    default: true,
  },
});

defineEmits(["select"]);

// section2 占位文件夹名：与 X=X3 决策的 5 个建议 folder（岗位画像 / 面试题库 / 项目经历 / 业务知识 / 个人复盘）一致，
// v0.2 lazy-init 真实 folder 后这里会被替换为递归 folder 树。
const folderPlaceholders = [
  "岗位画像",
  "面试题库",
  "项目经历",
  "业务知识",
  "个人复盘",
];

// 客户端聚合 visibility，避免向后端新增专门的聚合端点（v0.1 1 天交付边界）。
const publicCount = computed(() =>
  props.documents.filter((d) => (d.visibility || d.scope) === "public").length
);

const privateCount = computed(() =>
  props.documents.filter((d) => (d.visibility || d.scope) === "private").length
);

const totalChunks = computed(() =>
  props.documents.reduce((sum, d) => sum + (d.chunkCount || 0), 0)
);

// 维度信息：从第一份文档的 embeddingDimension 取（后端 Q8=B 三字段都暴露了）；空态时显示 ''
const embeddingLabel = computed(() => {
  const first = props.documents.find((d) => d.embeddingDimension > 0);
  return first ? `${first.embeddingDimension}d` : "";
});
</script>

<style scoped>
/* 容器：宽度跟随父 grid track，不写死像素（响应式约束 R） */
.kb-sidebar {
  display: flex;
  flex-direction: column;
  gap: clamp(0.875rem, 1.5vw, 1.25rem);
  padding: clamp(0.875rem, 1.5vw, 1.125rem) clamp(0.75rem, 1.25vw, 1rem) clamp(1rem, 1.75vw, 1.25rem);
  background:
    linear-gradient(180deg, rgba(16, 17, 22, 1) 0%, rgba(10, 11, 14, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.03) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  isolation: isolate;
  min-width: 0;
}

.kb-sidebar-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.kb-sidebar-section-disabled {
  opacity: 0.6;
}

.kb-sidebar-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 0.625rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.kb-sidebar-section-title {
  font: 600 clamp(11px, 0.78vw, 12px) var(--mono);
  color: var(--t3);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.kb-sidebar-section-tag {
  font: 500 clamp(10px, 0.7vw, 11px) var(--mono);
  color: rgba(220, 155, 90, 0.6);
  letter-spacing: 0.04em;
  padding: 2px 6px;
  border: 1px solid rgba(220, 155, 90, 0.2);
  border-radius: 999px;
  background: rgba(220, 155, 90, 0.05);
}

.kb-sidebar-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.kb-sidebar-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.5rem 0.625rem;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font: clamp(13px, 0.95vw, 15px) var(--sans);
  color: var(--t2);
  transition: color 0.2s ease, background-color 0.2s ease;
  min-width: 0;
}

.kb-sidebar-item:hover {
  color: var(--t);
  background: rgba(255, 255, 255, 0.03);
}

.kb-sidebar-item-active {
  color: var(--t);
  background: rgba(220, 155, 90, 0.06);
  position: relative;
}

.kb-sidebar-item-active::before {
  content: "";
  position: absolute;
  left: -0.625rem;
  top: 0.5rem;
  bottom: 0.5rem;
  width: 2px;
  background: rgba(220, 155, 90, 0.95);
  border-radius: 0 2px 2px 0;
}

.kb-sidebar-item-disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

.kb-sidebar-item-disabled:hover {
  color: var(--t2);
  background: transparent;
}

.kb-sidebar-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.kb-dot-public {
  background: rgba(108, 190, 255, 0.85);
  box-shadow: 0 0 6px rgba(108, 190, 255, 0.35);
}

.kb-dot-private {
  background: rgba(220, 155, 90, 0.95);
  box-shadow: 0 0 6px rgba(220, 155, 90, 0.45);
}

.kb-sidebar-folder-icon {
  font: 12px var(--mono);
  color: var(--t3);
  flex-shrink: 0;
}

.kb-sidebar-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.kb-sidebar-count {
  font: clamp(11px, 0.78vw, 12px) var(--mono);
  color: var(--t3);
  letter-spacing: 0.03em;
  flex-shrink: 0;
}

.kb-sidebar-stats {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.75rem 0 0;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.kb-sidebar-stat-line {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}

.kb-sidebar-stat-num {
  font: 600 clamp(13px, 0.95vw, 15px) var(--mono);
  color: var(--t);
}

.kb-sidebar-stat-lb {
  font: clamp(11px, 0.78vw, 12px) var(--mono);
  color: var(--t3);
  letter-spacing: 0.04em;
}
</style>
