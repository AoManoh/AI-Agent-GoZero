// Package logic 提供AI面试系统的核心业务逻辑处理层实现
// chatlogic.go 实现基于RAG增强的智能面试对话功能
//
// 主要功能:
//  1. 流式对话处理 - 实现实时的AI面试对话，支持流式响应和异步处理
//  2. RAG知识增强 - 集成向量数据库的知识检索，为AI提供相关背景知识
//  3. 状态驱动交互 - 基于面试状态管理，提供结构化的面试流程控制
//  4. 多轮对话记忆 - 维护完整的对话历史，确保上下文连贯性
//  5. 异步消息存储 - 并行处理用户消息存储和AI响应，提高系统响应速度
//
// 技术架构特性:
//   - 基于GoZero微服务框架，提供高性能的并发处理能力
//   - 集成OpenAI GPT模型，支持流式响应和自然语言理解
//   - 向量数据库存储，实现高效的语义搜索和知识检索
//   - Redis状态管理，支持分布式部署和状态持久化
//   - Channel异步通信，确保流式响应的实时性和稳定性
//
// RAG增强机制:
//
//	本文件实现了检索增强生成(RAG)的完整流程:
//	- 用户输入 -> 知识库检索 -> 历史对话获取 -> 上下文构建 -> AI生成 -> 流式输出
//	- 每次对话都基于最相关的知识片段，确保AI回答的准确性和专业性
//	- 智能内容截断和上下文长度控制，平衡信息完整性和响应性能
//
// 应用场景:
//   - 企业技术面试的自动化和标准化
//   - 在线教育平台的智能问答系统
//   - 知识库驱动的专业咨询服务
//   - 基于RAG的智能客服和助手系统
//
// 性能优化设计:
//   - 异步Goroutine处理避免阻塞用户请求
//   - 并行执行消息存储和知识检索操作
//   - 流式响应减少用户等待时间
//   - 智能上下文截断控制token消耗成本
package logic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	chatAuth "GoZero-AI/api/chat/internal/auth"
	"GoZero-AI/internal/chatflow"

	"github.com/sashabaranov/go-openai"
	"github.com/zeromicro/go-zero/core/logx"

	"GoZero-AI/api/chat/internal/interviewer"
	"GoZero-AI/api/chat/internal/svc"
	types2 "GoZero-AI/api/chat/internal/types"
)

// ChatLogic AI面试对话的核心业务逻辑处理器
// 实现基于RAG增强的智能面试对话功能，支持状态管理和流式响应
//
// 设计架构:
//  1. **依赖注入模式** - 通过ServiceContext获取所有外部依赖服务
//  2. **上下文传递** - 使用context.Context实现请求级别的生命周期管理
//  3. **结构化日志** - 集成GoZero的logx组件，支持上下文关联的日志记录
//  4. **单一职责** - 专注于对话逻辑处理，不承担其他业务职责
//
// 核心能力:
//   - 多轮对话管理: 维护完整的对话上下文和历史记忆
//   - RAG知识增强: 集成向量数据库的语义搜索和知识注入
//   - 状态驱动交互: 支持面试流程的状态进度管理和转换
//   - 异步流式处理: 实现高性能的实时对话响应
//   - 错误容错处理: 全面的异常处理和业务容错机制
//
// 结构字段说明:
//   - Logger: GoZero框架的结构化日志组件，自动包含请求上下文信息
//   - ctx: Go标准库的上下文对象，用于生命周期管理和取消传播
//   - svcCtx: GoZero的服务上下文，包含配置、数据库连接、外部服务客户端等
type ChatLogic struct {
	// GoZero 框架提供的日志记录器，自动包含了请求的上下文信息
	// 我们可以通过l.logger.Error("获取会话数据失败", err)获取日志记录
	logx.Logger

	// Go 语言标准库的上下文对象
	ctx context.Context

	// GoZero 的服务上下文，包含所有依赖的资源
	// 存储配置信息（如 OpenAI 配置）
	// 存储数据库连接、Redis 连接等资源
	// 存储业务逻辑依赖的服务实例
	svcCtx *svc.ServiceContext
}

