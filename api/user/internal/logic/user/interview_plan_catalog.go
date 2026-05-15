package user

import (
	"net/http"
	"strings"

	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/statuserr"
)

const (
	defaultInterviewPlanLimit = int64(8)
	maxInterviewPlanLimit     = int64(20)
)

type interviewPlanQuestionSeed struct {
	Key                  string
	DirectionKey         string
	FocusKey             string
	Title                string
	Prompt               string
	MinDifficulty        int64
	MaxDifficulty        int64
	ExpectedSignals      []string
	FollowUps            []string
	EvaluationDimensions []string
}

var interviewPlanQuestionCatalog = []interviewPlanQuestionSeed{
	{
		Key:           "go-concurrency-gmp",
		DirectionKey:  "go_backend",
		FocusKey:      "concurrency",
		Title:         "GMP 调度与阻塞处理",
		Prompt:        "讲讲 Go GMP 调度模型。一个 goroutine 发生系统调用或长时间阻塞时，P、M、G 会怎样变化？",
		MinDifficulty: 2,
		MaxDifficulty: 5,
		ExpectedSignals: []string{
			"能区分 G、M、P 的职责",
			"能说明 sysmon、work stealing 或 netpoll 的作用",
			"能联系线上阻塞、CPU 飙高或 goroutine 泄漏排查",
		},
		FollowUps: []string{
			"如果线上 goroutine 数持续上涨，你会先看哪些指标？",
			"什么时候应该用 worker pool，而不是无限制启动 goroutine？",
		},
		EvaluationDimensions: []string{"technical_depth", "engineering_practice", "communication"},
	},
	{
		Key:           "go-concurrency-cancel",
		DirectionKey:  "go_backend",
		FocusKey:      "concurrency",
		Title:         "Context 取消链路",
		Prompt:        "你在项目里如何保证请求取消、超时和下游 RPC 调用可以正确传递？",
		MinDifficulty: 2,
		MaxDifficulty: 5,
		ExpectedSignals: []string{
			"能说明 context 超时、取消和 value 的边界",
			"能解释下游 goroutine、DB、RPC 如何感知取消",
			"能指出不应把 context 存进结构体长期持有",
		},
		FollowUps: []string{
			"如果某个 goroutine 没有监听 ctx.Done，会造成什么问题？",
			"你如何在日志里定位是哪一层触发了超时？",
		},
		EvaluationDimensions: []string{"engineering_practice", "technical_depth", "communication"},
	},
	{
		Key:           "go-concurrency-lock",
		DirectionKey:  "go_backend",
		FocusKey:      "concurrency",
		Title:         "锁、原子操作与 channel 取舍",
		Prompt:        "同一个共享计数器场景下，你会如何在 mutex、atomic 和 channel 之间做选择？",
		MinDifficulty: 1,
		MaxDifficulty: 5,
		ExpectedSignals: []string{
			"能区分互斥、无锁计数和消息传递的适用场景",
			"能说明可读性、临界区大小和竞争检测",
			"能提到 race detector 或压测验证",
		},
		FollowUps: []string{
			"如果临界区里包含 IO，会发生什么风险？",
			"atomic 更快，为什么不总是用 atomic？",
		},
		EvaluationDimensions: []string{"technical_depth", "engineering_practice"},
	},
	{
		Key:           "go-db-index-transaction",
		DirectionKey:  "go_backend",
		FocusKey:      "database",
		Title:         "索引、事务与 SQL 优化",
		Prompt:        "一个分页查询在数据量上涨后变慢，你会如何定位索引、SQL 和事务隔离级别的问题？",
		MinDifficulty: 2,
		MaxDifficulty: 5,
		ExpectedSignals: []string{
			"能说明 explain、慢查询和索引选择",
			"能区分 offset 分页和游标分页",
			"能联系事务隔离、锁等待和连接池",
		},
		FollowUps: []string{
			"如果索引命中了但仍然慢，你会继续看什么？",
			"深分页为什么会拖慢查询？",
		},
		EvaluationDimensions: []string{"technical_depth", "engineering_practice", "architecture_sense"},
	},
	{
		Key:           "go-db-pgvector-rag",
		DirectionKey:  "go_backend",
		FocusKey:      "database",
		Title:         "pgvector 与 RAG 检索",
		Prompt:        "如果面试系统需要从简历和知识库中召回上下文，你会如何设计 pgvector 的写入、检索和过滤条件？",
		MinDifficulty: 3,
		MaxDifficulty: 5,
		ExpectedSignals: []string{
			"能说明 embedding 维度、向量距离和 topK",
			"能结合 user_id、doc_type、session_id 做权限和范围过滤",
			"能提到召回质量、重建索引和降级策略",
		},
		FollowUps: []string{
			"如果 embedding 模型升级导致维度变化，怎么迁移？",
			"如何避免把别人的私有简历召回到当前会话？",
		},
		EvaluationDimensions: []string{"architecture_sense", "technical_depth", "engineering_practice"},
	},
	{
		Key:           "go-db-pool",
		DirectionKey:  "go_backend",
		FocusKey:      "database",
		Title:         "连接池与背压",
		Prompt:        "高并发下数据库连接池打满，你会如何定位根因并设计背压或降级？",
		MinDifficulty: 3,
		MaxDifficulty: 5,
		ExpectedSignals: []string{
			"能区分连接池过小、慢 SQL 和事务未释放",
			"能说明超时、限流、熔断和排队策略",
			"能提出指标和告警口径",
		},
		FollowUps: []string{
			"连接池调大一定能解决问题吗？",
			"如何避免请求堆积把上游也拖垮？",
		},
		EvaluationDimensions: []string{"engineering_practice", "architecture_sense"},
	},
	{
		Key:           "go-system-gozero-etcd",
		DirectionKey:  "go_backend",
		FocusKey:      "system_design",
		Title:         "GoZero、ETCD 与服务发现",
		Prompt:        "你如何解释 go-zero 微服务里 API、RPC、ETCD、Redis 和 PostgreSQL 的职责边界？",
		MinDifficulty: 2,
		MaxDifficulty: 5,
		ExpectedSignals: []string{
			"能说清 API 服务、RPC 服务和注册发现的边界",
			"能说明 Redis 与 PostgreSQL 的不同职责",
			"能联系配置、超时、重试和部署拓扑",
		},
		FollowUps: []string{
			"ETCD 不可用时，已有连接和新实例发现会怎样？",
			"哪些数据适合放 Redis，哪些必须落 PostgreSQL？",
		},
		EvaluationDimensions: []string{"architecture_sense", "technical_depth", "communication"},
	},
	{
		Key:           "go-system-cache-breakdown",
		DirectionKey:  "go_backend",
		FocusKey:      "system_design",
		Title:         "缓存击穿与一致性",
		Prompt:        "热点 key 过期导致大量请求打到数据库，你会如何处理缓存击穿、穿透和雪崩？",
		MinDifficulty: 2,
		MaxDifficulty: 5,
		ExpectedSignals: []string{
			"能区分击穿、穿透、雪崩",
			"能说明互斥锁、singleflight、随机 TTL 和空值缓存",
			"能讨论一致性和失效策略",
		},
		FollowUps: []string{
			"如果缓存里的数据必须强一致，你会怎么设计？",
			"singleflight 在多实例下有什么限制？",
		},
		EvaluationDimensions: []string{"technical_depth", "architecture_sense"},
	},
	{
		Key:           "go-engineering-incident",
		DirectionKey:  "go_backend",
		FocusKey:      "engineering",
		Title:         "线上故障复盘",
		Prompt:        "讲一个你处理线上故障的过程。你是如何确认影响面、止损、定位和复盘的？",
		MinDifficulty: 1,
		MaxDifficulty: 5,
		ExpectedSignals: []string{
			"能按发现、止损、定位、修复、复盘组织回答",
			"能给出指标、日志或链路追踪证据",
			"能说明长期改进，而不是只描述修 bug",
		},
		FollowUps: []string{
			"当时有没有做临时降级？如何评估副作用？",
			"这个故障之后，你们补了哪些监控或测试？",
		},
		EvaluationDimensions: []string{"engineering_practice", "communication"},
	},
	{
		Key:           "go-engineering-test",
		DirectionKey:  "go_backend",
		FocusKey:      "engineering",
		Title:         "测试分层与回归保护",
		Prompt:        "你会如何给一个包含数据库、Redis 和外部模型调用的接口设计测试？",
		MinDifficulty: 2,
		MaxDifficulty: 5,
		ExpectedSignals: []string{
			"能区分单元测试、集成测试和 E2E",
			"能说明 mock、testcontainer 或本地依赖的取舍",
			"能覆盖失败路径和幂等性",
		},
		FollowUps: []string{
			"哪些测试必须进 CI，哪些可以放到手工闭测？",
			"模型调用不稳定时怎么让测试可靠？",
		},
		EvaluationDimensions: []string{"engineering_practice", "architecture_sense"},
	},
	{
		Key:                  "java-jvm-gc",
		DirectionKey:         "java_backend",
		FocusKey:             "engineering",
		Title:                "JVM 内存与 GC",
		Prompt:               "线上 Java 服务出现长时间 STW，你会如何定位 GC、堆内存和对象分配问题？",
		MinDifficulty:        2,
		MaxDifficulty:        5,
		ExpectedSignals:      []string{"能解释 GC 日志", "能联系对象分配和堆外内存", "能提出压测和告警指标"},
		FollowUps:            []string{"如何判断是内存泄漏还是流量尖峰？", "你会如何选择 GC 参数？"},
		EvaluationDimensions: []string{"technical_depth", "engineering_practice"},
	},
	{
		Key:                  "java-spring-transaction",
		DirectionKey:         "java_backend",
		FocusKey:             "database",
		Title:                "Spring 事务边界",
		Prompt:               "Spring 事务失效常见原因有哪些？你如何设计事务边界避免锁持有过久？",
		MinDifficulty:        2,
		MaxDifficulty:        5,
		ExpectedSignals:      []string{"能说明代理调用和传播行为", "能识别长事务风险", "能联系幂等和补偿"},
		FollowUps:            []string{"事务里调用外部 RPC 有什么问题？", "如何处理部分成功？"},
		EvaluationDimensions: []string{"technical_depth", "architecture_sense"},
	},
	{
		Key:                  "frontend-vue-state",
		DirectionKey:         "frontend_vue",
		FocusKey:             "frontend_arch",
		Title:                "Vue 状态与组件边界",
		Prompt:               "一个复杂工作台页面里，哪些状态应该放组件内，哪些应该放到 store 或 URL？",
		MinDifficulty:        1,
		MaxDifficulty:        5,
		ExpectedSignals:      []string{"能区分局部状态和跨页面状态", "能说明 URL 可恢复性", "能关注组件职责"},
		FollowUps:            []string{"筛选条件放 URL 有什么好处？", "如何避免 store 变成大杂烩？"},
		EvaluationDimensions: []string{"architecture_sense", "engineering_practice", "communication"},
	},
	{
		Key:                  "frontend-performance",
		DirectionKey:         "frontend_vue",
		FocusKey:             "performance",
		Title:                "前端性能定位",
		Prompt:               "页面首屏慢，你会从网络、构建产物、渲染和数据接口哪些方面排查？",
		MinDifficulty:        2,
		MaxDifficulty:        5,
		ExpectedSignals:      []string{"能区分加载和渲染瓶颈", "能使用性能面板和打包分析", "能提出懒加载和缓存策略"},
		FollowUps:            []string{"如何判断瓶颈在接口而不是渲染？", "大列表渲染怎么优化？"},
		EvaluationDimensions: []string{"engineering_practice", "technical_depth"},
	},
	{
		Key:                  "system-capacity",
		DirectionKey:         "system_design",
		FocusKey:             "system_design",
		Title:                "容量估算与限流",
		Prompt:               "如果一个面试系统需要支撑晚高峰并发面试，你会如何做容量估算和限流设计？",
		MinDifficulty:        2,
		MaxDifficulty:        5,
		ExpectedSignals:      []string{"能估算 QPS、并发和资源消耗", "能设计限流、排队和降级", "能考虑模型调用成本"},
		FollowUps:            []string{"SSE 长连接会如何影响容量？", "模型供应商限流时怎么兜底？"},
		EvaluationDimensions: []string{"architecture_sense", "engineering_practice"},
	},
	{
		Key:                  "system-observability",
		DirectionKey:         "system_design",
		FocusKey:             "observability",
		Title:                "可观测性设计",
		Prompt:               "你会为一次 AI 面试链路设计哪些日志、指标和链路追踪字段？",
		MinDifficulty:        2,
		MaxDifficulty:        5,
		ExpectedSignals:      []string{"能覆盖请求、检索、模型、评估和数据库", "能注意隐私脱敏", "能给出告警指标"},
		FollowUps:            []string{"如何定位模型慢还是数据库慢？", "日志里哪些内容不能直接打印？"},
		EvaluationDimensions: []string{"engineering_practice", "architecture_sense"},
	},
	{
		Key:                  "algorithm-lru",
		DirectionKey:         "algorithm",
		FocusKey:             "algorithm",
		Title:                "LRU 设计",
		Prompt:               "请设计一个 LRU Cache，并说明 get、put 的复杂度和边界条件。",
		MinDifficulty:        1,
		MaxDifficulty:        4,
		ExpectedSignals:      []string{"能组合哈希表和双向链表", "能处理容量为 0 或 key 覆盖", "能说明复杂度"},
		FollowUps:            []string{"如何让它并发安全？", "如果要加 TTL，会影响哪些逻辑？"},
		EvaluationDimensions: []string{"technical_depth", "communication"},
	},
	{
		Key:                  "algorithm-topk",
		DirectionKey:         "algorithm",
		FocusKey:             "algorithm",
		Title:                "TopK 与复杂度",
		Prompt:               "海量日志中找出现次数最多的 K 个关键词，你会如何设计？",
		MinDifficulty:        2,
		MaxDifficulty:        5,
		ExpectedSignals:      []string{"能说明哈希计数和堆", "能讨论内存不足时的分片或近似算法", "能分析复杂度"},
		FollowUps:            []string{"如果数据无法一次放入内存怎么办？", "K 很大时堆方案还合适吗？"},
		EvaluationDimensions: []string{"technical_depth", "architecture_sense"},
	},
}

