<template>
  <!--
    WorkbenchResume：简历管理。
    布局：顶部 hero + 上传 dropzone + 已上传简历卡片列表 + 选中简历的解析详情面板。
    后端契约：上传走 /users/resume/upload (multipart/form-data)；
    上传成功后自动触发 /users/resume/artifacts/:id/analysis/prepare 生成持久化评估。
  -->
  <WorkbenchLayout>
    <div class="wb-resume-content">
      <section class="wb-resume-hero">
        <div class="wb-eyebrow">
          <span class="wb-eyebrow-dot" aria-hidden="true"></span>
          <span>简历管理</span>
        </div>
        <h1 class="wb-resume-title">让 AI 读懂你的项目</h1>
        <p class="wb-resume-sub">上传简历后，AI 会基于项目细节做深度追问，模拟真实面试节奏。</p>
      </section>

      <!--
        三栏 shell（v3 布局：左评估 / 中原文 chunks / 右追问 + CTA）。
        详见 docs/requirements/2026-05-12-workbench-resume-redesign.md §6.1-6.4。
        本 commit（C2）仅落 S0 中栏 dropzone + 左右栏占位，后续 commit 填充：
          - C3 左栏 dropdown selector + 简历列表载入 + ?artifact= query 同步
          - C4 左栏 5 维评估 + tooltip + 总结/强项/风险/建议 + OverallScore 圆环
          - C5 中栏 S2 chunks 列表 + S3 失败提示
          - C6 右栏 AI 追问 + CTA + 降级声明 + chunk 联动
          - C7 S1 polling + 状态机 + 5 分钟硬超时 + 手动刷新
      -->
      <div class="wb-resume-shell" :data-state="resumeState">
        <!-- 左栏：简历版本切换 dropdown（C3 commit）+ 元数据 + 评估占位（C4 填充） -->
        <aside class="wb-resume-left" aria-label="简历评估">
          <!-- C3: 简历版本切换 selector。与 ?artifact= URL query 双向同步 -->
          <div
            v-if="resumes.length > 0"
            class="wb-resume-selector"
            :class="{ 'wb-resume-selector-open': selectorOpen }"
          >
            <button
              type="button"
              class="wb-resume-selector-trigger"
              :aria-expanded="selectorOpen"
              aria-haspopup="listbox"
              @click.stop="toggleSelector"
            >
              <span class="wb-resume-selector-name">{{ selectedResume?.name || '选择简历' }}</span>
              <span
                v-if="selectedResume?.overallScore != null && selectedResume?.evaluationStatus === 'ready'"
                class="wb-resume-selector-score"
              >{{ formatScore(selectedResume.overallScore) }} 分</span>
              <span class="wb-resume-selector-caret" aria-hidden="true">▾</span>
            </button>
            <ul v-if="selectorOpen" class="wb-resume-selector-menu" role="listbox">
              <li
                v-for="r in resumes"
                :key="r.id"
                class="wb-resume-selector-item"
                :class="{ 'wb-resume-selector-item-active': r.id === selectedId }"
                role="option"
                :aria-selected="r.id === selectedId"
                @click.stop="handleSelectArtifact(r.id)"
              >
                <span class="wb-resume-selector-item-name">{{ r.name }}</span>
                <span
                  v-if="r.overallScore != null && r.evaluationStatus === 'ready'"
                  class="wb-resume-selector-item-score"
                >{{ formatScore(r.overallScore) }} 分</span>
              </li>
            </ul>
          </div>

          <!-- 元数据条（选中后显示：技能数 / chunk 数 + 上传时间） -->
          <div v-if="selectedResume" class="wb-resume-meta-bar">
            <span>{{ selectedResume.skillCount ? `${selectedResume.skillCount} 技能` : (selectedResume.size || '—') }}</span>
            <span aria-hidden="true">·</span>
            <span>{{ selectedResume.uploadedAt }}</span>
          </div>

          <!-- C4: 评估卡。有评估数据时渲染（圆环 + 总结 + 5 维 + 强项 + 风险 + 建议），
               否则 fallback 到原占位。 -->
          <div v-if="hasEvaluation" class="wb-resume-eval-card">
            <!-- OverallScore 圆环 + Level 徽章 -->
            <div class="wb-overall-score-wrap">
              <div class="wb-overall-score">
                <svg viewBox="0 0 100 100" class="wb-overall-svg" aria-hidden="true">
                  <circle cx="50" cy="50" r="44" class="wb-overall-track" />
                  <circle
                    cx="50" cy="50" r="44"
                    class="wb-overall-fill"
                    :class="`wb-overall-fill-${selectedResume.level || 'mid'}`"
                    :style="overallCircleStyle"
                  />
                </svg>
                <div class="wb-overall-center">
                  <span class="wb-overall-num">{{ formatScore(selectedResume.overallScore) }}</span>
                  <span class="wb-overall-meta">/ 100</span>
                </div>
              </div>
              <div
                class="wb-overall-badge"
                :class="`wb-overall-badge-${selectedResume.level || 'mid'}`"
              >{{ levelLabel(selectedResume.level) }}</div>
            </div>

            <!-- AI 总结 -->
            <p v-if="selectedResume.summary" class="wb-resume-summary">{{ selectedResume.summary }}</p>

            <!-- 5 维评估（D-U8 删 target_alignment）。hover 显示评分标准 tooltip（D-U9） -->
            <div v-if="filteredDimensions.length > 0" class="wb-dimensions-list">
              <div
                v-for="dim in filteredDimensions"
                :key="dim.key"
                class="wb-dim-row"
                :aria-describedby="`wb-dim-tip-${dim.key}`"
              >
                <div class="wb-dim-head">
                  <span class="wb-dim-label">{{ dim.label }}</span>
                  <span class="wb-dim-score">{{ dim.score }}</span>
                </div>
                <div class="wb-dim-bar" aria-hidden="true">
                  <span :style="{ width: `${Math.min(100, Math.max(0, dim.score || 0))}%` }"></span>
                </div>
                <div
                  v-if="dim.summary"
                  :id="`wb-dim-tip-${dim.key}`"
                  class="wb-dim-tooltip"
                  role="tooltip"
                >{{ dim.summary }}</div>
              </div>
            </div>

            <!-- 强项 -->
            <div v-if="selectedResume.strengths?.length > 0" class="wb-eval-section">
              <div class="wb-eval-section-label">强项</div>
              <ul class="wb-eval-list">
                <li v-for="s in selectedResume.strengths" :key="s">{{ s }}</li>
              </ul>
            </div>

            <!-- 风险（三色 left border 按 severity） -->
            <div v-if="selectedResume.risks?.length > 0" class="wb-eval-section">
              <div class="wb-eval-section-label">风险</div>
              <ul class="wb-eval-risks">
                <li
                  v-for="r in selectedResume.risks"
                  :key="r.key || r.label"
                  :class="`wb-risk-${r.severity || 'medium'}`"
                >
                  <span class="wb-risk-label">{{ r.label }}</span>
                  <span v-if="r.suggestion" class="wb-risk-suggest">{{ r.suggestion }}</span>
                </li>
              </ul>
            </div>

            <!-- 建议 -->
            <div v-if="selectedResume.suggestions?.length > 0" class="wb-eval-section">
              <div class="wb-eval-section-label">建议</div>
              <ul class="wb-eval-list">
                <li v-for="s in selectedResume.suggestions" :key="s">{{ s }}</li>
              </ul>
            </div>
          </div>

          <!-- 评估未就绪时的 fallback 占位卡 -->
          <div v-else class="wb-resume-placeholder wb-resume-placeholder--left">
            <div class="wb-resume-placeholder-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4">
                <circle cx="12" cy="12" r="9" />
                <path d="M12 7v5l3 2" stroke-linecap="round" />
              </svg>
            </div>
            <p v-if="selectedResume?.evaluationLoading">评估生成中…</p>
            <p v-else-if="selectedResume">评估尚未生成<br><strong>点右栏 重新生成 AI 画像</strong></p>
            <p v-else>上传简历后这里会显示<br><strong>总评分 + 5 维评估</strong></p>
          </div>
        </aside>

        <!-- 中栏 flex 1：根据 resumeState 渲染对应内容 -->
        <main class="wb-resume-mid" aria-label="简历主内容区">
          <!-- S0 未上传：dropzone 垂直居中 -->
          <div v-if="resumeState === 'S0'" class="wb-resume-mid-empty">
            <div
              class="wb-dropzone wb-dropzone--centered"
              :class="{ 'wb-dropzone-dragging': isDragging, 'wb-dropzone-uploading': uploading }"
              @dragenter.prevent="handleDragEnter"
              @dragover.prevent
              @dragleave.prevent="handleDragLeave"
              @drop.prevent="handleDrop"
              @click="triggerFileInput"
            >
              <input
                ref="fileInputRef"
                type="file"
                accept=".pdf,application/pdf"
                class="wb-file-input"
                @change="handleFileChange"
              />
              <div class="wb-dropzone-inner">
                <div class="wb-dropzone-icon" aria-hidden="true">
                  <svg viewBox="0 0 64 64" fill="none">
                    <rect x="14" y="10" width="36" height="44" rx="3" stroke="currentColor" stroke-width="1.5" />
                    <line x1="22" y1="22" x2="42" y2="22" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
                    <line x1="22" y1="30" x2="42" y2="30" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
                    <line x1="22" y1="38" x2="34" y2="38" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
                    <circle cx="48" cy="48" r="10" fill="rgba(220,155,90,0.12)" stroke="rgba(220,155,90,0.85)" stroke-width="1.5" />
                    <line x1="48" y1="44" x2="48" y2="52" stroke="rgba(220,155,90,0.95)" stroke-width="1.5" stroke-linecap="round" />
                    <line x1="44" y1="48" x2="52" y2="48" stroke="rgba(220,155,90,0.95)" stroke-width="1.5" stroke-linecap="round" />
                  </svg>
                </div>
                <div class="wb-dropzone-text">
                  <div class="wb-dropzone-main">{{ uploading ? '正在上传...' : '上传你的简历' }}</div>
                  <div class="wb-dropzone-sub">
                    <span v-if="!uploading">拖拽 PDF 到此处或</span>
                    <span class="wb-dropzone-action" v-if="!uploading">点击选择文件</span>
                    <span class="wb-dropzone-meta">支持 PDF · 最大 10 MB · 解析约 30 秒</span>
                  </div>
                </div>
              </div>
              <div v-if="uploading" class="wb-upload-progress">
                <div class="wb-upload-bar" :style="{ width: uploadProgress + '%' }"></div>
              </div>
            </div>
            <p v-if="uploadError" class="wb-upload-error" role="alert">{{ uploadError }}</p>
          </div>

          <!-- S2 解析完成（C5 commit）：原文 chunks 列表 -->
          <div v-else-if="resumeState === 'S2'" class="wb-resume-mid-chunks">
            <!-- chunks 加载中。watch selectedId 会触发 loadResumeChunks，未返回前显示骨架。 -->
            <div v-if="chunksLoading && !selectedResume?.chunks?.length" class="wb-resume-mid-placeholder">
              <div class="wb-resume-placeholder-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4">
                  <circle cx="12" cy="12" r="9" />
                  <path d="M12 7v5" stroke-linecap="round" />
                </svg>
              </div>
              <p>加载简历原文中…</p>
            </div>

            <!-- 后端返回空 chunks（朗读不到文本 / 清洗后）。 -->
            <div v-else-if="!selectedResume?.chunks?.length" class="wb-resume-mid-placeholder">
              <div class="wb-resume-placeholder-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4">
                  <rect x="4" y="3" width="16" height="18" rx="2" />
                  <line x1="8" y1="8" x2="16" y2="8" stroke-linecap="round" />
                  <line x1="8" y1="12" x2="16" y2="12" stroke-linecap="round" />
                </svg>
              </div>
              <p>暂无可显示的原文片段</p>
              <p class="wb-resume-placeholder-meta">请文本型 PDF 或联系后端检查解析质量</p>
            </div>

            <!-- 正常 chunks 列表。按 chunk.index 升序渲染。 -->
            <ol v-else class="wb-chunks-list">
              <li
                v-for="chunk in selectedResume.chunks"
                :key="chunk.index"
                class="wb-chunk-card"
                :data-chunk-index="chunk.index"
              >
                <header class="wb-chunk-head">
                  <!-- TODO(phase2-resume-chunk-pagenum): 后端补 ResumeArtifactChunk.PageNumber 字段后改为 "chunk #N · 第 X 页"。
                       当前实现：仅显示 chunk 序号 #NN，后端 ResumeArtifactChunk 只有 Index 字段不是页码。
                       后端状态：未规划（需与后端确认 chunk 切分逻辑是否能映射页码）。
                       对齐目标：渲染 "chunk #NN · 第 X 页" 格式。
                       触发条件：@d:\Go-Project\GoZero-AI\api\user\internal\types\types.go 出现 ResumeArtifactChunk.PageNumber 字段。
                       起草日期：2026-05-12 -->
                  <span class="wb-chunk-num">chunk #{{ String(chunk.index ?? '?').padStart(2, '0') }}</span>
                </header>
                <p class="wb-chunk-content">{{ chunk.content }}</p>
              </li>
            </ol>
          </div>

          <!-- S1 解析中 / S3 解析失败 占位，由 C7 接管状态机后细化进度 / 失败详情 -->
          <div v-else class="wb-resume-mid-placeholder">
            <div class="wb-resume-placeholder-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4">
                <rect x="4" y="3" width="16" height="18" rx="2" />
                <line x1="8" y1="8" x2="16" y2="8" stroke-linecap="round" />
                <line x1="8" y1="12" x2="16" y2="12" stroke-linecap="round" />
                <line x1="8" y1="16" x2="13" y2="16" stroke-linecap="round" />
              </svg>
            </div>
            <p v-if="resumeState === 'S1'">解析中…</p>
            <p v-else-if="resumeState === 'S3'">解析失败</p>
            <p v-else>已选中：<strong>{{ selectedResume?.name }}</strong></p>
            <p class="wb-resume-placeholder-meta">当前状态 <code>{{ resumeState }}</code>；C7 commit 接入 polling 后细化</p>
          </div>
        </main>

        <!-- 右栏：AI 追问 + CTA（C6 commit） -->
        <aside class="wb-resume-right" aria-label="AI 追问与开始面试">
          <!-- 选中简历后渲染完整右栏；未选中时 fallback 到原占位 -->
          <div v-if="selectedResume" class="wb-resume-right-card">
            <!-- 方向匹配 chip 列表：FocusMatches[] -->
            <div v-if="(selectedResume.focusMatches?.length || 0) > 0" class="wb-focus-matches">
              <div class="wb-section-label">方向匹配</div>
              <div
                v-for="match in selectedResume.focusMatches"
                :key="match.key"
                class="wb-focus-chip"
              >
                <div class="wb-focus-chip-head">
                  <span class="wb-focus-chip-label">{{ match.label }}</span>
                  <span class="wb-focus-chip-score">{{ match.matchScore }}%</span>
                </div>
                <div class="wb-focus-chip-bar" aria-hidden="true">
                  <span :style="{ width: `${Math.min(100, Math.max(0, match.matchScore || 0))}%` }"></span>
                </div>
                <p v-if="(match.plannedQuestion || 0) > 0" class="wb-focus-chip-meta">打算追问 {{ match.plannedQuestion }} 题</p>
              </div>
            </div>

            <!-- AI 追问问题列表：SuggestedQuestions[]。点击 chunk 联动高亮（phase2-resume-chunk-question-link） -->
            <div v-if="(selectedResume.suggestedQuestions?.length || 0) > 0" class="wb-questions-list">
              <div class="wb-section-label">AI 追问问题</div>
              <button
                v-for="(q, i) in selectedResume.suggestedQuestions.slice(0, 6)"
                :key="q.key || `q-${i}`"
                type="button"
                class="wb-question-card"
                @click="handleQuestionClick(q)"
              >
                <div class="wb-question-head">
                  <span class="wb-question-num">#{{ String(i + 1).padStart(2, '0') }}</span>
                  <span
                    v-if="q.difficultyLabel"
                    class="wb-question-chip wb-question-chip-difficulty"
                  >{{ q.difficultyLabel }}</span>
                  <span
                    v-if="q.focusLabel"
                    class="wb-question-chip wb-question-chip-focus"
                  >{{ q.focusLabel }}</span>
                </div>
                <p class="wb-question-title">{{ q.title || q.prompt }}</p>
              </button>
            </div>

            <!-- 评估 fallback：后端未返 focusMatches / suggestedQuestions 时的提示 -->
            <div
              v-else-if="hasEvaluation && !(selectedResume.focusMatches?.length || 0)"
              class="wb-question-fallback"
            >
              <p class="wb-section-label">AI 追问问题</p>
              <p class="wb-question-fallback-text">当前评估未生成追问。点下方 重新生成 AI 画像 刷新评估。</p>
            </div>

            <!-- 操作区：重新生成 + 看完整详情 -->
            <div class="wb-actions">
              <button
                type="button"
                class="wb-action-regenerate"
                :disabled="selectedResume.evaluationLoading"
                @click="prepareResumeEvaluation(selectedResume.id, { force: true })"
              >
                <span class="wb-action-icon" aria-hidden="true">↻</span>
                <span>{{ selectedResume.evaluationLoading ? '生成中…' : '重新生成 AI 画像' }}</span>
              </button>
              <p v-if="selectedResume.evaluationMeta?.lastRefreshedAt" class="wb-action-meta">
                最后刷新：{{ formatRelativeTime(selectedResume.evaluationMeta.lastRefreshedAt) }}
                <span aria-hidden="true">·</span>
                来源：{{ scoreSourceLabel(selectedResume.evaluationMeta.scoreSource) }}
              </p>
              <RouterLink
                :to="`/workbench/resume/${selectedResume.id}`"
                class="wb-action-detail"
              >看完整详情 →</RouterLink>
            </div>

            <!-- TODO(phase2-resume-cta-disclaimer): 后端 CreateSession 接受 resumeId 后移除本声明。
                 当前实现：11px 灰字两行，明确告知用户追问仅供参考。
                 后端状态：正在开发（用户 2026-05-12 01:19 告知）。
                 对齐目标：本声明文本整块删除，CTA 上方不再需要提示。
                 触发条件：@d:\Go-Project\GoZero-AI\api\user\user.api 中 CreateSessionReq 出现 resumeId 字段。
                 起草日期：2026-05-12 -->
            <p class="wb-cta-disclaimer">
              面试会使用这份简历的方向偏好；追问列表当前仅供参考。
            </p>

            <!-- TODO(phase2-resume-cta-handler): 后端 CreateSession 接受 resumeId 后接入。
                 当前实现：仅 router.push('/workbench/new?resumeId=:id')，不能把 SuggestedQuestions 带进新面试。
                 后端状态：正在开发（用户 2026-05-12 01:19 告知）。
                 对齐目标：改为 apiService.user.createSession({ resumeId, directionKey, focusKeys })，
                          后端自动从 resume_evaluations.suggested_questions 取 [0] 作首题。
                 触发条件：@d:\Go-Project\GoZero-AI\api\user\user.api 中 CreateSessionReq 出现 resumeId 字段。
                 起草日期：2026-05-12 -->
            <button
              type="button"
              class="wb-cta-primary"
              @click="handleStartInterview"
            >
              <span>用这份简历开始面试</span>
              <span class="wb-cta-arrow" aria-hidden="true">→</span>
            </button>
          </div>

          <!-- 未选中简历时 fallback 占位 -->
          <div v-else class="wb-resume-placeholder wb-resume-placeholder--right">
            <div class="wb-resume-placeholder-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4">
                <rect x="3" y="5" width="18" height="14" rx="2" />
                <path d="M5 8l7 5 7-5" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </div>
            <p>上传简历后这里会出现<br><strong>AI 追问</strong> 和 <strong>开始面试</strong> 入口</p>
          </div>
        </aside>
      </div>
    </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