// NewChatLogic 创建对话逻辑处理器实例
// 遵循GoZero框架的工厂模式，为每个请求创建独立的逻辑处理器
//
// 初始化特性:
//  1. **请求级别隔离** - 每个请求都有独立的实例，避免状态共享和并发冲突
//  2. **上下文绑定** - 将HTTP请求的context传递到业务逻辑层，支持超时和取消
//  3. **日志关联** - 绑定请求上下文到日志系统，便于链路追踪
//  4. **资源注入** - 通过ServiceContext注入所需的外部依赖和配置
//
// 参数说明:
//
//	ctx: HTTP请求的上下文，包含超时、取消信号和请求元数据
//	svcCtx: 服务上下文，包含配置信息和外部服务客户端
//
// 返回值:
//
//	*ChatLogic: 配置完成的对话逻辑处理器实例
func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{
		Logger: logx.WithContext(ctx), // 创建带上下文的日志记录器，支持请求追踪
		ctx:    ctx,                   // 保存请求上下文，用于后续操作的超时控制
		svcCtx: svcCtx,                // 注入服务上下文，提供业务处理所需的依赖
	}
}

// Chat AI面试对话的核心处理方法
// 实现基于RAG增强和状态管理的智能面试对话功能
//
// 技术架构设计:
//  1. **异步流式处理** - 使用Goroutine和Channel实现非阻塞的流式响应
//  2. **RAG增强策略** - 集成知识检索和对话历史，为AI提供丰富上下文
//  3. **状态驱动流程** - 基于面试状态进行动态的交互策略调整
//  4. **错误容错机制** - 全面的异常处理，确保服务稳定性和可用性
//  5. **性能优化设计** - 并行处理消息存储和知识检索，提高响应速度
//
// 业务流程设计:
//
//	步靄1: 用户消息存储 -> 步靄2: 状态获取 -> 步靄3: 知识检索(RAG)
//	步靄4: 上下文构建 -> 步靄5: OpenAI请求 -> 步靄6: 流式响应处理
//	步靄7: AI回复存储 -> 步靄8: 状态更新 -> 步靄9: 响应结束
//
// 异步处理优势:
//   - 用户请求立即返回，不阻塞等待AI计算完成
//   - 支持多用户并发访问，提高系统吞吐量
//   - 流式响应减少用户等待时间，提升体验
//   - 容错设计确保单点失败不影响整体流程
//
// RAG增强机制:
//   - 自动检索与用户问题最相关的知识片段
//   - 动态注入到AI的系统提示词中
//   - 平衡知识丰富度和token消耗成本
//   - 确保AI回答的准确性和专业性
//
// 参数说明:
//
//	req: 面试对话请求，包含会话标识和用户消息
//
// 返回值:
//
//	<-chan *types.ChatRes: 只读的响应数据流，支持实时流式输出
//	error: 初始化阶段的错误，不包含异步处理中的错误
func (l *ChatLogic) Chat(req *types2.InterviewAppChatReq) (<-chan *types2.ChatRes, error) {
	// 1. 创建一个 channel 用于通信
	ch := make(chan *types2.ChatRes)
	var scopedUserID *int64
	if userID, ok := chatAuth.UserIDFromContext(l.ctx); ok {
		scopedUserID = &userID
	}
	effectiveMode, err := l.svcCtx.VectorStore.ResolveSessionMode(l.ctx, req.ChatId, scopedUserID, req.Mode)
	if err != nil {
		return nil, err
	}
	scope := ConversationScope{
		ChatID: req.ChatId,
		UserID: scopedUserID,
		Mode:   effectiveMode,
	}
	if err := l.svcCtx.VectorStore.ValidateSessionWrite(l.ctx, req.ChatId, scopedUserID); err != nil {
		return nil, err
	}

	// 2. 启动一个 goroutine 执行耗时操作，进行异步通信
	go func() {
		defer close(ch)

		stateManager := NewStateManager(l.ctx, l.svcCtx)
		if _, err := stateManager.UpdateExecutionState(scope, chatflow.ExecutionRetrieving, "incoming_request"); err != nil {
			l.Logger.Errorf("更新 flow 执行状态失败: %v", err)
		}

		// 2.1 首先就先将用户的对话内容，添加到向量存储中
		if err := l.svcCtx.VectorStore.SaveMessageWithUser(l.ctx, req.ChatId, openai.ChatMessageRoleUser, req.Message, scopedUserID, effectiveMode); err != nil {
			l.Errorf("保存用户消息失败: %v", err)
			_, _ = stateManager.UpdateExecutionState(scope, chatflow.ExecutionFailed, "user_message_persist_failed")
			ch <- &types2.ChatRes{
				Content:  "系统错误：无法保存会话消息",
				IsLatest: true,
			}
			return
		} else if err := stateManager.RecordTurn(scope, openai.ChatMessageRoleUser, "user_message_persisted"); err != nil {
			l.Logger.Errorf("记录用户 turn 失败: %v", err)
		}

		// 2.2 获取当前状态
		currentState, err := stateManager.GetCurrentState(scope)
		if err != nil {
			l.Logger.Errorf("获取状态失败: %v", err)
			currentState = types2.StateStart
		}

		// 2.2 新增：知识检索（RAG）
		// 从向量数据库中检索与用户消息最相关的知识片段
		knowledgeChunks, err := l.svcCtx.VectorStore.RetrieveKnowledgeScopedContext(l.ctx, req.Message, l.svcCtx.Config.VectorDB.Knowledge.TopK, scopedUserID, req.ChatId)
		if err != nil {
			l.Logger.Errorf("知识检索失败: %v", err)
			knowledgeChunks = []types2.KnowledgeChunk{} // 确保不为nil
		}

		// 2.3 构建消息
		if _, err := stateManager.UpdateExecutionState(scope, chatflow.ExecutionGenerating, "model_stream_start"); err != nil {
			l.Logger.Errorf("更新 flow 执行状态失败: %v", err)
		}
		var sessionConfig *svc.SessionInterviewConfig
		if loadedConfig, found, err := l.svcCtx.VectorStore.LoadSessionInterviewConfig(l.ctx, req.ChatId, scopedUserID); err == nil && found {
			sessionConfig = &loadedConfig
		} else if err != nil {
			l.Logger.Errorf("读取会话配置失败: %v", err)
		}

		openSession, err := l.buildMessagesWithState(req.ChatId, currentState, knowledgeChunks, scopedUserID, sessionConfig)
		if err != nil {
			l.Logger.Errorf("构建消息失败: %v", err)
			_, _ = stateManager.UpdateExecutionState(scope, chatflow.ExecutionFailed, "build_messages_failed")
			ch <- &types2.ChatRes{
				Content:  "系统错误：无法构建消息",
				IsLatest: true,
			}
			return
		}

		// 2.4 创建 OpenAI 的 API 请求，从 config.yaml 配置文件中读取配置
		request := openai.ChatCompletionRequest{
			Model:               l.svcCtx.Config.OpenAI.Model,
			Messages:            openSession,
			Stream:              true,
			MaxCompletionTokens: l.svcCtx.Config.OpenAI.MaxCompletionTokens,
			Temperature:         l.svcCtx.Config.OpenAI.Temperature,
		}

		// 2.5 创建流式响应
		stream, err := l.svcCtx.OpenAIClient.CreateChatCompletionStream(l.ctx, request)
		if err != nil {
			l.Errorf("创建聊天失败: %v", err)
			l.Errorf("请求配置: BaseURL=%s, Model=%s\n", l.svcCtx.Config.OpenAI.BaseURL, l.svcCtx.Config.OpenAI.Model)
			_, _ = stateManager.UpdateExecutionState(scope, chatflow.ExecutionFailed, "stream_start_failed")
			ch <- &types2.ChatRes{
				Content:  "系统错误：无法连接 OpenAI 的 API 请求",
				IsLatest: true,
			}
			return
		}
		defer stream.Close()

		// 2.6 处理流式响应
		var fullResponse strings.Builder
		for {
			select {
			case <-l.ctx.Done(): // 如果客户端断开连接，就取消上下文退出
				reason := "request_canceled"
				if errors.Is(l.ctx.Err(), context.DeadlineExceeded) {
					reason = "request_timeout"
				}
				_, _ = stateManager.UpdateExecutionState(scope, chatflow.ExecutionFailed, reason)
				return
			default:
				res, err := stream.Recv()
				if errors.Is(err, io.EOF) { // 如果流式响应结束，就退出
					// 退出前先保存助手回复
					// 流结束后处理状态更新
					finalRes := fullResponse.String()
					if _, markErr := stateManager.UpdateExecutionState(scope, chatflow.ExecutionPersisting, "assistant_message_persist"); markErr != nil {
						l.Logger.Errorf("更新 flow 执行状态失败: %v", markErr)
					}
					if fullResponse.String() != "" {
						if saveErr := l.svcCtx.VectorStore.SaveMessageWithUser(
							l.ctx,
							req.ChatId,
							openai.ChatMessageRoleAssistant,
							finalRes,
							scopedUserID,
							effectiveMode,
						); saveErr != nil {
							l.Errorf("保存助手消息失败: %v", saveErr)
							_, _ = stateManager.UpdateExecutionState(scope, chatflow.ExecutionFailed, "assistant_message_persist_failed")
							ch <- &types2.ChatRes{IsLatest: true}
							return
						} else if err := stateManager.RecordTurn(scope, openai.ChatMessageRoleAssistant, "assistant_message_persisted"); err != nil {
							l.Logger.Errorf("记录助手 turn 失败: %v", err)
						}
					}

					snapshot, err := stateManager.EvaluateAndUpdateState(scope, finalRes)
					if err != nil {
						l.Logger.Errorf("更新状态失败: %v", err)
					} else {
						l.Logger.Infof("状态更新: %s -> %s", currentState, snapshot.InterviewState)
					}

					ch <- &types2.ChatRes{IsLatest: true} // 流结束，发送结束标记
					return
				}
				if err != nil {
					reason := "stream_receive_failed"
					if errors.Is(err, context.Canceled) {
						reason = "request_canceled"
					} else if errors.Is(err, context.DeadlineExceeded) {
						reason = "request_timeout"
					}
					l.Logger.Errorf("接受流式响应失败: %v", err)
					_, _ = stateManager.UpdateExecutionState(scope, chatflow.ExecutionFailed, reason)
					if reason == "request_canceled" {
						return
					}
					ch <- &types2.ChatRes{
						Content:  "系统错误：无法接受流式响应",
						IsLatest: true,
					}
					return
				}

				// 处理有效响应数据
				if len(res.Choices) > 0 && res.Choices[0].Delta.Content != "" {
					content := res.Choices[0].Delta.Content
					fullResponse.WriteString(content)
					ch <- &types2.ChatRes{
						Content:  content,
						IsLatest: false,
					}
				}
			}
		}
	}()

	return ch, nil
}