func buildInterviewPlanPreviewConfig(req *types.InterviewPlanPreviewReq) (types.SessionConfigSnapshot, error) {
	createReq := types.CreateSessionReq{
		DirectionKey:     strings.TrimSpace(req.DirectionKey),
		Difficulty:       req.Difficulty,
		FocusKeys:        parseInterviewPlanFocusKeys(req.FocusKeys),
		InterviewerStyle: strings.TrimSpace(req.InterviewerStyle),
	}
	_, config, err := buildSessionCreateConfig(&createReq)
	return config, err
}

func parseInterviewPlanFocusKeys(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == ' ' || r == '\t' || r == '\n'
	})
	keys := make([]string, 0, len(fields))
	for _, field := range fields {
		key := strings.TrimSpace(field)
		if key != "" {
			keys = append(keys, key)
		}
	}
	return keys
}

func buildInterviewPlanResp(config types.SessionConfigSnapshot, rawLimit int64) types.InterviewPlanResp {
	questions := selectInterviewPlanQuestions(config, normalizeInterviewPlanLimit(rawLimit))
	return types.InterviewPlanResp{
		Config:     config,
		Questions:  questions,
		Milestones: buildInterviewPlanMilestones(int64(len(questions))),
		PlanMeta: types.ReportMeta{
			SchemaVersion: "interview-plan-v1",
			Available:     len(questions) > 0,
		},
	}
}

func normalizeInterviewPlanLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultInterviewPlanLimit
	}
	if limit > maxInterviewPlanLimit {
		return maxInterviewPlanLimit
	}
	return limit
}

func selectInterviewPlanQuestions(config types.SessionConfigSnapshot, limit int64) []types.InterviewPlanQuestion {
	focusSet := make(map[string]struct{}, len(config.FocusAreas))
	for _, area := range config.FocusAreas {
		if area.Key != "" {
			focusSet[area.Key] = struct{}{}
		}
	}

	selected := make([]types.InterviewPlanQuestion, 0, limit)
	seen := make(map[string]struct{}, len(interviewPlanQuestionCatalog))
	passes := []struct {
		direction  bool
		focus      bool
		difficulty bool
	}{
		{direction: true, focus: true, difficulty: true},
		{direction: true, focus: true},
		{direction: true},
		{focus: true},
		{},
	}

	for _, pass := range passes {
		for _, seed := range interviewPlanQuestionCatalog {
			if len(selected) >= int(limit) {
				return selected
			}
			if _, ok := seen[seed.Key]; ok {
				continue
			}
			if pass.direction && seed.DirectionKey != config.DirectionKey {
				continue
			}
			if pass.focus {
				if _, ok := focusSet[seed.FocusKey]; !ok {
					continue
				}
			}
			if pass.difficulty && !seed.matchesDifficulty(config.DifficultyLevel) {
				continue
			}

			seen[seed.Key] = struct{}{}
			selected = append(selected, seed.toResponse(config))
		}
	}

	return selected
}

