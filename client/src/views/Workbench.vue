<template>
  <!--
    Workbench 主页：登录后的工作总览。
    布局：Hero greeting + 3 个统计 → 4 张快捷入口卡 → 最近面试表 + 能力雷达
    视觉锚点：沿用 Home 的 aurora / display 字体 / demo-win 哑光黑卡 / 暖琥珀 accent；
            克制使用动画，仅 hover 提供反馈，避免主页 hero 的"剧院级"扫光。
  -->
  <WorkbenchLayout>
    <div class="wb-content">
      <!-- ============ Hero（v2 多态：左 60% 数据 + 右 40% 状态卡） ============ -->
      <!--
        左侧职责：时间锚 (eyebrow) + 个性化标题 + 副标题 + 4 metric 数据条
        右侧职责：根据用户状态（S1 首次 / S2 进行中 / S3 未复盘 / S4 推荐 / S0 兜底）
                 渲染对应卡片，恒定 4 层信息密度，避免空旷感
        视觉规范：沿用 --panel + --b 哑光黑卡 token、display/sans/mono 三字体、--radius-lg 圆角
                 暖琥珀 (rgba(220,155,90,0.9)) 仅在 S1/S2 主 CTA 与 chevron 上出现
       -->
      <section class="wb-hero" aria-label="工作台总览">
        <!-- 左栏：数据陈述 -->
        <div class="wb-hero-left">
          <div class="wb-eyebrow">
            <span class="wb-eyebrow-dot" aria-hidden="true"></span>
            <span>{{ heroEyebrow }}</span>
          </div>
          <h1 class="wb-title">
            <span class="wb-title-greet">{{ heroGreeting }}</span>
            <span class="wb-title-name">{{ displayName }}</span>
          </h1>
          <p class="wb-sub">{{ heroSubtitle }}</p>

          <!-- 4 metric 数据条：替代旧 3 联统计 -->
          <div class="wb-metrics" role="group" aria-label="本周练习数据">
            <div class="wb-metric" v-for="(m, i) in metricsRow" :key="m.key">
              <div class="wb-metric-num">{{ m.value }}</div>
              <div class="wb-metric-lb">{{ m.label }}</div>
              <div class="wb-metric-sep" v-if="i < metricsRow.length - 1" aria-hidden="true"></div>
            </div>
          </div>
        </div>

        <!-- 右栏：多态状态卡 -->
        <aside class="wb-hero-right" :data-state="heroState.kind" aria-label="下一步行动">
          <!-- S1 首次用户：3 步引导 + amber 主 CTA -->
          <div v-if="heroState.kind === 'S1'" class="wb-card wb-card-onboard">
            <div class="wb-card-eyebrow">开始你的旅程</div>
            <ol class="wb-onboard-steps">
              <li v-for="(s, idx) in onboardSteps" :key="s.key" class="wb-onboard-step" :class="{ 'is-done': s.done }">
                <span class="wb-onboard-num">{{ idx + 1 }}</span>
                <span class="wb-onboard-label">{{ s.label }}</span>
              </li>
            </ol>
            <div class="wb-card-foot">
              <router-link to="/workbench/new" class="wb-card-cta wb-card-cta-amber">
                开始第一场
                <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M5 3l5 5-5 5" />
                </svg>
              </router-link>
            </div>
          </div>

          <!-- S2 有进行中会话：continue 卡 + amber 主 CTA -->
          <div v-else-if="heroState.kind === 'S2'" class="wb-card wb-card-continue">
            <div class="wb-card-eyebrow">继续上次面试</div>
            <div class="wb-continue-progress" role="progressbar" :aria-valuenow="heroState.progress" aria-valuemin="0" aria-valuemax="100">
              <div class="wb-continue-bar" :style="{ width: heroState.progress + '%' }"></div>
            </div>
            <div class="wb-continue-meta">
              <span class="wb-continue-pct">{{ heroState.progress }}%</span>
              <span class="wb-continue-dim">已完成</span>
            </div>
            <div class="wb-continue-topic" :title="heroState.title">{{ heroState.title }}</div>
            <div class="wb-continue-tags">
              <span class="wb-tag">{{ heroState.directionLabel }}</span>
              <span class="wb-tag" v-if="heroState.difficultyLabel">{{ heroState.difficultyLabel }}</span>
            </div>
            <div class="wb-card-foot">
              <router-link :to="heroState.link" class="wb-card-cta wb-card-cta-amber">
                继续
                <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M5 3l5 5-5 5" />
                </svg>
              </router-link>
            </div>
          </div>

          <!-- S3 上次未复盘：score + tags + 文字链接 -->
          <div v-else-if="heroState.kind === 'S3'" class="wb-card wb-card-review">
            <div class="wb-card-eyebrow">查看上次复盘</div>
            <div class="wb-review-score">
              <span class="wb-review-num">{{ heroState.score }}</span>
              <span class="wb-review-max">/ {{ heroState.maxScore }}</span>
            </div>
            <div class="wb-review-title" :title="heroState.title">{{ heroState.title }}</div>
            <div class="wb-review-tags" v-if="heroState.tags.length">
              <span v-for="t in heroState.tags" :key="t.key" class="wb-tag" :class="`wb-tag-${t.level}`">{{ t.label }}</span>
            </div>
            <div class="wb-card-foot">
              <router-link :to="heroState.link" class="wb-card-cta wb-card-cta-text">
                {{ heroState.ctaLabel }}
                <svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M5 3l5 5-5 5" />
                </svg>
              </router-link>
            </div>
          </div>

          <!-- S4 常规回访：推荐项 + 描述 + 文字链接 -->
          <div v-else-if="heroState.kind === 'S4'" class="wb-card wb-card-next">
            <div class="wb-card-eyebrow">下一步建议</div>
            <div class="wb-next-title">{{ heroState.title }}</div>
            <p class="wb-next-desc">{{ heroState.description }}</p>
            <div class="wb-card-foot">
              <router-link :to="heroState.link" class="wb-card-cta wb-card-cta-text">
                {{ heroState.ctaLabel }}
                <svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M5 3l5 5-5 5" />
                </svg>
              </router-link>
            </div>
          </div>

          <!-- S0 兜底：未登录 / bootstrap 失败 -->
          <div v-else class="wb-card wb-card-fallback">
            <div class="wb-card-eyebrow">登录后查看完整工作台</div>
            <div class="wb-fallback-title">实时数据暂未就绪</div>
            <p class="wb-fallback-desc">登录或刷新后，这里会显示你最近的练习状态与下一步建议。</p>
            <div class="wb-card-foot">
              <router-link to="/workbench/new" class="wb-card-cta wb-card-cta-text">
                直接新建一场
                <svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M5 3l5 5-5 5" />
                </svg>
              </router-link>
            </div>
          </div>
        </aside>
      </section>

      <!-- ============ 4 业务领域状态卡 ============ -->
      <!--
        信息架构定位（详见 docs/requirements/2026-05-10-workbench-information-architecture.md）：
        - Hero 多态卡 = 时间维度（"我此刻该做什么"，已隐含面试领域陈述）
        - 4 业务卡 = 空间维度（"4 个非面试支撑业务的当前进度"）
        - 删除「继续上次」「新建面试」两旧卡（与 Hero S2 / 顶 nav CTA 重复）
        - 新增「报告中心」「知识库」两缺口卡（暴露后端已就绪但前端无入口的子产品）
        - 「简历」「题库」保留并升级为业务陈述卡，主数字 amber gradient text 作视觉锚
      -->
      <section class="wb-quick" aria-label="业务领域状态">
        <!-- ## 卡 1：报告中心（暴露 ReportCenter 子产品入口） ## -->
        <article class="wb-qcard wb-qcard-report">
          <div class="wb-qcard-head">
            <span class="wb-qcard-icon" aria-hidden="true">
              <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round">
                <path d="M2.5 13.5h11" stroke-linecap="round" />
                <rect x="3.5" y="8" width="2" height="5" />
                <rect x="7" y="5" width="2" height="8" />
                <rect x="10.5" y="9.5" width="2" height="3.5" />
                <path d="M3.5 4.5l3.5-2.5 3 2 3-3" stroke-linecap="round" />
              </svg>
            </span>
            <span class="wb-qcard-tag" :class="{ 'wb-tag-amber': hasReports }">{{ reportsTag }}</span>
          </div>
          <h3 class="wb-qcard-title">报告中心</h3>
          <p class="wb-qcard-desc" v-if="hasReports">
            <span class="wb-qcard-num">{{ stats.completed }}</span> 份报告 · 平均 {{ stats.avgScore }} 分
          </p>
          <p class="wb-qcard-desc" v-else>完成首场面试后，这里会显示能力分析与复盘建议。</p>
          <div class="wb-qcard-spacer"></div>
          <div class="wb-qcard-foot">
            <span class="wb-qcard-meta">{{ reportsMeta }}</span>
            <router-link :to="reportsLink" class="wb-qcard-link">{{ reportsCta }} →</router-link>
          </div>
        </article>

        <!-- ## 卡 2：简历库（消费 bootstrapData.resumeSummary） ## -->
        <article class="wb-qcard wb-qcard-resume">
          <div class="wb-qcard-head">
            <span class="wb-qcard-icon" aria-hidden="true">
              <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round">
                <path d="M3.5 1.5h6.5L13 4.5v10H3.5z" />
                <path d="M10 1.5V4.5h3" />
                <line x1="5.5" y1="8" x2="10.5" y2="8" stroke-linecap="round" />
                <line x1="5.5" y1="10.5" x2="10.5" y2="10.5" stroke-linecap="round" />
              </svg>
            </span>
            <span class="wb-qcard-tag">{{ resumeStatus }}</span>
          </div>
          <h3 class="wb-qcard-title">简历库</h3>
          <p class="wb-qcard-desc" v-if="hasResume">
            <span class="wb-qcard-num">{{ resumeTotal }}</span> 份
            <template v-if="resumeProjectsCount > 0"> · {{ resumeProjectsCount }} 个项目</template>
            <template v-else-if="resumeChunkCount > 0"> · {{ resumeChunkCount }} 片段</template>
            · 已分析
          </p>
          <p class="wb-qcard-desc" v-else>上传简历后，AI 会基于项目经历做深度追问。</p>
          <div class="wb-qcard-spacer"></div>
          <div class="wb-qcard-foot">
            <span class="wb-qcard-meta">{{ resumeMeta }}</span>
            <router-link to="/workbench/resume" class="wb-qcard-link">{{ resumeCta }} →</router-link>
          </div>
        </article>

        <!-- ## 卡 3：知识库（消费 bootstrapData.knowledgeSummary） ## -->
        <article class="wb-qcard wb-qcard-knowledge">
          <div class="wb-qcard-head">
            <span class="wb-qcard-icon" aria-hidden="true">
              <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round">
                <path d="M3 2.5h6.5L13 5.5v8H3z" />
                <path d="M3 11.5h10" />
                <path d="M5 5.5h3M5 8h5" stroke-linecap="round" />
              </svg>
            </span>
            <span class="wb-qcard-tag">{{ knowledgeTag }}</span>
          </div>
          <h3 class="wb-qcard-title">知识库</h3>
          <p class="wb-qcard-desc" v-if="hasKnowledge">
            <span class="wb-qcard-num">{{ knowledgeDocuments }}</span> 篇 · {{ knowledgeChunks }} 块
          </p>
          <p class="wb-qcard-desc" v-else>上传文档后，AI 会在面试时引用你提供的资料。</p>
          <div class="wb-qcard-spacer"></div>
          <div class="wb-qcard-foot">
            <span class="wb-qcard-meta">{{ knowledgeMeta }}</span>
            <router-link to="/workbench/knowledge" class="wb-qcard-link">{{ knowledgeCta }} →</router-link>
          </div>
        </article>

        <!-- ## 卡 4：题库（消费 interviewPresets.directions[].questionCount 累加） ## -->
        <article class="wb-qcard wb-qcard-bank">
          <div class="wb-qcard-head">
            <span class="wb-qcard-icon" aria-hidden="true">
              <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round">
                <rect x="2" y="3" width="2.5" height="11" rx="0.4" />
                <rect x="5.5" y="2" width="3" height="12" rx="0.4" />
                <path d="M9.6 4.2l3.4 .8v9l-3.4-.8z" />
              </svg>
            </span>
            <span class="wb-qcard-tag">题库</span>
          </div>
          <h3 class="wb-qcard-title">题库浏览</h3>
          <p class="wb-qcard-desc">
            <span class="wb-qcard-num">{{ bankTotalQuestions }}</span> 题 · {{ bankDirectionsCount }} 个方向
          </p>
          <div class="wb-qcard-spacer"></div>
          <div class="wb-qcard-foot">
            <span class="wb-qcard-meta">按方向 / 难度筛选</span>
            <router-link to="/workbench/bank" class="wb-qcard-link">浏览 →</router-link>
          </div>
        </article>
      </section>

      <!-- ============ 下方两栏：最近面试 + 能力雷达 ============ -->
      <section class="wb-bottom">
        <!-- 左栏：最近面试表 -->
        <div class="wb-card wb-sessions">
          <header class="wb-block-head">
            <h3 class="wb-block-title">最近面试</h3>
            <router-link to="/workbench/new" class="wb-more">查看全部 →</router-link>
          </header>

          <div v-if="recentSessions.length > 0" class="wb-table" role="table">
            <div class="wb-tr wb-thead" role="row">
              <div class="wb-th wb-col-title">主题</div>
              <div class="wb-th">方向</div>
              <div class="wb-th">难度</div>
              <div class="wb-th wb-col-num">分</div>
              <div class="wb-th wb-col-num">时长</div>
              <div class="wb-th">时间</div>
            </div>

            <router-link
              v-for="row in recentSessions"
              :key="row.id"
              :to="row.link"
              class="wb-tr wb-trow"
              role="row"
            >
              <div class="wb-td wb-col-title" :title="row.title">{{ row.title }}</div>
              <div class="wb-td">{{ row.direction }}</div>
              <div class="wb-td">
                <span class="wb-diff" :class="`wb-diff-${row.difficultyLevel}`">{{ row.difficulty }}</span>
              </div>
              <div class="wb-td wb-col-num">
                <span class="wb-score" :class="getScoreClass(row.score)">
                  <span class="wb-score-dot" aria-hidden="true"></span>{{ row.score }}
                </span>
              </div>
              <div class="wb-td wb-col-num">{{ row.duration }}m</div>
              <div class="wb-td wb-col-time">{{ row.time }}</div>
            </router-link>
          </div>

          <div v-else class="wb-empty">
            <div class="wb-empty-title">还没有面试记录</div>
            <div class="wb-empty-sub">完成第一场面试后，结果会出现在这里。</div>
            <router-link to="/workbench/new" class="wb-empty-cta">+ 新建面试</router-link>
          </div>
        </div>

        <!-- 右栏：能力雷达（有数据态 SVG / 0 数据态骨架） -->
        <div class="wb-card wb-radar">
          <header class="wb-block-head">
            <h3 class="wb-block-title">能力雷达</h3>
            <span class="wb-radar-meta">最近 5 场</span>
          </header>

          <!-- v-if hasRadarData：后端 abilityRadar 至少 1 维 score>0 才渲染雷达 SVG，
               避免「刚进页雷达多边形崩缩为中心点 + 78/90/72/65/82 假评分」的假现象。 -->
          <template v-if="hasRadarData">
            <!-- viewBox 220×220 + preserveAspectRatio 保证 SVG 在容器内等比缩放，
                 不会因 aspect-ratio 容器崩塌而变形。 -->
            <svg class="wb-radar-svg" viewBox="0 0 220 220" preserveAspectRatio="xMidYMid meet" role="img" aria-label="个人能力雷达图">
              <!-- 5 层正五边形网格 -->
              <g class="wb-radar-grid">
                <polygon
                  v-for="(scale, i) in [1, 0.8, 0.6, 0.4, 0.2]"
                  :key="`grid-${i}`"
                  :points="getPolygonPoints(scale)"
                  class="wb-radar-grid-line"
                />
              </g>

              <!-- 5 条径向线 -->
              <line
                v-for="(angle, i) in radarAngles"
                :key="`line-${i}`"
                x1="110"
                y1="110"
                :x2="110 + radarRadius * Math.cos(angle)"
                :y2="110 + radarRadius * Math.sin(angle)"
                class="wb-radar-line"
              />

              <!-- 用户能力多边形 -->
              <polygon :points="userPolygonPoints" class="wb-radar-user" />

              <!-- 顶点圆点 -->
              <circle
                v-for="(pt, i) in userPoints"
                :key="`pt-${i}`"
                :cx="pt.x"
                :cy="pt.y"
                r="3.5"
                class="wb-radar-dot"
              />

              <!-- 维度标签 -->
              <text
                v-for="(dim, i) in radarDims"
                :key="`label-${i}`"
                :x="110 + (radarRadius + 18) * Math.cos(radarAngles[i])"
                :y="110 + (radarRadius + 18) * Math.sin(radarAngles[i])"
                text-anchor="middle"
                dominant-baseline="middle"
                class="wb-radar-label"
              >{{ dim.label }}</text>
            </svg>

            <div class="wb-radar-weaks">
              <div class="wb-weaks-title">建议加强</div>
              <div class="wb-weaks-list">
                <span v-for="weak in weakSpots" :key="weak" class="wb-weak">{{ weak }}</span>
              </div>
            </div>
          </template>

          <!-- 0 数据态：雷达骨架 + 引导文案（完成首场后填满准确评分） -->
          <div v-else class="wb-radar-empty">
            <svg class="wb-radar-empty-svg" viewBox="0 0 220 220" preserveAspectRatio="xMidYMid meet" aria-hidden="true">
              <!-- 静态骨架五边形，5 层 grid + 5 条径线，用 muted 色调 -->
              <g class="wb-radar-empty-grid">
                <polygon v-for="(scale, i) in [1, 0.8, 0.6, 0.4, 0.2]" :key="`empty-grid-${i}`" :points="getPolygonPoints(scale)" />
              </g>
              <line
                v-for="(angle, i) in radarAngles"
                :key="`empty-line-${i}`"
                x1="110"
                y1="110"
                :x2="110 + radarRadius * Math.cos(angle)"
                :y2="110 + radarRadius * Math.sin(angle)"
                class="wb-radar-empty-line"
              />
              <!-- 5 个维度标签仍保留，用 muted 色 -->
              <text
                v-for="(dim, i) in radarDims"
                :key="`empty-label-${i}`"
                :x="110 + (radarRadius + 18) * Math.cos(radarAngles[i])"
                :y="110 + (radarRadius + 18) * Math.sin(radarAngles[i])"
                text-anchor="middle"
                dominant-baseline="middle"
                class="wb-radar-empty-label"
              >{{ dim.label }}</text>
            </svg>
            <div class="wb-radar-empty-text">
              <div class="wb-radar-empty-title">完成首场面试</div>
              <div class="wb-radar-empty-sub">5 个维度评分会填满雷达。</div>
            </div>
          </div>
        </div>
      </section>
    </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
