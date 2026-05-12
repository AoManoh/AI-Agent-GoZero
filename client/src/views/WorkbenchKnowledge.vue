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
      - POST /api/ai/knowledge/upload → 任何登录用户可上传，admin 上传 visibility=public，普通 user → private（Q7=B 角色路由）
      - GET /api/ai/knowledge/documents/:id/chunks → 单文档 chunks 数组

    本期不实现（v0.2 范围）:
      - folder CRUD、folderId 过滤、reader 模式（chunks concat + marked 渲染）、Test query tab

    决策来源: docs/requirements/2026-05-12-workbench-knowledge-base-redesign.md
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
            accept=".pdf,.md,.txt,.docx"
            multiple
            class="wb-kb-file-input"
            @change="handleUpload"
          />
        </div>
      </section>

      <div class="wb-kb-shell">
        <!-- 左：visibility sidebar（v0.1=visibility 二分类，v0.2 切到 folder-tree） -->
        <KnowledgeSidebar
          mode="visibility"
          :documents="documents"
          :active-key="activeKey"
          :has-private-access="hasPrivateAccess"
          @select="handleSidebarSelect"
        />

        <!-- 中：文档列表 -->
        <section class="wb-kb-list">
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
                <button
                  type="button"
                  class="wb-kb-doc-reader-btn"
                  :title="readerButtonHint"
                  disabled
                  aria-disabled="true"
                  @click.stop
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
        </section>

        <!-- 右：详情 + 片段预览 -->
        <aside class="wb-kb-detail">
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
        </aside>
      </div>
    </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
import KnowledgeSidebar from "../components/workbench/KnowledgeSidebar.vue";
import { apiService } from "../composables/useApi";

// === 上传 ===
const uploadInputRef = ref(null);
const triggerUpload = () => uploadInputRef.value?.click();

