package interviewer

import (
	"hash/fnv"
	"strings"
)

type Style struct {
	Key    string
	Label  string
	Prompt string
}

var styles = []Style{
	{
		Key:    "structured",
		Label:  "结构化面试官",
		Prompt: "问题清晰、节奏稳定，按“核心概念 -> 项目实践 -> 边界风险”推进；少铺垫，不跳题。",
	},
	{
		Key:    "senior",
		Label:  "资深技术官",
		Prompt: "像资深工程负责人一样推进面试，关注事实、边界、取舍和工程落地；追问稳定克制，不替候选人作答。",
	},
	{
		Key:    "pressure",
		Label:  "压力型面试官",
		Prompt: "节奏稍快，会挑战模糊表述，要求候选人给出证据、取舍和边界；保持尊重，禁止羞辱、讽刺或人身攻击。",
	},
	{
		Key:    "conversational",
		Label:  "对话式面试官",
		Prompt: "像资深同事做技术讨论，表达自然温和；可以短暂回应候选人的思路，但不要替候选人回答。",
	},
	{
		Key:    "humorous",
		Label:  "诙谐型面试官",
		Prompt: "可用少量轻松表达缓和氛围，但不讲长段子，不油腻，不削弱技术判断；每轮仍以一个有效问题收束。",
	},
	{
		Key:    "mentor",
		Label:  "导师型面试官",
		Prompt: "候选人卡住时给一个很小的提示，再追问推理过程；重点观察学习能力、表达修正能力和问题拆解能力。",
	},
	{
		Key:    "system_design",
		Label:  "架构深挖面试官",
		Prompt: "关注容量估算、瓶颈定位、降级策略、观测性、一致性取舍和演进路径；问题要贴近线上系统。",
	},
	{
		Key:    "behavioral",
		Label:  "行为面试官",
		Prompt: "偏 STAR 追问，围绕经历、冲突、决策、复盘、协作和责任心展开；少问纯概念定义。",
	},
}

var styleAliases = map[string]string{
	"engineer":      "senior",
	"engineering":   "senior",
	"工程师型":          "senior",
	"coach":         "mentor",
	"coaching":      "mentor",
	"guide":         "mentor",
	"guiding":       "mentor",
	"引导型":           "mentor",
	"deep_dive":     "system_design",
	"deepdive":      "system_design",
	"深挖型":           "system_design",
	"system-design": "system_design",
	"systemdesign":  "system_design",
}

func SelectStyle(chatID string) Style {
	trimmed := strings.TrimSpace(chatID)
	if trimmed == "" || len(styles) == 0 {
		return styles[0]
	}

	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(trimmed))
	return styles[int(hasher.Sum32())%len(styles)]
}

func SelectStyleByKey(key, chatID string) Style {
	normalized := normalizeKey(key)
	if normalized == "" {
		return SelectStyle(chatID)
	}
	if alias, ok := styleAliases[normalized]; ok {
		normalized = alias
	}
	for _, style := range styles {
		if style.Key == normalized {
			return style
		}
	}
	return SelectStyle(chatID)
}

func BuildStylePrompt(style Style) string {
	return "# 本轮面试官风格\n" +
		"- 风格: " + style.Label + "\n" +
		"- 执行方式: " + style.Prompt + "\n" +
		"- 边界: 风格只改变语气、追问节奏和观察角度，不改变专业、尊重、克制、一次只问一个问题的底线。"
}