// getSessionHistory 构建AI对话的完整上下文历史
// 作为RAG系统的核心方法，负责整合历史对话和检索知识，为AI提供丰富的上下文信息
//
// 功能职责:
//  1. **历史对话检索** - 从向量数据库获取指定会话的最近N条消息记录
//  2. **知识注入处理** - 将RAG检索到的相关知识片段整合到系统提示词中
//  3. **上下文构造** - 按OpenAI API格式构造完整的对话上下文
//  4. **内容长度控制** - 通过配置控制知识片段长度，避免超出token限制
//
// RAG增强机制:
//   - 系统消息 = 基础角色设定 + 动态注入的相关知识
//   - 知识片段按相似度排序，优先展示最相关内容
//   - 自动截断过长内容，平衡信息完整性和响应速度
//   - 保持对话历史的时间顺序，确保上下文逻辑连贯
//
// 技术实现特点:
//   - 支持多轮对话的上下文记忆和延续
//   - 动态知识注入，每次对话都基于最新相关知识
//   - 可配置的历史消息数量限制(当前10条)
//   - 智能内容截断，避免token超限导致的API调用失败
//
// ==============================
// 优化方案分析 (待评估)
// ==============================
//
// 方案一: 异步并行优化 ★★★★☆
// 问题: 当前历史检索和知识处理是串行执行，存在性能瓶颈
// 解决方案: 使用sync.WaitGroup并行执行GetMessage和知识处理
// 与估效果: 性能提升20-30%, 实现复杂度中等, 风险低
//
// 方案二: 智能缓存机制 ★★★★★
// 问题: 相同的历史检索可能重复执行，缺乏缓存机制
// 解决方案: 实现基于Redis的对话历史缓存，TTL 5-10分钟
// 与估效果: 性能提升40-60%, 实现复杂度高, 风险中等
//
// 方案三: 上下文窗口智能优化 ★★★★☆
// 问题: 固定10条消息限制不够灵活，未考虑token消耗优化
// 解决方案: 实现基于token数量的动态截断算法
// 与估效果: 成本优化15-25%, 实现复杂度中等, 风险低
//
// 方案四: 知识注入优先级优化 ★★★★☆
// 问题: 所有知识片段同等对待，未考虑相似度分数差异
// 解决方案: 实现基于相似度分数的加权注入策略
// 与估效果: 质量提升显著, 实现复杂度中等, 风险低
//
// 推荐实施优先级:
// 1. 方案三(上下文优化) - 直接效益，低风险
// 2. 方案四(知识优先级) - 提升质量，无性能损失
// 3. 方案一(异步并行) - 性能提升明显
// 4. 方案二(智能缓存) - 最大性能提升
//
// 参数说明:
//
//	chatId: 会话唯一标识符，用于检索特定会话的历史消息
//	knowledge: RAG系统检索到的相关知识片段，按相似度排序
//
// 返回值:
//
//	[]openai.ChatCompletionMessage: 符合OpenAI API格式的完整对话上下文
//	error: 历史消息检索失败或数据处理异常
func (l *ChatLogic) getSessionHistory(chatId string, knowledge []types2.KnowledgeChunk, userID *int64) ([]openai.ChatCompletionMessage, error) {
	// 步骤1: 从向量数据库检索历史对话记录
	// 限制为最近10条消息，避免上下文过长影响AI响应质量
	// VectorStore.GetMessage会自动按时间排序并返回正确的消息顺序
	vectorMessages, err := l.svcCtx.VectorStore.GetMessageWithUserContext(l.ctx, chatId, userID, 10)
	if err != nil {
		// 历史消息检索失败，可能原因:
		// - 数据库连接异常
		// - chatId不存在或格式错误
		// - 向量数据库查询语法问题
		return nil, fmt.Errorf("检索会话历史失败: %w", err)
	}

	prompt := interviewer.BuildPrompt(interviewer.BuildInput{
		ChatID:            chatId,
		State:             types2.StateQuestion,
		Knowledge:         buildInterviewerKnowledge(knowledge),
		MaxKnowledgeRunes: l.svcCtx.Config.VectorDB.Knowledge.MaxContextLength,
	})

	// 步骤4: 构造OpenAI API所需的消息格式
	// 系统消息必须放在最前面，定义AI的角色和可用知识
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem, // 系统角色，包含AI的行为指令和知识库
			Content: prompt.SystemMessage,         // 完整的系统提示词，包含角色定义和动态知识
		},
	}

	// 步骤5: 添加历史对话消息
	// 将向量数据库中的历史消息转换为OpenAI API格式
	// 保持原有的角色信息(user/assistant)和时间顺序
	for _, msg := range vectorMessages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,    // 保持原始角色(user用户/assistant助手)
			Content: msg.Content, // 保持原始消息内容
		})
	}

	// 返回完整的对话上下文，包含:
	// 1. 系统消息(角色定义 + 动态知识)
	// 2. 历史对话记录(按时间顺序)
	return messages, nil
}

