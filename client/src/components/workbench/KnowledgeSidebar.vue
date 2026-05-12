<template>
  <!--
    KnowledgeSidebar：私人知识库子页左栏抽象组件。

    Props/事件契约（保持稳定，v0.1 / v0.2 共用）:
      - mode: 'visibility' | 'folder-tree'
      - documents: KnowledgeDocumentItem[] - 后端返回的完整文档列表（v0.1 用于 visibility 聚合）
      - folders: KnowledgeFolderItem[] - 后端返回的当前用户目录列表
      - activeKey: 当前激活节点 key（'public' | 'private' | 'unfiled' | 'folder:{id}'）
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

    <!-- section2: 文件夹（v0.2 真实目录树 + CRUD） -->
    <section v-if="hasPrivateAccess" class="kb-sidebar-section">
      <header class="kb-sidebar-section-head">
        <span class="kb-sidebar-section-title">文件夹</span>
        <button
          v-if="!creatingFolder"
          class="kb-sidebar-icon-btn"
          type="button"
          title="新建文件夹"
          aria-label="新建文件夹"
          @click="startCreateFolder"
        >
          +
        </button>
      </header>
      <!-- 新建 folder 的 inline input：点 + 展开，Enter 提交，Esc 取消 -->
      <form
        v-if="creatingFolder"
        class="kb-sidebar-inline-form"
        @submit.prevent="submitCreateFolder"
      >
        <input
          ref="createInputRef"
          v-model="createDraft"
          class="kb-sidebar-inline-input"
          type="text"
          placeholder="文件夹名"
          maxlength="80"
          :disabled="busy"
          @keydown.esc.prevent="cancelCreateFolder"
        />
        <button
          class="kb-sidebar-inline-btn kb-sidebar-inline-primary"
          type="submit"
          :disabled="busy || !createDraft.trim()"
          aria-label="确认新建"
        >
          ✓
        </button>
        <button
          class="kb-sidebar-inline-btn"
          type="button"
          :disabled="busy"
          aria-label="取消新建"
          @click="cancelCreateFolder"
        >
          ✕
        </button>
      </form>
      <ul class="kb-sidebar-list">
        <li
          class="kb-sidebar-item"
          :class="{ 'kb-sidebar-item-active': activeKey === 'unfiled' }"
          @click="$emit('select', 'unfiled')"
        >
          <span class="kb-sidebar-folder-icon" aria-hidden="true">•</span>
          <span class="kb-sidebar-label">未归类</span>
          <span class="kb-sidebar-count">{{ unfiledCount }}</span>
        </li>
        <template v-for="folder in orderedFolders" :key="folder.id">
          <!-- 改名模式：item 整行被 inline input 替代，避免 click 与 input 冲突 -->
          <li
            v-if="renamingId === folder.id"
            class="kb-sidebar-item kb-sidebar-item-renaming"
            :style="{ paddingLeft: `${0.625 + folder.depth * 0.875}rem` }"
          >
            <span class="kb-sidebar-folder-icon" aria-hidden="true">▸</span>
            <form class="kb-sidebar-rename-form" @submit.prevent="submitRenameFolder(folder)">
              <input
                ref="renameInputRef"
                v-model="renameDraft"
                class="kb-sidebar-inline-input kb-sidebar-rename-input"
                type="text"
                maxlength="80"
                :disabled="busy"
                @keydown.esc.prevent="cancelRenameFolder"
                @click.stop
              />
              <button
                class="kb-sidebar-inline-btn kb-sidebar-inline-primary"
                type="submit"
                :disabled="busy || !renameDraft.trim() || renameDraft.trim() === folder.name"
                aria-label="确认改名"
                @click.stop
              >
                ✓
              </button>
              <button
                class="kb-sidebar-inline-btn"
                type="button"
                :disabled="busy"
                aria-label="取消改名"
                @click.stop="cancelRenameFolder"
              >
                ✕
              </button>
            </form>
          </li>
          <li
            v-else
            class="kb-sidebar-item kb-sidebar-item-folder"
            :class="{ 'kb-sidebar-item-active': activeKey === folder.key }"
            :style="{ paddingLeft: `${0.625 + folder.depth * 0.875}rem` }"
            @click="$emit('select', folder.key)"
          >
            <span class="kb-sidebar-folder-icon" aria-hidden="true">▸</span>
            <span class="kb-sidebar-label">{{ folder.name }}</span>
            <span class="kb-sidebar-count">{{ folder.documentCount || 0 }}</span>
            <span class="kb-sidebar-folder-actions" @click.stop>
              <button
                class="kb-sidebar-icon-btn kb-sidebar-action-btn"
                type="button"
                :title="`改名「${folder.name}」`"
                :aria-label="`改名「${folder.name}」`"
                :disabled="busy"
                @click.stop="startRenameFolder(folder)"
              >
                ✎
              </button>
              <button
                class="kb-sidebar-icon-btn kb-sidebar-action-btn kb-sidebar-action-danger"
                type="button"
                :title="`删除「${folder.name}」`"
                :aria-label="`删除「${folder.name}」`"
                :disabled="busy"
                @click.stop="confirmDeleteFolder(folder)"
              >
                ×
              </button>
            </span>
          </li>
        </template>
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
import { computed, nextTick, ref } from "vue";

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
  folders: {
    type: Array,
    default: () => [],
  },
  unfiledCount: {
    type: Number,
    default: -1,
  },
  activeKey: {
    type: String,
    default: "public",
  },
  hasPrivateAccess: {
    type: Boolean,
    default: true,
  },
  // busy: folder mutation 进行中互斥锁。父组件在 apiService 调用前后切换，
  // sidebar 据此 disable 所有按钮防止重复提交（emit 不能 await，所以必须由父组件控制）。
  busy: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["select", "create-folder", "rename-folder", "delete-folder"]);