import { apiService } from "../composables/useApi";

const route = useRoute();
const router = useRouter();

// === 上传 ===
const fileInputRef = ref(null);
const isDragging = ref(false);
const uploading = ref(false);
const uploadProgress = ref(0);
const uploadError = ref("");

const handleDragEnter = () => {
  isDragging.value = true;
};

const handleDragLeave = (e) => {
  // 只在真正离开 dropzone 时才取消高亮（不算冒泡到子元素）
  if (e.currentTarget === e.target) {
    isDragging.value = false;
  }
};

const handleDrop = (e) => {
  isDragging.value = false;
  const files = e.dataTransfer?.files;
  if (files && files.length > 0) {
    uploadFile(files[0]);
  }
};

const triggerFileInput = () => {
  if (uploading.value) return;
  fileInputRef.value?.click();
};

const handleFileChange = (e) => {
  const file = e.target.files?.[0];
  if (file) {
    uploadFile(file);
  }
};

const validateFile = (file) => {
  const maxBytes = 10 * 1024 * 1024;
  if (file.size > maxBytes) {
    return "文件过大，最大支持 10 MB";
  }
  if (!/\.pdf$/i.test(file.name || "")) {
    return "仅支持 PDF 格式";
  }
  return "";
};

const uploadFile = async (file) => {
  uploadError.value = "";
  const validationError = validateFile(file);
  if (validationError) {
    uploadError.value = validationError;
    return;
  }

  uploading.value = true;
  uploadProgress.value = 10;

  // 视觉伪进度：真实进度需要 axios upload progress event，
  // 当前 apiService.user.resumeUpload 直接返回 Promise，无 onUploadProgress 接入。
  // 后续可改造 endpoint 工厂支持 progress callback。
  const fakeTimer = setInterval(() => {
    if (uploadProgress.value < 85) {
      uploadProgress.value += 5;
    }
  }, 200);

  try {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("title", file.name);
    formData.append("mode", "Memory");
    const res = await apiService.user.resumeUpload(formData);
    uploadProgress.value = 100;

    // 成功后立即拉列表，让后端返回的 artifactId / status 为准。
    await loadResumes();
    const targetId = res?.artifactId || resumes.value[0]?.id || "";
    if (targetId) {
      selectedId.value = targetId;
      void prepareResumeEvaluation(targetId, { force: true });
    }
  } catch (error) {
    uploadError.value = error?.message || "上传失败，请稍后再试";
  } finally {
    clearInterval(fakeTimer);
    uploading.value = false;
    uploadProgress.value = 0;
    if (fileInputRef.value) {
      fileInputRef.value.value = "";
    }
  }
};