import { apiService } from "../composables/useApi";
import { useAuth } from "../composables/useAuth";

const { username: storedUsername } = useAuth();

// === 用户名 / 显示名 ===
// 登录态 username 来自 useAuth；profile 成功时用后端最新值覆盖展示。
const profileUsername = ref("");
const displayName = computed(() => profileUsername.value || storedUsername.value || "同学");

// === 统计数据：默认 0/—占位，onMounted 从 bootstrap.stats 与 sessions 聚合后覆盖 ===
// 原则 5（1 sourcing 4 cards principles）：后端没字段的地方走空态。
// 删除了原 bankCount=256 hardcode mock，题库数现从 interviewPresets.directions[].questionCount 派生。
const stats = ref({
  completed: 0,
  avgScore: "—",
  lastAt: "暂无",
});

// === bootstrap 原始响应快照：用于 heroState computed 派生 4 状态分支 ===
// 设计上保留指针/数组/嵌套对象的原貌，避免 stats 扁平化丢失字段。
const bootstrapData = ref(null);

// === 4 业务领域状态卡所需数据源 ===
// 旧的 resumeDirection/Difficulty/Progress/Link 已删除（功能被 Hero S2 完全覆盖）
// 旧的 resumeName/Status/ProjectsHint 已重构为基于 resumeSummary 派生的 computed

