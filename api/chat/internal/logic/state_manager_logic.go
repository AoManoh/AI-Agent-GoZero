// Package logic 提供AI面试系统的业务逻辑处理层实现
// state_manager_logic.go 实现面试流程的状态管理和转换逻辑
//
// 主要功能:
//  1. 面试状态生命周期管理 - 处理从开始到结束的完整面试流程状态
//  2. 智能状态转换逻辑 - 基于AI回复内容的关键词分析进行状态自动转换
//  3. Redis状态持久化 - 利用Redis实现分布式状态存储和过期管理
//  4. 状态异常恢复机制 - 处理状态获取失败时的默认状态初始化
//
// 技术特性:
//   - 基于Redis的分布式状态存储，支持多实例部署
//   - 智能关键词匹配算法，实现自然语言到状态的映射
//   - 状态TTL管理，自动清理过期的面试会话
//   - 原子性状态操作，确保状态转换的一致性
//
// 状态转换流程:
//
//	start -> question -> follow_up/evaluate -> end
//	每个状态都有明确的触发条件和转换规则
//
// 应用场景:
//   - AI面试官系统的核心状态控制
//   - 智能对话系统的流程管理
//   - 多轮对话中的上下文状态维护
//   - 面试质量评估和流程优化
package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/api/chat/internal/types"
)

// 状态管理相关常量定义
const (
	stateKeyPrefix = "chat_state:"  // Redis中状态数据的键前缀，用于命名空间隔离
	stateTTL       = 24 * time.Hour // 状态数据的过期时间，24小时后自动清理
)

// StateManager 面试状态管理器
// 负责管理AI面试系统中各个会话的状态转换和持久化
type StateManager struct {
	svcCtx *svc.ServiceContext // 服务上下文，提供Redis连接和配置信息访问
}

// NewStateManager 创建状态管理器实例
// 遵循GoZero框架的工厂模式，为每个请求创建状态管理器
func NewStateManager(svcCtx *svc.ServiceContext) *StateManager {
	return &StateManager{
		svcCtx: svcCtx, // 注入服务上下文，提供对Redis等外部服务的访问
	}
}

// GetCurrentState 获取指定会话的当前状态
// 实现状态的懒加载和默认初始化机制
func (sm *StateManager) GetCurrentState(chatId string) (string, error) {
	// 构造Redis中的状态存储键，使用前缀避免键冲突
	key := stateKeyPrefix + chatId

	// 步骤1: 尝试从Redis获取已存在的状态
	state, err := sm.svcCtx.RedisClient.Get(context.Background(), key).Result()
	if err == nil {
		// 状态存在，直接返回
		return state, nil
	}

	// 步骤2: 处理Redis中状态不存在的情况
	if errors.Is(err, redis.Nil) {
		// 状态不存在，需要初始化为默认状态
		if err := sm.svcCtx.RedisClient.Set(
			context.Background(),
			key,              // 状态存储键
			types.StateStart, // 初始状态值
			stateTTL,         // 过期时间，24小时后自动清理
		).Err(); err != nil {
			// 状态初始化失败，但仍返回默认状态保证系统可用
			return types.StateStart, fmt.Errorf("设置初始状态失败: %v", err)
		}
		// 初始化成功，返回初始状态
		return types.StateStart, nil
	}

	// 步骤3: 处理其他Redis错误（如连接失败等）
	// 返回默认状态，确保系统在网络异常时仍能正常运行
	return types.StateStart, fmt.Errorf("获取状态失败: %w", err)
}

// SetState 设置指定会话的状态值
// 实现状态的持久化存储和TTL管理
func (sm *StateManager) SetState(chatId, state string) error {
	// 构造Redis中的状态存储键，保持与获取方法一致
	key := stateKeyPrefix + chatId

	// 执行Redis SET操作，同时设置过期时间
	if err := sm.svcCtx.RedisClient.Set(
		context.Background(), // 使用背景上下文，无超时限制
		key,                  // 状态存储键
		state,                // 目标状态值
		stateTTL,             // 过期时间，24小时后自动清理
	).Err(); err != nil {
		// 状态设置失败，返回详细错误信息
		return fmt.Errorf("设置状态失败: %v", err)
	}
	// 状态设置成功
	return nil
}

