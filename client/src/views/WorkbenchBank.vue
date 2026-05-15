<template>
  <!--
    WorkbenchBank：题库浏览（对应设计图 3）。
    布局：左侧 filter sidebar (260px) + 右侧主区（toolbar + 题目 grid 卡片）。
    底部抽屉：选中题目后展开题目详情（移动端走全屏 modal）。
    数据契约：优先读取后端结构化题库；静态 JSON 和 plan preview 仅作为降级来源。
  -->
  <WorkbenchLayout>
    <div class="wb-bank-content">
      <section class="wb-bank-hero">
        <div class="wb-eyebrow">
          <span class="wb-eyebrow-dot" aria-hidden="true"></span>
          <span>题库浏览</span>
        </div>
        <p class="wb-bank-sub">{{ questionBankStatsLabel }}</p>
      </section>

      <div class="wb-bank-shell">
        <!-- 左侧 filter sidebar -->
        <aside class="wb-bank-filter">
          <div class="wb-filter-block">
            <h4 class="wb-filter-title">方向</h4>
            <label
              v-for="dir in directions"
              :key="dir.key"
              class="wb-filter-item"
            >
              <input
                type="checkbox"
                :value="dir.key"
                :checked="filters.directions.includes(dir.key)"
                @change="toggleFilter('directions', dir.key)"
              />
              <span class="wb-filter-check" aria-hidden="true">
                <svg viewBox="0 0 16 16" fill="none">
                  <polyline points="3,8 7,12 13,4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </span>
              <span class="wb-filter-label">{{ dir.label }}</span>
              <span class="wb-filter-count">{{ countByDirection(dir.key) }}</span>
            </label>
          </div>

          <div class="wb-filter-block">
            <h4 class="wb-filter-title">难度</h4>
            <label
              v-for="diff in difficulties"
              :key="diff.key"
              class="wb-filter-item"
            >
              <input
                type="checkbox"
                :value="diff.key"
                :checked="filters.difficulties.includes(diff.key)"
                @change="toggleFilter('difficulties', diff.key)"
              />
              <span class="wb-filter-check" aria-hidden="true">
                <svg viewBox="0 0 16 16" fill="none">
                  <polyline points="3,8 7,12 13,4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </span>
              <span class="wb-filter-label">{{ diff.label }}</span>
            </label>
          </div>

          <div class="wb-filter-block">
            <h4 class="wb-filter-title">能力</h4>
            <label
              v-for="cap in capabilities"
              :key="cap.key"
              class="wb-filter-item"
            >
              <input
                type="checkbox"
                :value="cap.key"
                :checked="filters.capabilities.includes(cap.key)"
                @change="toggleFilter('capabilities', cap.key)"
              />
              <span class="wb-filter-check" aria-hidden="true">
                <svg viewBox="0 0 16 16" fill="none">
                  <polyline points="3,8 7,12 13,4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </span>
              <span class="wb-filter-label">{{ cap.label }}</span>
            </label>
          </div>

          <button
            v-if="hasFilters"
            type="button"
            class="wb-filter-reset"
            @click="resetFilters"
          >清空筛选</button>
        </aside>

        <!-- 右侧主区 -->
        <section class="wb-bank-main">
          <!-- toolbar -->
          <div class="wb-bank-toolbar">
            <div class="wb-search">
              <svg class="wb-search-icon" viewBox="0 0 16 16" fill="none" aria-hidden="true">
                <circle cx="7" cy="7" r="5" stroke="currentColor" stroke-width="1.5" />
                <line x1="11" y1="11" x2="14" y2="14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
              </svg>
              <input
                v-model="searchQuery"
                type="search"
                class="wb-search-input"
                placeholder="搜索题目、关键词、技术栈..."
              />
            </div>
            <div class="wb-bank-sort">
              <span class="wb-sort-label">排序</span>
              <button
                v-for="sort in sortOptions"
                :key="sort.key"
                type="button"
                class="wb-sort-btn"
                :class="{ 'wb-sort-active': activeSort === sort.key }"
                @click="activeSort = sort.key"
              >{{ sort.label }}</button>
            </div>
          </div>

          <!-- 题目列表 -->
          <div v-if="filteredQuestions.length > 0" class="wb-bank-list">
            <article
              v-for="q in visibleQuestions"
              :key="q.id"
              class="wb-q"
              :class="{ 'wb-q-active': selectedId === q.id }"
              @click="selectedId = selectedId === q.id ? '' : q.id"
            >
              <div class="wb-q-head">
                <span class="wb-q-tag" :class="`wb-tag-dir-${q.direction}`">{{ getDirectionLabel(q.direction) }}</span>
                <span class="wb-q-diff" :class="`wb-diff-${q.difficulty}`">{{ getDifficultyLabel(q.difficulty) }}</span>
                <span class="wb-q-hot">
                  <span class="wb-q-hot-lb" aria-hidden="true">热度</span>
                  <span class="wb-q-hot-num">{{ q.hot }}</span>
                </span>
              </div>
              <h3 class="wb-q-title">{{ q.title }}</h3>
              <p class="wb-q-desc">{{ q.summary }}</p>
              <div class="wb-q-foot">
                <div class="wb-q-tags">
                  <span v-for="tag in q.tags" :key="tag" class="wb-q-tag-chip">{{ tag }}</span>
                </div>
                <button
                  type="button"
                  class="wb-q-action"
                  @click.stop="practiceQuestion(q)"
                >用此题练 →</button>
              </div>

              <!-- 展开详情 -->
              <div v-if="selectedId === q.id" class="wb-q-detail" @click.stop>
                <div class="wb-q-detail-block">
                  <div class="wb-detail-label">题目细节</div>
                  <p class="wb-q-detail-text">{{ q.detail }}</p>
                </div>
                <div class="wb-q-detail-block">
                  <div class="wb-detail-label">考察点</div>
                  <ul class="wb-q-points">
                    <li v-for="point in q.points" :key="point">{{ point }}</li>
                  </ul>
                </div>
                <div class="wb-q-detail-block">
                  <div class="wb-detail-label">AI 追问示例</div>
                  <ol class="wb-q-followups">
                    <li v-for="fu in q.followups" :key="fu">{{ fu }}</li>
                  </ol>
                </div>
              </div>
            </article>
          </div>

          <div v-else class="wb-empty">
            <div class="wb-empty-title">没有匹配的题目</div>
            <div class="wb-empty-sub">尝试清空筛选或换个关键词。</div>
            <button v-if="hasFilters" type="button" class="wb-empty-cta" @click="resetFilters">清空筛选</button>
          </div>
        </section>
      </div>
    </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
