// Package types 提供AI面试系统的状态管理类型定义
// redisState.go 定义面试流程中的状态转移常量
//
// 功能说明:
//
//	本文件定义了AI面试系统中用于状态管理的常量值，这些状态用于:
//	1. 控制面试流程的进度和阶段转换
//	2. 在Redis中存储和管理会话状态
//	3. 根据不同状态调整AI的行为和响应策略
//	4. 实现结构化的面试流程控制
//
// 状态流转逻辑:
//
//	start -> question -> follow_up/evaluate -> end
//	状态转换基于AI回复内容的关键词匹配和语义分析
//
// 技术实现:
//   - 配合Redis实现分布式状态存储
//   - 支持状态持久化和恢复
//   - 提供状态过期和自动清理机制
//   - 确保多实例部署时的状态一致性
package types

// Redis状态转移常量定义
// 用于AI面试系统的流程控制和状态管理
const (
	StateStart    = "start"     // 面试开始状态 - 初始化面试环境，发送欢迎信息
	StateQuestion = "question"  // 核心问题阶段 - 提出专业技术问题，评估基础能力
	StateFollowUp = "follow_up" // 追问深入阶段 - 针对回答进行深入追问，考察细节掌握
	StateEvaluate = "evaluate"  // 评估总结阶段 - 对面试表现进行评估和总结
	StateEnd      = "end"       // 面试结束状态 - 完成面试流程，提供最终反馈
)