// 简历摘要：从 bootstrapData.resumeSummary 派生，0 数据态在 computed 处理
// projectsCount 字段由后端 0e22383 commit 交付（feat(workbench): 补充简历项目数摘要），
// 表示 AI 分析出的项目数量（比 chunkCount 在用户语义上更可读）。
const resumeSummary = ref({ total: 0, latestTitle: "", chunkCount: 0, projectsCount: 0, latestUpdatedAt: "" });

// 知识库摘要：bootstrapData.knowledgeSummary 派生（之前完全没消费）
const knowledgeSummary = ref({ documents: 0, chunks: 0, latestTitle: "", latestAddedAt: "" });

// 面试 presets（题库卡数据源）：从 /api/users/interview/presets 拉，directions[].questionCount 累加
// 后端入库 1614 题后 questionCount 自动从 DB 派生，前端代码无需改
const interviewPresets = ref({ directions: [], difficulties: [], focusOptions: [] });

// === 最近面试列表：从 /users/sessions 与 bootstrap.recentSessions 拉 ===
// 原则 5：接口失败/返回空列表时不造假数据，模板 wb-empty 分支会呈现「还没有面试记录 + 新建面试」。
// 删除了原 mockSessions 5 条假数据（m1-m5），避免 401/网络失败时误导用户「已有记录」。
const recentSessions = ref([]);

// === 能力雷达 ===
// 5 个能力维度，覆盖技术深度 / 表达 / 沟通：
// 项目深度（projects）/ 语言基础（lang）/ 算法（algo）/ 架构（arch）/ 表达（comm）
// 默认 5 维骨架 value=0，applyAbilityRadar 从 bootstrap.abilityRadar 拿真实评分覆盖。
// 原则 5：不再 hardcode 78/90/72/65/82 假评分，0 数据态走雷达空态分支。
const radarDims = ref([
  { key: "project", label: "项目深度", value: 0, maxScore: 100 },
  { key: "lang", label: "语言基础", value: 0, maxScore: 100 },
  { key: "algo", label: "算法", value: 0, maxScore: 100 },
  { key: "arch", label: "架构", value: 0, maxScore: 100 },
  { key: "comm", label: "表达", value: 0, maxScore: 100 },
]);

// 雷达是否有真实数据：至少 1 维 score > 0 才算有。模板用这个 computed 在 SVG 雷达与「完成首场」空态之间切。
const hasRadarData = computed(() => radarDims.value.some((d) => (d.value || 0) > 0));

const radarRadius = 90;

// 5 个角度从顶部 (-90°) 起顺时针均布。
// SVG y 轴向下，但极坐标 sin 在第 1/2 象限是负，刚好让 -π/2 = 顶部。
const radarAngles = computed(() =>
  radarDims.value.map((_, i) => -Math.PI / 2 + (i * 2 * Math.PI) / radarDims.value.length)
);

const userPoints = computed(() =>
  radarDims.value.map((dim, i) => {
    const r = (dim.value / (dim.maxScore || 100)) * radarRadius;
    return {
      x: 110 + r * Math.cos(radarAngles.value[i]),
      y: 110 + r * Math.sin(radarAngles.value[i]),
    };
  })
);

const userPolygonPoints = computed(() =>
  userPoints.value.map((p) => `${p.x.toFixed(2)},${p.y.toFixed(2)}`).join(" ")
);

const getPolygonPoints = (scale) => {
  return radarDims.value
    .map((_, i) => {
      const r = scale * radarRadius;
      const angle = radarAngles.value[i];
      const x = 110 + r * Math.cos(angle);
      const y = 110 + r * Math.sin(angle);
      return `${x.toFixed(2)},${y.toFixed(2)}`;
    })
    .join(" ");
};

// 弱项标签：取 radarDims 中 value < 75 的维度
const weakSpots = computed(() => {
  const weaks = radarDims.value
    .filter((d) => d.value < 75)
    .map((d) => d.label + (d.value < 70 ? " 加强" : ""));
  // 静态补充几个具体话题点，让弱项不仅是维度名，更具行动指向。
  if (weaks.length === 0) {
    return ["分布式事务", "Raft 协议", "Linux 调度"];
  }
  return [...weaks, "分布式事务", "Raft 协议"].slice(0, 4);
});

// 分数色彩分级：≥85 暖琥珀（高分）/ 75-84 中性白 / <75 冷蓝（待加强）
// 用于最近面试表 .wb-score 的 class 映射。
const getScoreClass = (score) => {
  if (score >= 85) return "wb-score-high";
  if (score >= 75) return "wb-score-mid";
  return "wb-score-low";
};

// === Hero 多态卡：派生字段（依赖 stats / bootstrapData / radarDims） ===

// 当前 ISO 周次（W23 这种锚点比"实时同步"更具时间感）
const weekISOLabel = computed(() => {
  const d = new Date();
  // ISO 周计算：从今年第一个周四起算
  const jan4 = new Date(d.getFullYear(), 0, 4);
  const dayOfYear = Math.floor((d - jan4) / 86400000);
  const week = Math.max(1, Math.ceil((dayOfYear + jan4.getDay() + 1) / 7));
  return `本周 · W${week}`;
});

const heroEyebrow = computed(() => weekISOLabel.value);

// 问候语：根据时间段切换"早 / 午 / 晚"
const heroGreeting = computed(() => {
  const h = new Date().getHours();
  if (h < 6) return "夜深了，";
  if (h < 11) return "早，";
  if (h < 14) return "午安，";
  if (h < 18) return "下午好，";
  return "晚上好，";
});

// 副标题：根据状态分支给出不同短语，避免空洞文案
const heroSubtitle = computed(() => {
  const k = heroState.value.kind;
  if (k === "S1") return "三步完成第一场练习，建立你的能力基线。";
  if (k === "S2") return "上次的对话还在等你，回到原节奏继续。";
  if (k === "S3") return "刚结束的练习评估已就绪，看看哪些维度可以再打磨。";
  if (k === "S4") return "根据本周表现，这是接下来最值得练的方向。";
  return "登录后这里会显示你的最近练习与下一步建议。";
});

// 4 metric 数据条：替代旧 3 联统计，所有字段都有 0/— 兜底
// 数据来源：stats（完成场次、平均分）+ recentSessions 聚合 + bootstrapData 周锚（如有）
const metricsRow = computed(() => {
  // 总练习时长：当前后端 SessionItem 不返回 duration 字段，求和恒为 0 → durLabel = "—"。
  // 待 SessionItem 补 duration（或 bootstrap.stats.totalDurationMinutes）后接入；
  // 不在前端伪造每场 30m 假时长（原 mock 兜底已在 toRecentSessionRow 移除）。
  const totalMin = recentSessions.value.reduce((acc, s) => acc + (s.duration || 0), 0);
  const durLabel =
    totalMin <= 0
      ? "—"
      : totalMin < 60
        ? `${totalMin}m`
        : `${Math.floor(totalMin / 60)}h${totalMin % 60 ? totalMin % 60 + "m" : ""}`;

  // 连续周数：当前没有后端字段，先静态展示"本周"占位，后端补 stats.weeklyStreak 后接入
  const streakLabel = bootstrapData.value?.stats?.weeklyStreak
    ? `${bootstrapData.value.stats.weeklyStreak} 周`
    : "本周";

  return [
    { key: "completed", value: stats.value.completed ?? 0, label: "已完成" },
    { key: "duration", value: durLabel, label: "总时长" },
    { key: "avg", value: stats.value.avgScore ?? "—", label: "平均分" },
    { key: "streak", value: streakLabel, label: "连续" },
  ];
});

// S1 onboard 步骤：根据 resumeSummary / completedInterviews 标记完成状态
const onboardSteps = computed(() => {
  const resumeDone = (bootstrapData.value?.resumeSummary?.total ?? 0) > 0;
  const completedDone = (bootstrapData.value?.stats?.completedInterviews ?? 0) > 0;
  return [
    { key: "resume", label: "上传一份简历", done: resumeDone },
    { key: "direction", label: "选择练习方向", done: false },
    { key: "first", label: "完成第一场练习", done: completedDone },
  ];
});