import { apiService } from "../composables/useApi";

const router = useRouter();

// 后端 difficulty.level (int) 与本地 key (string) 双向映射
const DIFFICULTY_KEY_BY_LEVEL = { 1: "intro", 2: "junior", 3: "mid", 4: "senior", 5: "expert" };
const DIFFICULTY_LEVEL_BY_KEY = { intro: 1, junior: 2, mid: 3, senior: 4, expert: 5 };
const QUESTION_BANK_ASSET_URL = "/data/interview-question-bank.json";
const VISIBLE_QUESTION_LIMIT = 240;

// === Filter 选项（本地 fallback；onMounted 接入 interviewPresets 后覆盖） ===
const directions = ref([
  { key: "go_backend", label: "Go 后端" },
  { key: "java_backend", label: "Java 后端" },
  { key: "frontend_vue", label: "前端 Vue" },
  { key: "system_design", label: "系统设计" },
  { key: "algorithm", label: "算法基础" },
]);

const difficulties = ref([
  { key: "intro", label: "入门" },
  { key: "junior", label: "初级" },
  { key: "mid", label: "中级" },
  { key: "senior", label: "资深" },
  { key: "expert", label: "专家" },
]);

const capabilities = ref([
  { key: "concurrency", label: "并发与调度" },
  { key: "database", label: "数据库" },
  { key: "system_design", label: "系统设计" },
  { key: "engineering", label: "工程实践" },
  { key: "network", label: "网络协议" },
  { key: "performance", label: "性能优化" },
  { key: "algorithm", label: "算法基础" },
  { key: "communication", label: "表达沟通" },
  { key: "frontend_arch", label: "前端架构" },
  { key: "observability", label: "可观测性" },
]);

const filters = ref({
  directions: [],
  difficulties: [],
  capabilities: [],
});

const hasFilters = computed(() => {
  return (
    filters.value.directions.length > 0 ||
    filters.value.difficulties.length > 0 ||
    filters.value.capabilities.length > 0 ||
    searchQuery.value.trim().length > 0
  );
});

const toggleFilter = (group, key) => {
  const arr = filters.value[group];
  const i = arr.indexOf(key);
  if (i >= 0) {
    arr.splice(i, 1);
  } else {
    arr.push(key);
  }
};

const resetFilters = () => {
  filters.value.directions = [];
  filters.value.difficulties = [];
  filters.value.capabilities = [];
  searchQuery.value = "";
};

// === 搜索 + 排序 ===
const searchQuery = ref("");
const activeSort = ref("hot");