const formatBytes = (bytes) => {
  if (!bytes || bytes < 0) return "—";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
};

const createEmptyEvaluationState = () => ({
  evaluationStatus: "missing",
  evaluationLoading: false,
  evaluationLoaded: false,
  overallScore: null,
  level: "",
  summary: "",
  dimensions: [],
  strengths: [],
  risks: [],
  suggestions: [],
  evidence: [],
  // C5: chunks 缓存。选中某份简历后调 resumeArtifactDetail 拉取一次，不重复拉。
  chunks: [],
  chunksLoaded: false,
  // C6: 合后端 ResumeArtifactAnalysisResp.FocusMatches / SuggestedQuestions / EvaluationMeta
  focusMatches: [],
  suggestedQuestions: [],
  evaluationMeta: null,
});

// === 简历列表（mock first，onMounted 异步接入 resumeArtifacts 覆盖） ===
const resumes = ref([
  {
    id: "r-v3",
    name: "Resume_v3.pdf",
    size: "1.2 MB",
    uploadedAt: "2 天前",
    projectCount: 12,
    skillCount: 28,
    status: "parsed",
    primary: true,
    skills: ["Go", "Vue", "Postgres", "Redis", "Docker", "ETCD", "gRPC", "RAG"],
    projects: [
      { name: "GoZero-AI 个人面试官", stack: "Go-Zero · Vue 3 · pgvector" },
      { name: "微服务订单系统", stack: "Go · Kafka · MySQL" },
      { name: "实时协作白板", stack: "WebSocket · CRDT · Redis" },
    ],
    ...createEmptyEvaluationState(),
    evaluationStatus: "ready",
    overallScore: 82,
    summary: "项目素材完整，适合围绕微服务、RAG 和工程实践展开追问。",
    strengths: ["技术栈清晰", "项目素材丰富"],
    suggestions: ["补充核心项目的量化指标。"],
    risks: [],
  },
  {
    id: "r-v2",
    name: "Resume_v2.pdf",
    size: "1.0 MB",
    uploadedAt: "1 周前",
    projectCount: 10,
    skillCount: 22,
    status: "parsed",
    primary: false,
    skills: ["Go", "Vue", "Postgres", "Docker"],
    projects: [
      { name: "面试系统 v1", stack: "Go · MySQL · Vue" },
      { name: "blog 后端", stack: "Go · MongoDB" },
    ],
    ...createEmptyEvaluationState(),
    evaluationStatus: "ready",
    overallScore: 74,
    summary: "已有基础项目线索，仍需补充职责边界和结果证据。",
    strengths: ["方向明确"],
    suggestions: ["补充项目职责和优化结果。"],
    risks: [],
  },
]);

const selectedId = ref("");

const selectedResume = computed(() => {
  return resumes.value.find((r) => r.id === selectedId.value) || null;
});

// 简历状态机（C2 commit 落静态判定，C7 commit 接入 polling 实时刷新）
// 详见 docs/requirements/2026-05-12-workbench-resume-redesign.md §7.1
//   S0 未上传：无简历或未选中
//   S1 解析中：artifact.status === "parsing"
//   S2 解析完成：artifact.status === "parsed" && evaluationStatus 有效
//   S3 解析失败：artifact.status === "failed" 或 evaluationStatus 为 noData/insufficient_data
const resumeState = computed(() => {
  const r = selectedResume.value;
  if (!r) return "S0";
  if (r.status === "parsing") return "S1";
  if (r.status === "failed") return "S3";
  if (r.status === "parsed") {
    const evalStatus = r.evaluationStatus || "missing";
    if (evalStatus === "insufficient_data" || evalStatus === "noData") return "S3";
    return "S2";
  }
  return "S0";
});