// buildMessagesWithState 构建带状态感知的AI对话上下文
// 基于面试状态动态调整AI的行为策略和系统提示词，实现状态驱动的智能交互
//
// 功能职责:
//  1. **状态感知系统消息构建** - 根据当前面试状态动态生成不同的AI行为指令
//  2. **RAG知识增强集成** - 将检索到的相关知识片段注入到状态特定的系统提示词中
//  3. **历史对话上下文整合** - 获取并整合会话历史，确保状态转换的连贯性
//  4. **OpenAI API格式适配** - 将所有上下文信息转换为OpenAI API标准格式
//
// 状态驱动策略设计:
//   - **StateStart**: 欢迎和开场策略，建立友好的面试氛围
//   - **StateQuestion**: 技术问题策略，重点考察核心概念和基础知识
//   - **StateFollowUp**: 追问策略，深入挖掘候选人的理解深度和实践经验
//   - **StateEvaluate**: 评估策略，全面分析候选人的技术水平和综合能力
//   - **StateEnd**: 结束策略，提供建设性反馈和面试总结
//
// 技术实现特点:
//   - 状态特定的目标设定，让AI明确当前阶段的重点任务
//   - 动态知识注入，确保AI基于最新相关知识进行状态感知的对话
//   - 历史消息保持，维护状态转换过程中的上下文连贯性
//   - 统一的消息格式，简化上层调用复杂度
//
// 与getSessionHistory的区别:
//   - getSessionHistory: 通用的历史对话构建，不感知业务状态
//   - buildMessagesWithState: 业务状态驱动的上下文构建，针对面试场景优化
//
// 参数说明:
//
//	chatId: 会话唯一标识符，用于检索特定会话的历史消息
//	currentState: 当前面试状态，决定AI的行为策略和目标设定
//	knowledge: RAG系统检索到的相关知识片段，将被注入到状态特定的系统提示词中
//
// 返回值:
//
//	[]openai.ChatCompletionMessage: 包含状态感知系统消息和历史对话的完整上下文
//	error: 历史消息检索失败或上下文构建异常
func (l *ChatLogic) buildMessagesWithState(chatId, currentState string, knowledge []types2.KnowledgeChunk, userID *int64, sessionConfig *svc.SessionInterviewConfig) ([]openai.ChatCompletionMessage, error) {
	prompt := interviewer.BuildPrompt(interviewer.BuildInput{
		ChatID:            chatId,
		State:             currentState,
		Session:           buildInterviewerSessionConfig(sessionConfig),
		Knowledge:         buildInterviewerKnowledge(knowledge),
		MaxKnowledgeRunes: l.svcCtx.Config.VectorDB.Knowledge.MaxContextLength,
	})

	// --- 阶段三：组合最终消息列表 ---
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt.SystemMessage,
		},
	}
	// 将历史消息转换为OpenAI API格式并添加到上下文
	// 保持原有的角色信息和时间顺序，确保状态转换的逻辑性
	history, err := l.svcCtx.VectorStore.GetMessageWithUserContext(l.ctx, chatId, userID, 10)
	if err != nil {
		return nil, err
	}
	for _, msg := range history {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return messages, nil
}

func buildInterviewerSessionConfig(config *svc.SessionInterviewConfig) *interviewer.SessionConfig {
	if config == nil {
		return nil
	}
	return &interviewer.SessionConfig{
		DirectionKey:          config.DirectionKey,
		DirectionLabel:        config.DirectionLabel,
		DifficultyLevel:       config.DifficultyLevel,
		DifficultyLabel:       config.DifficultyLabel,
		InterviewerStyle:      config.InterviewerStyle,
		InterviewerStyleLabel: config.InterviewerStyleLabel,
		FocusAreas:            interviewer.ParseFocusAreas(config.FocusAreas),
		FollowUpDepth:         config.FollowUpDepth,
		EstimatedMinutes:      config.EstimatedMinutes,
		ProgressPercent:       config.ProgressPercent,
	}
}

func buildInterviewerKnowledge(chunks []types2.KnowledgeChunk) []interviewer.KnowledgeChunk {
	if len(chunks) == 0 {
		return nil
	}
	knowledge := make([]interviewer.KnowledgeChunk, 0, len(chunks))
	for _, chunk := range chunks {
		knowledge = append(knowledge, interviewer.KnowledgeChunk{
			Title:   chunk.Title,
			Content: chunk.Content,
		})
	}
	return knowledge
}
