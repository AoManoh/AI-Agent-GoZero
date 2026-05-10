<template>
  <!--
    WorkbenchKnowledge：知识库管理（对应设计图 4）。
    布局：左 tree (240px) + 中 list (1fr) + 右 detail (340px)
    后端契约：当前 mock；后续接 GET /users/knowledge/categories + GET /users/knowledge/docs?categoryId=
            上传走 /ai/knowledge/upload (已存在的 chat API)
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
            <span>上传文档</span>
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
        <!-- 左：分类 tree -->
        <aside class="wb-kb-tree">
          <div class="wb-kb-tree-head">
            <span class="wb-kb-tree-title">分类</span>
            <button type="button" class="wb-kb-tree-add" title="新建分类" @click="addCategory">+</button>
          </div>
          <ul class="wb-kb-cat-list">
            <li
              v-for="cat in categories"
              :key="cat.id"
              class="wb-kb-cat"
              :class="{ 'wb-kb-cat-active': activeCategoryId === cat.id }"
              @click="activeCategoryId = cat.id"
            >
              <span class="wb-kb-cat-dot" :style="{ background: cat.color }" aria-hidden="true"></span>
              <span class="wb-kb-cat-label">{{ cat.label }}</span>
              <span class="wb-kb-cat-count">{{ getDocCount(cat.id) }}</span>
            </li>
          </ul>
          <div class="wb-kb-tree-stats">
            <div class="wb-kb-stat-line">
              <span class="wb-kb-stat-num">{{ documents.length }}</span>
              <span class="wb-kb-stat-lb">文档</span>
            </div>
            <div class="wb-kb-stat-line">
              <span class="wb-kb-stat-num">{{ totalChunks }}</span>
              <span class="wb-kb-stat-lb">片段</span>
            </div>
            <div class="wb-kb-stat-line">
              <span class="wb-kb-stat-num">{{ totalVectorMb }}</span>
              <span class="wb-kb-stat-lb">向量 MB</span>
            </div>
          </div>
        </aside>

        <!-- 中：文档列表 -->
        <section class="wb-kb-list">
          <header class="wb-block-head">
            <h3 class="wb-block-title">{{ activeCategoryLabel }}</h3>
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
                  <span>{{ doc.size }}</span>
                  <span aria-hidden="true">·</span>
                  <span>{{ doc.chunkCount }} 片段</span>
                  <span aria-hidden="true">·</span>
                  <span>{{ doc.uploadedAt }}</span>
                </div>
              </div>
              <span class="wb-kb-doc-status">
                <span class="wb-kb-doc-dot" :class="`wb-status-${doc.status}`" aria-hidden="true"></span>
                {{ getDocStatusLabel(doc.status) }}
              </span>
            </article>
          </div>

          <div v-else class="wb-empty">
            <div class="wb-empty-title">这个分类还没有文档</div>
            <div class="wb-empty-sub">上传 PDF / Markdown / TXT，AI 会自动切片并向量化。</div>
            <button type="button" class="wb-empty-cta" @click="triggerUpload">+ 上传文档</button>
          </div>
        </section>

        <!-- 右：详情 -->
        <aside class="wb-kb-detail">
          <div v-if="selectedDoc" class="wb-kb-detail-inner">
            <header class="wb-kb-detail-head">
              <div class="wb-kb-detail-icon" aria-hidden="true">{{ getDocTypeLabel(selectedDoc.type) }}</div>
              <div class="wb-kb-detail-meta">
                <h4 class="wb-kb-detail-name">{{ selectedDoc.name }}</h4>
                <div class="wb-kb-detail-info">{{ selectedDoc.uploadedAt }} · {{ selectedDoc.size }}</div>
              </div>
            </header>

            <div class="wb-kb-detail-stats">
              <div class="wb-kb-detail-stat">
                <span class="wb-kb-detail-stat-num">{{ selectedDoc.chunkCount }}</span>
                <span class="wb-kb-detail-stat-lb">片段</span>
              </div>
              <div class="wb-kb-detail-stat">
                <span class="wb-kb-detail-stat-num">{{ selectedDoc.embedding }}</span>
                <span class="wb-kb-detail-stat-lb">维度</span>
              </div>
              <div class="wb-kb-detail-stat">
                <span class="wb-kb-detail-stat-num">{{ selectedDoc.queryCount }}</span>
                <span class="wb-kb-detail-stat-lb">命中</span>
              </div>
            </div>

            <div class="wb-kb-detail-block">
              <div class="wb-detail-label">片段预览</div>
              <div class="wb-kb-chunks">
                <div
                  v-for="(chunk, i) in selectedDoc.chunks"
                  :key="`chunk-${i}`"
                  class="wb-kb-chunk"
                >
                  <div class="wb-kb-chunk-num">#{{ String(i + 1).padStart(2, '0') }}</div>
                  <p class="wb-kb-chunk-text">{{ chunk }}</p>
                </div>
              </div>
            </div>

            <div class="wb-kb-detail-actions">
              <button type="button" class="wb-kb-action-btn">重新切片</button>
              <button type="button" class="wb-kb-action-btn wb-kb-action-danger">删除文档</button>
            </div>
          </div>

          <div v-else class="wb-kb-detail-empty">
            <div class="wb-empty-title">选择一份文档</div>
            <div class="wb-empty-sub">查看片段预览、向量统计与命中频率。</div>
          </div>
        </aside>
      </div>
    </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
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
      // 后端 KnowledgeUpload 仅要求 file part；本地分类仅供前端 UI 使用。
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