// === C4: 评估卡 helpers ===
// 删除的维度 keys（统一保持前端 5 维展示）：
// - target_alignment（D-U8）：依赖 DirectionKey 输入，同一简历在不同目标方向下分数
//   不稳定，且与右栏 FocusMatches chip 语义重复。
// - interview_readiness：仅在后端 heuristic fallback 模式返回，语义与右栏
//   SuggestedQuestions 数量重复（"可追问度" ≈ "AI 能问几题"），前端不重复展示。
//   LLM 模式不返回该维度，过滤是 idempotent。
const DIMENSION_OMITTED_KEYS = new Set(["target_alignment", "interview_readiness"]);

// 判定评估是否就绪：有 dimensions 或 summary 即可认为有可渲染评估内容。
const hasEvaluation = computed(() => {
  const r = selectedResume.value;
  if (!r) return false;
  if (r.evaluationStatus !== "ready") return false;
  return (Array.isArray(r.dimensions) && r.dimensions.length > 0) || !!r.summary;
});

// 5 维评估（过滤 D-U8 删除的 target_alignment）。
const filteredDimensions = computed(() => {
  const dims = selectedResume.value?.dimensions || [];
  return dims.filter((d) => !DIMENSION_OMITTED_KEYS.has(d.key));
});

// OverallScore SVG 圆环：stroke-dasharray + dashoffset 实现进度环。
// circumference = 2 * π * r = 2 * π * 44 ≈ 276.46
const overallCircleStyle = computed(() => {
  const score = Number(selectedResume.value?.overallScore) || 0;
  const clamped = Math.max(0, Math.min(100, score));
  const circumference = 2 * Math.PI * 44;
  const dashOffset = circumference * (1 - clamped / 100);
  return {
    strokeDasharray: String(circumference),
    strokeDashoffset: String(dashOffset),
  };
});

// Level 中文标签。后端返回 strong/mid/weak 等英文枚举。
const levelLabel = (level) => {
  switch ((level || "").toLowerCase()) {
    case "strong":
    case "high":
      return "表现优秀";
    case "weak":
    case "low":
      return "需加强";
    case "mid":
    case "medium":
      return "表现中等";
    default:
      return "评估完成";
  }
};

// === C3: 简历版本切换 dropdown selector ===
// 设计要点：
//   - selector trigger 点击后弹出 menu，点选项后 selectedId 更新 + URL ?artifact= 同步
//   - 点击 selector 外部自动关闭menu（onMounted/onUnmounted 绑定 document click）
//   - 浏览器前进后退（改变 ?artifact=）同步 selectedId（watch route.query.artifact）
//   - 与 D-U3 钻深页 push 路由策略分离：本处用 router.replace（不污染历史）
const selectorOpen = ref(false);

const toggleSelector = () => {
  selectorOpen.value = !selectorOpen.value;
};

const handleSelectArtifact = (id) => {
  if (id === selectedId.value) {
    selectorOpen.value = false;
    return;
  }
  selectedId.value = id;
  selectorOpen.value = false;
  router.replace({ query: { ...route.query, artifact: id } });
};

// 点击 selector 外部关闭 dropdown
const handleClickOutsideSelector = (e) => {
  if (!e.target.closest('.wb-resume-selector')) {
    selectorOpen.value = false;
  }
};

// URL query 变化（浏览器后退 / 外部链接）→ 同步 selectedId
watch(
  () => route.query.artifact,
  (queryId) => {
    const id = String(queryId || '');
    if (id && id !== selectedId.value && resumes.value.find((r) => r.id === id)) {
      selectedId.value = id;
    }
  }
);

// 后端 status (string) → 本地状态表（UI：parsed/parsing/failed）
const mapArtifactStatus = (raw) => {
  if (!raw) return "parsing";
  const s = String(raw).toLowerCase();
  if (s.includes("ready") || s.includes("parsed") || s.includes("success")) return "parsed";
  if (s.includes("fail") || s.includes("error")) return "failed";
  return "parsing";
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

// 列表拉取：后端返回资产与评估摘要，详情分块在选中时 lazy load。
const loadResumes = async () => {
  try {
    const res = await apiService.user.resumeArtifacts();
    const list = Array.isArray(res?.artifacts) ? res.artifacts : [];
    if (list.length === 0) return; // 保留 mock

    resumes.value = list.map((it, i) => ({
      id: it.artifactId,
      name: it.title || it.filename || `简历 v${it.version}`,
      size: it.chunkCount > 0 ? `${it.chunkCount} 片段` : "—",
      uploadedAt: formatRelativeTime(it.updatedAt || it.uploadedAt),
      projectCount: it.projectCount || 0,
      skillCount: it.skillCount || 0,
      status: mapArtifactStatus(it.status),
      primary: i === 0,
      skills: [],
      projects: [],
      ...createEmptyEvaluationState(),
      evaluationStatus: it.evaluationStatus || "missing",
      overallScore: typeof it.overallScore === "number" ? it.overallScore : null,
      level: it.level || "",
    }));
  } catch (error) {
    // 静默降级；mock 列表已可用
  }
};

const updateResume = (id, patch) => {
  const idx = resumes.value.findIndex((r) => r.id === id);
  if (idx < 0) return;
  resumes.value[idx] = {
    ...resumes.value[idx],
    ...patch,
  };
};

const applyResumeAnalysis = (id, res) => {
  const idx = resumes.value.findIndex((r) => r.id === id);
  if (idx < 0 || !res) return;
  const target = resumes.value[idx];

  const skills = Array.isArray(res.skills) ? res.skills.map((s) => s.label).filter(Boolean) : [];
  const projects = Array.isArray(res.projects)
    ? res.projects.map((p) => ({
      name: p.title || "未命名项目",
      stack: p.summary || (Array.isArray(p.evidence) ? p.evidence.slice(0, 2).join(" · ") : ""),
    }))
    : [];

  resumes.value[idx] = {
    ...target,
    skills,
    projects,
    skillCount: skills.length,
    projectCount: projects.length,
    evaluationStatus: res.evaluationStatus || target.evaluationStatus || "missing",
    overallScore: typeof res.overallScore === "number" ? res.overallScore : target.overallScore,
    level: res.level || target.level || "",
    summary: res.summary || target.summary || "",
    dimensions: Array.isArray(res.dimensions) ? res.dimensions : [],
    strengths: Array.isArray(res.strengths) ? res.strengths : [],
    risks: Array.isArray(res.risks) ? res.risks : [],
    suggestions: Array.isArray(res.suggestions) ? res.suggestions : [],
    evidence: Array.isArray(res.evidence) ? res.evidence : [],
    // C6: 右栏需要的 FocusMatches / SuggestedQuestions / EvaluationMeta
    focusMatches: Array.isArray(res.focusMatches) ? res.focusMatches : [],
    suggestedQuestions: Array.isArray(res.suggestedQuestions) ? res.suggestedQuestions : [],
    evaluationMeta: res.evaluationMeta || null,
    evaluationLoaded: true,
  };
};

// === C6: 右栏交互 helpers ===
// findRelatedChunk：根据 SuggestedQuestion.expectedSignals[] 与 chunk.content 做关键词匹配。
// 返回匹配率最高的 chunk，命中率 < 50% 返 null（需求文档 §12 风险降级项）。
const findRelatedChunk = (question) => {
  const signals = Array.isArray(question?.expectedSignals) ? question.expectedSignals : [];
  const chunks = selectedResume.value?.chunks || [];
  if (!signals.length || !chunks.length) return null;

  let bestChunk = null;
  let bestScore = 0;
  chunks.forEach((chunk) => {
    const content = String(chunk.content || '').toLowerCase();
    let score = 0;
    signals.forEach((sig) => {
      const signal = String(sig || '').toLowerCase();
      if (signal && content.includes(signal)) score++;
    });
    if (score > bestScore) {
      bestScore = score;
      bestChunk = chunk;
    }
  });

  // 命中率 < 50% 不联动（防止误高亮不相关 chunk）
  if (bestScore / signals.length < 0.5) {
    console.warn('[resume] chunk-question-link match rate too low (', bestScore, '/', signals.length, ') for question:', question?.title);
    return null;
  }
  return bestChunk;
};

// handleQuestionClick：点追问 → 中栏滚动到关联 chunk + 0.6s 黄色高亮。
// 命中率 < 50% 时 console.warn 且不滚动（D-U4 单向联动 + 需求文档 §12）。
const handleQuestionClick = (question) => {
  const chunk = findRelatedChunk(question);
  if (!chunk) return;
  const el = document.querySelector(`[data-chunk-index="${chunk.index}"]`);
  if (!el) return;
  el.scrollIntoView({ behavior: 'smooth', block: 'center' });
  el.classList.add('wb-chunk-highlight');
  setTimeout(() => el.classList.remove('wb-chunk-highlight'), 600);
};

// handleStartInterview：金色 CTA。
// TODO(phase2-resume-cta-handler): 后端 CreateSession 接受 resumeId 后改为直接
//   apiService.user.createSession({ resumeId: id, directionKey, focusKeys, ... })，
//   后端从 resume_evaluations.suggested_questions 拿 [0] 作首题。
//   触发条件：user.api CreateSessionReq 出现 resumeId 字段。
const handleStartInterview = () => {
  const id = selectedResume.value?.id;
  if (!id) return;
  router.push({ path: '/workbench/new', query: { resumeId: id } });
};

// scoreSourceLabel：后端 ScoreSource 枚举 → 中文标签。
const scoreSourceLabel = (source) => {
  switch ((source || '').toLowerCase()) {
    case 'llm':
      return 'LLM';
    case 'heuristic':
    case 'fallback':
      return '规则降级';
    default:
      return '未知';
  }
};

// === C5: 中栏 chunks 拉取 ===
// resumeArtifactDetail 返回 chunks 列表（index + content）。选中后 lazy load 一次，
// 缓存在 selectedResume.chunks；后续重选同一份简历不重复拉。
// chunksLoading 仅在首次拉取时为 true，当前州帝总体加载状态。
const chunksLoading = ref(false);

const loadResumeChunks = async (id) => {
  if (!id) return;
  const idx = resumes.value.findIndex((r) => r.id === id);
  if (idx < 0) return;
  const target = resumes.value[idx];
  if (target.chunksLoaded) return; // 已拉过不重复

  chunksLoading.value = true;
  try {
    const res = await apiService.user.resumeArtifactDetail(id);
    const chunks = Array.isArray(res?.chunks) ? res.chunks : [];
    resumes.value[idx] = {
      ...resumes.value[idx],
      chunks,
      chunksLoaded: true,
    };
  } catch (error) {
    // 静默降级：拱失败不报错，mid placeholder 会进 "暂无可显示的原文片段" 分支。
    console.warn('[resume] resumeArtifactDetail failed for', id, error);
  } finally {
    chunksLoading.value = false;
  }
};

// 选中某份简历后 lazy 拉 analysis 覆盖 skills/projects。
const loadResumeAnalysis = async (id) => {
  if (!id) return;
  const idx = resumes.value.findIndex((r) => r.id === id);
  if (idx < 0) return;
  const target = resumes.value[idx];
  if (target.evaluationLoaded) return; // 已拉过不重复

  try {
    const res = await apiService.user.resumeArtifactAnalysis(id, { limit: 6 });
    applyResumeAnalysis(id, res);
  } catch (error) {
    // 静默降级；mock 字段保留
  }
};

const prepareResumeEvaluation = async (id, options = {}) => {
  if (!id) return;
  updateResume(id, {
    evaluationLoading: true,
    evaluationStatus: "evaluating",
  });
  try {
    const res = await apiService.user.resumeArtifactAnalysisPrepare(id, {
      force: Boolean(options.force),
      limit: 6,
    });
    applyResumeAnalysis(id, res);
  } catch (error) {
    updateResume(id, {
      evaluationStatus: "failed",
      suggestions: [error?.message || "评估失败，请稍后重试"],
    });
  } finally {
    updateResume(id, {
      evaluationLoading: false,
      evaluationLoaded: true,
    });
  }
};

// 选中变化时拉取详情 + chunks（C5 增加 chunks 拉取）
watch(selectedId, (id) => {
  if (!id) return;
  loadResumeAnalysis(id);
  loadResumeChunks(id);
});

// C3: onMounted 黑 selector 外部点击监听 + 按顺序拉列表 → 选中默认简历 → 同步 URL
onMounted(async () => {
  document.addEventListener('click', handleClickOutsideSelector);
  await loadResumes();
  // 优先采用 ?artifact=:id query；未命中时 fallback 默认选中第一份简历
  const queryArtifact = String(route.query.artifact || '');
  const matched = queryArtifact && resumes.value.find((r) => r.id === queryArtifact);
  if (matched) {
    selectedId.value = queryArtifact;
  } else if (resumes.value.length > 0 && !selectedId.value) {
    selectedId.value = resumes.value[0].id;
    // 如果 URL 上没有 ?artifact=，同步上去，避免刷新后选中丢失
    if (selectedId.value !== queryArtifact) {
      router.replace({ query: { ...route.query, artifact: selectedId.value } });
    }
  }
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutsideSelector);
});