const sortOptions = [
  { key: "hot", label: "热度" },
  { key: "new", label: "最新" },
  { key: "diff", label: "难度" },
];

// === 题库数据（本地 fallback；优先由 interviewQuestions / 静态资产 / plan preview 覆盖） ===
const questions = ref([
  {
    id: "q-001",
    title: "Go map 并发访问会怎么样？如何安全使用？",
    direction: "go_backend",
    difficulty: "mid",
    capability: "concurrency",
    hot: 1280,
    summary: "考察 Go runtime 的 map 并发检测机制和 sync.Map / RWMutex 使用边界。",
    detail: "Go 1.6 后内置 map 并发写检测，会 fatal 而非 panic。本题要求展开 sync.Map 与 mutex+map 的性能权衡场景。",
    points: ["map 并发 fatal 的运行时检测", "sync.Map 的读多写少优化", "RWMutex vs Mutex 的吞吐对比", "shard map 的常见实现"],
    followups: ["sync.Map 的 amended 字段是干什么的？", "1024 个 goroutine 同时写 map 的最差表现？", "你设计的 LRU 缓存如何兼顾并发？"],
    tags: ["map", "并发", "sync.Map"],
  },
  {
    id: "q-002",
    title: "Vue 3 的响应式原理与 Vue 2 有什么本质区别？",
    direction: "frontend_vue",
    difficulty: "mid",
    capability: "frontend_arch",
    hot: 956,
    summary: "Proxy 替代 Object.defineProperty 后，深度监听、新增属性、数组索引都不再是问题。",
    detail: "本题要求从 reactive、ref、effect 三个层次分别说清，并指出 Proxy 的局限（部分内置对象不能代理）。",
    points: ["Proxy 拦截 vs defineProperty 改写", "track / trigger 的依赖收集", "ref 与 reactive 的差异", "scheduler 与异步更新批处理"],
    followups: ["effect 是如何实现 lazy 的？", "computed 与 watchEffect 的内部差异？", "shallowRef 在什么场景必须使用？"],
    tags: ["Vue 3", "响应式", "Proxy"],
  },
  {
    id: "q-003",
    title: "如何设计一个分布式 ID 生成器？",
    direction: "system_design",
    difficulty: "senior",
    capability: "system_design",
    hot: 1834,
    summary: "Snowflake、号段、Redis incr 三种主流方案各有取舍。本题考察设计权衡能力。",
    detail: "需要从趋势递增、全局唯一、可扩展、容错四个维度展开，并能说出每种方案的具体边界场景。",
    points: ["Snowflake 时钟回拨问题", "号段模式的双 Buffer 优化", "Redis incr 的高可用设计", "ID 长度与索引性能的权衡"],
    followups: ["机器 ID 用完了怎么办？", "你怎么保证号段服务挂了不影响业务？", "MySQL 的 auto_increment 在分库分表后还能用吗？"],
    tags: ["分布式", "ID", "Snowflake"],
  },
  {
    id: "q-004",
    title: "Linux 进程调度算法 CFS 是如何工作的？",
    direction: "algorithm",
    difficulty: "senior",
    capability: "algorithm",
    hot: 624,
    summary: "完全公平调度器通过 vruntime 红黑树维护可运行进程，每次取最左节点。",
    detail: "本题考察对内核调度的理解，重点是 vruntime 计算公式、weight、period 的关系，以及与 O(1) 调度器的对比。",
    points: ["vruntime = delta_exec * weight_0 / weight_p", "红黑树的最左节点选择", "nice 值与权重的映射", "调度延迟与吞吐的权衡"],
    followups: ["进程睡眠后醒来 vruntime 会变吗？", "CPU 密集型 vs IO 密集型在 CFS 下的表现差异？", "为什么 idle 进程的 vruntime 是最大值？"],
    tags: ["Linux", "调度", "CFS"],
  },
  {
    id: "q-005",
    title: "讲讲你做过的最有挑战的项目，遇到的最大瓶颈和如何突破？",
    direction: "system_design",
    difficulty: "mid",
    capability: "communication",
    hot: 2156,
    summary: "STAR 法则下展开：项目背景、个人角色、关键决策、可量化结果。AI 会针对每个环节深度追问。",
    detail: "本题是行为面试经典题，AI 不接受空泛回答；会针对你说的每一个技术决策追问『为什么不是 X』和『你怎么衡量这个选择是对的』。",
    points: ["项目的真实业务背景", "你个人对哪些决策有 ownership", "决策时考虑了哪些备选方案", "结果如何量化与归因"],
    followups: ["为什么不用 X 方案？", "如果重来一次会怎么改？", "这个数据是如何统计出来的？", "团队对这个决策有不同意见吗？"],
    tags: ["项目深度", "STAR", "行为面试"],
  },
  {
    id: "q-006",
    title: "Redis 持久化 RDB 与 AOF 的取舍",
    direction: "go_backend",
    difficulty: "mid",
    capability: "database",
    hot: 894,
    summary: "RDB 快照恢复快但有数据丢失窗口，AOF 命令日志数据安全但文件大且恢复慢。",
    detail: "考察对 Redis 持久化机制的理解，能否说清 fork 时的 COW、AOF 重写时机、混合持久化的工作机制。",
    points: ["bgsave 的 fork 与 COW", "AOF appendfsync 三种策略", "AOF 重写时为什么不阻塞主线程", "Redis 4.0 混合持久化"],
    followups: ["fork 在大内存实例下的代价？", "AOF 重写期间又有写命令怎么办？", "RDB 文件损坏时还能用 AOF 恢复吗？"],
    tags: ["Redis", "持久化", "AOF"],
  },
]);