// containsAny 判断字符串是否包含任意一个子字符串
// 用于状态转换逻辑中的关键词匹配和语义分析
func containsAny(s string, subStrings []string) bool {
	// 遍历所有候选子字符串
	for _, sub := range subStrings {
		// 使用strings.Contains进行子字符串匹配
		if strings.Contains(s, sub) {
			// 找到匹配，立即返回true（早期返回）
			return true
		}
	}
	// 所有子字符串都不匹配
	return false
}

// TransitionState 执行状态转换逻辑的核心方法
// 基于AI回复内容的关键词分析，决定是否需要进行状态转换
func (sm *StateManager) TransitionState(currentState, aiRes string) string {
	// 将AI回复转为小写，实现大小写不敏感的匹配
	lowerRes := strings.ToLower(aiRes)

	// 状态转换机：根据当前状态和关键词决定下一个状态
	switch currentState {
	case types.StateStart:
		// 开始状态 -> 问题状态: 检测欢迎词和面试开始信号
		if containsAny(lowerRes, []string{"你好", "欢迎", "面试开始"}) {
			return types.StateQuestion
		}

	case types.StateQuestion:
		// 问题状态 -> 追问状态: 检测追问和深入探究的信号
		if containsAny(lowerRes, []string{"追问", "详细说明", "为什么", "怎么实现"}) {
			return types.StateFollowUp
		}
		// 问题状态 -> 评估状态: 检测评估和总结的信号
		if containsAny(lowerRes, []string{"评估", "总结", "表现", "优缺点"}) {
			return types.StateEvaluate
		}

	case types.StateFollowUp:
		// 追问状态 -> 评估状态: 检测评估和总结的信号
		if containsAny(lowerRes, []string{"评估", "总结", "表现", "优缺点"}) {
			return types.StateEvaluate
		}
		// 追问状态 -> 问题状态: 检测继续提问的信号
		if containsAny(lowerRes, []string{"下一个问题", "新问题"}) {
			return types.StateQuestion
		}

	case types.StateEvaluate:
		// 评估状态 -> 结束状态: 检测面试结束的信号
		if containsAny(lowerRes, []string{"结束", "再见", "感谢参加"}) {
			return types.StateEnd
		}
		// 评估状态 -> 问题状态: 检测继续面试的信号
		if containsAny(lowerRes, []string{"继续", "下一个问题"}) {
			return types.StateQuestion
		}

	case types.StateEnd:
		// 结束状态保持不变，不再接受状态转换
	}

	// 默认情况: 没有匹配到任何转换规则，保持当前状态
	return currentState
}

// EvaluateAndUpdateState 统一的状态评估和更新接口
// 封装了状态获取、转换判断和持久化的完整流程
func (sm *StateManager) EvaluateAndUpdateState(chatId, aiResponse string) (string, error) {
	// 步骤1: 获取当前状态，包含默认初始化逻辑
	currentState, err := sm.GetCurrentState(chatId)
	if err != nil {
		// 状态获取失败但仍继续执行，使用当前值作为备选
		return currentState, err
	}

	// 步骤2: 根据AI回复内容计算新状态
	newState := sm.TransitionState(currentState, aiResponse)

	// 步骤3: 仅在状态发生变化时才执行持久化操作
	if newState != currentState {
		// 状态已变化，需要持久化到Redis
		if err := sm.SetState(chatId, newState); err != nil {
			// 状态设置失败，返回原状态保证数据一致性
			return currentState, err
		}
	}

	// 步骤4: 返回最终状态（可能是新状态或原状态）
	return newState, nil
}