const getStatusLabel = (status) => {
  switch (status) {
    case "parsed":
      return "已解析";
    case "parsing":
      return "解析中";
    case "failed":
      return "解析失败";
    default:
      return "待处理";
  }
};

const getEvaluationStatusLabel = (status) => {
  switch (status) {
    case "ready":
      return "评估完成";
    case "stale":
      return "需刷新";
    case "evaluating":
      return "评估中";
    case "insufficient_data":
      return "资料不足";
    case "failed":
      return "评估失败";
    case "missing":
      return "待评估";
    default:
      return "待评估";
  }
};

const formatScore = (score) => {
  if (typeof score !== "number" || Number.isNaN(score)) return "—";
  return Math.round(score);
};
</script>

<style scoped>
/* ============ Layout ============ */
/* 主框架：响应式宽度，避免硬限 1440px。max-width 1680 防大屏拉伸过分；
   padding 用 clamp(20, 4vw, 56) 跟视口缩放，小屏上 20px、中屏按 4vw、大屏封顶 56px。 */
.wb-resume-content {
  width: 100%;
  max-width: 1680px;
  margin: 0 auto;
  padding: 0 clamp(20px, 4vw, 56px) 80px;
}

/* ============ 三栏 Shell（C2 commit）============ */
/* 详见 docs/requirements/2026-05-12-workbench-resume-redesign.md §6.1 整体网格。
   C2.1 修订：从固定 px + 多 breakpoint 跳变改为百分比 + minmax 响应式。
   三栏比例 22% / 1fr / 28%：中栏主导，左右栏对称感。
   minmax 保护：左栏 ≥ 200px（6 维进度条可读下限）、右栏 ≥ 260px（追问卡不振压下限）。
   gap 也用 clamp 随视口缩放（16 至 28）。超小屏（< 900px）才堆叠。 */
.wb-resume-shell {
  display: grid;
  grid-template-columns:
    minmax(200px, 22%)
    minmax(0, 1fr)
    minmax(260px, 28%);
  gap: clamp(16px, 1.6vw, 28px);
  align-items: start;
  margin-top: 8px;
}

/* C2.2 取消 sticky：原 sticky top:100px 让左右栏“贴顶”、中栏 flex center 让 dropzone “漂浮”，
   三栏完全错位。现在三栏都从顶端自然对齐。
   sticky 逻辑推迟到 C5（中栏 chunks 可滚动后不同检讨）。 */
.wb-resume-left,
.wb-resume-right {
  min-width: 0;
}

.wb-resume-mid {
  min-width: 0;
}

/* ============ 简历版本切换 dropdown selector（C3 commit）============ */
/* 位于左栏顶部，点击弹出 menu。trigger 与 .wb-resume-card 同渐变底 + gradient border-box，
   使顶部控件与下方评估卡视觉语言一致。 */
.wb-resume-selector {
  position: relative;
  margin-bottom: 12px;
}

.wb-resume-selector-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 12px 14px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.92) 0%, rgba(11, 12, 16, 0.92) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.025) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  color: var(--t);
  font: 600 13px var(--sans);
  cursor: pointer;
  transition: border-color .2s ease;
  text-align: left;
  isolation: isolate;
}

.wb-resume-selector-trigger:hover,
.wb-resume-selector-open .wb-resume-selector-trigger {
  border-color: rgba(220, 155, 90, 0.4);
}

.wb-resume-selector-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.wb-resume-selector-score {
  font: 600 12px var(--mono);
  color: rgba(220, 155, 90, 0.95);
  letter-spacing: .04em;
  flex-shrink: 0;
}

.wb-resume-selector-caret {
  font-size: 14px;
  color: var(--t3);
  transition: transform .2s ease;
  flex-shrink: 0;
}

.wb-resume-selector-open .wb-resume-selector-caret {
  transform: rotate(180deg);
}

/* menu 在 trigger 下方弹出。max-height + overflow-y 避免简历过多时 menu 溢出页面。 */
.wb-resume-selector-menu {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  right: 0;
  z-index: 10;
  list-style: none;
  margin: 0;
  padding: 6px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.96) 0%, rgba(11, 12, 16, 0.96) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.025) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.4);
  isolation: isolate;
  max-height: 320px;
  overflow-y: auto;
}