// 核心：heroState computed —— 优先级 S1 > S2 > S3 > S4 > S0
// 设计原则：bootstrap 失败时走 S0（诚实的「连接中」空态），不再用 mock 调出 S4 假装「已有进度」。
// （原则 5：后端没字段就空态「不造假」）
const heroState = computed(() => {
  const b = bootstrapData.value;

  // bootstrap 完全失败（含 401 / 网络）时直接走 S0。
  // S0 自身包含引导文案与 CTA，比「用静态推荐词填 S4」更诚实。
  if (!b) {
    return { kind: "S0" };
  }

  // S1 首次用户：完成数 0 且没有 session
  const completedCnt = b.stats?.completedInterviews ?? 0;
  const sessionsCnt = Array.isArray(b.recentSessions) ? b.recentSessions.length : 0;
  if (completedCnt === 0 && sessionsCnt === 0) {
    return { kind: "S1" };
  }

  // S2 有进行中会话
  const cont = b.continueSession;
  if (cont) {
    const cfg = cont.config || {};
    const sid = cont.session?.sessionId || cont.session?.id;
    return {
      kind: "S2",
      title: cont.session?.title || "未命名练习",
      directionLabel: cfg.directionLabel || "未指定方向",
      difficultyLabel: cfg.difficultyLabel || "",
      progress: Math.max(0, Math.min(100, cont.progress ?? 0)),
      link: sid ? `/chat?sessionId=${encodeURIComponent(sid)}` : "/workbench/new",
    };
  }

  // S3 上次刚结束、近 24h 未跳转报告
  const last = (b.recentSessions || [])[0];
  if (last?.completedAt) {
    const completedAt = new Date(last.completedAt).getTime();
    const within24h = !Number.isNaN(completedAt) && Date.now() - completedAt < 86400000;
    if (within24h) {
      // 维度 tag：从 abilityRadar 派生最弱 / 最强
      const radar = Array.isArray(b.abilityRadar) ? b.abilityRadar : [];
      const sorted = [...radar].sort((a, c) => (a.score ?? 0) - (c.score ?? 0));
      const tags = [];
      if (sorted[0]) tags.push({ key: "weak", label: `${sorted[0].label} 待加强`, level: "weak" });
      if (sorted[sorted.length - 1] && sorted.length > 1) {
        tags.push({ key: "strong", label: `${sorted[sorted.length - 1].label} 出色`, level: "strong" });
      }
      // 评分：暂从 stats.averageScore 借用（后端 SessionItem 没单独 score 字段）
      const score = Math.round(b.stats?.averageScore ?? 0) || 0;
      // 关键验证：score=0 表示后端 OverallScore 还是 insufficient_data 或 draft状态（见论文 ch05），
      // 不应在 hero 卡中展示「0/100」；此时跳过 S3，让后续逻辑走 S4 推荐。
      if (score > 0) {
        const sid = last.sessionId || last.id || "";
        // 死链修复：vue-router 没有注册 /reports/:id，跳过去会被 catch-all吃掉。
        // CTA 应跳会话回看页 /chat?sessionId=xxx（现有路由）。
        return {
          kind: "S3",
          title: last.title || "上次练习",
          score,
          maxScore: 100,
          tags,
          link: sid ? `/chat?sessionId=${encodeURIComponent(sid)}` : "/workbench",
          ctaLabel: "回看本次对话",
        };
      }
    }
  }

  // S4 推荐：优先使用 nextActions[0]，否则 fallback 到通用文案
  const next = (b.nextActions || [])[0];
  if (next) {
    return {
      kind: "S4",
      title: next.label || "下一步建议",
      description: next.description || "",
      ctaLabel: "立即开始",
      link: normalizeWorkbenchRoute(next.route || "/workbench/new"),
    };
  }
  return {
    kind: "S4",
    title: "保持节奏：来一场新练习",
    description: "每周 2-3 场是最佳频率，可以让你看到稳定的进步曲线。",
    ctaLabel: "新建一场",
    link: "/workbench/new",
  };
});

const normalizeWorkbenchRoute = (route) => {
  const value = String(route || "").trim();
  const routeMap = {
    "/interview/new": "/workbench/new",
    "/resume": "/workbench/resume",
    "/knowledge": "/workbench/knowledge",
    // /reports 与 /workbench/reports 都映射到新增的报告中心路由
    "/reports": "/workbench/reports",
    "/workbench/reports": "/workbench/reports",
  };
  return routeMap[value] || value || "/workbench/new";
};

// ============ 4 业务领域状态卡 computed 派生 ============
// 设计原则：每张卡都有 hasXxx 判断 0 数据态，文案 / CTA / link 都通过 computed 派生
// 让 template 保持声明式，所有业务逻辑（如"完成首场后跳哪里"）集中在 script 里维护。

// === 卡 1：报告中心 ===
// 0 数据态（决策 4=B）：显示「完成首场面试后查看分析」+ CTA 跳 /workbench/new
// 有数据态：显示「N 份报告 · 平均 X 分」+ CTA 跳 /workbench/reports（决策 5=A：占位 SFC）
const hasReports = computed(() => stats.value.completed > 0);
const reportsTag = computed(() => (hasReports.value ? "已生成" : "等待数据"));
const reportsLink = computed(() => (hasReports.value ? "/workbench/reports" : "/workbench/new"));
const reportsCta = computed(() => (hasReports.value ? "看分析" : "去面试"));
const reportsMeta = computed(() => {
  if (!hasReports.value) return "暂未生成";
  return stats.value.lastAt && stats.value.lastAt !== "暂无"
    ? `上次 ${stats.value.lastAt}`
    : "已就绪";
});

// === 卡 2：简历库 ===
// 0 数据态：显示「上传简历后 AI 基于项目经历追问」+ CTA "上传 →"
// 有数据态：显示「N 份 · M 片段已入库」+ CTA "管理 →"
const hasResume = computed(() => resumeSummary.value.total > 0);
const resumeStatus = computed(() => (hasResume.value ? "已上传" : "未上传"));
const resumeTotal = computed(() => resumeSummary.value.total || 0);
const resumeChunkCount = computed(() => resumeSummary.value.chunkCount || 0);
// projectsCount（0e22383 后端交付）：模板优先用「N 个项目」文案，0 时 fallback 到 chunkCount。
const resumeProjectsCount = computed(() => resumeSummary.value.projectsCount || 0);
const resumeMeta = computed(() => {
  if (!hasResume.value) return "上传后开启项目深度追问";
  return resumeSummary.value.latestTitle
    ? `最新：${resumeSummary.value.latestTitle}`
    : "已分析";
});
const resumeCta = computed(() => (hasResume.value ? "管理" : "上传"));

// === 卡 3：知识库 ===
// 0 数据态：显示「上传文档后 AI 引用资料」+ CTA "上传 →"
// 有数据态：显示「N 篇 · M 块」+ CTA "管理 →"
const hasKnowledge = computed(() => knowledgeSummary.value.documents > 0);
const knowledgeTag = computed(() => (hasKnowledge.value ? "已入库" : "未上传"));
const knowledgeDocuments = computed(() => knowledgeSummary.value.documents || 0);
const knowledgeChunks = computed(() => knowledgeSummary.value.chunks || 0);
const knowledgeMeta = computed(() => {
  if (!hasKnowledge.value) return "支持 PDF / 文本资料";
  return knowledgeSummary.value.latestTitle
    ? `最新：${knowledgeSummary.value.latestTitle}`
    : "已就绪";
});
const knowledgeCta = computed(() => (hasKnowledge.value ? "管理" : "上传"));

// === 卡 4：题库 ===
// 数据源：interviewPresets.directions[].questionCount 累加
// 后端入库 1614 题后 questionCount 自动从 DB 派生，前端代码无需改
const bankTotalQuestions = computed(() => {
  const dirs = interviewPresets.value.directions || [];
  if (dirs.length === 0) return stats.value.bankCount || 0;
  return dirs.reduce((sum, d) => sum + (typeof d.questionCount === "number" ? d.questionCount : 0), 0);
});
const bankDirectionsCount = computed(() => (interviewPresets.value.directions || []).length);

// === 时间格式化（绝对时间戳 → "2 小时前" / "昨天" / "3 天前" 等） ===
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

// === 异步加载真实数据 ===
// 设计原则（与原则 5 对齐）：网络失败 / 401 时静默降级到空态，不预填 mock。
// 让 UI 真实反映「是否拿到数据」，避免造「已有内容」的假象。
// （删除了 inferDifficultyLevel helper：唯一调用方 toRecentSessionRow 已不再传入 difficulty
//  字段，因为后端 SessionItem 不返回 difficulty/difficultyLabel，inferDifficultyLevel
//  永远基于 undefined 返回 "mid"，是死代码。）
const loadProfile = async () => {
  try {
    const profile = await apiService.user.profile();
    if (profile?.username) {
      profileUsername.value = profile.username;
    }
    // profile 里如果有简历快照字段，可以在此覆盖 resumeName/resumeStatus
    // 当前后端契约里 profile 不返回简历，留给后续 Resume 子页接管。
  } catch (error) {
    // 静默降级：保留 localStorage 里的 username 兜底。
  }
};

// 将后端 SessionItem 转换为表格行数据。
// 后端 SessionItem 当前只返回 10 个字段（sessionId/title/mode/modeKey/messageCount/
// isActive/createdAt/updatedAt/lastMessageAt/completedAt），不返回 direction/difficulty/
// score/duration。原则 5：不造 hardcode mock 兜底（原 "未指定"/"中级"/30m 已删除），
// 空字段让模板侧 v-if 决定显示与否，待独立 UX 任务做模板列重构（6 列 → 5 列对齐契约）。
// 过渡期允许「破列」呈现：方向空、难度空 span、分 0、时长 0m，比 mock 假数据诚实。
const toRecentSessionRow = (s, i) => {
  const id = s.sessionId || s.id || `s-${i}`;
  return {
    id,
    title: s.title || s.topic || "未命名会话",
    direction: s.direction || s.directionLabel || "",
    difficulty: s.difficulty || s.difficultyLabel || "",
    difficultyLevel: "",
    score: typeof s.score === "number" ? s.score : 0,
    duration: typeof s.duration === "number" ? s.duration : (s.durationMinutes || 0),
    time: formatRelativeTime(s.updatedAt || s.createdAt || s.lastMessageAt || s.lastActiveAt),
    link: `/chat?sessionId=${encodeURIComponent(id)}`,
  };
};

const loadSessions = async () => {
  try {
    const res = await apiService.user.sessions();
    const list = Array.isArray(res?.sessions)
      ? res.sessions
      : Array.isArray(res)
        ? res
        : [];
    // 空列表也写入空数组：明确「没有数据」而不是「还在加载」。
    // 原则 5：不再「列表为空时保留 mock」，模板 wb-empty 会接手呈现「还没有面试记录」。
    recentSessions.value = list.slice(0, 5).map(toRecentSessionRow);

    // 仅用 list.length 更新 completed；avgScore 完全交给 bootstrap.stats.averageScore，
    // 不做基于 row.score 的前端二次计算（删除原 scoredList.reduce 死代码：后端 SessionItem
    // 不返回 score 字段，row.score 永远为 0，scoredList 永远为空，二次计算永远不执行）。
    stats.value.completed = list.length;
    if (recentSessions.value[0]) {
      stats.value.lastAt = recentSessions.value[0].time;
    }
  } catch (error) {
    // 接口失败时保持 recentSessions 为空数组，模板 wb-empty 分支生效。
    // 不再造 mock 假数据，让 0 数据态如实呈现。
  }
};