const formatBytes = (bytes) => {
  if (!bytes) return "—";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
};

const inferFileType = (name) => {
  const ext = name.toLowerCase().match(/\.(pdf|docx?|md|txt)$/);
  if (!ext) return "file";
  return ext[1] === "doc" ? "docx" : ext[1];
};

// === 分类 ===
// color 与 WorkbenchNew 的方向语义色同源，做到跨页面色彩一致。
const categories = ref([
  { id: "c-go", label: "Go 后端", color: "#4cd6a8" },
  { id: "c-vue", label: "前端", color: "#6eb6ff" },
  { id: "c-arch", label: "系统设计", color: "#b599ff" },
  { id: "c-system", label: "Linux / 网络", color: "rgba(255, 255, 255, 0.55)" },
  { id: "c-personal", label: "个人项目笔记", color: "#ffd770" },
]);

// 新建分类时随机分配语义色，避免引入手动选色 UI。
const CATEGORY_COLOR_PALETTE = ["#4cd6a8", "#6eb6ff", "#b599ff", "#ffd770", "#ff9966"];

const activeCategoryId = ref("c-go");

const activeCategoryLabel = computed(() => {
  return categories.value.find((c) => c.id === activeCategoryId.value)?.label || "全部";
});

const addCategory = () => {
  const name = window.prompt("新建分类名称");
  if (!name?.trim()) return;
  const id = `c-${Date.now()}`;
  const color = CATEGORY_COLOR_PALETTE[Math.floor(Math.random() * CATEGORY_COLOR_PALETTE.length)];
  categories.value.push({ id, label: name.trim(), color });
  activeCategoryId.value = id;
};