// === Folder CRUD inline state ===
//
// 状态扁平：creatingFolder / renamingId 互斥（startCreate 时取消 rename，反之亦然）。
// 提交后立即收起 inline UI（创建/改名 emit 后乐观关闭），不等待父组件 ack；
// 失败时父组件可显示 alert，但 inline 状态已收起，避免悬挂。
const creatingFolder = ref(false);
const createDraft = ref("");
const createInputRef = ref(null);
const renamingId = ref(0);
const renameDraft = ref("");
const renameInputRef = ref(null);

const startCreateFolder = async () => {
  if (props.busy) return;
  cancelRenameFolder();
  creatingFolder.value = true;
  createDraft.value = "";
  await nextTick();
  createInputRef.value?.focus();
};

const cancelCreateFolder = () => {
  if (props.busy) return;
  creatingFolder.value = false;
  createDraft.value = "";
};

const submitCreateFolder = () => {
  const name = createDraft.value.trim();
  if (!name || props.busy) return;
  emit("create-folder", { name });
  // 乐观关闭 inline form：父组件 reload folders 后会拉到新条目
  creatingFolder.value = false;
  createDraft.value = "";
};

const startRenameFolder = async (folder) => {
  if (props.busy) return;
  cancelCreateFolder();
  renamingId.value = Number(folder.id);
  renameDraft.value = folder.name || "";
  await nextTick();
  // renameInputRef 在 v-for 中可能是数组（取第一个，因为同一时刻只渲染一个 rename input）
  const input = Array.isArray(renameInputRef.value) ? renameInputRef.value[0] : renameInputRef.value;
  input?.focus();
  input?.select?.();
};

const cancelRenameFolder = () => {
  if (props.busy) return;
  renamingId.value = 0;
  renameDraft.value = "";
};

const submitRenameFolder = (folder) => {
  const name = renameDraft.value.trim();
  if (!name || name === folder.name || props.busy) return;
  emit("rename-folder", { id: Number(folder.id), name });
  renamingId.value = 0;
  renameDraft.value = "";
};

const confirmDeleteFolder = (folder) => {
  if (props.busy) return;
  // 用 native confirm 简化：后端会在事务内把直接子项提升到父级。
  const ok = window.confirm(`确认删除文件夹「${folder.name}」吗？\n\n其中的文档和直接子文件夹会提升到父级目录，顶级目录会提升到未归类。`);
  if (!ok) return;
  emit("delete-folder", { id: Number(folder.id) });
};

// 客户端聚合 visibility，避免向后端新增专门的聚合端点（v0.1 1 天交付边界）。
const publicCount = computed(() =>
  props.documents.filter((d) => (d.visibility || d.scope) === "public").length
);

const privateCount = computed(() =>
  props.documents.filter((d) => (d.visibility || d.scope) === "private").length
);

const unfiledCount = computed(() => {
  if (props.unfiledCount >= 0) return props.unfiledCount;
  return props.documents.filter((d) => (d.visibility || d.scope) === "private" && !Number(d.folderId || 0)).length;
});