func (q interviewPlanQuestionSeed) matchesDifficulty(level int64) bool {
	if level == 0 {
		level = 3
	}
	return level >= q.MinDifficulty && level <= q.MaxDifficulty
}

func (q interviewPlanQuestionSeed) toResponse(config types.SessionConfigSnapshot) types.InterviewPlanQuestion {
	focusLabel := q.FocusKey
	if option, ok := findFocusOption(q.FocusKey); ok {
		focusLabel = option.Label
	}

	return types.InterviewPlanQuestion{
		Key:                  q.Key,
		DirectionKey:         q.DirectionKey,
		FocusKey:             q.FocusKey,
		FocusLabel:           focusLabel,
		DifficultyLevel:      config.DifficultyLevel,
		DifficultyLabel:      config.DifficultyLabel,
		Title:                q.Title,
		Prompt:               q.Prompt,
		ExpectedSignals:      append([]string(nil), q.ExpectedSignals...),
		FollowUps:            append([]string(nil), q.FollowUps...),
		EvaluationDimensions: append([]string(nil), q.EvaluationDimensions...),
	}
}

func buildInterviewPlanMilestones(questionCount int64) []types.InterviewPlanMilestone {
	if questionCount <= 0 {
		return []types.InterviewPlanMilestone{}
	}

	milestones := []types.InterviewPlanMilestone{
		{
			Key:             "opening",
			Label:           "开场校准",
			Description:     "确认候选人的项目背景、表达方式和当前方向。",
			QuestionIndexes: []int64{1},
		},
	}

	if questionCount > 2 {
		milestones = append(milestones, types.InterviewPlanMilestone{
			Key:             "deep_dive",
			Label:           "核心深挖",
			Description:     "围绕方向、难度和侧重点持续追问关键技术细节。",
			QuestionIndexes: buildQuestionIndexRange(2, questionCount-1),
		})
	}

	if questionCount > 1 {
		milestones = append(milestones, types.InterviewPlanMilestone{
			Key:             "wrap_up",
			Label:           "复盘收束",
			Description:     "用最后一题观察候选人的总结、权衡和补充能力。",
			QuestionIndexes: []int64{questionCount},
		})
	}

	return milestones
}

func buildQuestionIndexRange(start, end int64) []int64 {
	if start > end {
		return []int64{}
	}
	values := make([]int64, 0, end-start+1)
	for i := start; i <= end; i++ {
		values = append(values, i)
	}
	return values
}

func validateInterviewPlanLimit(limit int64) error {
	if limit < 0 {
		return statuserr.New(http.StatusBadRequest, "计划题目数量不能为负数")
	}
	return nil
}