// === 文档（mock） ===
const documents = ref([
  {
    id: "d-1",
    categoryId: "c-go",
    name: "Go 并发编程模式.pdf",
    type: "pdf",
    size: "2.4 MB",
    chunkCount: 48,
    uploadedAt: "3 天前",
    status: "ready",
    embedding: 1024,
    queryCount: 26,
    chunks: [
      "Go 的并发模型基于 CSP（Communicating Sequential Processes）：goroutine 之间通过 channel 传递所有权而非共享内存。",
      "select 语句允许 goroutine 同时等待多个 channel 操作，配合 default 分支可实现非阻塞通信。",
      "context.Context 是跨 goroutine 传递取消信号、deadline 和 trace 的标准方式，应当作为函数首个参数传入。",
    ],
  },
  {
    id: "d-2",
    categoryId: "c-go",
    name: "GoZero 微服务实战.md",
    type: "md",
    size: "180 KB",
    chunkCount: 32,
    uploadedAt: "1 周前",
    status: "ready",
    embedding: 1024,
    queryCount: 18,
    chunks: [
      "GoZero 通过 goctl 工具从 .api 文件生成完整的 HTTP 服务骨架，包括 handler / logic / svc / config 各层。",
      "rpc 服务使用 etcd 作为服务注册中心，client 通过 etcd watch 自动感知 endpoint 变化并负载均衡。",
    ],
  },
  {
    id: "d-3",
    categoryId: "c-vue",
    name: "Vue 3 响应式源码笔记.md",
    type: "md",
    size: "92 KB",
    chunkCount: 21,
    uploadedAt: "2 天前",
    status: "ready",
    embedding: 1024,
    queryCount: 12,
    chunks: [
      "Vue 3 使用 Proxy 实现响应式：reactive() 把对象转成 Proxy，所有 get/set 操作经过 trapper 拦截，触发 track / trigger。",
      "ref() 把基本类型包成对象，用 .value 访问；template 中自动 unwrap，但在 reactive 里也会自动 unwrap。",
    ],
  },
  {
    id: "d-4",
    categoryId: "c-arch",
    name: "分布式事务方案对比.pdf",
    type: "pdf",
    size: "1.6 MB",
    chunkCount: 36,
    uploadedAt: "5 天前",
    status: "ready",
    embedding: 1024,
    queryCount: 8,
    chunks: [
      "TCC（Try-Confirm-Cancel）依赖业务侧自己实现幂等三阶段，强一致但侵入业务；适合金融、电商核心。",
      "Saga 把长事务拆成一系列本地事务 + 对应补偿事务；最终一致，对业务侵入小，但需要小心补偿失败场景。",
    ],
  },
  {
    id: "d-5",
    categoryId: "c-system",
    name: "Linux 网络栈速查.txt",
    type: "txt",
    size: "48 KB",
    chunkCount: 12,
    uploadedAt: "2 周前",
    status: "ready",
    embedding: 1024,
    queryCount: 4,
    chunks: [
      "TCP 三次握手：SYN -> SYN+ACK -> ACK，连接进入 ESTABLISHED；四次挥手 FIN -> ACK -> FIN -> ACK 后进入 TIME_WAIT。",
    ],
  },
  {
    id: "d-6",
    categoryId: "c-personal",
    name: "GoZero-AI 项目复盘.md",
    type: "md",
    size: "240 KB",
    chunkCount: 56,
    uploadedAt: "昨天",
    status: "processing",
    embedding: 1024,
    queryCount: 0,
    chunks: [],
  },
]);

const filteredDocs = computed(() => {
  return documents.value.filter((d) => d.categoryId === activeCategoryId.value);
});

const getDocCount = (catId) => {
  return documents.value.filter((d) => d.categoryId === catId).length;
};

const totalChunks = computed(() =>
  documents.value.reduce((sum, d) => sum + (d.chunkCount || 0), 0)
);

const totalVectorMb = computed(() => {
  // 粗略估算：每个 chunk × 1024 dim × 4 bytes (float32) / 1024 / 1024
  const bytes = totalChunks.value * 1024 * 4;
  return (bytes / (1024 * 1024)).toFixed(1);
});

const selectedDocId = ref("d-1");