.wb-resume-selector-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 9px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background-color .15s ease, color .15s ease;
  font: 13px var(--sans);
  color: var(--t2);
}

.wb-resume-selector-item:hover {
  background: rgba(255, 255, 255, 0.04);
  color: var(--t);
}

.wb-resume-selector-item-active {
  background: rgba(220, 155, 90, 0.10);
  color: rgba(255, 224, 190, 0.98);
}

.wb-resume-selector-item-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.wb-resume-selector-item-score {
  font: 600 11px var(--mono);
  color: rgba(220, 155, 90, 0.85);
  letter-spacing: .04em;
  flex-shrink: 0;
}

/* 元数据条：简历描述。mono 体 + 字间距 .04em，与项目现有 mono meta 样式一致。 */
.wb-resume-meta-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
  margin: 0 0 14px;
  padding: 0 4px;
}

/* ============ 左栏评估卡（C4 commit）============ */
/* 代替原占位卡的真实评估区。使用与 .wb-resume-card 统一的卡片视觉语言，
   内部从上到下：OverallScore 圆环 + Level 徽章 → AI 总结 → 5 维评估 → 强项 → 风险 → 建议。 */
.wb-resume-eval-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 22px 20px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.85) 0%, rgba(11, 12, 16, 0.85) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.025) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  isolation: isolate;
}

/* OverallScore 圆环 + Level 徽章 */
.wb-overall-score-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.wb-overall-score {
  position: relative;
  width: 120px;
  height: 120px;
}

.wb-overall-svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.wb-overall-track {
  fill: none;
  stroke: rgba(255, 255, 255, 0.06);
  stroke-width: 6;
}

.wb-overall-fill {
  fill: none;
  stroke-width: 6;
  stroke-linecap: round;
  transition: stroke-dashoffset .5s ease;
}

.wb-overall-fill-strong,
.wb-overall-fill-high { stroke: rgba(220, 155, 90, 0.95); }
.wb-overall-fill-mid,
.wb-overall-fill-medium { stroke: rgba(230, 200, 130, 0.90); }
.wb-overall-fill-weak,
.wb-overall-fill-low { stroke: rgba(220, 110, 90, 0.85); }

.wb-overall-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1px;
}

.wb-overall-num {
  font: 700 32px/1 var(--display);
  color: var(--t);
  letter-spacing: -.02em;
}

.wb-overall-meta {
  font: 10px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

.wb-overall-badge {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 10px;
  font: 600 11px var(--sans);
  letter-spacing: .04em;
  border-radius: var(--radius-pill);
  border: 1px solid transparent;
}

.wb-overall-badge-strong,
.wb-overall-badge-high {
  color: rgba(255, 224, 190, 0.98);
  background: rgba(220, 155, 90, 0.14);
  border-color: rgba(220, 155, 90, 0.40);
}

.wb-overall-badge-mid,
.wb-overall-badge-medium {
  color: rgba(245, 225, 175, 0.95);
  background: rgba(230, 200, 130, 0.10);
  border-color: rgba(230, 200, 130, 0.32);
}

.wb-overall-badge-weak,
.wb-overall-badge-low {
  color: rgba(255, 195, 180, 0.95);
  background: rgba(220, 110, 90, 0.12);
  border-color: rgba(220, 110, 90, 0.36);
}

/* AI 总结 */
.wb-resume-summary {
  margin: 0;
  font: 13px/1.7 var(--sans);
  color: var(--t2);
  text-align: left;
}

/* 5 维评估。hover tooltip 显示评分依据。 */
.wb-dimensions-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wb-dim-row {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 6px;
  cursor: help;
}

.wb-dim-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 8px;
}

.wb-dim-label {
  font: 600 12px var(--sans);
  color: var(--t);
}

.wb-dim-score {
  font: 700 12px var(--mono);
  color: rgba(220, 155, 90, 0.95);
  letter-spacing: .04em;
}

.wb-dim-bar {
  height: 6px;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.06);
  overflow: hidden;
}

.wb-dim-bar span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, rgba(76, 214, 168, 0.85), rgba(220, 155, 90, 0.95));
  transition: width .5s ease;
}

/* hover tooltip（D-U9）。200ms transition delay 避免鼠标顺势跳动时快闪。 */
.wb-dim-tooltip {
  position: absolute;
  bottom: calc(100% + 8px);
  left: 50%;
  transform: translateX(-50%);
  width: max-content;
  max-width: 240px;
  padding: 8px 12px;
  background: rgba(0, 0, 0, 0.88);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: var(--radius-sm);
  color: var(--t);
  font: 12px/1.5 var(--sans);
  text-align: center;
  pointer-events: none;
  opacity: 0;
  visibility: hidden;
  transition: opacity .2s ease .12s, visibility .2s ease .12s;
  z-index: 5;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
}

.wb-dim-row:hover .wb-dim-tooltip,
.wb-dim-row:focus-within .wb-dim-tooltip {
  opacity: 1;
  visibility: visible;
}

/* 强项 / 风险 / 建议 三个叙述 section */
.wb-eval-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wb-eval-section-label {
  font: 600 10px var(--mono);
  color: var(--t3);
  letter-spacing: .08em;
  text-transform: uppercase;
}

.wb-eval-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.wb-eval-list li {
  font: 12px/1.6 var(--sans);
  color: var(--t2);
  padding: 7px 10px;
  background: rgba(255, 255, 255, 0.025);
  border-radius: var(--radius-sm);
}

.wb-eval-risks {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.wb-eval-risks li {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 7px 10px;
  background: rgba(255, 255, 255, 0.025);
  border-radius: var(--radius-sm);
  border-left: 3px solid rgba(255, 255, 255, 0.15);
}

/* ============ 右栏 追问 + CTA（C6 commit）============ */
/* 详见需求文档 §6.4。从上到下：方向匹配 chip + AI 追问问题列表 + 操作区（重新生成 + 看完整详情）
   + 降级声明 + 金色 CTA。右栏使用与左栏 .wb-resume-eval-card 同视觉语言。 */
.wb-resume-right-card {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 22px 20px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.85) 0%, rgba(11, 12, 16, 0.85) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.025) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  isolation: isolate;
}

.wb-section-label {
  font: 600 10px var(--mono);
  color: var(--t3);
  letter-spacing: .08em;
  text-transform: uppercase;
  margin: 0 0 10px;
}

/* 方向匹配 chip（FocusMatches[]） */
.wb-focus-matches {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wb-focus-chip {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.wb-focus-chip-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 8px;
}

.wb-focus-chip-label {
  font: 600 12px var(--sans);
  color: var(--t);
}

.wb-focus-chip-score {
  font: 700 11px var(--mono);
  color: rgba(220, 155, 90, 0.95);
  letter-spacing: .04em;
}

.wb-focus-chip-bar {
  height: 4px;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.06);
  overflow: hidden;
}

.wb-focus-chip-bar span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, rgba(76, 214, 168, 0.85), rgba(220, 155, 90, 0.95));
  transition: width .5s ease;
}

.wb-focus-chip-meta {
  margin: 0;
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
}

/* AI 追问问题列表。点击 chunk 联动（phase2-resume-chunk-question-link） */
.wb-questions-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wb-question-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 11px 13px;
  background: rgba(255, 255, 255, 0.025);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-md);
  cursor: pointer;
  text-align: left;
  font: inherit;
  color: inherit;
  transition: background-color .2s ease, border-color .2s ease;
}

.wb-question-card:hover {
  background: rgba(220, 155, 90, 0.08);
  border-color: rgba(220, 155, 90, 0.30);
}

