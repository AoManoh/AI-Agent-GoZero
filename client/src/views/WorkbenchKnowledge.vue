<template>
  <!--
    WorkbenchKnowledge：私人知识库管理（v0.1 实现）。

    布局（响应式三栏，约束 R）:
      - xs/sm (<768px): 单列堆叠，左栏暂直接堆在中栏上方（v0.2 改为 drawer）
      - md (768-1024px): 三栏紧凑（左 minmax(200px,220px) / 中 1fr / 右 minmax(280px,320px)）
      - lg (1024-1440px): 三栏标准（左 minmax(240px,260px) / 中 1fr / 右 minmax(320px,360px)）
      - xl (≥1440px): 三栏宽松（左 minmax(280px,300px) / 中 minmax(0,1200px) / 右 minmax(360px,400px)）

    后端契约（2026-05-12 v0.1 已就位）:
      - GET /api/ai/knowledge/documents → KnowledgeDocumentItem[]，含 visibility / sizeBytes / embeddingDimension / embeddingModel
      - GET /api/ai/knowledge/folders → 当前用户目录树
      - POST /api/ai/knowledge/upload → 任何登录用户可上传，admin 上传 visibility=public，普通 user → private（Q7=B 角色路由），可选 folderId
      - GET /api/ai/knowledge/documents/:id/chunks → 单文档 chunks 数组

    v0.3 已并入:
      - reader 模式（chunks concat + marked 渲染 + DOMPurify 净化），点击卡片「全文阅读」按钮触发，
        中栏 mode 切换 'list' → 'reader'，右栏 tabs 在 reader 模式下隐藏（专注阅读）。
        chunks 一次拉满（limit=500）保证完整长度，渲染前用 DOMPurify 净化避免 XSS。

    决策来源: docs/requirements/2026-05-12-workbench-knowledge-base-redesign.md §7.1 F7
  -->
  <WorkbenchLayout>
    <div class="wb-kb-content">
      <section class="wb-kb-hero">
        <div class="wb-eyebrow">
          <span class="wb-eyebrow-dot" aria-hidden="true"></span>
          <span>知识库</span>
        </div>
        <div class="wb-kb-hero-row">
          <div class="wb-kb-hero-text">
            <h1 class="wb-kb-title">让 AI 学你的知识</h1>
            <p class="wb-kb-sub">上传文档后，AI 会基于检索增强 (RAG) 在面试时引用你提供的资料。</p>
          </div>
          <button type="button" class="wb-kb-upload-btn" @click="triggerUpload">
            <span class="wb-kb-upload-plus" aria-hidden="true">+</span>
            <span>上传到我的私人资料库</span>
          </button>
          <input
            ref="uploadInputRef"
            type="file"
            accept=".pdf"
            multiple
            class="wb-kb-file-input"
            @change="handleUpload"
          />
        </div>
      </section>

      <div class="wb-kb-shell">
        <!-- 左：知识范围 + 目录树（含 v0.2 folder CRUD） -->
        <KnowledgeSidebar
          mode="folder-tree"
          :documents="sidebarDocuments"
          :folders="folders"
          :unfiled-count="folderMeta.unfiledCount"
          :active-key="activeKey"
          :has-private-access="hasPrivateAccess"
          :busy="folderMutating"
          @select="handleSidebarSelect"
          @create-folder="handleCreateFolder"
          @rename-folder="handleRenameFolder"
          @delete-folder="handleDeleteFolder"
        />

        <!-- 阅读进度条：reader 模式下 fixed 顶部，跨整页宽度，跟随 window scroll；非 reader 模式不渲染 -->
        <Teleport to="body">
          <div
            v-if="viewMode === 'reader'"
            class="wb-kb-reader-progress"
            role="progressbar"
            :aria-valuenow="readerProgress"
            aria-valuemin="0"
            aria-valuemax="100"
            :aria-label="`阅读进度 ${readerProgress}%`"
          >
            <div
              class="wb-kb-reader-progress-bar"
              :style="{ width: `${readerProgress}%` }"
            ></div>
          </div>
        </Teleport>

        <!-- 中：文档列表 / Reader 模式（v0.3 F7） -->
        <section
          class="wb-kb-list"
          :class="{ 'wb-kb-list-reader': viewMode === 'reader' }"
          ref="readerScrollRef"
        >
          <!-- Reader 模式：跨两列全屏阅读 -->
          <template v-if="viewMode === 'reader' && selectedDoc">
            <header class="wb-kb-reader-head">
              <button
                type="button"
                class="wb-kb-reader-back"
                aria-label="返回文档列表"
                @click="exitReaderMode"
              >
                ← 返回
              </button>
              <div class="wb-kb-reader-meta">
                <h3 class="wb-kb-reader-title">{{ selectedDoc.name }}</h3>
                <div class="wb-kb-reader-sub">
                  <span>{{ selectedDoc.uploadedAt }}</span>
                  <span aria-hidden="true">·</span>
                  <span>{{ formatBytes(selectedDoc.sizeBytes) }}</span>
                  <span aria-hidden="true">·</span>
                  <span>{{ selectedDoc.chunkCount }} 片段</span>
                </div>
              </div>
            </header>

            <div v-if="readerLoading" class="wb-kb-reader-state">加载文档全文中…</div>
            <div v-else-if="readerError" class="wb-kb-reader-state wb-kb-reader-error">
              {{ readerError }}
            </div>
            <template v-else>
              <article
                v-if="readerHtml"
                class="wb-kb-reader-body"
                v-html="readerHtml"
              ></article>
              <div v-else class="wb-kb-reader-state">这份文档没有可阅读的内容。</div>
              <!-- §7.1 F7 备选：阅读时也能临时查看切片质检（默认折叠，details/summary 原生交互） -->
              <details
                v-if="readerChunksList.length > 0"
                class="wb-kb-reader-chunks"
              >
                <summary class="wb-kb-reader-chunks-summary">
                  查看切片质检（共 {{ readerChunksList.length }} 条）
                </summary>
                <ol class="wb-kb-reader-chunks-list">
                  <li
                    v-for="(chunk, i) in readerChunksList"
                    :key="`rc-${i}`"
                    class="wb-kb-reader-chunks-item"
                  >
                    <span class="wb-kb-reader-chunks-num">#{{ i + 1 }}</span>
                    <p class="wb-kb-reader-chunks-text">{{ chunk }}</p>
                  </li>
                </ol>
              </details>
            </template>
          </template>

          <!-- List 模式（默认） -->
          <template v-else>
          <header class="wb-block-head">
            <h3 class="wb-block-title">{{ activeScopeLabel }}</h3>
            <span class="wb-block-meta">{{ filteredDocs.length }} 份文档</span>
          </header>

          <div v-if="filteredDocs.length > 0" class="wb-kb-docs">
            <article
              v-for="doc in filteredDocs"
              :key="doc.id"
              class="wb-kb-doc"
              :class="{ 'wb-kb-doc-active': selectedDocId === doc.id }"
              @click="selectedDocId = doc.id"
            >
              <div class="wb-kb-doc-icon" aria-hidden="true">{{ getDocTypeLabel(doc.type) }}</div>
              <div class="wb-kb-doc-meta">
                <h4 class="wb-kb-doc-name">{{ doc.name }}</h4>
                <div class="wb-kb-doc-info">
                  <template v-for="(item, i) in buildDocChips(doc)" :key="`chip-${doc.id}-${i}`">
                    <span v-if="i > 0" class="wb-kb-doc-info-sep" aria-hidden="true">·</span>
                    <span class="wb-kb-doc-info-chip">{{ item }}</span>
                  </template>
                </div>
              </div>
              <div class="wb-kb-doc-actions">
                <!-- v0.2：仅私人文档显示「移动到目录」select；公共文档由 admin 管理，不走此 UI -->
                <label
                  v-if="(doc.visibility || doc.scope) === 'private'"
                  class="wb-kb-doc-move"
                  :title="docMoveHint"
                  @click.stop
                >
                  <span class="wb-kb-doc-move-icon" aria-hidden="true">📁</span>
                  <select
                    class="wb-kb-doc-move-select"
                    :value="String(doc.folderId || 0)"
                    :disabled="docMoving === doc.id"
                    @change="handleMoveDocument(doc, $event.target.value)"
                  >
                    <option value="0">未归类</option>
                    <option
                      v-for="folder in folderOptions"
                      :key="folder.id"
                      :value="String(folder.id)"
                    >
                      {{ `${"　".repeat(folder.depth || 0)}${folder.name}` }}
                    </option>
                  </select>
                </label>
                <button
                  type="button"
                  class="wb-kb-doc-reader-btn"
                  :title="readerButtonHint"
                  :disabled="doc.status !== 'ready'"
                  @click.stop="enterReaderMode(doc)"
                >
                  全文阅读
                </button>
                <span class="wb-kb-doc-status">
                  <span class="wb-kb-doc-dot" :class="`wb-status-${doc.status}`" aria-hidden="true"></span>
                  {{ getDocStatusLabel(doc.status) }}
                </span>
              </div>
            </article>
          </div>

          <div v-else class="wb-empty">
            <div class="wb-empty-title">{{ emptyStateTitle }}</div>
            <div class="wb-empty-sub">{{ emptyStateSub }}</div>
            <button v-if="emptyStateCTA" type="button" class="wb-empty-cta" @click="triggerUpload">
              + 上传文档
            </button>
          </div>
          </template>
        </section>

        <!-- 右：详情 + 双 tab（Chunks 预览 / Test query 召回）；reader 模式下隐藏让中栏专注阅读 -->
        <aside v-show="viewMode === 'list'" class="wb-kb-detail">
          <!-- Tab 切换栏：在 detail-head 上方，跨整列宽 -->
          <nav class="wb-kb-tabs" role="tablist" aria-label="知识库右栏视图切换">
            <button
              v-for="tab in tabDefs"
              :key="tab.id"
              type="button"
              class="wb-kb-tab"
              :class="{ 'wb-kb-tab-active': activeTab === tab.id }"
              role="tab"
              :aria-selected="activeTab === tab.id"
              @click="activeTab = tab.id"
            >
              {{ tab.label }}
            </button>
          </nav>

          <!-- Tab 1: Chunks 预览（默认） -->
          <div v-if="activeTab === 'chunks'" class="wb-kb-tabpanel" role="tabpanel">
            <div v-if="selectedDoc" class="wb-kb-detail-inner">
              <header class="wb-kb-detail-head">
                <div class="wb-kb-detail-icon" aria-hidden="true">{{ getDocTypeLabel(selectedDoc.type) }}</div>
                <div class="wb-kb-detail-meta">
                  <h4 class="wb-kb-detail-name">{{ selectedDoc.name }}</h4>
                  <div class="wb-kb-detail-info">{{ selectedDoc.uploadedAt }} · {{ formatBytes(selectedDoc.sizeBytes) }}</div>
                </div>
              </header>

              <div class="wb-kb-detail-stats">
                <div class="wb-kb-detail-stat">
                  <span class="wb-kb-detail-stat-num">{{ selectedDoc.chunkCount }}</span>
                  <span class="wb-kb-detail-stat-lb">片段</span>
                </div>
                <div class="wb-kb-detail-stat">
                  <span class="wb-kb-detail-stat-num">{{ selectedDoc.embeddingDimension || "—" }}</span>
                  <span class="wb-kb-detail-stat-lb">维度</span>
                </div>
                <div class="wb-kb-detail-stat">
                  <span class="wb-kb-detail-stat-num">v{{ selectedDoc.version || 1 }}</span>
                  <span class="wb-kb-detail-stat-lb">版本</span>
                </div>
              </div>

              <div v-if="selectedDoc.embeddingModel" class="wb-kb-detail-meta-line">
                <span class="wb-detail-label">向量模型</span>
                <span class="wb-kb-detail-meta-value">{{ selectedDoc.embeddingModel }}</span>
              </div>

              <div class="wb-kb-detail-block">
                <div class="wb-detail-label">片段预览（Chunks）</div>
                <div v-if="selectedDoc.chunks?.length > 0" class="wb-kb-chunks">
                  <div
                    v-for="(chunk, i) in selectedDoc.chunks"
                    :key="`chunk-${i}`"
                    class="wb-kb-chunk"
                  >
                    <div class="wb-kb-chunk-num">#{{ String(i + 1).padStart(2, '0') }}</div>
                    <p class="wb-kb-chunk-text">{{ chunk }}</p>
                  </div>
                </div>
                <div v-else-if="selectedDoc.status === 'processing'" class="wb-kb-chunks-empty">
                  文档正在解析中…
                </div>
                <div v-else-if="selectedDoc.status === 'failed'" class="wb-kb-chunks-empty">
                  文档解析失败，请重新上传
                </div>
                <div v-else class="wb-kb-chunks-empty">点击文档卡片以加载片段。</div>
              </div>
            </div>

            <div v-else class="wb-kb-detail-empty">
              <div class="wb-empty-title">选择一份文档</div>
              <div class="wb-empty-sub">查看片段预览、向量维度和元信息。</div>
            </div>
          </div>

          <!-- Tab 2: Test query 召回 -->
          <div v-else-if="activeTab === 'test-query'" class="wb-kb-tabpanel" role="tabpanel">
            <div class="wb-kb-test-block">
              <label class="wb-detail-label" for="kb-test-query-input">召回测试 Query</label>
              <div class="wb-kb-test-input-row">
                <input
                  id="kb-test-query-input"
                  v-model="testQuery"
                  type="text"
                  class="wb-kb-test-input"
                  placeholder="输入一段话或问题，验证向量检索 TopK 命中"
                  :disabled="testLoading"
                  @keydown.enter.prevent="runTestQuery"
                  @keydown.ctrl.enter.prevent="runTestQuery"
                />
                <select
                  v-model.number="testTopK"
                  class="wb-kb-test-topk"
                  :disabled="testLoading"
                  aria-label="TopK"
                >
                  <option :value="3">TopK 3</option>
                  <option :value="5">TopK 5</option>
                  <option :value="10">TopK 10</option>
                </select>
                <button
                  type="button"
                  class="wb-kb-test-btn"
                  :disabled="testLoading || !testQuery.trim()"
                  @click="runTestQuery"
                >
                  {{ testLoading ? "检索中…" : "检索" }}
                </button>
              </div>
              <div class="wb-kb-test-meta">
                <span>当前知识范围：{{ activeScopeLabel }}</span>
                <span aria-hidden="true">·</span>
                <span>TopK = {{ testTopK }}</span>
                <span v-if="testError" class="wb-kb-test-error">{{ testError }}</span>
              </div>
            </div>

            <div class="wb-kb-detail-block">
              <div class="wb-detail-label">召回结果</div>
              <div v-if="testResults.length > 0" class="wb-kb-chunks">
                <div
                  v-for="(item, i) in testResults"
                  :key="`tr-${i}-${item.chunkId}`"
                  class="wb-kb-chunk wb-kb-test-result"
                >
                  <div class="wb-kb-test-result-head">
                    <span class="wb-kb-chunk-num">#{{ String(i + 1).padStart(2, '0') }}</span>
                    <span
                      class="wb-kb-score-chip"
                      :class="`wb-kb-score-${getScoreLevel(item.score)}`"
                      :title="`相似度 ${formatScore(item.score)}`"
                    >
                      {{ formatScore(item.score) }}
                    </span>
                  </div>
                  <h5 v-if="item.title" class="wb-kb-test-result-title">{{ item.title }}</h5>
                  <p class="wb-kb-chunk-text">{{ item.content }}</p>
                </div>
              </div>
              <div v-else-if="testQueriedOnce" class="wb-kb-chunks-empty">
                没有找到与查询相关的片段。请尝试不同的关键词或先上传更多资料。
              </div>
              <div v-else class="wb-kb-chunks-empty">
                输入查询后，向量检索会返回 TopK 相似的 chunks 与相似度分数。
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { marked } from "marked";
import DOMPurify from "dompurify";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
import KnowledgeSidebar from "../components/workbench/KnowledgeSidebar.vue";
import { apiService } from "../composables/useApi";
import { useAuth } from "../composables/useAuth";