const usingQuestionBankAsset = ref(false);
const usingBackendQuestionBank = ref(false);
const questionBankMeta = ref({
  total: questions.value.length,
  generatedAt: "",
  source: "fallback",
});

const questionBankStatsLabel = computed(() => {
  const sourceLabel = usingBackendQuestionBank.value
    ? "数据库题库"
    : usingQuestionBankAsset.value
      ? "离线题库资产"
      : "预览题库";
  const total = questionBankMeta.value.total || questions.value.length;
  return `${filteredQuestions.value.length} / ${total} 道题，来源：${sourceLabel}，按方向、难度、能力筛选，看 AI 如何追问。`;
});

// === 过滤 + 排序 ===
const filteredQuestions = computed(() => {
  let list = questions.value;

  // 方向
  if (filters.value.directions.length > 0) {
    list = list.filter((q) => filters.value.directions.includes(q.direction));
  }
  // 难度
  if (filters.value.difficulties.length > 0) {
    list = list.filter((q) => filters.value.difficulties.includes(q.difficulty));
  }
  // 能力
  if (filters.value.capabilities.length > 0) {
    list = list.filter((q) => filters.value.capabilities.includes(q.capability));
  }
  // 搜索
  const q = searchQuery.value.trim().toLowerCase();
  if (q) {
    list = list.filter(
      (item) =>
        item.title.toLowerCase().includes(q) ||
        item.summary.toLowerCase().includes(q) ||
        (item.detail || "").toLowerCase().includes(q) ||
        item.tags.some((t) => t.toLowerCase().includes(q))
    );
  }
  // 排序
  list = [...list];
  if (activeSort.value === "hot") {
    list.sort((a, b) => b.hot - a.hot);
  } else if (activeSort.value === "new") {
    list.sort((a, b) => (b.sequence || 0) - (a.sequence || 0));
  } else if (activeSort.value === "diff") {
    const diffOrder = { intro: 1, junior: 2, mid: 3, senior: 4, expert: 5 };
    list.sort((a, b) => diffOrder[a.difficulty] - diffOrder[b.difficulty]);
  }
  return list;
});

const visibleQuestions = computed(() => filteredQuestions.value.slice(0, VISIBLE_QUESTION_LIMIT));

const countByDirection = (dirKey) => {
  return questions.value.filter((q) => q.direction === dirKey).length;
};

// === 选中题目 ===
const selectedId = ref("");

const getDirectionLabel = (key) => directions.value.find((d) => d.key === key)?.label || key;
const getDifficultyLabel = (key) => difficulties.value.find((d) => d.key === key)?.label || key;

const practiceQuestion = (q) => {
  router.push({
    path: "/chat",
    query: {
      direction: q.direction,
      difficulty: q.difficulty,
      questionId: q.questionKey || q.id,
      from: "workbench-bank",
    },
  });
};

// === 远程预设 + 题目拉取 ===