.wb-question-head {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.wb-question-num {
  font: 600 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

.wb-question-chip {
  display: inline-flex;
  align-items: center;
  height: 18px;
  padding: 0 7px;
  font: 600 10px var(--sans);
  letter-spacing: .04em;
  border-radius: var(--radius-pill);
  border: 1px solid transparent;
}

.wb-question-chip-difficulty {
  background: rgba(76, 174, 230, 0.10);
  border-color: rgba(76, 174, 230, 0.30);
  color: rgba(165, 215, 250, 0.95);
}

.wb-question-chip-focus {
  background: rgba(220, 155, 90, 0.10);
  border-color: rgba(220, 155, 90, 0.30);
  color: rgba(255, 224, 190, 0.95);
}

.wb-question-title {
  margin: 0;
  font: 13px/1.55 var(--sans);
  color: var(--t);
}

/* 评估未生成追问时的 fallback */
.wb-question-fallback {
  padding: 12px 14px;
  background: rgba(255, 255, 255, 0.022);
  border: 1px dashed rgba(255, 255, 255, 0.10);
  border-radius: var(--radius-md);
}

.wb-question-fallback-text {
  margin: 0;
  font: 12px/1.55 var(--sans);
  color: var(--t3);
}

/* 操作区：重新生成 + 看完整详情 */
.wb-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wb-action-regenerate {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 9px 14px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.10);
  border-radius: var(--radius-sm);
  color: var(--t2);
  font: 600 12px var(--sans);
  cursor: pointer;
  transition: border-color .2s ease, color .2s ease, background-color .2s ease;
}

.wb-action-regenerate:hover:not(:disabled) {
  color: rgba(255, 224, 190, 0.95);
  border-color: rgba(220, 155, 90, 0.40);
  background: rgba(220, 155, 90, 0.08);
}

.wb-action-regenerate:disabled {
  cursor: wait;
  opacity: 0.6;
}

.wb-action-icon {
  font-size: 13px;
  line-height: 1;
}

.wb-action-meta {
  margin: 0;
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  flex-wrap: wrap;
}

.wb-action-detail {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font: 12px var(--sans);
  color: rgba(220, 155, 90, 0.95);
  text-decoration: none;
  padding: 6px 0;
  transition: color .2s ease;
}

.wb-action-detail:hover {
  color: rgba(255, 200, 140, 1);
  text-decoration: underline;
}

/* 降级声明（D-U10 / phase2-resume-cta-disclaimer） */
.wb-cta-disclaimer {
  margin: 0;
  font: 11px/1.55 var(--sans);
  color: var(--t3);
  text-align: center;
  opacity: 0.75;
  padding: 0 4px;
}

/* 金色 CTA（phase2-resume-cta-handler） */
.wb-cta-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  height: 56px;
  padding: 0 22px;
  background: linear-gradient(135deg, rgba(220, 155, 90, 0.95), rgba(200, 130, 65, 0.95));
  border: 1px solid rgba(220, 155, 90, 0.6);
  border-radius: var(--radius-md);
  color: rgba(20, 12, 6, 0.95);
  font: 700 15px var(--sans);
  letter-spacing: .01em;
  cursor: pointer;
  transition: transform .2s ease, box-shadow .2s ease, opacity .2s ease;
  box-shadow: 0 4px 16px rgba(220, 155, 90, 0.20);
}

.wb-cta-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 22px rgba(220, 155, 90, 0.32);
  opacity: 0.96;
}

.wb-cta-arrow {
  font-size: 16px;
  line-height: 1;
  transition: transform .2s ease;
}

.wb-cta-primary:hover .wb-cta-arrow {
  transform: translateX(2px);
}

/* 风险三色 left border（D-U9 + 需求文档 §6.2 第 7 项） */
.wb-risk-high { border-left-color: rgba(220, 100, 80, 0.85); }
.wb-risk-medium { border-left-color: rgba(230, 165, 100, 0.85); }
.wb-risk-low { border-left-color: rgba(220, 200, 130, 0.80); }

.wb-risk-label {
  font: 600 12px var(--sans);
  color: var(--t);
}

.wb-risk-suggest {
  font: 11px/1.55 var(--sans);
  color: var(--t3);
}

/* 占位样式：C2.2 卡片化。
   从 dashed border + 微弱 background 改为与项目现有 .wb-resume-card 一致的
   实色渐变 background + gradient border-box。这让占位看起来是“等待填充的
   真实卡片”而不是“空虚线框”，与中栏 dropzone 视觉语言统一。
   min-height 180 → 360，跟中栏 dropzone 高度接近，三栏顶端从同一行开始平齐展开。 */
.wb-resume-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding: 32px 22px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.85) 0%, rgba(11, 12, 16, 0.85) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.025) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  text-align: center;
  font: 13px/1.7 var(--sans);
  color: var(--t3);
  min-height: 360px;
  isolation: isolate;
}

.wb-resume-placeholder p {
  margin: 0;
}

.wb-resume-placeholder strong {
  color: rgba(220, 155, 90, 0.92);
  font-weight: 600;
}

.wb-resume-placeholder-icon {
  width: 36px;
  height: 36px;
  color: var(--t3);
  opacity: 0.5;
}

.wb-resume-placeholder-icon svg {
  width: 100%;
  height: 100%;
}

.wb-resume-placeholder-meta {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
  opacity: 0.55;
}

/* S0 未上传 中栏 dropzone 容器。
   C2.2：取消 flex center + min-height:480，不再让 dropzone 在中栏“垂直居中漂浮”。
   dropzone 现在直接从中栏顶端开始，与左右栏占位卡顶端平齐。 */
.wb-resume-mid-empty {
  width: 100%;
}

.wb-dropzone--centered {
  width: 100%;
  min-height: 360px;
}

/* S1/S2/S3 中栏占位。C2.2 卡片化，与左右栏占位同高。 */
.wb-resume-mid-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  min-height: 360px;
  padding: 32px 24px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.85) 0%, rgba(11, 12, 16, 0.85) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.025) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  text-align: center;
  font: 14px/1.7 var(--sans);
  color: var(--t3);
  isolation: isolate;
}

.wb-resume-mid-placeholder p {
  margin: 0;
}

.wb-resume-mid-placeholder strong {
  color: var(--t);
  font-weight: 600;
}

.wb-resume-mid-placeholder code {
  font: 11px var(--mono);
  padding: 2px 8px;
  background: rgba(220, 155, 90, 0.10);
  border: 1px solid rgba(220, 155, 90, 0.25);
  border-radius: 4px;
  color: rgba(220, 155, 90, 0.95);
  letter-spacing: .04em;
}

.wb-resume-hero {
  padding: 0 0 40px;
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

.wb-resume-title {
  font: 800 clamp(30px, 2.8vw, 42px) var(--display);
  color: var(--t);
  letter-spacing: -.02em;
  margin: 0 0 14px;
}

.wb-resume-sub {
  font-size: 15px;
  color: var(--t3);
  line-height: 1.7;
  margin: 0;
  max-width: 560px;
}

/* ============ Dropzone ============ */
.wb-upload {
  margin-bottom: 48px;
}

.wb-dropzone {
  position: relative;
  display: block;
  padding: 56px 32px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.7) 0%, rgba(11, 12, 16, 0.7) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.03) 100%) border-box;
  border: 1.5px dashed rgba(255, 255, 255, 0.14);
  border-radius: var(--radius-lg);
  cursor: pointer;
  text-align: center;
  transition: border-color .25s ease, background-color .25s ease;
  overflow: hidden;
}

.wb-dropzone:hover {
  border-color: rgba(220, 155, 90, 0.5);
}

.wb-dropzone-dragging {
  border-color: rgba(220, 155, 90, 0.85);
  background: rgba(220, 155, 90, 0.05);
}

.wb-dropzone-uploading {
  cursor: progress;
}

.wb-file-input {
  display: none;
}

.wb-dropzone-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.wb-dropzone-icon {
  width: 80px;
  height: 80px;
  color: var(--t3);
}

.wb-dropzone-icon svg {
  width: 100%;
  height: 100%;
}

.wb-dropzone-main {
  font: 600 18px var(--display);
  color: var(--t);
  letter-spacing: -.01em;
  margin-bottom: 6px;
}

.wb-dropzone-sub {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font: 13px var(--sans);
  color: var(--t3);
  flex-wrap: wrap;
  justify-content: center;
}

.wb-dropzone-action {
  color: rgba(220, 155, 90, 0.95);
  font-weight: 600;
}

.wb-dropzone-meta {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
  margin-left: 6px;
  padding-left: 8px;
  border-left: 1px solid rgba(255, 255, 255, 0.1);
}

.wb-upload-progress {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: rgba(255, 255, 255, 0.06);
}

.wb-upload-bar {
  height: 100%;
  background: linear-gradient(90deg, rgba(220, 155, 90, 0.6), rgba(220, 155, 90, 0.95));
  transition: width .3s ease;
}

.wb-upload-error {
  margin-top: 12px;
  font: 13px var(--sans);
  color: #ef6660;
}

/* ============ Resume Cards ============ */
.wb-resumes {
  margin-bottom: 32px;
}

.wb-block-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
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

.wb-block-close {
  background: none;
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--t3);
  width: 28px;
  height: 28px;
  border-radius: 50%;
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color .2s ease, border-color .2s ease;
}

.wb-block-close:hover {
  color: var(--t);
  border-color: rgba(255, 255, 255, 0.25);
}

.wb-resumes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 14px;
}

.wb-resume-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 18px 20px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 1) 0%, rgba(11, 12, 16, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.03) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: transform .25s ease, box-shadow .25s ease, border-color .25s ease;
  isolation: isolate;
}

.wb-resume-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.35);
}

