package user

import (
	"net/http"
	"strings"

	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"
)

var interviewDirections = []types.InterviewDirectionPreset{
	{
		Key:             "go_backend",
		Label:           "Go 后端",
		Description:     "围绕 Go 语言、并发、数据库、微服务和工程实践进行追问。",
		QuestionCount:   256,
		DifficultyRange: []int64{1, 5},
		FocusKeys:       []string{"concurrency", "database", "system_design", "engineering"},
	},
	{
		Key:             "java_backend",
		Label:           "Java 后端",
		Description:     "覆盖 JVM、Spring、并发、数据库和分布式系统设计。",
		QuestionCount:   224,
		DifficultyRange: []int64{1, 5},
		FocusKeys:       []string{"system_design", "database", "network", "engineering"},
	},
	{
		Key:             "frontend_vue",
		Label:           "前端 Vue",
		Description:     "关注组件设计、工程化、性能优化和浏览器基础。",
		QuestionCount:   180,
		DifficultyRange: []int64{1, 5},
		FocusKeys:       []string{"frontend_arch", "performance", "engineering"},
	},
	{
		Key:             "system_design",
		Label:           "系统设计",
		Description:     "围绕高并发、缓存、消息队列、一致性和可观测性展开。",
		QuestionCount:   168,
		DifficultyRange: []int64{2, 5},
		FocusKeys:       []string{"system_design", "database", "network", "observability"},
	},
	{
		Key:             "algorithm",
		Label:           "算法基础",
		Description:     "覆盖数据结构、复杂度分析、编码表达和边界用例。",
		QuestionCount:   210,
		DifficultyRange: []int64{1, 5},
		FocusKeys:       []string{"algorithm", "communication"},
	},
}

var interviewDifficulties = []types.InterviewDifficultyPreset{
	{Level: 1, Label: "入门", Description: "确认基本概念和术语理解。", DefaultFollowUpDepth: "N+1"},
	{Level: 2, Label: "初级", Description: "检查常见场景和基础工程经验。", DefaultFollowUpDepth: "N+2"},
	{Level: 3, Label: "中级", Description: "要求解释机制、取舍和常见故障。", DefaultFollowUpDepth: "N+3"},
	{Level: 4, Label: "资深", Description: "进入深度追问，关注架构判断和工程边界。", DefaultFollowUpDepth: "N+5"},
	{Level: 5, Label: "专家", Description: "模拟高压技术面，要求系统化论证。", DefaultFollowUpDepth: "N+7"},
}

var interviewFocusOptions = []types.InterviewFocusOption{
	{Key: "concurrency", Label: "并发与调度", Description: "goroutine、线程模型、锁、调度与资源竞争。"},
	{Key: "database", Label: "数据库", Description: "事务、索引、SQL 优化、连接池和一致性。"},
	{Key: "system_design", Label: "系统设计", Description: "容量、缓存、限流、降级和可扩展性。"},
	{Key: "engineering", Label: "工程实践", Description: "框架、部署、监控、测试和故障处理。"},
	{Key: "network", Label: "网络协议", Description: "HTTP、RPC、连接复用、超时和重试。"},
	{Key: "performance", Label: "性能优化", Description: "定位瓶颈、压测、内存和延迟优化。"},
	{Key: "algorithm", Label: "算法基础", Description: "复杂度、数据结构和代码边界。"},
	{Key: "communication", Label: "表达沟通", Description: "结构化表达、权衡说明和追问承接。"},
	{Key: "frontend_arch", Label: "前端架构", Description: "组件拆分、状态管理、构建和可维护性。"},
	{Key: "observability", Label: "可观测性", Description: "日志、指标、链路追踪和告警设计。"},
}

var interviewStyles = []types.InterviewStyleOption{
	{Key: "senior", Label: "资深技术官", Description: "稳重、追问准确，强调工程边界和系统性。"},
	{Key: "pressure", Label: "压力型", Description: "追问更紧，要求快速补充漏洞和细节。"},
	{Key: "humorous", Label: "诙谐型", Description: "氛围较轻松，但仍保留关键技术追问。"},
	{Key: "mentor", Label: "导师型", Description: "偏引导和复盘，适合练习表达。"},
}

func defaultSessionConfigSnapshot() types.SessionConfigSnapshot {
	return types.SessionConfigSnapshot{
		DirectionKey:          "go_backend",
		DirectionLabel:        "Go 后端",
		DifficultyLevel:       3,
		DifficultyLabel:       "中级",
		InterviewerStyle:      "senior",
		InterviewerStyleLabel: "资深技术官",
		FocusAreas:            defaultFocusAreasForDirection("go_backend"),
		FollowUpDepth:         "N+3",
		EstimatedMinutes:      30,
		ProgressPercent:       0,
		DurationSeconds:       0,
	}
}