const { isAuthenticated } = useAuth();

// marked 全局开关：开 GFM + breaks（对中文 PDF 解析后的弱结构 markdown 友好）
marked.setOptions({ gfm: true, breaks: true });

// === 上传 ===
const uploadInputRef = ref(null);
const triggerUpload = () => uploadInputRef.value?.click();

const handleUpload = async (e) => {
  const files = Array.from(e.target.files || []);
  if (files.length === 0) return;
  let anySuccess = false;
  for (const file of files) {
    try {
      if (!String(file.name || "").toLowerCase().endsWith(".pdf")) {
        throw new Error("当前仅支持 PDF 文件");
      }
      const formData = new FormData();
      formData.append("file", file);
      if (activeFolderId.value > 0) {
        formData.append("folderId", String(activeFolderId.value));
      }
      await apiService.chat.knowledgeUpload(formData);
      anySuccess = true;
    } catch (error) {
      // 静默：上传失败不阻塞 UI；后续可加 toast
    }
  }
  if (uploadInputRef.value) {
    uploadInputRef.value.value = "";
  }
  // 只要有成功上传，重拉文档列表以后端返回为准。
  if (anySuccess) {
    await Promise.all([loadFolders(), loadSidebarDocuments(), loadDocuments()]);
  }
};

// 字节数格式化：暴露给模板使用 selectedDoc.sizeBytes
const formatBytes = (bytes) => {
  if (!bytes || bytes <= 0) return "—";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
};