// 后端 InterviewPlanQuestion → 本地题目卡数据。
const toQuestionCard = (q, i, total = 0) => {
  const difficultyLevel = Number(q.difficultyLevel) || 3;
  const diffKey = DIFFICULTY_KEY_BY_LEVEL[difficultyLevel] || "mid";
  const summary = (q.prompt || "").length > 80
    ? `${(q.prompt || "").slice(0, 80)}…`
    : (q.prompt || "");
  // 静态题库没有真实热度，按批次、难度和顺序生成稳定估值，避免每次渲染跳动。
  const sequence = Number(q.sequence) || i + 1;
  const batchWeight = Number(q.batchSequence || 1) * 80;
  const usageCount = Number(q.usageCount) || 0;
  const hot = usageCount > 0
    ? usageCount
    : 600 + batchWeight + difficultyLevel * 45 + Math.max(0, (total || 0) - i);
  const rawTags = Array.isArray(q.tags) ? q.tags : [];
  return {
    id: q.key || q.questionKey || String(q.id || `q-${i}`),
    numericId: q.id || 0,
    questionKey: q.key || q.questionKey || "",
    title: q.title || "未命名题目",
    direction: q.directionKey || "",
    difficulty: diffKey,
    capability: q.focusKey || "",
    hot: hot > 0 ? hot : 100,
    sequence,
    batch: q.batch || "",
    summary,
    detail: q.prompt || "",
    points: Array.isArray(q.expectedSignals) ? q.expectedSignals : [],
    followups: Array.isArray(q.followUps) ? q.followUps : [],
    tags: [...new Set([q.focusLabel, q.difficultyLabel, q.batchLabel, ...rawTags].filter(Boolean))],
  };
};

const loadBackendQuestionBank = async () => {
  try {
    const res = await apiService.user.interviewQuestions({ limit: 2000, sort: "hot" });
    const list = Array.isArray(res?.questions) ? res.questions : [];
    if (list.length === 0) {
      return false;
    }
    questions.value = list.map((item, index) => toQuestionCard(item, index, Number(res.total) || list.length));
    questionBankMeta.value = {
      total: Number(res.total) || list.length,
      generatedAt: res?.questionMeta?.schemaVersion || "",
      source: "database",
    };
    usingBackendQuestionBank.value = true;
    usingQuestionBankAsset.value = false;
    return true;
  } catch (error) {
    usingBackendQuestionBank.value = false;
    return false;
  }
};

const loadQuestionBankAsset = async () => {
  try {
    const response = await fetch(QUESTION_BANK_ASSET_URL, { cache: "no-cache" });
    if (!response.ok) {
      return false;
    }
    const data = await response.json();
    const list = Array.isArray(data?.questions) ? data.questions : [];
    if (list.length === 0) {
      return false;
    }

    questions.value = list.map((item, index) => toQuestionCard(item, index, list.length));
    questionBankMeta.value = {
      total: Number(data.total) || list.length,
      generatedAt: data.generatedAt || "",
      source: "asset",
    };
    usingQuestionBankAsset.value = true;
    usingBackendQuestionBank.value = false;
    return true;
  } catch (error) {
    return false;
  }
};

// 拉预设（方向 / 难度 / focus）。错误静默降级到本地 fallback。
const loadPresets = async () => {
  try {
    const res = await apiService.user.interviewPresets();
    if (!res) return;
    if (Array.isArray(res.directions) && res.directions.length > 0) {
      directions.value = res.directions.map((d) => ({ key: d.key, label: d.label }));
    }
    if (Array.isArray(res.difficulties) && res.difficulties.length > 0) {
      difficulties.value = res.difficulties.map((dif) => ({
        key: DIFFICULTY_KEY_BY_LEVEL[dif.level] || "mid",
        label: dif.label,
      }));
    }
    if (Array.isArray(res.focusOptions) && res.focusOptions.length > 0) {
      capabilities.value = res.focusOptions.map((f) => ({ key: f.key, label: f.label }));
    }
  } catch (error) {
    // 静默降级
  }
};

// 拉题目。静态资产不可用时，允许传入选中的方向 / 难度 / focusKeys 去后端过滤；
// 本地 questions 列表仍会按 filters 再过一道筛选，以保证多选场景可用。
const loadQuestions = async (params = {}) => {
  if (usingBackendQuestionBank.value || usingQuestionBankAsset.value) {
    return;
  }
  try {
    const res = await apiService.user.interviewPlanPreview({ limit: 50, ...params });
    const list = Array.isArray(res?.questions) ? res.questions : [];
    if (list.length === 0) return; // 保留本地 fallback
    questions.value = list.map((item, index) => toQuestionCard(item, index, list.length));
    questionBankMeta.value = {
      total: list.length,
      generatedAt: "",
      source: "api-preview",
    };
    usingBackendQuestionBank.value = false;
    usingQuestionBankAsset.value = false;
  } catch (error) {
    // 静默降级
  }
};