const orderedFolders = computed(() => {
  const flattenTree = (nodes, depth = 0, output = []) => {
    const sorted = [...nodes].sort((a, b) => {
      const sortDiff = Number(a.sortOrder || 0) - Number(b.sortOrder || 0);
      if (sortDiff !== 0) return sortDiff;
      return String(a.name || "").localeCompare(String(b.name || ""), "zh-CN");
    });
    for (const folder of sorted) {
      const id = Number(folder.id || 0);
      if (!id) continue;
      output.push({
        ...folder,
        id,
        key: `folder:${id}`,
        depth,
      });
      if (Array.isArray(folder.children) && folder.children.length > 0) {
        flattenTree(folder.children, depth + 1, output);
      }
    }
    return output;
  };

  if (props.folders.some((folder) => Array.isArray(folder.children) && folder.children.length > 0)) {
    return flattenTree(props.folders);
  }

  const byParent = new Map();
  for (const folder of props.folders) {
    const parentId = Number(folder.parentId || 0);
    if (!byParent.has(parentId)) byParent.set(parentId, []);
    byParent.get(parentId).push(folder);
  }
  for (const group of byParent.values()) {
    group.sort((a, b) => {
      const sortDiff = Number(a.sortOrder || 0) - Number(b.sortOrder || 0);
      if (sortDiff !== 0) return sortDiff;
      return String(a.name || "").localeCompare(String(b.name || ""), "zh-CN");
    });
  }

  const output = [];
  const visited = new Set();
  const walk = (parentId, depth) => {
    for (const folder of byParent.get(parentId) || []) {
      const id = Number(folder.id || 0);
      if (!id || visited.has(id)) continue;
      visited.add(id);
      output.push({
        ...folder,
        id,
        key: `folder:${id}`,
        depth,
      });
      walk(id, depth + 1);
    }
  };
  walk(0, 0);
  return output;
});

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
  font: 600 clamp(var(--fs-2xs), 0.78vw, var(--fs-xs)) var(--mono);
  color: var(--t3);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.kb-sidebar-section-tag {
  font: 500 clamp(var(--fs-3xs), 0.7vw, var(--fs-2xs)) var(--mono);
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
  font: clamp(var(--fs-sm), 0.95vw, var(--fs-lg)) var(--sans);
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
  font: var(--fs-xs) var(--mono);
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
  font: clamp(var(--fs-2xs), 0.78vw, var(--fs-xs)) var(--mono);
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
  font: 600 clamp(var(--fs-sm), 0.95vw, var(--fs-lg)) var(--mono);
  color: var(--t);
}

.kb-sidebar-stat-lb {
  font: clamp(var(--fs-2xs), 0.78vw, var(--fs-xs)) var(--mono);
  color: var(--t3);
  letter-spacing: 0.04em;
}

/* === Folder CRUD: 新建按钮 + 行内编辑表单 + folder hover 操作（v0.2） === */
.kb-sidebar-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  padding: 0;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.03);
  color: var(--t2);
  font: 600 14px var(--sans);
  line-height: 1;
  cursor: pointer;
  transition: color 0.15s ease, background-color 0.15s ease, border-color 0.15s ease;
  flex-shrink: 0;
}

.kb-sidebar-icon-btn:hover:not(:disabled) {
  color: var(--t);
  background: rgba(220, 155, 90, 0.12);
  border-color: rgba(220, 155, 90, 0.35);
}

.kb-sidebar-icon-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

/* 行内 form：新建 / 改名共用一套样式 */
.kb-sidebar-inline-form {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.5rem;
  border: 1px solid rgba(220, 155, 90, 0.25);
  border-radius: var(--radius-sm);
  background: rgba(220, 155, 90, 0.04);
}

.kb-sidebar-inline-input {
  flex: 1;
  min-width: 0;
  font: var(--fs-sm) var(--sans);
  color: var(--t);
  background: transparent;
  border: none;
  outline: none;
  padding: 2px 4px;
}

.kb-sidebar-inline-input::placeholder {
  color: var(--t3);
}

.kb-sidebar-inline-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.03);
  color: var(--t2);
  font: 600 12px var(--sans);
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s ease;
}

.kb-sidebar-inline-btn:hover:not(:disabled) {
  color: var(--t);
  background: rgba(255, 255, 255, 0.06);
}

.kb-sidebar-inline-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.kb-sidebar-inline-primary {
  border-color: rgba(220, 155, 90, 0.4);
  color: rgba(220, 155, 90, 0.95);
}

.kb-sidebar-inline-primary:hover:not(:disabled) {
  background: rgba(220, 155, 90, 0.15);
  border-color: rgba(220, 155, 90, 0.6);
}

/* 改名 inline form：嵌在 folder item 里，去掉外部 padding */
.kb-sidebar-rename-form {
  display: flex;
  flex: 1;
  align-items: center;
  gap: 0.375rem;
  min-width: 0;
}

.kb-sidebar-rename-input {
  font: clamp(var(--fs-sm), 0.95vw, var(--fs-lg)) var(--sans);
}

.kb-sidebar-item-renaming {
  background: rgba(220, 155, 90, 0.06);
}

/* folder hover 显示操作按钮 */
.kb-sidebar-folder-actions {
  display: none;
  align-items: center;
  gap: 0.25rem;
  margin-left: 0.25rem;
  flex-shrink: 0;
}

.kb-sidebar-item-folder:hover .kb-sidebar-folder-actions,
.kb-sidebar-item-folder:focus-within .kb-sidebar-folder-actions {
  display: inline-flex;
}

.kb-sidebar-action-btn {
  width: 20px;
  height: 20px;
  font-size: 11px;
}

.kb-sidebar-action-danger {
  color: rgba(239, 138, 115, 0.85);
}

.kb-sidebar-action-danger:hover:not(:disabled) {
  color: rgba(239, 138, 115, 1);
  background: rgba(239, 102, 96, 0.12);
  border-color: rgba(239, 102, 96, 0.4);
}
</style>