const inferFileType = (name) => {
  const ext = name.toLowerCase().match(/\.(pdf|docx?|md|txt)$/);
  if (!ext) return "file";
  return ext[1] === "doc" ? "docx" : ext[1];
};

// === 知识范围与目录树 ===
//
// activeKey:
//   - 'public'：公共知识
//   - 'private'：我的私人资料，跨目录查看
//   - 'unfiled'：私人资料中 folder_id 为空的未归类文档
//   - 'folder:{id}'：指定目录
const activeKey = ref("public");

const hasPrivateAccess = computed(() => isAuthenticated.value);

const handleSidebarSelect = (key) => {
  activeKey.value = key;
};

// === v0.2 Folder CRUD handlers ===
//
// folderMutating 互斥锁：任一 create/rename/delete 进行中时禁用 sidebar 内全部操作按钮，
// 避免并发请求与目录列表的 race condition。每个 handler 在进入时拉锁、finally 释放，
// 成功后并行 reload folders + sidebarDocuments + 主列表（如目标目录被影响）。
const folderMutating = ref(false);

const handleCreateFolder = async ({ name }) => {
  if (folderMutating.value) return;
  folderMutating.value = true;
  try {
    await apiService.chat.knowledgeCreateFolder({ name });
    await Promise.all([loadFolders(), loadSidebarDocuments()]);
  } catch (error) {
    console.error("创建文件夹失败:", error);
    window.alert(`创建文件夹失败：${error?.message || "请稍后重试"}`);
  } finally {
    folderMutating.value = false;
  }
};

const handleRenameFolder = async ({ id, name }) => {
  if (folderMutating.value) return;
  folderMutating.value = true;
  try {
    await apiService.chat.knowledgeUpdateFolder(id, { name });
    await Promise.all([loadFolders(), loadSidebarDocuments()]);
  } catch (error) {
    console.error("重命名文件夹失败:", error);
    window.alert(`重命名文件夹失败：${error?.message || "请稍后重试"}`);
  } finally {
    folderMutating.value = false;
  }
};

const handleDeleteFolder = async ({ id }) => {
  if (folderMutating.value) return;
  folderMutating.value = true;
  try {
    // 后端只允许删除空目录，避免文档或子目录被隐式移动。
    await apiService.chat.knowledgeDeleteFolder(id);
    // 如果当前选中的是被删除的目录，回退到「我的私人资料」
    if (activeKey.value === `folder:${id}`) {
      activeKey.value = "private";
    }
    await Promise.all([loadFolders(), loadSidebarDocuments(), loadDocuments()]);
  } catch (error) {
    console.error("删除文件夹失败:", error);
    window.alert(`删除文件夹失败：${error?.message || "请稍后重试"}`);
  } finally {
    folderMutating.value = false;
  }
};

// === 文档移动到目录（v0.2）===
//
// docMoving 记录当前正在移动的文档 ID，避免用户快速切换 select 时产生并发请求。
// select 选中即触发，乐观等后端 200 + reload；失败时 alert 提示并保持 select 显示旧值（因为 value 绑定 doc.folderId，reload 会纠正）。
const docMoving = ref("");
const docMoveHint = "移动到目录（目录必须先在左栏创建）";

const handleMoveDocument = async (doc, newFolderIdRaw) => {
  const newFolderId = Number(newFolderIdRaw || 0);
  const currentFolderId = Number(doc.folderId || 0);
  if (newFolderId === currentFolderId) return; // 无变化
  if (docMoving.value) return;

  docMoving.value = doc.id;
  try {
    await apiService.chat.knowledgeMoveDocumentFolder(Number(doc.id), { folderId: newFolderId });
    await Promise.all([loadFolders(), loadSidebarDocuments(), loadDocuments()]);
  } catch (error) {
    console.error("移动文档失败:", error);
    window.alert(`移动文档失败：${error?.message || "请稍后重试"}`);
  } finally {
    docMoving.value = "";
  }
};

const matchVisibility = (doc, key) => {
  const v = (doc.visibility || doc.scope || "").toLowerCase();
  if (key === "public") return v === "public";
  if (key === "private") return v === "private";
  if (key === "unfiled") return v === "private" && !Number(doc.folderId || 0);
  if (key.startsWith("folder:")) {
    return v === "private" && Number(doc.folderId || 0) === Number(key.slice("folder:".length));
  }
  return true;
};

const activeScopeLabel = computed(() => {
  if (activeKey.value === "private") return "我的私人资料";
  if (activeKey.value === "unfiled") return "未归类";
  if (activeKey.value.startsWith("folder:")) {
    const id = Number(activeKey.value.slice("folder:".length));
    const folder = folderOptions.value.find((item) => Number(item.id) === id);
    return folder?.name || "目录";
  }
  return "公共知识";
});

const activeFolderId = computed(() => {
  if (!activeKey.value.startsWith("folder:")) return 0;
  const id = Number(activeKey.value.slice("folder:".length));
  return Number.isFinite(id) && id > 0 ? id : 0;
});

const buildScopeParams = () => {
  if (activeKey.value === "public") {
    return { visibility: "public" };
  }
  if (activeKey.value === "private") {
    return { visibility: "private" };
  }
  if (activeKey.value === "unfiled") {
    return { visibility: "private", folderScoped: true, folderId: 0 };
  }
  if (activeFolderId.value > 0) {
    return { visibility: "private", folderScoped: true, folderId: activeFolderId.value };
  }
  return {};
};

// === 文档（仅来自后端，无前端 mock） ===
const documents = ref([]);
const sidebarDocuments = ref([]);
const folders = ref([]);
const folderMeta = ref({ unfiledCount: 0, totalCount: 0, initialized: false });

const flattenFolders = (nodes, depth = 0, output = []) => {
  for (const node of nodes || []) {
    const id = Number(node.id || 0);
    if (!id) continue;
    output.push({
      ...node,
      id,
      depth,
    });
    if (Array.isArray(node.children) && node.children.length > 0) {
      flattenFolders(node.children, depth + 1, output);
    }
  }
  return output;
};

const folderOptions = computed(() => flattenFolders(folders.value));

const filteredDocs = computed(() =>
  documents.value.filter((d) => matchVisibility(d, activeKey.value))
);

const selectedDocId = ref("");

const selectedDoc = computed(() => {
  if (!selectedDocId.value) return null;
  return documents.value.find((d) => d.id === selectedDocId.value) || null;
});

// 后端 status (string) → 本地状态标
const mapKnowledgeStatus = (raw) => {
  if (!raw) return "processing";
  const s = String(raw).toLowerCase();
  if (s.includes("ready") || s.includes("success") || s.includes("active")) return "ready";
  if (s.includes("fail") || s.includes("error")) return "failed";
  return "processing";
};

// 绝对时间戳 → 相对时间
const formatRelativeTime = (timestamp) => {
  if (!timestamp) return "近期";
  const ts = typeof timestamp === "number" ? timestamp : new Date(timestamp).getTime();
  if (Number.isNaN(ts)) return "近期";
  const diff = Date.now() - ts;
  const min = 60 * 1000;
  const hour = 60 * min;
  const day = 24 * hour;
  if (diff < hour) return `${Math.max(1, Math.floor(diff / min))} 分钟前`;
  if (diff < day) return `${Math.floor(diff / hour)} 小时前`;
  if (diff < 2 * day) return "昨天";
  if (diff < 7 * day) return `${Math.floor(diff / day)} 天前`;
  if (diff < 30 * day) return `${Math.floor(diff / (7 * day))} 周前`;
  return new Date(ts).toLocaleDateString("zh-CN");
};