// applyContinueSession 已删除：旧 4 卡的「继续上次面试」与 Hero S2 多态卡 100% 功能重复，
// 现在只由 heroState computed 处理 continueSession，4 卡区不再消费此字段。

// 用后端返回的能力雷达覆盖本地 5 维默认。维持 5 个项以避免雷达多边形变形。
const applyAbilityRadar = (points) => {
  if (!Array.isArray(points) || points.length === 0) return;
  const next = points.slice(0, 5).map((p) => ({
    key: p.key || p.label || "未命名",
    label: p.label || "未命名",
    value: typeof p.score === "number" ? p.score : 0,
    maxScore: typeof p.maxScore === "number" && p.maxScore > 0 ? p.maxScore : 100,
  }));
  // 不足 5 个时补零位维持多边形闭合
  while (next.length < 5) {
    next.push({ key: `pad-${next.length}`, label: "\u2014", value: 0, maxScore: 100 });
  }
  radarDims.value = next;
};

// 用后端简历摘要覆盖卡片数据源。新版 4 卡的简历库卡通过 computed 派生展示文案。
const applyResumeSummary = (summary) => {
  if (!summary) return;
  resumeSummary.value = {
    total: typeof summary.total === "number" ? summary.total : 0,
    latestTitle: summary.latestTitle || "",
    chunkCount: typeof summary.chunkCount === "number" ? summary.chunkCount : 0,
    // 0e22383 后端补充的项目数摘要：用户语义级别指标，比 chunkCount（embed 内部细节）更可读。
    projectsCount: typeof summary.projectsCount === "number" ? summary.projectsCount : 0,
    latestUpdatedAt: summary.latestUpdatedAt || "",
  };
};

// 用后端知识库摘要覆盖卡片数据源（之前完全没消费这个字段）。
const applyKnowledgeSummary = (summary) => {
  if (!summary) return;
  knowledgeSummary.value = {
    documents: typeof summary.documents === "number" ? summary.documents : 0,
    chunks: typeof summary.chunks === "number" ? summary.chunks : 0,
    latestTitle: summary.latestTitle || "",
    latestAddedAt: summary.latestAddedAt || "",
  };
};

// 拉 interviewPresets，给题库卡提供方向数 + 题数累加。
// 这个端点在 WorkbenchNew.vue 也用，但工作台首页是 4 卡的题库卡专用入口，
// 因此 onMounted 中独立调用一次（小数据 + 高复用，不需要 store 缓存）。
const loadInterviewPresets = async () => {
  try {
    const res = await apiService.user.interviewPresets();
    if (!res) return;
    interviewPresets.value = {
      directions: Array.isArray(res.directions) ? res.directions : [],
      difficulties: Array.isArray(res.difficulties) ? res.difficulties : [],
      focusOptions: Array.isArray(res.focusOptions) ? res.focusOptions : [],
    };
  } catch (error) {
    // 静默降级：computed 在 directions=[] 时显示「采集中」占位文案。
  }
};

// 首选：一个接口拿首屏全部数据。失败后 fallback 到 profile + sessions 双调用。
const loadWorkbenchBootstrap = async () => {
  try {
    const res = await apiService.user.workbenchBootstrap();
    if (!res) throw new Error("empty bootstrap");

    // 保存原始响应快照：heroState computed 依赖 continueSession / nextActions /
    // recentSessions[].completedAt / abilityRadar 等嵌套字段，扁平 stats 不够。
    bootstrapData.value = res;

    if (res.user?.username) profileUsername.value = res.user.username;

    if (res.stats) {
      if (typeof res.stats.completedInterviews === "number") {
        stats.value.completed = res.stats.completedInterviews;
      }
      if (typeof res.stats.averageScore === "number" && res.stats.averageScore > 0) {
        stats.value.avgScore = Math.round(res.stats.averageScore);
      }
      if (res.stats.lastPracticeAt) {
        stats.value.lastAt = formatRelativeTime(res.stats.lastPracticeAt);
      }
    }

    // continueSession 由 heroState computed 直接读 bootstrapData，不再有 applyContinueSession
    applyAbilityRadar(res.abilityRadar);
    applyResumeSummary(res.resumeSummary);
    applyKnowledgeSummary(res.knowledgeSummary);

    if (Array.isArray(res.recentSessions) && res.recentSessions.length > 0) {
      recentSessions.value = res.recentSessions.slice(0, 5).map(toRecentSessionRow);
      if (recentSessions.value[0]) {
        stats.value.lastAt = recentSessions.value[0].time;
      }
    }
    return true;
  } catch (error) {
    return false;
  }
};

onMounted(async () => {
  // 原则 5：不预填 mock。所有数据从 bootstrap / sessions / presets 拉，失败走空态。
  // 让 UI 真实反映「是否拿到数据」，不造「已有内容」的假象。
  // 代价：首次进入会有 ~200ms 「空态 → 有数据」的闪烁（可接受），换取「0 数据态不造假」的系统诚实性。

  // 首选 bootstrap 一次拿完；失败时 fallback 到 profile + sessions 双调用。
  // interviewPresets 与 bootstrap 并行拉（独立端点 + 给题库卡用），失败不影响主流程。
  const [ok] = await Promise.all([
    loadWorkbenchBootstrap(),
    loadInterviewPresets(),
  ]);
  if (!ok) {
    loadProfile();
    loadSessions();
  }
});
</script>

<style scoped>
/* ============ Layout 容器 ============ */
.wb-content {
  max-width: 1320px;
  margin: 0 auto;
  padding: 0 44px 80px;
  /* z-index 留给 main 容器统一管理；本层不要自己抢 stacking。 */
}

/* ============ Hero（v2 多态：左 60% 数据 + 右 40% 状态卡） ============ */
/* Hero v2 多态卡：grid 50:50（应用户反馈）、左右等重。
   左侧是数据陈述主角、右侧是动作迷你卡，等宽布局让右卡有足够的身体到达 editorial 质感。 */
.wb-hero {
  /* hero-scoped 质感 token：提炼自概念图 Vision DNA（详见届于同事 256-workbench-hero-v2s3 的提取报告）。
     这些变量只在 .wb-hero 作用域生效，不污染 Home 或其他页面。 */
  --hero-card-bg-1: rgba(22, 20, 18, 1);
  --hero-card-bg-2: rgba(20, 18, 16, 1);
  --hero-card-border: rgba(255, 255, 255, 0.08);
  --hero-card-radius: 20px;
  --hero-card-shadow:
    0 24px 70px rgba(0, 0, 0, 0.42),
    0 2px 10px rgba(0, 0, 0, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.05);
  --hero-text-warm: #f3eee7;
  --hero-text-muted: rgba(243, 238, 231, 0.62);
  --hero-text-soft: rgba(243, 238, 231, 0.46);
  --hero-amber: rgba(240, 180, 60, 0.95);
  --hero-amber-soft: rgba(240, 180, 60, 0.14);
  --hero-green: rgba(140, 220, 160, 0.95);

  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 40px;
  align-items: stretch;
  padding: 0 0 56px;
  /* 卡加大后需要更高的底限，给 hero 整体仍保持不闪烁 */
  min-height: 360px;
}

.wb-hero-left {
  display: flex;
  flex-direction: column;
  /* justify-content: center 让左侧文本组在 grid 行高内垂直居中，
     与右侧卡片的视觉中线对齐（陷阱 #11：grid stretch 默认会让子项顶起）。 */
  justify-content: center;
  /* min-width: 0 让长 username 能省略而非撑破网格。 */
  min-width: 0;
}

.wb-hero-right {
  display: flex;
  align-items: center;
  /* min-width: 0 同上。 */
  min-width: 0;
}

/* 响应式：< 1024px 右侧卡下沉到左侧文本之下，单列堆叠避免挤压。 */
@media (max-width: 1024px) {
  .wb-hero {
    grid-template-columns: 1fr;
    gap: 32px;
    min-height: 0;
  }
  .wb-hero-right {
    align-items: flex-start;
  }
}

.wb-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font: 12px var(--mono);
  /* 文字从 var(--t2) 中性灰白 → 暖米调：与 amber dot 同体系，
     让「本周·W19」看起来是一个集成的 amber 状态指示器而不是中性文字+颜色标。 */
  color: rgba(243, 230, 210, 0.78);
  border: 1px solid rgba(240, 180, 60, 0.18);
  border-radius: var(--radius-pill);
  padding: 6px 14px;
  margin-bottom: 22px;
  letter-spacing: .04em;
  background: rgba(240, 180, 60, 0.04);
  backdrop-filter: blur(8px);
  white-space: nowrap;
  width: fit-content;
}

.wb-eyebrow-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  /* dot 从纯上色加 box-shadow glow：让 dot 不是「安静色点」而是「低调发光体」，
     与 username halo 形成「小点 → 大名」的同质感颜色语言。 */
  background: rgba(240, 180, 60, 0.95);
  box-shadow: 0 0 6px rgba(240, 180, 60, 0.45);
  animation: wb-edot 2.6s ease-in-out infinite;
}

@keyframes wb-edot {
  0%, 100% { opacity: 1; }
  50% { opacity: .35; }
}

/* hero-title 比 Home 收敛：clamp(32px, 3vw, 48px)，不要主页 hero 那种 66px 巨字。
   工作台是工具属性，标题不应抢镜；display 字体保留以维持品牌识别。 */
.wb-title {
  font-size: clamp(32px, 3vw, 48px);
  font-weight: 800;
  font-family: var(--display);
  line-height: 1.18;
  letter-spacing: -.02em;
  color: var(--t);
  margin: 0 0 14px;
}