.wb-resume-card-active {
  border-color: rgba(220, 155, 90, 0.5);
  background:
    linear-gradient(180deg, rgba(28, 22, 18, 1) 0%, rgba(18, 14, 11, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(220, 155, 90, 0.4) 0%, rgba(220, 155, 90, 0.1) 100%) border-box;
}

.wb-resume-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.wb-resume-card-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--t2);
}

.wb-resume-card-icon svg {
  width: 16px;
  height: 16px;
  display: block;
}

.wb-resume-card-tag {
  font: 10px var(--mono);
  color: rgba(220, 155, 90, 0.95);
  background: rgba(220, 155, 90, 0.1);
  border: 1px solid rgba(220, 155, 90, 0.3);
  border-radius: var(--radius-pill);
  padding: 2px 8px;
  letter-spacing: .04em;
}

.wb-resume-card-name {
  font: 600 14px var(--sans);
  color: var(--t);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-resume-card-meta {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
  display: flex;
  gap: 8px;
}

.wb-resume-card-stats {
  display: flex;
  gap: 16px;
  padding: 8px 0;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.wb-resume-stat {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.wb-resume-stat-num {
  font: 700 18px var(--mono);
  color: var(--t);
}

.wb-resume-stat-lb {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

.wb-resume-card-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.wb-resume-card-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

.wb-resume-card-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.wb-status-parsed {
  background: rgba(155, 209, 168, 0.85);
  box-shadow: 0 0 6px rgba(155, 209, 168, 0.45);
}

.wb-status-parsing {
  background: rgba(220, 155, 90, 0.85);
  animation: wb-edot 1.4s ease-in-out infinite;
}

.wb-status-failed {
  background: #ef6660;
}

.wb-resume-card-action {
  font: 12px var(--mono);
  color: var(--t2);
  letter-spacing: .04em;
}

.wb-resume-card-active .wb-resume-card-action {
  color: rgba(220, 155, 90, 0.95);
}

/* ============ 中栏 S2 chunks 列表（C5 commit）============ */
/* 详见需求文档 §6.3。按 chunk.index 升序渲染。每个 chunk 是卡片，
   头部 chunk #NN（C8 拓开页码后会加 "· 第 X 页"，见 phase2-resume-chunk-pagenum TODO），
   底部原文 13/1.7 sans 保护阅读节奏。卡片使用与 .wb-resume-card 同渐变底 + gradient border-box。 */
.wb-chunks-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.wb-chunk-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px 18px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.85) 0%, rgba(11, 12, 16, 0.85) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0.025) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  isolation: isolate;
  transition: border-color .2s ease;
}

.wb-chunk-card:hover {
  border-color: rgba(220, 155, 90, 0.25);
}

/* TODO(phase2-resume-chunk-question-link): C6 commit 会接入 "点右栏追问 → 中栏滚动到关联 chunk 高亮黄色".
   本 commit (C5) 仅加入 .wb-chunk-highlight 状态样式雏形，C6 才绑定 JS 联动逻辑。 */
.wb-chunk-card.wb-chunk-highlight {
  border-color: rgba(220, 155, 90, 0.55);
  box-shadow: 0 0 12px rgba(220, 155, 90, 0.25);
  animation: wb-chunk-flash 0.6s ease;
}

@keyframes wb-chunk-flash {
  0% { background-color: rgba(220, 155, 90, 0.18); }
  100% { background-color: transparent; }
}

.wb-chunk-head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wb-chunk-num {
  font: 600 11px var(--mono);
  color: rgba(220, 155, 90, 0.95);
  letter-spacing: .04em;
}

.wb-chunk-content {
  margin: 0;
  font: 13px/1.7 var(--sans);
  color: var(--t2);
  white-space: pre-wrap;
  word-break: break-word;
}

/* === Empty === */
.wb-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 40px 20px;
  gap: 8px;
}

.wb-empty-icon {
  font-size: 32px;
  opacity: .5;
}

.wb-empty-title {
  font: 600 15px var(--display);
  color: var(--t);
}

.wb-empty-sub {
  font-size: 13px;
  color: var(--t3);
}

/* ============ Detail ============ */
.wb-resume-detail {
  padding: 24px 26px;
  background:
    linear-gradient(180deg, rgba(16, 17, 22, 1) 0%, rgba(10, 11, 14, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.03) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  isolation: isolate;
}

.wb-detail-actions {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.wb-refresh-btn {
  height: 30px;
  padding: 0 12px;
  border-radius: var(--radius-sm);
  border: 1px solid rgba(220, 155, 90, 0.35);
  background: rgba(220, 155, 90, 0.10);
  color: rgba(255, 224, 190, 0.95);
  font: 600 12px var(--sans);
  cursor: pointer;
}

.wb-refresh-btn:disabled {
  cursor: wait;
  opacity: .62;
}

.wb-eval-overview {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr);
  gap: 22px;
  align-items: stretch;
  margin-top: 20px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.025);
  border-radius: var(--radius-sm);
}

.wb-eval-score {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 92px;
}

.wb-eval-score-num {
  font: 700 42px/1 var(--sans);
  color: var(--t);
  letter-spacing: 0;
}

.wb-eval-score-label {
  margin-top: 8px;
  font: 12px var(--mono);
  color: var(--t3);
}

.wb-eval-summary {
  min-width: 0;
}

.wb-eval-summary p {
  margin: 10px 0 0;
  color: var(--t2);
  font: 14px/1.7 var(--sans);
}

.wb-eval-status {
  display: inline-flex;
  align-items: center;
  height: 24px;
  padding: 0 9px;
  border-radius: var(--radius-pill);
  font: 600 12px var(--sans);
  color: rgba(255, 255, 255, 0.88);
  background: rgba(255, 255, 255, 0.07);
}

.wb-eval-ready {
  background: rgba(76, 214, 168, 0.12);
  color: rgba(155, 242, 213, 0.95);
}

.wb-eval-stale,
.wb-eval-missing,
.wb-eval-evaluating {
  background: rgba(255, 215, 112, 0.12);
  color: rgba(255, 230, 160, 0.95);
}

.wb-eval-failed,
.wb-eval-insufficient_data {
  background: rgba(255, 120, 120, 0.12);
  color: rgba(255, 185, 185, 0.95);
}

.wb-dimensions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 18px;
}

.wb-dimension {
  min-width: 0;
  padding: 14px;
  background: rgba(255, 255, 255, 0.025);
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: var(--radius-sm);
}

.wb-dimension-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font: 600 13px var(--sans);
  color: var(--t);
}

.wb-dimension-head strong {
  font: 700 13px var(--mono);
  color: rgba(220, 155, 90, 0.95);
}

.wb-dimension-bar {
  height: 5px;
  margin-top: 10px;
  overflow: hidden;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.08);
}

.wb-dimension-bar span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, rgba(76, 214, 168, 0.9), rgba(220, 155, 90, 0.95));
}

.wb-dimension p {
  margin: 9px 0 0;
  color: var(--t3);
  font: 12px/1.6 var(--sans);
}

.wb-detail-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.3fr);
  gap: 32px;
  margin-top: 20px;
}

.wb-detail-col {
  min-width: 0;
}

.wb-detail-label {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .06em;
  text-transform: uppercase;
  margin-bottom: 12px;
}

.wb-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.wb-tag {
  font: 12px var(--sans);
  color: var(--t2);
  padding: 4px 10px;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.wb-projects {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wb-project {
  padding: 10px 14px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-sm);
}

.wb-project-name {
  font: 600 13px var(--sans);
  color: var(--t);
  margin-bottom: 2px;
}

.wb-project-stack {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
}

.wb-eval-lists {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
  margin-top: 24px;
}

.wb-eval-list ul {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.wb-eval-list li {
  color: var(--t2);
  font: 13px/1.6 var(--sans);
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.022);
}

@media (max-width: 1024px) {
  .wb-detail-grid {
    grid-template-columns: 1fr;
    gap: 24px;
  }

  .wb-dimensions,
  .wb-eval-lists {
    grid-template-columns: 1fr;
  }
}

/* ============ 响应式 fallback（D-R5修订）============ */
/* C2.1：取消 1279/1023 两个 breakpoint 跳变，minmax + clamp 已代替连续响应。
   仅保留超小屏堆叠逻辑：当三栏总小宽（200 + 260 + 中栏最少 ~280） > viewport 时堆叠。 */
@media (max-width: 899px) {
  .wb-resume-shell {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  .wb-resume-placeholder,
  .wb-resume-mid-placeholder {
    min-height: 200px;
    padding: 24px;
  }
  .wb-dropzone--centered {
    min-height: 280px;
  }
}

@media (max-width: 768px) {
  .wb-dropzone {
    padding: 36px 20px;
  }
}
</style>