func buildSessionCreateConfig(req *types.CreateSessionReq) (model.SessionCreateConfig, types.SessionConfigSnapshot, error) {
	config := defaultSessionConfigSnapshot()

	if directionKey := strings.TrimSpace(req.DirectionKey); directionKey != "" {
		direction, ok := findDirectionPreset(directionKey)
		if !ok {
			return model.SessionCreateConfig{}, types.SessionConfigSnapshot{}, statuserr.New(http.StatusBadRequest, "不支持的面试方向")
		}
		config.DirectionKey = direction.Key
		config.DirectionLabel = direction.Label
		config.FocusAreas = defaultFocusAreasForDirection(direction.Key)
	}

	if req.Difficulty != 0 {
		difficulty, ok := findDifficultyPreset(req.Difficulty)
		if !ok {
			return model.SessionCreateConfig{}, types.SessionConfigSnapshot{}, statuserr.New(http.StatusBadRequest, "面试难度必须在 1 到 5 之间")
		}
		config.DifficultyLevel = difficulty.Level
		config.DifficultyLabel = difficulty.Label
		config.FollowUpDepth = difficulty.DefaultFollowUpDepth
	}

	if styleKey := strings.TrimSpace(req.InterviewerStyle); styleKey != "" {
		style, ok := findStyleOption(styleKey)
		if !ok {
			return model.SessionCreateConfig{}, types.SessionConfigSnapshot{}, statuserr.New(http.StatusBadRequest, "不支持的面试官风格")
		}
		config.InterviewerStyle = style.Key
		config.InterviewerStyleLabel = style.Label
	}

	if req.EstimatedMinutes > 0 {
		if req.EstimatedMinutes < 5 || req.EstimatedMinutes > 180 {
			return model.SessionCreateConfig{}, types.SessionConfigSnapshot{}, statuserr.New(http.StatusBadRequest, "预计面试时长必须在 5 到 180 分钟之间")
		}
		config.EstimatedMinutes = req.EstimatedMinutes
	}

	if len(req.FocusKeys) > 0 {
		focusAreas, err := buildFocusAreaSelections(req.FocusKeys)
		if err != nil {
			return model.SessionCreateConfig{}, types.SessionConfigSnapshot{}, err
		}
		config.FocusAreas = focusAreas
	}

	return model.SessionCreateConfig{
		DirectionKey:          config.DirectionKey,
		DirectionLabel:        config.DirectionLabel,
		DifficultyLevel:       config.DifficultyLevel,
		DifficultyLabel:       config.DifficultyLabel,
		InterviewerStyle:      config.InterviewerStyle,
		InterviewerStyleLabel: config.InterviewerStyleLabel,
		FocusAreas:            config.FocusAreas,
		FollowUpDepth:         config.FollowUpDepth,
		EstimatedMinutes:      config.EstimatedMinutes,
	}, config, nil
}

func buildFocusAreaSelections(keys []string) ([]types.FocusAreaSelection, error) {
	seen := make(map[string]struct{}, len(keys))
	focusAreas := make([]types.FocusAreaSelection, 0, len(keys))
	for _, rawKey := range keys {
		key := strings.TrimSpace(rawKey)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		option, ok := findFocusOption(key)
		if !ok {
			return nil, statuserr.New(http.StatusBadRequest, "不支持的考察侧重")
		}
		seen[key] = struct{}{}
		focusAreas = append(focusAreas, types.FocusAreaSelection{
			Key:   option.Key,
			Label: option.Label,
		})
	}
	if len(focusAreas) == 0 {
		return nil, statuserr.New(http.StatusBadRequest, "考察侧重不能为空")
	}
	return focusAreas, nil
}

func defaultFocusAreasForDirection(directionKey string) []types.FocusAreaSelection {
	direction, ok := findDirectionPreset(directionKey)
	if !ok {
		direction = interviewDirections[0]
	}
	keys := direction.FocusKeys
	if len(keys) > 3 {
		keys = keys[:3]
	}
	focusAreas, err := buildFocusAreaSelections(keys)
	if err != nil {
		return []types.FocusAreaSelection{
			{Key: "concurrency", Label: "并发与调度"},
			{Key: "database", Label: "数据库"},
			{Key: "system_design", Label: "系统设计"},
		}
	}
	return focusAreas
}

func findDirectionPreset(key string) (types.InterviewDirectionPreset, bool) {
	for _, item := range interviewDirections {
		if item.Key == key {
			return item, true
		}
	}
	return types.InterviewDirectionPreset{}, false
}

func findDifficultyPreset(level int64) (types.InterviewDifficultyPreset, bool) {
	for _, item := range interviewDifficulties {
		if item.Level == level {
			return item, true
		}
	}
	return types.InterviewDifficultyPreset{}, false
}

func findFocusOption(key string) (types.InterviewFocusOption, bool) {
	for _, item := range interviewFocusOptions {
		if item.Key == key {
			return item, true
		}
	}
	return types.InterviewFocusOption{}, false
}

func findStyleOption(key string) (types.InterviewStyleOption, bool) {
	for _, item := range interviewStyles {
		if item.Key == key {
			return item, true
		}
	}
	return types.InterviewStyleOption{}, false
}