// 卡片底部 chip 排版（Q8=B）：
//   "{embeddingDimension}d · {embeddingModel} · {sizeBytes 格式化} · {chunkCount} chunks · {uploadedAt}"
//   字段缺失时跳过（embeddingDimension=0 / embeddingModel='' 时不渲染）
const buildDocChips = (doc) => {
  const chips = [];
  if (doc.embeddingDimension > 0) chips.push(`${doc.embeddingDimension}d`);
  if (doc.embeddingModel) chips.push(doc.embeddingModel);
  if (doc.sizeBytes > 0) chips.push(formatBytes(doc.sizeBytes));
  if (doc.chunkCount > 0) chips.push(`${doc.chunkCount} 片段`);
  if (doc.uploadedAt) chips.push(doc.uploadedAt);
  return chips;
};

// 中栏空态文案：依据 activeKey 给出不同提示
const emptyStateTitle = computed(() => {
  if (activeKey.value === "private") return "私人资料库还没有文档";
  if (activeKey.value === "unfiled") return "还没有未归类文档";
  if (activeKey.value.startsWith("folder:")) return "这个目录还没有文档";
  return "公共知识库为空";
});

const emptyStateSub = computed(() => {
  if (activeKey.value !== "public") {
    return "上传 PDF 资料，AI 会自动切片并向量化，作为面试时的私人 RAG 数据源。";
  }
  return "公共知识库由管理员维护，当前没有可阅读的内容。";
});

// 仅在 private 视图下显示「上传」CTA（公共知识库由 admin 通过同端点上传，本子页 UI 不暴露 admin 切换 — Q7=C）
const emptyStateCTA = computed(() => activeKey.value !== "public");

// === Reader 模式（v0.3 F7）===
//
// viewMode='list' 默认显示文档列表 + 右栏双 tab；
// viewMode='reader' 切到中栏 reader 全屏（覆盖文档列表与右栏 tabs），渲染 chunks concat 后的 markdown。
//
// chunks 拉满 limit=500（v0.1 lazy 拉时只拉 6 条预览），保证 reader 完整性；
// 已拉过 chunks 的文档不重拉（fullChunksLoaded 标记）。
const viewMode = ref("list");
const readerLoading = ref(false);
const readerError = ref("");
const readerScrollRef = ref(null);
// 阅读进度：0-100 整数百分比。基于 scroll 容器的 scrollTop / (scrollHeight - clientHeight) 换算。
// 非滚动态（内容不足一屏）保持 100%，避免顶部进度条显得永远不满。
const readerProgress = ref(0);

const readerButtonHint = "全文阅读：把所有切片按顺序拼接后用 markdown 渲染";

// 注意：mapKnowledgeDocument 的 chunks 字段在 v0.1 仅用 preview 占位（一条），
// reader 模式需要全部 chunks 内容，因此用独立的 fullChunksByDoc Map 做缓存，避免冲掉列表预览。
const fullChunksByDoc = ref(new Map());

// reader 阅读位置持久化：同 tab 内（sessionStorage）按文档 id 记忆 scrollY，下次进入同一文档自动恢复。
// 跨 tab 不共享避免不同窗口干扰；用户退出登录或关闭页签后自然清空。
const READER_SCROLL_KEY_PREFIX = "wb-kb-reader-scroll-";
const readReaderScrollFor = (docId) => {
  try {
    const raw = sessionStorage.getItem(READER_SCROLL_KEY_PREFIX + docId);
    if (!raw) return 0;
    const n = Number(raw);
    return Number.isFinite(n) && n >= 0 ? n : 0;
  } catch {
    return 0;
  }
};
const writeReaderScrollFor = (docId, top) => {
  try {
    sessionStorage.setItem(READER_SCROLL_KEY_PREFIX + docId, String(Math.max(0, Math.round(top))));
  } catch {
    // 隐私模式 / 配额满 → 静默降级，不影响阅读
  }
};

const enterReaderMode = async (doc) => {
  if (!doc?.id) return;
  selectedDocId.value = doc.id;
  viewMode.value = "reader";
  readerError.value = "";
  readerProgress.value = 0;
  // 上一次阅读到的 scrollY；首次进入或新文档 → 0
  const restoreTop = readReaderScrollFor(doc.id);
  await nextTick();
  window.scrollTo({ top: restoreTop, behavior: "auto" });
  // 已缓存则跳过 fetch
  if (fullChunksByDoc.value.has(doc.id)) {
    await nextTick();
    // chunks 已就位，再尝试恢复一次（避免 reader-body 已经渲染时 nextTick 没等到布局）
    window.scrollTo({ top: restoreTop, behavior: "auto" });
    updateReaderProgress();
    return;
  }
  readerLoading.value = true;
  try {
    const res = await apiService.chat.knowledgeDocumentChunks(doc.id, { limit: 500 });
    const chunks = Array.isArray(res?.chunks) ? res.chunks : [];
    // 按 createdAt 升序拼接（后端已 ORDER BY created_at ASC）；保留每个 chunk 的 content 原文
    fullChunksByDoc.value.set(doc.id, chunks.map((c) => c?.content || "").filter(Boolean));
  } catch (error) {
    console.error("加载文档全文失败:", error);
    readerError.value = error?.message || "加载文档全文失败，请稍后重试";
  } finally {
    readerLoading.value = false;
    await nextTick();
    // 内容渲染完成后再恢复一次，确保 scrollHeight 已经撑开
    window.scrollTo({ top: restoreTop, behavior: "auto" });
    updateReaderProgress();
  }
};

const exitReaderMode = () => {
  // 兜底写入最后一次 scrollY，避免 throttle 边缘的最后一次微小位移没被持久化
  const docId = selectedDocId.value;
  if (docId) {
    writeReaderScrollFor(docId, window.scrollY || document.documentElement.scrollTop || 0);
  }
  viewMode.value = "list";
  readerProgress.value = 0;
};

// 直接读 document.documentElement 计算 window-level scroll 进度。
// .wb-kb-list 不是独立 scroll 容器（无 overflow-y），整个页面在 page level 滚动，所以监听对象是 window。
const updateReaderProgress = () => {
  const root = document.documentElement;
  const max = root.scrollHeight - root.clientHeight;
  if (max <= 0) {
    // 内容不足一屏：视为已读完，进度条满，不影响主视觉
    readerProgress.value = 100;
    return;
  }
  const top = window.scrollY || root.scrollTop || 0;
  const ratio = Math.min(1, Math.max(0, top / max));
  readerProgress.value = Math.round(ratio * 100);
};

// 写入持久化的 throttle：scroll 事件高频触发，每 250ms 写一次 sessionStorage 即可。
// 用 timestamp 比较而非 setTimeout，避免 timer 累积；最后一次滚动也由 exitReaderMode 兜底。
let lastReaderScrollSaveAt = 0;
const handleReaderScroll = () => {
  if (viewMode.value !== "reader") return;
  updateReaderProgress();
  const docId = selectedDocId.value;
  if (!docId) return;
  const now = Date.now();
  if (now - lastReaderScrollSaveAt < 250) return;
  lastReaderScrollSaveAt = now;
  writeReaderScrollFor(docId, window.scrollY || document.documentElement.scrollTop || 0);
};

// 把 chunks 数组拼成 markdown 字符串。chunks 之间用空行分隔，避免段落黏合。
const readerMarkdown = computed(() => {
  if (!selectedDoc.value) return "";
  const chunks = fullChunksByDoc.value.get(selectedDoc.value.id);
  if (!Array.isArray(chunks) || chunks.length === 0) return "";
  return chunks.join("\n\n");
});

// 渲染前用 DOMPurify 净化，禁止 inline event 与 javascript: 协议；保留常见 markdown 安全标签。
const readerHtml = computed(() => {
  const md = readerMarkdown.value;
  if (!md) return "";
  const rawHtml = marked.parse(md);
  return DOMPurify.sanitize(rawHtml, {
    USE_PROFILES: { html: true },
    FORBID_ATTR: ["onerror", "onload", "onclick"],
  });
});

// 切换 visibility/folder 或选中其他文档时退出 reader 防止显示陈旧内容
watch(activeKey, () => {
  if (viewMode.value === "reader") exitReaderMode();
});

// reader 模式 chunks 切片折叠面板（需求文档 §7.1 F7 备选：阅读时也能临时查看切片质检）
//
// 用 fullChunksByDoc 现成缓存，不再额外 fetch；首次打开自动展开，复用 details/summary 原生交互。
const readerChunksList = computed(() => {
  if (!selectedDoc.value) return [];
  return fullChunksByDoc.value.get(selectedDoc.value.id) || [];
});