const handleUpload = async (e) => {
  const files = Array.from(e.target.files || []);
  if (files.length === 0) return;
  let anySuccess = false;
  for (const file of files) {
    try {
      const formData = new FormData();
      formData.append("file", file);
      // Q4=C 上下文感知（v0.1：folderId 参数尚未上线，留空；v0.2 注入 activeFolderId）
      // formData.append("folderId", activeFolderId ?? "");
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
    await loadDocuments();
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

// === visibility 二分类（Q5=B 决策派生） ===
//
// activeKey: 'public' | 'private'
//   - 默认 'public'，用户切到「我的私人资料」后变 'private'
//   - 未登录用户只能看 'public'，sidebar 不渲染 private 节点
//
// 不再保留前端 mock 5 分类（c-go/c-vue/c-arch/c-db/c-personal）— 完全消除原则 5 违规。
const activeKey = ref("public");

// 是否有私人资料访问权限：依据 documents 中是否出现过 visibility=private 的条目判定。
// 真实落地是「登录态 + 已上传过私人文档」；未登录用户后端只返 public，sidebar 自动隐藏 private 节点。
const hasPrivateAccess = computed(() =>
  documents.value.some((d) => (d.visibility || d.scope) === "private")
);

const handleSidebarSelect = (key) => {
  activeKey.value = key;
  // 切换 visibility 后，自动选中该分组下第一篇文档（如果有）
  const firstDoc = documents.value.find((d) => matchVisibility(d, key));
  selectedDocId.value = firstDoc ? firstDoc.id : "";
};

const matchVisibility = (doc, key) => {
  const v = (doc.visibility || doc.scope || "").toLowerCase();
  if (key === "public") return v === "public";
  if (key === "private") return v === "private";
  return true;
};

const activeScopeLabel = computed(() => {
  if (activeKey.value === "private") return "我的私人资料";
  return "公共知识";
});

// === 文档（仅来自后端，无前端 mock） ===
const documents = ref([]);

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
  return "公共知识库为空";
});

const emptyStateSub = computed(() => {
  if (activeKey.value === "private") {
    return "上传 PDF 资料，AI 会自动切片并向量化，作为面试时的私人 RAG 数据源。";
  }
  return "公共知识库由管理员维护，当前没有可阅读的内容。";
});

// 仅在 private 视图下显示「上传」CTA（公共知识库由 admin 通过同端点上传，本子页 UI 不暴露 admin 切换 — Q7=C）
const emptyStateCTA = computed(() => activeKey.value === "private");

// 全文阅读按钮 hint（v0.2 启用）
const readerButtonHint = "全文阅读模式将在 v0.2 启用（reader = chunks concat + marked 渲染）";

// 拉取文档列表：后端返回 KnowledgeDocumentItem[]，含 visibility / sizeBytes / embeddingDimension / embeddingModel
const loadDocuments = async () => {
  try {
    const res = await apiService.chat.knowledgeDocuments({ limit: 50 });
    const list = Array.isArray(res?.documents) ? res.documents : [];
    documents.value = list.map((d) => ({
      id: String(d.documentId),
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
    }));

    // 自动选中：默认显示有内容的分组的首篇文档
    if (documents.value.length > 0) {
      // 如果当前 activeKey 下没有 doc，自动切到有 doc 的分组
      if (!documents.value.some((d) => matchVisibility(d, activeKey.value))) {
        const fallback = hasPrivateAccess.value && documents.value.some((d) => d.visibility === "private")
          ? "private"
          : "public";
        activeKey.value = fallback;
      }
      const firstDoc = documents.value.find((d) => matchVisibility(d, activeKey.value));
      if (firstDoc) selectedDocId.value = firstDoc.id;
    } else {
      selectedDocId.value = "";
    }
  } catch (error) {
    // 静默降级：保持 documents.value=[] 触发空态
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
  loadDocuments();
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
  font: 12px var(--mono);
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
  font-size: 15px;
  color: var(--t3);
  line-height: 1.7;
  margin: 0;
  max-width: 560px;
}

.wb-kb-upload-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font: 600 14px var(--sans);
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
  font-size: 16px;
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

.wb-block-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.wb-block-title {
  font: 700 17px var(--display);
  color: var(--t);
  margin: 0;
  letter-spacing: -.01em;
}

.wb-block-meta {
  font: 12px var(--mono);
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
  font: 600 11px var(--mono);
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
  font: 600 14px var(--sans);
  color: var(--t);
  margin: 0 0 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-kb-doc-info {
  font: clamp(11px, 0.78vw, 12px) var(--mono);
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
  font: 600 clamp(11px, 0.78vw, 12px) var(--sans);
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

.wb-kb-doc-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font: 11px var(--mono);
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

.wb-kb-detail-head {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.wb-kb-detail-icon {
  font: 700 12px var(--mono);
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
  font: 700 clamp(14px, 1.1vw, 16px) var(--display);
  color: var(--t);
  margin: 0 0 4px;
  word-break: break-word;
}

.wb-kb-detail-info {
  font: 11px var(--mono);
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
  font: 700 clamp(16px, 1.3vw, 20px) var(--mono);
  color: var(--t);
  line-height: 1;
}

.wb-kb-detail-stat-lb {
  font: 11px var(--mono);
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
  font: 11px var(--mono);
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
  font: 10px var(--mono);
  color: rgba(220, 155, 90, 0.85);
  letter-spacing: .06em;
  margin-bottom: 4px;
}

.wb-kb-chunk-text {
  font-size: 12px;
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
  font: 600 clamp(11px, 0.78vw, 12px) var(--mono);
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
  font: 12px var(--mono);
  color: var(--t3);
  letter-spacing: 0.03em;
  padding: 1rem 0.75rem;
  text-align: center;
  border: 1px dashed rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.015);
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
  font: 700 16px var(--display);
  color: var(--t);
}

.wb-empty-sub {
  font-size: 13px;
  color: var(--t3);
  margin-bottom: 12px;
}

.wb-empty-cta {
  font: 600 13px var(--sans);
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