// 筛选变化时重拉：后端接受单个 directionKey / difficulty / focusKeys (csv)。
// 多选场景下拉首项后，本地 filteredQuestions 会再按全部选中项过滤。
const rebuildPreviewParams = () => {
  const params = {};
  if (filters.value.directions.length > 0) {
    params.directionKey = filters.value.directions[0];
  }
  if (filters.value.difficulties.length > 0) {
    params.difficulty = DIFFICULTY_LEVEL_BY_KEY[filters.value.difficulties[0]] || 0;
  }
  if (filters.value.capabilities.length > 0) {
    params.focusKeys = filters.value.capabilities.join(",");
  }
  return params;
};

watch(
  () => [
    [...filters.value.directions],
    [...filters.value.difficulties],
    [...filters.value.capabilities],
  ],
  () => {
    if (!usingBackendQuestionBank.value && !usingQuestionBankAsset.value) {
      loadQuestions(rebuildPreviewParams());
    }
  },
  { deep: true },
);

onMounted(async () => {
  await loadPresets();
  const backendLoaded = await loadBackendQuestionBank();
  if (backendLoaded) {
    return;
  }
  const assetLoaded = await loadQuestionBankAsset();
  if (!assetLoaded) {
    loadQuestions();
  }
});
</script>

<style scoped>
.wb-bank-content {
  width: 100%;
  max-width: min(1440px, 100%);
  margin: 0 auto;
  min-height: calc(100svh - 80px);
  display: flex;
  flex-direction: column;
  padding: 0 clamp(20px, 4vw, 56px) 80px;
}

.wb-bank-hero {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px 18px;
  padding: 0 0 clamp(20px, 2.5vw, 32px);
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
  letter-spacing: .04em;
  background: rgba(255, 255, 255, 0.025);
  backdrop-filter: blur(8px);
  width: fit-content;
  flex-shrink: 0;
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

.wb-bank-sub {
  font-size: var(--fs-lg);
  color: var(--t2);
  line-height: 1.65;
  margin: 0;
  flex: 1 1 560px;
}

/* ============ Shell：双列布局 ============ */
.wb-bank-shell {
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(220px, 22%) minmax(0, 1fr);
  gap: clamp(16px, 1.6vw, 28px);
  align-items: start;
  margin-top: 8px;
}

/* ============ 左侧 filter ============ */
.wb-bank-filter {
  position: sticky;
  top: 100px; /* 80 header + 20 留白 */
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 28px;
  padding: 22px 20px;
  background:
    linear-gradient(180deg, rgba(16, 17, 22, 1) 0%, rgba(10, 11, 14, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.03) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  isolation: isolate;
}

.wb-filter-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.wb-filter-title {
  font: 600 var(--fs-xs) var(--mono);
  color: var(--t3);
  letter-spacing: .08em;
  text-transform: uppercase;
  margin: 0 0 4px;
}

.wb-filter-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 8px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background-color .2s ease;
}

.wb-filter-item:hover {
  background: rgba(255, 255, 255, 0.03);
}

.wb-filter-item input[type="checkbox"] {
  position: absolute;
  width: 1px;
  height: 1px;
  opacity: 0;
  pointer-events: none;
}

.wb-filter-check {
  width: 16px;
  height: 16px;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: rgba(255, 255, 255, 0.04);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: transparent;
  transition: background-color .2s ease, border-color .2s ease, color .2s ease;
}

.wb-filter-check svg {
  width: 12px;
  height: 12px;
}

.wb-filter-item input:checked + .wb-filter-check {
  background: rgba(220, 155, 90, 0.95);
  border-color: rgba(220, 155, 90, 0.95);
  color: var(--bg);
}

.wb-filter-label {
  flex: 1;
  font: var(--fs-sm) var(--sans);
  color: var(--t2);
  min-width: 0;
}

.wb-filter-item input:checked ~ .wb-filter-label {
  color: var(--t);
}

.wb-filter-count {
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
  flex-shrink: 0;
}

.wb-filter-reset {
  margin-top: 4px;
  font: var(--fs-xs) var(--sans);
  color: rgba(220, 155, 90, 0.95);
  background: none;
  border: 1px solid rgba(220, 155, 90, 0.3);
  border-radius: var(--radius-sm);
  padding: 8px 12px;
  cursor: pointer;
  transition: background-color .2s ease, border-color .2s ease;
}

.wb-filter-reset:hover {
  background: rgba(220, 155, 90, 0.08);
  border-color: rgba(220, 155, 90, 0.5);
}

/* ============ 主区 ============ */
.wb-bank-main {
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* === Toolbar === */
.wb-bank-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.wb-search {
  position: relative;
  flex: 1;
  min-width: 240px;
  max-width: 480px;
}

.wb-search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: var(--t3);
  pointer-events: none;
}

.wb-search-input {
  width: 100%;
  padding: 10px 14px 10px 40px;
  font: var(--fs-md) var(--sans);
  color: var(--t);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-md);
  outline: none;
  transition: border-color .2s ease, background-color .2s ease;
}