// reader 模式键盘快捷键：Esc 退出阅读
//
// 用 window 级 keydown 监听，避免 reader 容器 focus 不在按钮上时按键失效。
// 仅在 viewMode='reader' 时响应，其他模式下保持系统默认行为（如关闭浏览器搜索框）。
const handleReaderKeydown = (event) => {
  if (viewMode.value !== "reader") return;
  if (event.key === "Escape") {
    event.preventDefault();
    exitReaderMode();
  }
};

// === 右栏双 tab：Chunks 预览 / Test query 召回（F4） ===
//
// 决策来源 §7.1 F4：
//   - 默认 tab='chunks'，命中文档详情 + 片段预览（保持现有交互不变）
//   - tab='test-query' 调用 POST /api/ai/knowledge/test-query 验证当前 visibility 范围内的 TopK 检索
//   - score chip 三色阈值：≥0.85 高（绿） / 0.70-0.85 中（金） / <0.70 低（红），配合 hover title 显示精确分数
const tabDefs = [
  { id: "chunks", label: "Chunks 预览" },
  { id: "test-query", label: "Test query" },
];
const activeTab = ref("chunks");

// Test query 状态
const testQuery = ref("");
const testTopK = ref(3);
const testResults = ref([]);
const testLoading = ref(false);
const testError = ref("");
const testQueriedOnce = ref(false);

const runTestQuery = async () => {
  const q = testQuery.value.trim();
  if (!q || testLoading.value) return;
  testLoading.value = true;
  testError.value = "";
  try {
    const res = await apiService.chat.knowledgeTestQuery({
      query: q,
      topK: testTopK.value,
      ...buildScopeParams(),
    });
    testResults.value = Array.isArray(res?.results) ? res.results : [];
    testQueriedOnce.value = true;
  } catch (error) {
    testResults.value = [];
    testError.value = error?.message || "检索失败，请稍后重试";
    testQueriedOnce.value = true;
  } finally {
    testLoading.value = false;
  }
};

// 切换 visibility 时清空上次 test query 结果，避免显示与新分组不匹配的旧结果
watch(activeKey, () => {
  selectedDocId.value = "";
  loadDocuments();
  if (testQueriedOnce.value) {
    testResults.value = [];
    testQueriedOnce.value = false;
    testError.value = "";
  }
});

// 相似度分数格式化为 0.00 - 1.00 三位定点
const formatScore = (score) => {
  if (typeof score !== "number" || Number.isNaN(score)) return "—";
  return score.toFixed(2);
};

// 三色阈值映射：返回 'high' | 'mid' | 'low'，CSS 类名拼接 wb-kb-score-{level}
const getScoreLevel = (score) => {
  if (typeof score !== "number" || Number.isNaN(score)) return "low";
  if (score >= 0.85) return "high";
  if (score >= 0.7) return "mid";
  return "low";
};

const mapKnowledgeDocument = (d) => ({
  id: String(d.documentId),
  folderId: Number(d.folderId || 0),
  // visibility/scope 同步保留，sidebar 与 filterDocs 都用它做分组
  visibility: d.visibility || d.scope || "private",
  scope: d.scope || d.visibility || "private",
  name: d.title || `文档 ${d.documentId}`,
  type: inferFileType(d.title || ""),
  sizeBytes: d.sizeBytes || 0,
  chunkCount: d.chunkCount || 0,
  embeddingDimension: d.embeddingDimension || 0,
  embeddingModel: d.embeddingModel || "",
  version: d.version || 1,
  status: mapKnowledgeStatus(d.status),
  uploadedAt: formatRelativeTime(d.updatedAt || d.createdAt),
  // chunks 在选中时 lazy 拉；preview 先作为占位
  chunks: d.preview ? [d.preview] : [],
  chunksLoaded: false,
  summary: d.preview || "",
});

// 拉取文档列表：后端返回 KnowledgeDocumentItem[]，含 visibility / folderId / sizeBytes / embeddingDimension / embeddingModel
const loadDocuments = async () => {
  try {
    const res = await apiService.chat.knowledgeDocuments({ limit: 50, ...buildScopeParams() });
    const list = Array.isArray(res?.documents) ? res.documents : [];
    documents.value = list.map(mapKnowledgeDocument);

    if (documents.value.length > 0) {
      const firstDoc = documents.value.find((d) => matchVisibility(d, activeKey.value));
      if (firstDoc) selectedDocId.value = firstDoc.id;
    } else {
      selectedDocId.value = "";
    }
  } catch (error) {
    // 静默降级：保持 documents.value=[] 触发空态
  }
};

const loadSidebarDocuments = async () => {
  try {
    const res = await apiService.chat.knowledgeDocuments({ limit: 100 });
    const list = Array.isArray(res?.documents) ? res.documents : [];
    sidebarDocuments.value = list.map(mapKnowledgeDocument);
  } catch (error) {
    sidebarDocuments.value = [];
  }
};

const loadFolders = async () => {
  try {
    const res = await apiService.chat.knowledgeFolders();
    folders.value = Array.isArray(res?.folders) ? res.folders : [];
    folderMeta.value = {
      unfiledCount: Number(res?.unfiledCount || 0),
      totalCount: Number(res?.totalCount || res?.total || 0),
      initialized: Boolean(res?.initialized),
    };
  } catch (error) {
    folders.value = [];
    folderMeta.value = { unfiledCount: 0, totalCount: 0, initialized: false };
  }
};

// 选中某份文档后 lazy 拉 chunks
const loadDocumentChunks = async (id) => {
  if (!id) return;
  const idx = documents.value.findIndex((d) => d.id === id);
  if (idx < 0) return;
  const target = documents.value[idx];
  if (target.chunksLoaded) return; // 不重拉

  try {
    const res = await apiService.chat.knowledgeDocumentChunks(id, { limit: 6 });
    const chunks = Array.isArray(res?.chunks) ? res.chunks : [];
    if (chunks.length === 0) return;

    // 不可变更新：避免 watch 重入
    documents.value[idx] = {
      ...target,
      chunks: chunks.map((c) => c.content).filter(Boolean),
      chunksLoaded: true,
    };
  } catch (error) {
    // 静默降级：保留 preview 嵌入的默认 chunk
  }
};

watch(selectedDocId, (id) => {
  if (id) loadDocumentChunks(id);
});

onMounted(() => {
  loadFolders();
  loadSidebarDocuments();
  loadDocuments();
  // reader 模式 Esc 退出快捷键：window 级监听以避免 focus 不在按钮时失效
  window.addEventListener("keydown", handleReaderKeydown);
  // 阅读进度条：listening on window 因为 .wb-kb-list 无 overflow，page-level 才是真正的 scroll 容器；
  // passive 提升滚动性能；handler 内部 viewMode 守卫保证 list 模式下不计算
  window.addEventListener("scroll", handleReaderScroll, { passive: true });
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", handleReaderKeydown);
  window.removeEventListener("scroll", handleReaderScroll);
});

// 返回 mono 文件类型标签，替代原 emoji 图标
const getDocTypeLabel = (type) => {
  if (!type) return "FILE";
  const map = { pdf: "PDF", md: "MD", txt: "TXT", docx: "DOC" };
  return map[type] || String(type).toUpperCase();
};

const getDocStatusLabel = (status) => {
  switch (status) {
    case "ready":
      return "已就绪";
    case "processing":
      return "处理中";
    case "failed":
      return "失败";
    default:
      return "—";
  }
};
</script>

<style scoped>
/*
  WorkbenchKnowledge 样式（v0.1 响应式重做，约束 R）

  改造要点:
    1. 容器宽度：max-width 用 clamp + 100% 流体收敛，不写死 1320px 决定布局
    2. 三栏 grid-template-columns 用 minmax(min, max) + 1fr 表达，断点收敛到 640/768/1024/1440 标准
    3. 内边距、间距改用 rem 与 clamp() 流体化，仅 1px 描边/圆角等视觉细节保留 px
    4. 动态视口 100dvh 替代 100vh 避免移动端地址栏抖动
    5. 删除原 .wb-kb-tree* / .wb-kb-cat* 样式（已被 KnowledgeSidebar 组件接管）
*/

.wb-kb-content {
  max-width: min(1440px, 100%);
  margin: 0 auto;
  padding: 0 clamp(1rem, 3vw, 2.75rem) clamp(2.5rem, 5vw, 5rem);
}