.wb-title-name {
  /* 用户名是 hero 视觉主焦点，应用「流光四溢」动态特效让"你是谁"的辨识度最大化：
     1. background-clip: text 把渐变剪成字形，作为流光填充层
     2. -webkit-text-stroke 1px amber 描边，让字符外圈有暖琥珀立体感
     3. drop-shadow filter 给字体一圈温暖光晕（halo），从字体本身发散
     4. linear-gradient amber→暖白→amber→暖白（200% 宽）+ 8s 循环动画 = 横向流光

     性能注：filter drop-shadow 在 Chrome/Firefox/Safari 都 GPU 加速，60fps 无负担。
     8s 周期偏长不打扰但能感知，避免成为干扰主视觉的"卡通效果"。 */
  position: relative;
  font-weight: 800;
  /* 流光填充：高光段从 1.0 降到 0.85，使「光斜过」更为动得头床、不刮眼，
     与页面其他低饱和 amber 语言肤调一致 */
  background: linear-gradient(
    100deg,
    rgba(220, 155, 90, 0.95) 0%,
    rgba(255, 230, 200, 0.85) 18%,
    rgba(240, 180, 60, 0.92) 35%,
    rgba(255, 230, 200, 0.85) 52%,
    rgba(220, 155, 90, 0.95) 70%,
    rgba(240, 180, 60, 0.92) 100%
  );
  background-size: 200% 100%;
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  color: transparent;
  /* amber 描边：降为 0.28 opacity，从“金属雕刻”调到“软描边”，避免物质打压其他元素 */
  -webkit-text-stroke: 1px rgba(240, 180, 60, 0.28);
  /* drop-shadow halo 降一档：从双层 (6+14px) 收为单层 8px，opacity 0.22，
     halo 仍在但不再「四溢”。但仃是个背景上可辨识的暖点。 */
  filter: drop-shadow(0 0 8px rgba(240, 180, 60, 0.22));
  /* 横向流光循环：8s linear 让光"从右往左"扫过 */
  animation: wb-name-shimmer 8s linear infinite;
}

/* prefers-reduced-motion: 关闭流光动画，保留静态 amber 渐变填充以维持视觉品质 */
@media (prefers-reduced-motion: reduce) {
  .wb-title-name {
    animation: none;
  }
}

@keyframes wb-name-shimmer {
  0% {
    background-position: 200% 50%;
  }
  100% {
    background-position: 0% 50%;
  }
}

.wb-sub {
  font-size: 16px;
  color: var(--t3);
  line-height: 1.7;
  margin: 0 0 32px;
  max-width: 560px;
}

/* 标题 greet 段：弱化"早，"等问候，让 username 成为主焦点 */
.wb-title-greet {
  color: var(--t2);
  font-weight: 600;
  margin-right: 6px;
}

/* ============ 左侧 4 metric 数据条 ============ */
/* 沿用 .wb-eyebrow 同款 panel 视觉语法（半透白底 + 1px 暗边 + 8px 圆角），
   不另立第三种"卡风格"，保持 Hero 区视觉词汇收敛。 */
.wb-metrics {
  display: inline-flex;
  align-items: center;
  gap: 24px;
  padding: 16px 24px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-md);
  /* width: fit-content 让数据条按内容收缩，左对齐而非撑满左栏 */
  width: fit-content;
  max-width: 100%;
  flex-wrap: wrap;
}

.wb-metric {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.wb-metric-num {
  font: 700 22px var(--mono);
  color: var(--t);
  letter-spacing: -.02em;
  line-height: 1;
  white-space: nowrap;
}

/* 第一个 metric（「已完成」）字号加 amber gradient text fill：
   让「本周最重要的成果数」与 username 、eyebrow dot 、右侧卡 amber accent 形成贯穿全页的
   amber accent 节奏。只 accent 一个数字避免过载，后三个进入中性白表达「参考数据」。 */
.wb-metric:first-child .wb-metric-num {
  background: linear-gradient(135deg, rgba(255, 230, 200, 0.95) 0%, rgba(240, 180, 60, 0.92) 100%);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  color: transparent;
}

.wb-metric-lb {
  font-size: 12px;
  color: var(--t3);
  letter-spacing: .02em;
  white-space: nowrap;
}

.wb-metric-sep {
  position: absolute;
  right: -12px;
  top: 50%;
  transform: translateY(-50%);
  width: 1px;
  height: 24px;
  background: rgba(255, 255, 255, 0.08);
}

/* ============ 右侧多态卡通用容器 ============ */
/* 质感取向：editorial-luxury 取向 — 哑光暖黑 + hairline 边框 + 软阴影 + grain texture。
   与下方 .wb-qcard 快捷卡区分：hero 卡更厚、更安静、更「叙事」。 */
.wb-card {
  position: relative;
  display: flex;
  flex-direction: column;
  width: 100%;
  /* 加高到 320 让内部 padding 染上后不会压迫，给 score 88px 足够展示空间 */
  min-height: 320px;
  /* padding 从原 24/26/22 加到 36/36/32，朝概念图 56/52 靠拢，
     但考虑右卡状态下 1440 视口实际宽度约 600px，36 是舒适上限。 */
  padding: 36px 36px 32px;
  background:
    linear-gradient(180deg,
      var(--hero-card-bg-1) 0%,
      var(--hero-card-bg-2) 100%
    ) padding-box,
    linear-gradient(160deg,
      rgba(255, 255, 255, 0.10) 0%,
      rgba(255, 255, 255, 0.03) 50%,
      rgba(255, 255, 255, 0.06) 100%
    ) border-box;
  border: 1px solid transparent;
  border-radius: var(--hero-card-radius);
  box-shadow: var(--hero-card-shadow);
  isolation: isolate;
  overflow: hidden;
}

/* Grain texture overlay：使用 inline SVG noise 避免静态资源依赖。
   noise opacity 0.08 + soft-light blend，仅在迫近看能感知，底下提升质感。 */
.wb-card::after {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='160' height='160'><filter id='n'><feTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='2' stitchTiles='stitch'/><feColorMatrix values='0 0 0 0 1  0 0 0 0 1  0 0 0 0 1  0 0 0 0.5 0'/></filter><rect width='100%' height='100%' filter='url(%23n)'/></svg>");
  background-size: 160px 160px;
  opacity: 0.08;
  mix-blend-mode: soft-light;
  border-radius: inherit;
  z-index: 0;
}

/* 卡内所有子元素该在 grain 层之上 */
.wb-card > * {
  position: relative;
  z-index: 1;
}

/* eyebrow：所有 4 状态共用的小 amber 标签。
   editorial 质感关键：letter-spacing 加到 .24em（原 .12em 太紧），
   font-weight 从 600 缓到 500，代价是在深色背景上这种肥鲑赢中作肥文本需要颜色颜颜亮。 */
.wb-card-eyebrow {
  font: 500 12px var(--mono);
  letter-spacing: .24em;
  text-transform: uppercase;
  color: var(--hero-amber);
  margin-bottom: 24px;
}

/* foot：CTA 区，统一对齐到右下，给主体内容腾出垂直空间 */
.wb-card-foot {
  margin-top: auto;
  display: flex;
  justify-content: flex-end;
  padding-top: 24px;
}

/* CTA 主按钮（amber 填充）：S1/S2 用，最高优先级动作。
   加大 padding 从 10/18 到 13/24 让主 CTA 更咨定、更「难以忽略」。 */
.wb-card-cta-amber {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 13px 24px;
  background: var(--hero-amber);
  color: #1a140d;
  font: 600 15px var(--sans);
  letter-spacing: .01em;
  border-radius: 12px;
  text-decoration: none;
  transition: transform .2s ease, opacity .2s ease, box-shadow .2s ease;
  /* 暖色阴影呼应 amber 主调，加应 + 沉在卡里的微微外推推感 */
  box-shadow:
    0 4px 14px rgba(240, 180, 60, 0.22),
    0 8px 24px rgba(240, 180, 60, 0.10);
}

.wb-card-cta-amber:hover {
  transform: translateY(-1px);
  opacity: .94;
  box-shadow: 0 6px 18px rgba(220, 155, 90, 0.28);
}

/* CTA 文字 + chevron：S3/S4 用，次优先级，与 amber 主按钮视觉权重区分。
   editorial-mode：font-weight 从 600 缓到 500，letter-spacing 加 .01em，字号从 14 加到 16。 */
.wb-card-cta-text {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 4px;
  color: var(--hero-text-muted);
  font: 500 16px var(--sans);
  letter-spacing: .01em;
  text-decoration: none;
  transition: color .2s ease, gap .2s ease;
}

.wb-card-cta-text:hover {
  color: var(--hero-text-warm);
  gap: 12px;
}

/* ============ S1 onboard：3 步引导 ============ */
/* 步骤间贴合更紧（gap 0）但每 step 上下加 padding，
   负黑废上看不出差异却能在“line连接”伪元素上连贯。 */
.wb-onboard-steps {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  /* gap 0 + step padding 11px 达到 “贴合但不拥挤”，同时让競竖连接线能连贯 */
  gap: 0;
}

.wb-onboard-step {
  position: relative;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 11px 0;
  font-size: 14px;
  color: var(--t2);
}

/* 竹形连接线：每个 step 底部 伸出一根短竹进下一个 step，
   最后一个 step 不画。竹线 x 坐标跟 num 圈中心对齐（16 - 0.5 = 15.5px）。 */
.wb-onboard-step:not(:last-child)::before {
  content: '';
  position: absolute;
  left: 11.5px;
  top: 33px;
  width: 1px;
  height: 14px;
  background: rgba(255, 255, 255, 0.12);
}

.wb-onboard-num {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  border-radius: var(--radius-circle);
  border: 1px solid rgba(255, 255, 255, 0.18);
  background: var(--bg);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font: 600 12px var(--mono);
  color: var(--t3);
  /* 圈需要遮住背后竹线的中间部分，则透明背景干扰 */
  position: relative;
  z-index: 1;
}

.wb-onboard-step.is-done .wb-onboard-num {
  background: rgba(220, 155, 90, 0.12);
  border-color: rgba(220, 155, 90, 0.5);
  color: rgba(220, 155, 90, 0.95);
}

.wb-onboard-step.is-done .wb-onboard-label {
  color: var(--t);
  text-decoration: line-through;
  text-decoration-color: rgba(255, 255, 255, 0.18);
}

/* ============ S2 continue：进度条 + topic + tag ============ */
.wb-continue-progress {
  width: 100%;
  height: 6px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-pill);
  overflow: hidden;
  margin-bottom: 8px;
}