.wb-search-input::placeholder {
  color: var(--t3);
}

.wb-search-input:focus {
  border-color: rgba(220, 155, 90, 0.5);
  background: rgba(255, 255, 255, 0.05);
}

.wb-bank-sort {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.wb-sort-label {
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .06em;
  text-transform: uppercase;
}

.wb-sort-btn {
  font: var(--fs-sm) var(--sans);
  color: var(--t3);
  background: none;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-sm);
  padding: 6px 12px;
  cursor: pointer;
  transition: color .2s ease, background-color .2s ease, border-color .2s ease;
}

.wb-sort-btn:hover {
  color: var(--t);
}

.wb-sort-active {
  color: rgba(220, 155, 90, 0.95);
  background: rgba(220, 155, 90, 0.06);
  border-color: rgba(220, 155, 90, 0.3);
}

/* === 题目卡片列表 === */
.wb-bank-list {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wb-q {
  padding: 20px 22px;
  background:
    linear-gradient(180deg, rgba(16, 17, 22, 1) 0%, rgba(10, 11, 14, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0.02) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: border-color .25s ease, transform .25s ease, box-shadow .25s ease;
  isolation: isolate;
}

.wb-q:hover {
  transform: translateY(-2px);
  border-color: rgba(255, 255, 255, 0.14);
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.3);
}

.wb-q-active {
  border-color: rgba(220, 155, 90, 0.4);
  background:
    linear-gradient(180deg, rgba(22, 18, 14, 1) 0%, rgba(14, 11, 8, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(220, 155, 90, 0.4) 0%, rgba(220, 155, 90, 0.08) 100%) border-box;
}

.wb-q-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.wb-q-tag {
  font: var(--fs-2xs) var(--mono);
  padding: 3px 9px;
  border-radius: var(--radius-pill);
  letter-spacing: .04em;
}

.wb-tag-dir-go-backend,
.wb-tag-dir-go_backend {
  color: #4cd6a8;
  background: rgba(76, 214, 168, 0.08);
  border: 1px solid rgba(76, 214, 168, 0.25);
}

.wb-tag-dir-frontend,
.wb-tag-dir-frontend_vue {
  color: #6eb6ff;
  background: rgba(110, 182, 255, 0.08);
  border: 1px solid rgba(110, 182, 255, 0.25);
}

.wb-tag-dir-java_backend {
  color: #f0c46f;
  background: rgba(240, 196, 111, 0.08);
  border: 1px solid rgba(240, 196, 111, 0.25);
}

.wb-tag-dir-fullstack {
  color: #b599ff;
  background: rgba(181, 153, 255, 0.08);
  border: 1px solid rgba(181, 153, 255, 0.25);
}

.wb-tag-dir-devops {
  color: #ffd770;
  background: rgba(255, 215, 112, 0.08);
  border: 1px solid rgba(255, 215, 112, 0.25);
}

.wb-tag-dir-data,
.wb-tag-dir-algorithm {
  color: #ff9966;
  background: rgba(255, 153, 102, 0.08);
  border: 1px solid rgba(255, 153, 102, 0.25);
}

.wb-tag-dir-system,
.wb-tag-dir-system_design {
  color: var(--t2);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.wb-q-diff {
  font: var(--fs-2xs) var(--mono);
  padding: 3px 9px;
  border-radius: var(--radius-pill);
  letter-spacing: .04em;
}

.wb-diff-junior {
  color: #9bd1a8;
  background: rgba(155, 209, 168, 0.08);
  border: 1px solid rgba(155, 209, 168, 0.18);
}

.wb-diff-mid {
  color: var(--t2);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.wb-diff-senior {
  color: rgba(220, 155, 90, 0.95);
  background: rgba(220, 155, 90, 0.08);
  border: 1px solid rgba(220, 155, 90, 0.25);
}

.wb-diff-expert {
  color: #ef8a73;
  background: rgba(239, 138, 115, 0.08);
  border: 1px solid rgba(239, 138, 115, 0.25);
}

.wb-q-hot {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
  margin-left: auto;
}

.wb-q-hot-lb {
  text-transform: uppercase;
  letter-spacing: .08em;
  opacity: .7;
}

.wb-q-hot-num {
  color: var(--t2);
  font-weight: 600;
}

.wb-q-title {
  font: 700 var(--fs-xl) var(--display);
  color: var(--t);
  margin: 0 0 8px;
  letter-spacing: 0;
  line-height: 1.4;
}

.wb-q-desc {
  font-size: var(--fs-sm);
  color: var(--t3);
  line-height: 1.65;
  margin: 0 0 14px;
}

.wb-q-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.wb-q-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.wb-q-tag-chip {
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-sm);
  padding: 2px 8px;
  letter-spacing: .03em;
}

.wb-q-action {
  font: 600 var(--fs-xs) var(--sans);
  color: var(--t);
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-sm);
  padding: 6px 12px;
  cursor: pointer;
  transition: background-color .2s ease, border-color .2s ease, color .2s ease;
  white-space: nowrap;
}

.wb-q-action:hover {
  background: rgba(220, 155, 90, 0.1);
  border-color: rgba(220, 155, 90, 0.4);
  color: rgba(220, 155, 90, 0.95);
}

/* === 展开详情 === */
.wb-q-detail {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  flex-direction: column;
  gap: 18px;
  cursor: default;
}

.wb-q-detail-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wb-detail-label {
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .06em;
  text-transform: uppercase;
}

.wb-q-detail-text {
  font-size: var(--fs-sm);
  color: var(--t2);
  line-height: 1.7;
  margin: 0;
}

.wb-q-points,
.wb-q-followups {
  margin: 0;
  padding-left: 20px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.wb-q-points li,
.wb-q-followups li {
  font: var(--fs-sm) var(--sans);
  color: var(--t2);
  line-height: 1.7;
}

.wb-q-followups li {
  color: rgba(220, 155, 90, 0.85);
}

/* === Empty === */
.wb-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 60px 20px;
  gap: 10px;
}

.wb-empty-icon {
  font-size: 36px;
  opacity: .5;
  margin-bottom: 4px;
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

/* === 响应式 === */
@media (min-width: 900px) {
  .wb-bank-content {
    height: calc(100svh - 80px);
    min-height: 640px;
    overflow: hidden;
    padding-bottom: clamp(28px, 4vw, 48px);
  }

  .wb-bank-hero {
    flex: 0 0 auto;
    padding: clamp(20px, 3vw, 28px) 0 clamp(18px, 2.4vw, 28px);
  }

  .wb-bank-shell {
    flex: 1 1 auto;
    align-items: stretch;
    overflow: hidden;
  }

  .wb-bank-filter,
  .wb-bank-main {
    min-height: 0;
    overflow: hidden;
  }

  .wb-bank-filter {
    position: static;
    top: auto;
    overflow-y: auto;
    overscroll-behavior: contain;
    scrollbar-width: thin;
    scrollbar-color: rgba(255, 255, 255, 0.16) transparent;
    scrollbar-gutter: stable;
  }

  .wb-bank-toolbar {
    flex: 0 0 auto;
  }

  .wb-bank-list,
  .wb-empty {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
    overscroll-behavior: contain;
    scrollbar-width: thin;
    scrollbar-color: rgba(255, 255, 255, 0.16) transparent;
    scrollbar-gutter: stable;
  }

  .wb-empty {
    justify-content: center;
  }

  .wb-bank-filter::-webkit-scrollbar,
  .wb-bank-list::-webkit-scrollbar,
  .wb-empty::-webkit-scrollbar {
    width: 6px;
  }

  .wb-bank-filter::-webkit-scrollbar-track,
  .wb-bank-list::-webkit-scrollbar-track,
  .wb-empty::-webkit-scrollbar-track {
    background: transparent;
  }

  .wb-bank-filter::-webkit-scrollbar-thumb,
  .wb-bank-list::-webkit-scrollbar-thumb,
  .wb-empty::-webkit-scrollbar-thumb {
    border-radius: var(--radius-pill);
    background: rgba(255, 255, 255, 0.16);
  }

  .wb-bank-filter::-webkit-scrollbar-thumb:hover,
  .wb-bank-list::-webkit-scrollbar-thumb:hover,
  .wb-empty::-webkit-scrollbar-thumb:hover {
    background: rgba(220, 155, 90, 0.32);
  }
}

@media (max-width: 899px) {
  .wb-bank-shell {
    grid-template-columns: 1fr;
  }

  .wb-bank-filter {
    position: static;
  }
}

@media (max-width: 768px) {
  .wb-bank-content {
    padding: 0 20px 60px;
  }
  .wb-bank-toolbar {
    flex-direction: column;
    align-items: stretch;
  }
  .wb-search {
    max-width: none;
  }
}
</style>