/* === Hero === */
.wb-kb-hero {
  padding: 0 0 clamp(1.25rem, 2.5vw, 2rem);
}

.wb-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font: var(--fs-xs) var(--mono);
  color: var(--t2);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-pill);
  padding: 6px 14px;
  margin-bottom: 22px;
  letter-spacing: .04em;
  background: rgba(255, 255, 255, 0.025);
  backdrop-filter: blur(8px);
  width: fit-content;
}

.wb-eyebrow-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgba(220, 155, 90, 0.9);
  animation: wb-edot 2.6s ease-in-out infinite;
}

@keyframes wb-edot {
  0%, 100% { opacity: 1; }
  50% { opacity: .35; }
}

.wb-kb-hero-row {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  flex-wrap: wrap;
}

.wb-kb-hero-text {
  flex: 1;
  min-width: 320px;
}

.wb-kb-title {
  font: 800 clamp(30px, 2.8vw, 42px) var(--display);
  color: var(--t);
  letter-spacing: -.02em;
  margin: 0 0 12px;
}

.wb-kb-sub {
  font-size: var(--fs-lg);
  color: var(--t3);
  line-height: 1.7;
  margin: 0;
  max-width: 560px;
}

.wb-kb-upload-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font: 600 var(--fs-md) var(--sans);
  color: var(--bg);
  background: var(--t);
  border: none;
  cursor: pointer;
  padding: 11px 22px;
  border-radius: var(--radius-md);
  transition: opacity .2s ease, transform .2s ease, box-shadow .2s ease;
  box-shadow: 0 4px 16px rgba(255, 255, 255, 0.08);
}

.wb-kb-upload-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(255, 255, 255, 0.16);
  opacity: .94;
}

.wb-kb-upload-plus {
  font-weight: 700;
  font-size: var(--fs-md);
  line-height: 1;
}

.wb-kb-file-input {
  display: none;
}

/*
  === Shell：响应式三栏底盘 ===

  默认 (≥1024px lg)：左 minmax(240px,260px) / 中 1fr / 右 minmax(320px,360px)
  窄屏断点：
    - md (768-1024px): 三栏紧凑（左/右收窄）
    - sm (640-768px): 双列（左 + 中合并堆叠，右栏跨整行下沉）
    - xs (<640px): 单列，sidebar / 中栏 / 右栏依次堆叠

  align-items: start 让三栏顶端对齐，避免左/右栏因 sticky 与中栏卡片高度不同造成视觉错位（C2.2 决策）。
*/
.wb-kb-shell {
  display: grid;
  grid-template-columns: minmax(240px, 260px) minmax(0, 1fr) minmax(320px, 360px);
  gap: clamp(0.875rem, 1.5vw, 1.25rem);
  align-items: start;
}

/* 左栏 KnowledgeSidebar：sticky 滚动跟随 */
.wb-kb-shell > :first-child {
  position: sticky;
  top: 6.25rem;
}

/* === 中间文档列表 === */
.wb-kb-list {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

/* Reader 模式：跨右栏占用，让阅读区域更宽（lg/xl/md 三栏 grid 时生效；sm/xs 已堆叠不需要） */
.wb-kb-list-reader {
  grid-column: 2 / -1;
}

/* === Reader 模式（v0.3 F7）=== */

/* 阅读进度条：reader 模式下 fixed 在 viewport 顶部，跨整页宽度，z-index 高于 SiteHeader 保证全程可见 */
.wb-kb-reader-progress {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: rgba(0, 0, 0, 0.35);
  z-index: 100;
  overflow: hidden;
  pointer-events: none;
}

.wb-kb-reader-progress-bar {
  height: 100%;
  background: linear-gradient(
    90deg,
    rgba(220, 155, 90, 0.55) 0%,
    rgba(220, 155, 90, 0.95) 100%
  );
  transition: width 0.12s linear;
  will-change: width;
}

.wb-kb-reader-head {
  display: flex;
  align-items: flex-start;
  gap: clamp(0.625rem, 1vw, 0.875rem);
  padding-bottom: 0.875rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  margin-bottom: 0.5rem;
}

.wb-kb-reader-back {
  font: 600 clamp(var(--fs-2xs), 0.78vw, var(--fs-xs)) var(--sans);
  color: var(--t2);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-sm);
  padding: 6px 10px;
  cursor: pointer;
  flex-shrink: 0;
  transition: color 0.15s ease, background-color 0.15s ease, border-color 0.15s ease;
}

.wb-kb-reader-back:hover {
  color: var(--t);
  background: rgba(220, 155, 90, 0.08);
  border-color: rgba(220, 155, 90, 0.3);
}

.wb-kb-reader-meta {
  flex: 1;
  min-width: 0;
}

.wb-kb-reader-title {
  font: 700 clamp(var(--fs-lg), 1.4vw, var(--fs-2xl)) var(--display);
  color: var(--t);
  margin: 0 0 4px;
  letter-spacing: -.01em;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-kb-reader-sub {
  font: clamp(var(--fs-2xs), 0.78vw, var(--fs-xs)) var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
  align-items: baseline;
}

.wb-kb-reader-state {
  font: clamp(var(--fs-sm), 0.95vw, var(--fs-md)) var(--sans);
  color: var(--t3);
  text-align: center;
  padding: 2.5rem 1rem;
  border: 1px dashed rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.015);
}

.wb-kb-reader-error {
  color: rgba(239, 138, 115, 0.9);
  border-color: rgba(239, 102, 96, 0.25);
  background: rgba(239, 102, 96, 0.04);
}

/* Reader body：阅读最佳宽度 ~70 字符；用 max-inline-size + auto margin 居中 */
.wb-kb-reader-body {
  max-inline-size: 72ch;
  margin: 0 auto;
  padding: 0.5rem 0.5rem 4rem;
  font: clamp(var(--fs-sm), 1vw, var(--fs-md)) var(--sans);
  color: var(--t);
  line-height: 1.75;
}

.wb-kb-reader-body :deep(h1),
.wb-kb-reader-body :deep(h2),
.wb-kb-reader-body :deep(h3),
.wb-kb-reader-body :deep(h4) {
  font-family: var(--display);
  color: var(--t);
  margin: 1.5em 0 0.5em;
  line-height: 1.3;
  letter-spacing: -.01em;
}
.wb-kb-reader-body :deep(h1) { font-size: clamp(var(--fs-xl), 1.6vw, var(--fs-2xl)); }
.wb-kb-reader-body :deep(h2) { font-size: clamp(var(--fs-lg), 1.3vw, var(--fs-xl)); }
.wb-kb-reader-body :deep(h3) { font-size: clamp(var(--fs-md), 1.1vw, var(--fs-lg)); }
.wb-kb-reader-body :deep(h4) { font-size: clamp(var(--fs-sm), 0.95vw, var(--fs-md)); }

.wb-kb-reader-body :deep(p) {
  margin: 0 0 1em;
}

.wb-kb-reader-body :deep(ul),
.wb-kb-reader-body :deep(ol) {
  margin: 0 0 1em;
  padding-inline-start: 1.5em;
}

.wb-kb-reader-body :deep(li) {
  margin-bottom: 0.375em;
}

.wb-kb-reader-body :deep(blockquote) {
  margin: 1em 0;
  padding: 0.5em 1em;
  border-inline-start: 3px solid rgba(220, 155, 90, 0.4);
  background: rgba(220, 155, 90, 0.03);
  color: var(--t2);
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
}

.wb-kb-reader-body :deep(code) {
  font-family: var(--mono);
  font-size: 0.9em;
  background: rgba(255, 255, 255, 0.06);
  padding: 1px 4px;
  border-radius: 3px;
}

.wb-kb-reader-body :deep(pre) {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-sm);
  padding: 0.875em 1em;
  overflow-x: auto;
  margin: 1em 0;
}

.wb-kb-reader-body :deep(pre code) {
  background: transparent;
  padding: 0;
  font-size: 0.85em;
  color: var(--t);
}

.wb-kb-reader-body :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1em 0;
  font-size: 0.9em;
}

.wb-kb-reader-body :deep(th),
.wb-kb-reader-body :deep(td) {
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 0.5em 0.75em;
  text-align: start;
}

.wb-kb-reader-body :deep(th) {
  background: rgba(255, 255, 255, 0.04);
  font-weight: 600;
}