.wb-continue-bar {
  height: 100%;
  background: rgba(220, 155, 90, 0.85);
  border-radius: var(--radius-pill);
  transition: width .4s ease;
}

.wb-continue-meta {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 14px;
}

.wb-continue-pct {
  font: 700 18px var(--mono);
  color: var(--t);
}

.wb-continue-dim {
  font-size: 12px;
  color: var(--t3);
}

.wb-continue-topic {
  font: 600 16px var(--sans);
  color: var(--t);
  line-height: 1.4;
  margin-bottom: 10px;
  /* 长标题省略，避免撑破卡 */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.wb-continue-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

/* ============ S3 review：score + tags ============ */
.wb-review-score {
  display: flex;
  align-items: baseline;
  gap: 4px;
  margin-bottom: 10px;
}

.wb-review-num {
  /* 从 60 跳到 88：概念图是 128-144，但实际全全面渲染时右卡宽 600、hero 高限制下
     88 是「dominant 但不破坏布局」的平衡点。 */
  font: 600 88px var(--display);
  color: var(--hero-text-warm);
  letter-spacing: -.04em;
  line-height: .9;
}

.wb-review-max {
  font: 400 22px var(--mono);
  color: var(--hero-text-soft);
  margin-left: 6px;
}

.wb-review-title {
  font: 600 14px var(--sans);
  color: var(--t2);
  line-height: 1.4;
  margin-bottom: 12px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.wb-review-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

/* ============ S4 next：推荐项 ============ */
.wb-next-title {
  /* serif display：从 26 加到 30px，加在 Noto Serif SC 中文衢线字体上能明显体现 editorial 感。 */
  font: 500 30px var(--display);
  color: var(--hero-text-warm);
  line-height: 1.25;
  letter-spacing: -.02em;
  margin-bottom: 14px;
}

.wb-next-desc {
  font-size: 15px;
  color: var(--hero-text-muted);
  line-height: 1.7;
  margin: 0 0 18px;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* ============ S0 fallback ============ */
.wb-fallback-title {
  font: 600 18px var(--display);
  color: var(--t);
  margin-bottom: 8px;
}

.wb-fallback-desc {
  font-size: 14px;
  color: var(--t3);
  line-height: 1.65;
  margin: 0;
}

/* ============ 通用 tag（S2/S3 共用） ============ */
/* editorial chip：outline 式、透明 fill、中型 radius，不是 pill，比原型更「杂志」。 */
.wb-tag {
  display: inline-flex;
  align-items: center;
  padding: 6px 14px;
  font: 500 13px var(--sans);
  color: var(--hero-text-muted);
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.10);
  border-radius: 10px;
  letter-spacing: .01em;
}

.wb-tag-weak {
  color: var(--hero-amber);
  border-color: rgba(240, 180, 60, 0.45);
  background: rgba(240, 180, 60, 0.04);
}

.wb-tag-strong {
  color: var(--hero-green);
  border-color: rgba(140, 220, 160, 0.35);
  background: rgba(140, 220, 160, 0.04);
}

/* ============ Hero stats（保留兼容旧实现，v2 重构后不再使用，可在下一轮移除） ============ */
.wb-stats {
  display: inline-flex;
  align-items: center;
  gap: 28px;
  padding: 18px 28px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.9) 0%, rgba(11, 12, 16, 0.85) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.02) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  backdrop-filter: blur(12px);
}

.wb-stat {
  display: flex;
  flex-direction: column;
  gap: 4px;
  /* min-width: 0 让 grid 子项可以收缩，避免长数字撑破容器（陷阱 #12）。 */
  min-width: 0;
}

.wb-stat-num {
  font: 700 24px var(--mono);
  color: var(--t);
  letter-spacing: -.02em;
  line-height: 1;
}

.wb-stat-lb {
  font-size: 12px;
  color: var(--t3);
  letter-spacing: .02em;
}

.wb-stat-sep {
  width: 1px;
  height: 24px;
  background: rgba(255, 255, 255, 0.1);
}

/* ============ 4 业务领域状态卡 ============ */
/* 设计定位（详见 docs/requirements/2026-05-10-workbench-information-architecture.md）：
   - 与 Hero v2 的 wb-card 同源质感（哑光黑 + 多层 shadow + radius 20）但薄一点（min-height 220 变 240）
   - 保持"hero 厚 / 4 卡薄"的视觉层级
   - amber accent 仅在「主数字」上使用 gradient text，让 1614 / 348 / N 这些量化指标成为视觉锚 */
.wb-quick {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 18px;
  margin-bottom: 56px;
}

/* 单卡：与 Hero v2 的 wb-card 同质感（哑光暖黑 + hairline border + 多层 shadow） */
.wb-qcard {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 14px;
  /* min-height 240：比 hero card 320 薄点，但比原 220 厚，让 4 卡与 hero 状态卡视觉节奏对齐。 */
  min-height: 240px;
  padding: 28px 30px 24px;
  background:
    linear-gradient(180deg,
      rgba(22, 20, 18, 1) 0%,
      rgba(18, 16, 14, 1) 60%,
      rgba(13, 12, 10, 1) 100%
    ) padding-box,
    linear-gradient(160deg,
      rgba(255, 255, 255, 0.10) 0%,
      rgba(255, 255, 255, 0.03) 30%,
      rgba(255, 255, 255, 0.02) 70%,
      rgba(255, 255, 255, 0.06) 100%
    ) border-box;
  border: 1px solid transparent;
  /* radius 20 与 hero card 对齐，让 4 卡与 hero 会话同质感 */
  border-radius: 20px;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.04),
    0 18px 44px rgba(0, 0, 0, 0.36),
    0 2px 8px rgba(0, 0, 0, 0.18);
  /* isolation 让 hover translateY 不与 aurora 父层叠加产生 stacking 异常。 */
  isolation: isolate;
  transition: transform .28s ease, box-shadow .28s ease, border-color .28s ease;
}

.wb-qcard:hover {
  transform: translateY(-3px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.07),
    0 24px 60px rgba(0, 0, 0, 0.46),
    0 2px 10px rgba(0, 0, 0, 0.22);
}

/* === 各卡 amber accent 区分。 ===
   设计原则：4 卡 amber 浓度梯度上升让「资源量」反映在色彩层级。
   - report (报告中心)：amber Z+1（深明） — 是头牌业务、应该最冸
   - resume (简历库)：amber Z（中明）
   - knowledge (知识库)：amber Z（中明）
   - bank (题库)：amber Z（中明）
   amber Z+1 / Z 的差异在 hover 边框上呈现，静态保持一致 hairline 以免锁锁热闹。 */
.wb-qcard-report:hover {
  border-color: rgba(240, 180, 60, 0.42);
}

.wb-qcard-resume:hover,
.wb-qcard-knowledge:hover,
.wb-qcard-bank:hover {
  border-color: rgba(240, 180, 60, 0.28);
}

/* card head：图标 + tag 横排。 */
.wb-qcard-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.wb-qcard-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: var(--t);
  flex-shrink: 0;
}

.wb-qcard-icon svg {
  width: 18px;
  height: 18px;
  display: block;
}

/* 报告中心卡的 icon 包中 amber，让头牌卡鬼象靠不住。 */
.wb-qcard-report .wb-qcard-icon {
  background: rgba(240, 180, 60, 0.10);
  border-color: rgba(240, 180, 60, 0.30);
  color: rgba(240, 180, 60, 0.95);
}

.wb-qcard-tag {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
  padding: 3px 9px;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  white-space: nowrap;
  text-transform: uppercase;
}

.wb-tag-amber {
  color: rgba(220, 155, 90, 0.95);
  background: rgba(220, 155, 90, 0.08);
  border-color: rgba(220, 155, 90, 0.25);
}

/* === 卡「主数字」 amber gradient text === */
/* 1614 / 348 / 5 / N 这些量化数据是各卡的视觉锚，
   用 background-clip: text + amber gradient 让数字从「描述文」中跳出来、
   成为「你现在拥有多少」的反身。与 radar 各能力点、metric 那一些 amber 变体形成同一色彩语言。 */
.wb-qcard-num {
  font: 700 22px var(--display);
  background: linear-gradient(
    135deg,
    rgba(255, 230, 160, 0.95) 0%,
    rgba(240, 180, 60, 0.95) 55%,
    rgba(220, 145, 65, 0.92) 100%
  );
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: -.01em;
  /* 微薄 letter-spacing 让 4 位以上数字（1614）可读性更好 */
  margin-right: 2px;
}