const selectedDoc = computed(() => {
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

// 后端 scope ("public" / "private") → 本地默认分类
const mapScopeToCategory = (scope) => {
  if (!scope) return "c-go";
  const s = String(scope).toLowerCase();
  if (s === "private" || s === "user") return "c-personal";
  return "c-go";
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

// 拉取文档列表：后端返回 KnowledgeDocumentItem[] 。
const loadDocuments = async () => {
  try {
    const res = await apiService.chat.knowledgeDocuments({ limit: 50 });
    const list = Array.isArray(res?.documents) ? res.documents : [];
    if (list.length === 0) return; // 保留 mock
    documents.value = list.map((d) => ({
      id: String(d.documentId),
      categoryId: mapScopeToCategory(d.scope),
      name: d.title || `文档 ${d.documentId}`,
      type: inferFileType(d.title || ""),
      // 后端未返回原始字节数；用片段数作为可读代替。
      size: d.chunkCount > 0 ? `${d.chunkCount} 片段` : "—",
      chunkCount: d.chunkCount || 0,
      uploadedAt: formatRelativeTime(d.updatedAt || d.createdAt),
      status: mapKnowledgeStatus(d.status),
      embedding: 1024,
      queryCount: 0, // 后端后续可拓展 hits 字段
      // chunks 在选中时 lazy 拉，preview 先作为掩护首屏
      chunks: d.preview ? [d.preview] : [],
      chunksLoaded: false,
      summary: d.preview || "",
    }));
    // 选中首项，避免依然指向已不存在的 mock id
    if (documents.value[0] && !documents.value.find((d) => d.id === selectedDocId.value)) {
      selectedDocId.value = documents.value[0].id;
      activeCategoryId.value = documents.value[0].categoryId;
    }
  } catch (error) {
    // 静默降级
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
    // 静默降级：保留 preview 嵌入的入默认 chunk
  }
};

watch(selectedDocId, (id) => {
  if (id) loadDocumentChunks(id);
});

onMounted(() => {
  loadDocuments();
});

// 返回 mono 文件类型标签，替代原 emoji 图标，与 Home 页 mono 字符语言对齐。
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
.wb-kb-content {
  max-width: 1320px;
  margin: 0 auto;
  padding: 0 44px 80px;
}

/* === Hero === */
.wb-kb-hero {
  padding: 0 0 32px;
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

/* === Shell：三列 === */
.wb-kb-shell {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr) 320px;
  gap: 16px;
  align-items: start;
}

/* === 左侧 tree === */
.wb-kb-tree {
  position: sticky;
  top: 100px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 18px 16px 20px;
  background:
    linear-gradient(180deg, rgba(16, 17, 22, 1) 0%, rgba(10, 11, 14, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.03) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  isolation: isolate;
}

.wb-kb-tree-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.wb-kb-tree-title {
  font: 600 12px var(--mono);
  color: var(--t3);
  letter-spacing: .08em;
  text-transform: uppercase;
}

.wb-kb-tree-add {
  width: 22px;
  height: 22px;
  font: 700 14px var(--sans);
  line-height: 1;
  color: var(--t3);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color .2s ease, background-color .2s ease;
}

.wb-kb-tree-add:hover {
  color: var(--t);
  background: rgba(220, 155, 90, 0.08);
  border-color: rgba(220, 155, 90, 0.3);
}

.wb-kb-cat-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.wb-kb-cat {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font: 13px var(--sans);
  color: var(--t2);
  transition: color .2s ease, background-color .2s ease;
}

.wb-kb-cat:hover {
  color: var(--t);
  background: rgba(255, 255, 255, 0.03);
}

.wb-kb-cat-active {
  color: var(--t);
  background: rgba(220, 155, 90, 0.06);
  position: relative;
}

.wb-kb-cat-active::before {
  content: '';
  position: absolute;
  left: -16px;
  top: 8px;
  bottom: 8px;
  width: 2px;
  background: rgba(220, 155, 90, 0.95);
  border-radius: 0 2px 2px 0;
}

.wb-kb-cat-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  opacity: .85;
}

.wb-kb-cat-active .wb-kb-cat-dot {
  opacity: 1;
  box-shadow: 0 0 0 2px rgba(220, 155, 90, 0.18);
}

.wb-kb-cat-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-kb-cat-count {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
  flex-shrink: 0;
}

.wb-kb-tree-stats {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px 0 0;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.wb-kb-stat-line {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}

.wb-kb-stat-num {
  font: 600 14px var(--mono);
  color: var(--t);
}

.wb-kb-stat-lb {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

/* === 中间文档列表 === */
.wb-kb-list {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
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
  gap: 14px;
  padding: 14px 16px;
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
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
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
  top: 100px;
  padding: 22px 22px 24px;
  background:
    linear-gradient(180deg, rgba(16, 17, 22, 1) 0%, rgba(10, 11, 14, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.03) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  isolation: isolate;
  max-height: calc(100vh - 120px);
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
  font: 700 14px var(--display);
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
  font: 700 18px var(--mono);
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

.wb-kb-detail-actions {
  display: flex;
  gap: 8px;
  margin-top: 18px;
  padding-top: 14px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.wb-kb-action-btn {
  flex: 1;
  font: 600 12px var(--sans);
  color: var(--t2);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-sm);
  padding: 8px 10px;
  cursor: pointer;
  transition: color .2s ease, background-color .2s ease, border-color .2s ease;
}

.wb-kb-action-btn:hover {
  color: var(--t);
  border-color: rgba(255, 255, 255, 0.16);
}

.wb-kb-action-danger {
  color: #ef8a73;
}

.wb-kb-action-danger:hover {
  color: #ef6660;
  border-color: rgba(239, 102, 96, 0.4);
  background: rgba(239, 102, 96, 0.06);
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

/* === 响应式 === */
@media (max-width: 1200px) {
  .wb-kb-shell {
    grid-template-columns: 200px minmax(0, 1fr);
  }
  .wb-kb-detail {
    grid-column: 1 / -1;
    position: static;
    max-height: none;
  }
}

@media (max-width: 900px) {
  .wb-kb-shell {
    grid-template-columns: 1fr;
  }
  .wb-kb-tree {
    position: static;
  }
}

@media (max-width: 768px) {
  .wb-kb-content {
    padding: 0 20px 60px;
  }
  .wb-kb-detail-stats {
    grid-template-columns: repeat(3, 1fr);
  }
}
</style>