.wb-kb-reader-body :deep(a) {
  color: rgba(220, 155, 90, 0.95);
  text-decoration: none;
  border-bottom: 1px solid rgba(220, 155, 90, 0.3);
}

.wb-kb-reader-body :deep(a:hover) {
  border-bottom-color: rgba(220, 155, 90, 0.95);
}

.wb-kb-reader-body :deep(hr) {
  border: none;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  margin: 2em 0;
}

/* Reader chunks 折叠面板：保持与 reader-body 同宽且居中，复用 details/summary */
.wb-kb-reader-chunks {
  max-inline-size: 72ch;
  margin: 0 auto 4rem;
  padding: 0 0.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding-top: 1rem;
}

.wb-kb-reader-chunks-summary {
  font: 600 clamp(var(--fs-2xs), 0.8vw, var(--fs-xs)) var(--mono);
  color: var(--t3);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  cursor: pointer;
  user-select: none;
  padding: 6px 8px;
  border-radius: var(--radius-sm);
  transition: color 0.15s ease, background-color 0.15s ease;
  list-style: none;
}

.wb-kb-reader-chunks-summary::-webkit-details-marker { display: none; }
.wb-kb-reader-chunks-summary::before {
  content: "▸ ";
  display: inline-block;
  margin-right: 0.25rem;
  transition: transform 0.15s ease;
}

.wb-kb-reader-chunks[open] .wb-kb-reader-chunks-summary::before {
  transform: rotate(90deg);
}

.wb-kb-reader-chunks-summary:hover {
  color: var(--t2);
  background: rgba(255, 255, 255, 0.03);
}

.wb-kb-reader-chunks-list {
  list-style: none;
  margin: 0.875rem 0 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.wb-kb-reader-chunks-item {
  display: flex;
  gap: 0.625rem;
  padding: 0.625rem 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.015);
}

.wb-kb-reader-chunks-num {
  font: 600 clamp(var(--fs-2xs), 0.78vw, var(--fs-xs)) var(--mono);
  color: rgba(220, 155, 90, 0.7);
  letter-spacing: 0.04em;
  flex-shrink: 0;
  padding-top: 2px;
}

.wb-kb-reader-chunks-text {
  margin: 0;
  font: clamp(var(--fs-sm), 0.95vw, var(--fs-md)) var(--sans);
  color: var(--t2);
  line-height: 1.65;
  white-space: pre-wrap;
  word-break: break-word;
}

.wb-block-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.wb-block-title {
  font: 700 var(--fs-xl) var(--display);
  color: var(--t);
  margin: 0;
  letter-spacing: -.01em;
}

.wb-block-meta {
  font: var(--fs-xs) var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

.wb-kb-docs {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wb-kb-doc {
  display: flex;
  align-items: center;
  gap: clamp(0.625rem, 1vw, 0.875rem);
  padding: clamp(0.75rem, 1.25vw, 0.875rem) clamp(0.875rem, 1.5vw, 1rem);
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: border-color .2s ease, background-color .2s ease, transform .2s ease;
  isolation: isolate;
}

.wb-kb-doc:hover {
  border-color: rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.04);
  transform: translateY(-1px);
}

.wb-kb-doc-active {
  border-color: rgba(220, 155, 90, 0.4);
  background: rgba(220, 155, 90, 0.05);
}

.wb-kb-doc-icon {
  font: 600 var(--fs-2xs) var(--mono);
  letter-spacing: .08em;
  color: var(--t2);
  width: 40px;
  height: 36px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.wb-kb-doc-active .wb-kb-doc-icon {
  color: rgba(220, 155, 90, 0.95);
  border-color: rgba(220, 155, 90, 0.3);
}

.wb-kb-doc-meta {
  flex: 1;
  min-width: 0;
}

.wb-kb-doc-name {
  font: 600 var(--fs-md) var(--sans);
  color: var(--t);
  margin: 0 0 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-kb-doc-info {
  font: clamp(var(--fs-2xs), 0.78vw, var(--fs-xs)) var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
  display: flex;
  gap: 0.375rem;
  flex-wrap: wrap;
  align-items: baseline;
}

.wb-kb-doc-info-chip {
  white-space: nowrap;
}

.wb-kb-doc-info-sep {
  opacity: .5;
  user-select: none;
}

/* 卡片右侧操作列：reader 按钮 + 状态徽标垂直堆叠 */
.wb-kb-doc-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.375rem;
  flex-shrink: 0;
}

.wb-kb-doc-reader-btn {
  font: 600 clamp(var(--fs-2xs), 0.78vw, var(--fs-xs)) var(--sans);
  color: var(--t3);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-sm);
  padding: 0.3125rem 0.625rem;
  cursor: not-allowed;
  transition: color .2s ease, border-color .2s ease, background-color .2s ease;
  letter-spacing: 0.02em;
}

.wb-kb-doc-reader-btn:not(:disabled):hover {
  color: var(--t);
  background: rgba(220, 155, 90, 0.06);
  border-color: rgba(220, 155, 90, 0.3);
}

.wb-kb-doc-reader-btn[disabled] {
  opacity: 0.55;
}

/* v0.2 移动到目录 select：label 包 icon + native select，保留键盘可访问性 */
.wb-kb-doc-move {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 3px 6px 3px 8px;
  font: 500 clamp(var(--fs-2xs), 0.78vw, var(--fs-xs)) var(--sans);
  color: var(--t2);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: color 0.15s ease, background-color 0.15s ease, border-color 0.15s ease;
}

.wb-kb-doc-move:hover {
  color: var(--t);
  border-color: rgba(220, 155, 90, 0.3);
  background: rgba(220, 155, 90, 0.06);
}

.wb-kb-doc-move-icon {
  font-size: 11px;
  opacity: 0.85;
}

.wb-kb-doc-move-select {
  font: inherit;
  color: inherit;
  background: transparent;
  border: none;
  outline: none;
  cursor: pointer;
  min-width: 0;
  max-width: 7.5rem;
  padding: 0 0.125rem;
  appearance: none;
}

.wb-kb-doc-move-select:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

/* select 下拉里的 option：浏览器会按 native 主题渲染，统一写 dark 背景避免白底 */
.wb-kb-doc-move-select option {
  background: #16181d;
  color: var(--t);
}

.wb-kb-doc-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
  flex-shrink: 0;
}

.wb-kb-doc-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.wb-status-ready {
  background: rgba(155, 209, 168, 0.85);
  box-shadow: 0 0 6px rgba(155, 209, 168, 0.45);
}

.wb-status-processing {
  background: rgba(220, 155, 90, 0.85);
  animation: wb-edot 1.4s ease-in-out infinite;
}

.wb-status-failed {
  background: #ef6660;
}

/* === 右侧详情 === */
.wb-kb-detail {
  position: sticky;
  top: 6.25rem;
  padding: clamp(1rem, 1.75vw, 1.375rem) clamp(1rem, 1.75vw, 1.375rem) clamp(1.125rem, 1.875vw, 1.5rem);
  background:
    linear-gradient(180deg, rgba(16, 17, 22, 1) 0%, rgba(10, 11, 14, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.03) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  isolation: isolate;
  max-height: calc(100dvh - 7.5rem);
  overflow-y: auto;
}

/* === 右栏 tab 切换 === */
.wb-kb-tabs {
  display: flex;
  gap: 0.25rem;
  padding: 0.25rem;
  margin-bottom: clamp(0.875rem, 1.5vw, 1.125rem);
  background: rgba(255, 255, 255, 0.025);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-pill);
}

.wb-kb-tab {
  flex: 1;
  font: 600 clamp(var(--fs-2xs), 0.85vw, var(--fs-sm)) var(--sans);
  color: var(--t3);
  background: transparent;
  border: none;
  border-radius: var(--radius-pill);
  padding: 0.4375rem 0.75rem;
  cursor: pointer;
  transition: color 0.2s ease, background-color 0.2s ease;
  letter-spacing: 0.01em;
}

.wb-kb-tab:hover:not(.wb-kb-tab-active) {
  color: var(--t);
  background: rgba(255, 255, 255, 0.03);
}

.wb-kb-tab-active {
  color: rgba(220, 155, 90, 0.95);
  background: rgba(220, 155, 90, 0.08);
  box-shadow: inset 0 0 0 1px rgba(220, 155, 90, 0.18);
}

.wb-kb-tabpanel {
  display: flex;
  flex-direction: column;
  gap: clamp(0.875rem, 1.5vw, 1.125rem);
}

