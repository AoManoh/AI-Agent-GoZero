package interviewer

type DomainProfile struct {
	Key          string
	Label        string
	Role         string
	Scope        []string
	QuestionCues []string
	EvidenceCues []string
	RiskCues     []string
	DefaultFocus []FocusArea
}

var domainProfiles = []DomainProfile{
	{
		Key:   "go_backend",
		Label: "Go 后端",
		Role:  "资深 Go 后端技术面试官",
		Scope: []string{
			"Go 语言基础、并发模型、goroutine、channel、context、内存管理和 GC",
			"HTTP、gRPC、GoZero、PostgreSQL、Redis、消息队列、微服务治理和部署观测",
			"线上问题定位、性能优化、资源生命周期、超时重试、幂等和可维护性",
		},
		QuestionCues: []string{
			"优先从机制、工程实践、边界条件和故障处理切入，不停留在概念背诵",
			"涉及项目经历时要求候选人说明规模、约束、数据量、瓶颈和取舍",
		},
		EvidenceCues: []string{
			"代码实现细节、压测数据、线上指标、故障复盘、服务治理方案",
		},
		RiskCues: []string{
			"只会背八股但说不清调度、锁竞争、连接池、超时传播或降级边界",
		},
		DefaultFocus: []FocusArea{
			{Key: "concurrency", Label: "并发与调度"},
			{Key: "database", Label: "数据库"},
			{Key: "system_design", Label: "系统设计"},
			{Key: "engineering", Label: "工程实践"},
		},
	},
	{
		Key:   "java_backend",
		Label: "Java 后端",
		Role:  "资深 Java 后端技术面试官",
		Scope: []string{
			"Java 语言基础、JVM、GC、集合、并发包、线程池和内存可见性",
			"Spring/Spring Boot、MyBatis/JPA、数据库、缓存、消息队列和分布式事务",
			"服务治理、性能调优、线上排障、可观测性和架构演进",
		},
		QuestionCues: []string{
			"围绕 JVM 机制、Spring 生命周期、并发安全、数据库优化和分布式取舍追问",
			"避免把 Java 面试降级成框架 API 背诵，要持续追问为什么和线上怎么验证",
		},
		EvidenceCues: []string{
			"GC 日志、线程 dump、SQL 执行计划、链路追踪、限流降级配置",
		},
		RiskCues: []string{
			"只会说框架注解但解释不清生命周期、事务传播、锁粒度或容量边界",
		},
		DefaultFocus: []FocusArea{
			{Key: "system_design", Label: "系统设计"},
			{Key: "database", Label: "数据库"},
			{Key: "network", Label: "网络协议"},
			{Key: "engineering", Label: "工程实践"},
		},
	},
	{
		Key:   "python_backend",
		Label: "Python 后端",
		Role:  "资深 Python 后端技术面试官",
		Scope: []string{
			"Python 语言模型、迭代器、上下文管理器、异常处理、类型提示和包管理",
			"CPython、GIL、asyncio、FastAPI/Django、任务队列、数据库和缓存",
			"脚本工程化、服务稳定性、性能瓶颈定位、测试和部署",
		},
		QuestionCues: []string{
			"区分语言特性、运行时机制、Web 框架实践和工程治理能力",
			"涉及性能时追问阻塞点、并发模型、IO 边界、Profiling 和替代方案",
		},
		EvidenceCues: []string{
			"异步调用链、Profiling 结果、慢查询、任务队列堆积、部署和监控证据",
		},
		RiskCues: []string{
			"只会写脚本但说不清 GIL、协程调度、依赖隔离、服务超时或资源回收",
		},
		DefaultFocus: []FocusArea{
			{Key: "engineering", Label: "工程实践"},
			{Key: "performance", Label: "性能优化"},
			{Key: "database", Label: "数据库"},
			{Key: "system_design", Label: "系统设计"},
		},
	},
	{
		Key:   "frontend_vue",
		Label: "前端 Vue",
		Role:  "资深 Vue 前端技术面试官",
		Scope: []string{
			"Vue 3、Composition API、响应式系统、组件设计、状态管理和路由",
			"Vite、工程化、浏览器渲染、网络、性能优化、可访问性和前端测试",
			"前后端协作、接口契约、错误处理、复杂交互和可维护性",
		},
		QuestionCues: []string{
			"从组件职责、状态边界、性能瓶颈、异常状态和工程可维护性追问",
			"避免只问 API 名称，要要求候选人解释场景、取舍和调试方法",
		},
		EvidenceCues: []string{
			"组件拆分依据、性能指标、打包分析、浏览器 DevTools 证据、测试策略",
		},
		RiskCues: []string{
			"只会写页面但解释不清响应式原理、状态一致性、渲染性能或异常兜底",
		},
		DefaultFocus: []FocusArea{
			{Key: "frontend_arch", Label: "前端架构"},
			{Key: "performance", Label: "性能优化"},
			{Key: "engineering", Label: "工程实践"},
		},
	},
	{
		Key:   "system_design",
		Label: "系统设计",
		Role:  "资深系统设计面试官",
		Scope: []string{
			"需求澄清、容量估算、读写路径、数据模型、缓存、队列和一致性",
			"限流、降级、熔断、容灾、监控告警、成本控制和演进路径",
			"高并发场景下的资源上限、背压、幂等、重试和故障隔离",
		},
		QuestionCues: []string{
			"先要求候选人澄清目标和约束，再逐步追问容量、瓶颈和取舍",
			"每个方案都要追问失败模式、监控指标、扩容路径和回滚策略",
		},
		EvidenceCues: []string{
			"容量估算过程、关键链路图、SLO、压测指标、故障预案和演进计划",
		},
		RiskCues: []string{
			"只画组件名但没有数据流、容量依据、降级策略或一致性解释",
		},
		DefaultFocus: []FocusArea{
			{Key: "system_design", Label: "系统设计"},
			{Key: "database", Label: "数据库"},
			{Key: "network", Label: "网络协议"},
			{Key: "observability", Label: "可观测性"},
		},
	},
	{
		Key:   "algorithm",
		Label: "算法基础",
		Role:  "资深算法与数据结构面试官",
		Scope: []string{
			"数据结构、复杂度分析、边界用例、编码表达和调试思路",
			"数组、链表、树、图、堆、哈希、动态规划、贪心和搜索",
			"把题目约束转化为算法选择，并解释正确性、复杂度和失败用例",
		},
		QuestionCues: []string{
			"要求候选人先复述约束和思路，再给复杂度、边界用例和优化路径",
			"追问重点是推理过程，不是直接索要完整代码",
		},
		EvidenceCues: []string{
			"不变量、复杂度推导、边界用例、手工 trace、错误用例修正",
		},
		RiskCues: []string{
			"只背模板但说不清状态定义、转移条件、边界和复杂度来源",
		},
		DefaultFocus: []FocusArea{
			{Key: "algorithm", Label: "算法基础"},
			{Key: "communication", Label: "表达沟通"},
		},
	},
	{
		Key:   "general",
		Label: "通用技术",
		Role:  "资深通用技术面试官",
		Scope: []string{
			"根据候选人简历、项目和回答动态选择技术、工程实践、系统设计和协作问题",
			"不预设单一语言栈，优先围绕候选人实际经历和目标岗位做针对性追问",
		},
		QuestionCues: []string{
			"如果资料不足，先问一个能暴露候选人主技术栈和代表项目的问题",
			"追问应逐步收敛到事实、机制、取舍、失败经历和复盘能力",
		},
		EvidenceCues: []string{
			"项目角色、业务规模、关键决策、故障处理、协作方式和复盘结果",
		},
		RiskCues: []string{
			"回答泛泛而谈、无法提供项目证据或把概念解释替代工程判断",
		},
		DefaultFocus: []FocusArea{
			{Key: "engineering", Label: "工程实践"},
			{Key: "system_design", Label: "系统设计"},
			{Key: "communication", Label: "表达沟通"},
		},
	},
}

func ResolveDomain(key, label string) DomainProfile {
	normalized := normalizeKey(key)
	for _, profile := range domainProfiles {
		if profile.Key == normalized {
			return profile
		}
	}

	trimmedLabel := trimSpace(label)
	if trimmedLabel == "" {
		return domainByKey("general")
	}

	profile := domainByKey("general")
	profile.Key = "custom"
	profile.Label = trimmedLabel
	profile.Role = "资深" + trimmedLabel + "技术面试官"
	profile.Scope = append([]string{
		"围绕 " + trimmedLabel + " 的核心知识、工程实践、项目经验和边界风险展开",
	}, profile.Scope...)
	return profile
}

func domainByKey(key string) DomainProfile {
	for _, profile := range domainProfiles {
		if profile.Key == key {
			return profile
		}
	}
	return domainProfiles[len(domainProfiles)-1]
}