.wb-qcard-title {
  font: 700 18px var(--display);
  color: var(--t);
  margin: 0;
  letter-spacing: -.01em;
  /* min-width: 0 + ellipsis 防止超长标题撑破卡。 */
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-qcard-desc {
  font-size: 13px;
  color: var(--t3);
  line-height: 1.6;
  margin: 0;
  /* 2 行截断 */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.wb-qcard-spacer {
  flex: 1;
}

.wb-qcard-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: auto;
}

.wb-qcard-meta {
  font: 12px var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
}

.wb-qcard-link {
  font: 600 13px var(--sans);
  color: var(--t);
  text-decoration: none;
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  transition: background-color .2s ease, border-color .2s ease, color .2s ease;
  white-space: nowrap;
}

.wb-qcard-link:hover {
  background: rgba(220, 155, 90, 0.12);
  border-color: rgba(220, 155, 90, 0.4);
  color: rgba(220, 155, 90, 0.95);
}

/* 报告中心卡的 CTA 采用塑质感：白底 + 黑字，与「报告 = 核心产出」的业务重量匹配。
   其他 3 卡保持 hairline outline。 */
.wb-qcard-report .wb-qcard-link {
  background: rgba(240, 180, 60, 0.92);
  color: rgba(20, 18, 14, 1);
  border-color: transparent;
  font-weight: 700;
}

.wb-qcard-report .wb-qcard-link:hover {
  background: rgba(255, 200, 100, 1);
  color: rgba(20, 18, 14, 1);
}

/* 0 数据态下报告 CTA 不加 amber wash（避免「让你去面试」提示过于强烈） */
.wb-qcard-report:not(:has(.wb-qcard-num)) .wb-qcard-link {
  background: rgba(255, 255, 255, 0.05);
  color: var(--t);
  font-weight: 600;
}

/* ============ 下方两栏：sessions + radar ============ */
.wb-bottom {
  display: grid;
  grid-template-columns: minmax(0, 1.6fr) minmax(0, 1fr);
  gap: 16px;
}

.wb-card {
  background:
    linear-gradient(180deg,
      rgba(16, 17, 22, 1) 0%,
      rgba(10, 11, 14, 1) 50%,
      rgba(7, 8, 11, 1) 100%
    ) padding-box,
    linear-gradient(160deg,
      rgba(255, 255, 255, 0.10) 0%,
      rgba(255, 255, 255, 0.03) 30%,
      rgba(255, 255, 255, 0.02) 70%,
      rgba(255, 255, 255, 0.06) 100%
    ) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  padding: 24px 26px 26px;
  /* 不要 overflow: hidden；要让 hover 上抬的卡 / Tooltip 这类 children
     在父盒外可见。表格行 hover 不需要溢出，仅文字溢出靠 ellipsis 控制。 */
  isolation: isolate;
}

.wb-block-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
  padding-bottom: 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.wb-block-title {
  font: 700 17px var(--display);
  color: var(--t);
  margin: 0;
  letter-spacing: -.01em;
}

.wb-more {
  font: 12px var(--mono);
  color: var(--t3);
  text-decoration: none;
  letter-spacing: .04em;
  transition: color .2s ease;
}

.wb-more:hover {
  color: rgba(220, 155, 90, 0.95);
}

.wb-radar-meta {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
  text-transform: uppercase;
}

/* ============ 最近面试表（grid 模拟 table）============ */
/* 使用 grid 而非 table 元素：避免 table 的 td white-space: nowrap 与百分比宽度
   组合时换行不可控；grid 列定义清晰，每列 min-width: 0 配合 ellipsis 安全收缩。 */
.wb-table {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.wb-tr {
  display: grid;
  grid-template-columns: minmax(0, 2.4fr) minmax(0, 1.2fr) minmax(0, 1fr) 60px 60px minmax(0, 1fr);
  gap: 16px;
  align-items: center;
  padding: 12px 0;
  text-decoration: none;
  color: inherit;
  border-radius: var(--radius-sm);
  transition: background-color .2s ease, padding .2s ease;
}

.wb-thead {
  /* 表头：mono / 小字 / 字距更宽，与表行视觉分层。 */
  padding: 0 0 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}

.wb-th {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .06em;
  text-transform: uppercase;
}

.wb-col-num {
  text-align: right;
}

.wb-col-time {
  text-align: right;
}

.wb-trow {
  cursor: pointer;
}

.wb-trow:hover {
  background: rgba(220, 155, 90, 0.05);
  padding: 12px 10px;
}

.wb-td {
  font-size: 13px;
  color: var(--t2);
  /* min-width: 0 + ellipsis 防止 grid 子项被内容撑爆（陷阱 #12 / #16） */
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-col-title {
  color: var(--t);
  font-weight: 500;
}

/* 难度 pill：低/中/高三色，与暖琥珀 accent 区分开。 */
.wb-diff {
  display: inline-flex;
  align-items: center;
  font: 11px var(--mono);
  padding: 2px 8px;
  border-radius: var(--radius-pill);
  letter-spacing: .04em;
}

.wb-diff-low {
  color: #9bd1a8;
  background: rgba(155, 209, 168, 0.08);
  border: 1px solid rgba(155, 209, 168, 0.18);
}

.wb-diff-mid {
  color: var(--t2);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.wb-diff-high {
  color: rgba(220, 155, 90, 0.95);
  background: rgba(220, 155, 90, 0.08);
  border: 1px solid rgba(220, 155, 90, 0.25);
}

/* 分数：dot + 数字。dot 颜色根据 wb-score-* class 切换。 */
.wb-score {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font: 600 13px var(--mono);
  color: var(--t);
}

.wb-score-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--t3);
}

.wb-score-high .wb-score-dot {
  background: rgba(220, 155, 90, 0.95);
  box-shadow: 0 0 8px rgba(220, 155, 90, 0.4);
}

.wb-score-mid .wb-score-dot {
  background: rgba(255, 255, 255, 0.65);
}

.wb-score-low .wb-score-dot {
  background: rgba(155, 175, 220, 0.7);
}

/* === 空态 === */
.wb-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 40px 20px;
  gap: 10px;
}

.wb-empty-icon {
  font-size: 32px;
  opacity: .5;
  margin-bottom: 6px;
}

.wb-empty-title {
  font: 600 16px var(--display);
  color: var(--t);
}

.wb-empty-sub {
  font-size: 13px;
  color: var(--t3);
  margin-bottom: 14px;
}

.wb-empty-cta {
  font: 600 13px var(--sans);
  color: var(--bg);
  background: var(--t);
  text-decoration: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  transition: opacity .2s ease;
}

.wb-empty-cta:hover {
  opacity: .9;
}

/* ============ 能力雷达 SVG ============ */
.wb-radar-svg {
  width: 100%;
  /* 限定显示高度，配合 viewBox 等比缩放。
     不用 aspect-ratio 1/1 是为了避免 grid 父盒在窄屏下被 SVG 强行撑高。 */
  max-height: 240px;
  display: block;
  margin: 0 auto;
}

.wb-radar-grid-line {
  fill: none;
  stroke: rgba(255, 255, 255, 0.06);
  stroke-width: 1;
}

.wb-radar-line {
  stroke: rgba(255, 255, 255, 0.04);
  stroke-width: 1;
}

.wb-radar-user {
  /* fill 加 从 0.18 → 0.24 + stroke 从 0.85 → 0.92：雷达 polygon 同步上调 amber 浓度，
     让雷达 「你的能力肢体」成为下方区域的 amber 错点，与 hero 商 amber 节奏连贯。 */
  fill: rgba(240, 180, 60, 0.24);
  stroke: rgba(240, 180, 60, 0.92);
  stroke-width: 1.6;
  stroke-linejoin: round;
}

.wb-radar-dot {
  fill: rgba(240, 180, 60, 0.95);
  stroke: var(--bg);
  stroke-width: 1.5;
  /* 雷达 顶点加 amber glow：与 eyebrow-dot 形成「点状发光体」同体系，
     让整页 amber accent 不仅是「静态着色」而是「低调辐射」。 */
  filter: drop-shadow(0 0 4px rgba(240, 180, 60, 0.45));
}

.wb-radar-label {
  font: 11px var(--mono);
  fill: var(--t2);
  letter-spacing: .04em;
  /* SVG 的 text 不被 select 影响布局，但加 user-select: none 防止误选 */
  user-select: none;
}

.wb-radar-weaks {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.wb-weaks-title {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .06em;
  text-transform: uppercase;
  margin-bottom: 10px;
}

.wb-weaks-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.wb-weak {
  font: 12px var(--sans);
  color: rgba(220, 155, 90, 0.95);
  padding: 4px 10px;
  border-radius: var(--radius-pill);
  background: rgba(220, 155, 90, 0.06);
  border: 1px solid rgba(220, 155, 90, 0.2);
  cursor: default;
  transition: background-color .2s ease, border-color .2s ease;
}

.wb-weak:hover {
  background: rgba(220, 155, 90, 0.12);
  border-color: rgba(220, 155, 90, 0.4);
}

/* ============ 雷达 0 数据态：骨架 + 引导文案 ============ */
/* 设计原则（原则 5：不造 mock）：
   后端 abilityRadar 返回空 / 全 0 时，不再画一个塌缩到中心的多边形 + 78/90 假评分；
   而是显示「完成首场面试」骨架，让 5 维标签仍可见，但不暗示「已有评分」。 */
.wb-radar-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 18px 12px 14px;
  gap: 16px;
}

.wb-radar-empty-svg {
  width: 100%;
  max-width: 280px;
  height: auto;
  /* 整体 muted opacity 让它在视觉上像是「等待数据填充」的占位 */
  opacity: 0.42;
  user-select: none;
}

/* 骨架五边形：比真实雷达更淡的网格线 */
.wb-radar-empty-grid polygon {
  fill: none;
  stroke: rgba(255, 255, 255, 0.06);
  stroke-width: 1;
}

.wb-radar-empty-line {
  stroke: rgba(255, 255, 255, 0.05);
  stroke-width: 1;
}

/* 维度标签保留，但 muted；让用户知道未来雷达会展示哪 5 维 */
.wb-radar-empty-label {
  font: 500 11px var(--sans);
  fill: rgba(255, 255, 255, 0.40);
  letter-spacing: .02em;
  user-select: none;
}

.wb-radar-empty-text {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  text-align: center;
  /* 不要 margin-top：上方 svg gap 已经给了节奏 */
}

.wb-radar-empty-title {
  font: 600 14px var(--display);
  color: var(--t2);
  letter-spacing: .01em;
}

.wb-radar-empty-sub {
  font-size: 12.5px;
  color: var(--t3);
  line-height: 1.55;
}

/* ============ 响应式 ============ */
@media (max-width: 1100px) {
  .wb-quick {
    /* 中屏 2x2 */
    grid-template-columns: repeat(2, 1fr);
  }
  .wb-bottom {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .wb-content {
    padding: 0 20px 60px;
  }
  .wb-quick {
    grid-template-columns: 1fr;
    gap: 12px;
  }
  .wb-stats {
    width: 100%;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 18px;
  }
  .wb-tr {
    grid-template-columns: minmax(0, 2fr) minmax(0, 1fr) 48px 60px;
    gap: 10px;
  }
  /* 移动端隐藏不重要的列：方向 / 时长，保留主题、难度、分、时间 */
  .wb-th:nth-child(2),
  .wb-td:nth-child(2),
  .wb-th:nth-child(5),
  .wb-td:nth-child(5) {
    display: none;
  }
}
</style>