.wb-kb-detail-head {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.wb-kb-detail-icon {
  font: 700 var(--fs-xs) var(--mono);
  letter-spacing: .1em;
  color: rgba(220, 155, 90, 0.95);
  width: 48px;
  height: 44px;
  border-radius: var(--radius-sm);
  background: rgba(220, 155, 90, 0.06);
  border: 1px solid rgba(220, 155, 90, 0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.wb-kb-detail-meta {
  flex: 1;
  min-width: 0;
}

.wb-kb-detail-name {
  font: 700 clamp(var(--fs-md), 1.1vw, var(--fs-lg)) var(--display);
  color: var(--t);
  margin: 0 0 4px;
  word-break: break-word;
}

.wb-kb-detail-info {
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
}

.wb-kb-detail-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  padding: 16px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.wb-kb-detail-stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
  text-align: center;
}

.wb-kb-detail-stat-num {
  font: 700 clamp(var(--fs-md), 1.3vw, var(--fs-2xl)) var(--mono);
  color: var(--t);
  line-height: 1;
}

.wb-kb-detail-stat-lb {
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
}

.wb-kb-detail-block {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.wb-detail-label {
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .06em;
  text-transform: uppercase;
}

.wb-kb-chunks {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wb-kb-chunk {
  padding: 10px 12px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: var(--radius-sm);
  border-left: 2px solid rgba(220, 155, 90, 0.5);
}

.wb-kb-chunk-num {
  font: var(--fs-3xs) var(--mono);
  color: rgba(220, 155, 90, 0.85);
  letter-spacing: .06em;
  margin-bottom: 4px;
}

.wb-kb-chunk-text {
  font-size: var(--fs-xs);
  color: var(--t2);
  line-height: 1.7;
  margin: 0;
}

/* embedding 模型行：在 stats 与 chunks 之间显示 */
.wb-kb-detail-meta-line {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.625rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.wb-kb-detail-meta-value {
  font: 600 clamp(var(--fs-2xs), 0.78vw, var(--fs-xs)) var(--mono);
  color: var(--t2);
  letter-spacing: 0.02em;
  text-align: right;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* chunks 空态：当 selectedDoc 为 processing/failed 或还未加载时显示 */
.wb-kb-chunks-empty {
  font: var(--fs-xs) var(--mono);
  color: var(--t3);
  letter-spacing: 0.03em;
  padding: 1rem 0.75rem;
  text-align: center;
  border: 1px dashed rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.015);
}

/* === Test query tab 输入区 === */
.wb-kb-test-block {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
  padding-bottom: 0.875rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.wb-kb-test-input-row {
  display: flex;
  gap: 0.5rem;
  align-items: stretch;
}

.wb-kb-test-input {
  flex: 1;
  min-width: 0;
  font: clamp(var(--fs-xs), 0.9vw, var(--fs-md)) var(--sans);
  color: var(--t);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-sm);
  padding: 0.5rem 0.75rem;
  outline: none;
  transition: border-color 0.2s ease, background-color 0.2s ease;
}

.wb-kb-test-input:focus {
  border-color: rgba(220, 155, 90, 0.45);
  background: rgba(255, 255, 255, 0.05);
}

.wb-kb-test-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.wb-kb-test-input::placeholder {
  color: var(--t3);
}

.wb-kb-test-topk {
  min-width: 80px;
  font: 600 clamp(var(--fs-xs), 0.85vw, var(--fs-sm)) var(--sans);
  color: var(--t);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-sm);
  padding: 0 0.5rem;
  outline: none;
}

.wb-kb-test-topk:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.wb-kb-test-btn {
  font: 600 clamp(var(--fs-xs), 0.9vw, var(--fs-sm)) var(--sans);
  color: var(--bg);
  background: rgba(220, 155, 90, 0.95);
  border: none;
  border-radius: var(--radius-sm);
  padding: 0 0.875rem;
  cursor: pointer;
  flex-shrink: 0;
  transition: background-color 0.2s ease, opacity 0.2s ease;
}

.wb-kb-test-btn:hover:not(:disabled) {
  background: rgba(232, 173, 110, 1);
}

.wb-kb-test-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.wb-kb-test-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.375rem;
  font: clamp(var(--fs-3xs), 0.7vw, var(--fs-2xs)) var(--mono);
  color: var(--t3);
  letter-spacing: 0.04em;
}

.wb-kb-test-error {
  color: #ef8a73;
}

/* === Test query 召回结果 === */
.wb-kb-test-result-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.375rem;
}

.wb-kb-test-result-title {
  font: 600 clamp(var(--fs-xs), 0.9vw, var(--fs-sm)) var(--sans);
  color: var(--t2);
  margin: 0 0 0.25rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* score chip 三色阈值（≥0.85 高 / 0.70-0.85 中 / <0.70 低） */
.wb-kb-score-chip {
  font: 700 var(--fs-2xs) var(--mono);
  letter-spacing: 0.04em;
  padding: 2px 8px;
  border-radius: 999px;
  flex-shrink: 0;
  border: 1px solid transparent;
}

.wb-kb-score-high {
  color: rgba(155, 209, 168, 0.95);
  background: rgba(155, 209, 168, 0.1);
  border-color: rgba(155, 209, 168, 0.3);
}

.wb-kb-score-mid {
  color: rgba(220, 155, 90, 0.95);
  background: rgba(220, 155, 90, 0.08);
  border-color: rgba(220, 155, 90, 0.3);
}

.wb-kb-score-low {
  color: rgba(239, 138, 115, 0.95);
  background: rgba(239, 102, 96, 0.08);
  border-color: rgba(239, 102, 96, 0.3);
}

.wb-kb-detail-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 40px 16px;
  gap: 8px;
}

/* === Empty (中间列空态) === */
.wb-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 60px 20px;
  gap: 10px;
}

.wb-empty-icon {
  font-size: 32px;
  opacity: .5;
}

.wb-empty-title {
  font: 700 var(--fs-md) var(--display);
  color: var(--t);
}

.wb-empty-sub {
  font-size: var(--fs-sm);
  color: var(--t3);
  margin-bottom: 12px;
}

.wb-empty-cta {
  font: 600 var(--fs-sm) var(--sans);
  color: rgba(220, 155, 90, 0.95);
  background: none;
  border: 1px solid rgba(220, 155, 90, 0.3);
  border-radius: var(--radius-sm);
  padding: 8px 16px;
  cursor: pointer;
  transition: background-color .2s ease;
}

.wb-empty-cta:hover {
  background: rgba(220, 155, 90, 0.08);
}

/*
  === 响应式断点（约束 R 标准）===
  采用与 docs/requirements/2026-05-12-workbench-knowledge-base-redesign.md §响应式布局规范 一致的 5 断点：
    - 375px (xs base)：超小手机，单列堆叠
    - 768px (sm)：竖屏平板，左 + 中合并双列，右栏跨行
    - 1024px (md)：横屏平板，三栏紧凑
    - 1440px (lg)：桌面标准（默认）
    - 1920px (xl)：大显示器（不展开，仅放宽边距）
  全部使用 max-width 形式收敛，与现有项目其他子页保持一致。
*/

/* md (≤1440px)：三栏紧凑（默认 grid 已生效，仅细微收敛） */
@media (max-width: 1440px) {
  .wb-kb-shell {
    grid-template-columns: minmax(220px, 240px) minmax(0, 1fr) minmax(300px, 340px);
  }
}

/* sm (≤1024px)：两栏（左 + 中），右栏跨整行下沉 */
@media (max-width: 1024px) {
  .wb-kb-shell {
    grid-template-columns: minmax(200px, 220px) minmax(0, 1fr);
  }
  .wb-kb-detail {
    grid-column: 1 / -1;
    position: static;
    max-height: none;
  }
}

/* xs (≤768px)：单列堆叠（sidebar / list / detail 依次） */
@media (max-width: 768px) {
  .wb-kb-shell {
    grid-template-columns: minmax(0, 1fr);
  }
  .wb-kb-shell > :first-child {
    position: static;
    top: auto;
  }
  .wb-kb-detail {
    position: static;
    max-height: none;
  }
}

/* 超小手机 (≤640px)：进一步收敛 padding 与字号 */
@media (max-width: 640px) {
  .wb-kb-detail-stats {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0.5rem;
  }
  .wb-kb-doc {
    flex-wrap: wrap;
  }
  .wb-kb-doc-actions {
    width: 100%;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
