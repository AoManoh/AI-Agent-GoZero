// Package logic 提供业务逻辑处理层实现
// knowledge_upload.go 实现RAG知识库上传的核心业务逻辑
//
// 主要功能:
//  1. 文档分块处理 - 将大文档按配置大小智能分割为可处理的小块
//  2. 向量化协调 - 协调VectorStore进行知识块的向量化和存储
//  3. 批量存储管理 - 确保所有知识块都成功存储到向量数据库
//  4. 业务状态反馈 - 向上层返回处理结果和统计信息
//
// 核心价值:
//   - 实现文档到可检索知识的转换，为RAG系统提供知识底座
//   - 处理大文档的分块策略，平衡检索精度和存储效率
//   - 提供事务性的批量存储，确保知识库数据一致性
//   - 统一的业务逻辑封装，便于上层Handler调用
//
// 技术特性:
//   - 配置驱动的分块策略，支持不同场景的优化调整
//   - 错误快速失败机制，确保数据质量
//   - 统计信息收集，便于监控和分析
//   - 上下文感知的日志记录
package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/api/chat/internal/types"
	"GoZero-AI/api/chat/internal/utils"
)

// KnowledgeUploadLogic RAG知识库上传业务逻辑处理器
// 作为Handler层和VectorStore层之间的业务协调者，负责文档分块和存储流程管理
//
// 职责定义:
//  1. **分块策略执行** - 根据配置的MaxChunkSize对文档进行智能分割
//  2. **存储流程控制** - 协调VectorStore完成知识块的向量化和持久化
//  3. **错误处理机制** - 确保存储失败时能够快速响应和回滚
//  4. **统计信息管理** - 收集和返回处理过程中的关键指标
//
// 设计模式:
//   - 采用GoZero框架的Logic层模式，实现业务逻辑与数据访问的分离
//   - 依赖注入模式，通过ServiceContext获取所需的服务组件
//   - 失败快速返回模式，确保任何环节失败都能及时中断并反馈
//
// **后续**扩展能力:
//   - 支持异步批量处理大文档上传
//   - 添加重试机制处理临时性存储失败
//   - 实现分块去重避免重复知识存储
//   - 集成进度回调支持前端实时进度展示
type KnowledgeUploadLogic struct {
	svcCtx      *svc.ServiceContext // 服务上下文，提供配置信息和依赖组件访问
	logx.Logger                     // GoZero日志组件，支持结构化日志和上下文传递
	ctx         context.Context     // 请求上下文，用于超时控制和取消传播
}

// NewKnowledgeUploadLogic 创建知识库上传逻辑处理器实例
// 遵循GoZero框架的工厂模式，为每个请求创建独立的逻辑处理器
//
// 设计理念:
//  1. **请求隔离** - 每个HTTP请求都有独立的Logic实例，避免并发冲突
//  2. **上下文传递** - 将HTTP请求的context传递到业务逻辑层，支持超时和取消
//  3. **依赖注入** - 通过ServiceContext注入所需的服务组件和配置
//  4. **日志关联** - 绑定请求上下文到日志系统，便于链路追踪
//
// **后续**调用时机:
//   - 在KnowledgeUploadHandler中为每个上传请求创建
//   - 实例生命周期与单次请求绑定，请求结束后自动回收
//   - 支持请求级别的配置覆盖和状态管理
//
// 参数说明:
//
//	ctx: HTTP请求的上下文，包含超时、取消信号和请求元数据
//	svcCtx: 服务上下文，包含数据库连接、外部API客户端和配置信息
//
// 返回值:
//
//	*KnowledgeUploadLogic: 配置完成的业务逻辑处理器实例
func NewKnowledgeUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeUploadLogic {
	return &KnowledgeUploadLogic{
		Logger: logx.WithContext(ctx), // 创建带上下文的日志记录器，支持请求追踪
		ctx:    ctx,                   // 保存请求上下文，用于后续操作的超时控制
		svcCtx: svcCtx,                // 注入服务上下文，提供业务处理所需的依赖
	}
}

// KnowledgeUpload 执行知识库文档上传的核心业务逻辑
// 将Handler层提取的PDF文本内容转换为可检索的向量化知识块
//
// 业务流程设计:
//  1. **文档分块** - 使用配置化的分块策略将长文档切分为最优大小的片段
//  2. **批量存储** - 循环处理每个知识块，确保全部成功存储到向量数据库
//  3. **失败处理** - 任何环节失败都立即停止并返回错误，保证数据一致性
//  4. **结果统计** - 返回成功处理的知识块数量，便于用户了解处理效果
//
// 分块策略说明:
//   - 使用utils.SplitText按MaxChunkSize配置进行智能分块
//   - 分块大小影响检索精度：块过大可能稀释关键信息，块过小可能丢失上下文
//   - 当前采用固定大小分块，**后续**可优化为语义感知分块
//
// 存储机制特点:
//   - 每个知识块独立向量化，提高检索的粒度和准确性
//   - 保留文档标题信息，便于检索结果的溯源和展示
//   - 采用同步存储确保一致性，**后续**可优化为异步批量处理
//
// **后续**优化方向:
//   - 实现文档去重检测，避免重复知识入库
//   - 添加分块质量评估，过滤无效或低质量片段
//   - 支持增量更新，允许文档内容的局部修改
//   - 集成分块预览，让用户确认分块效果
//
// 参数说明:
//
//	req: 知识上传请求，包含文档标题和从PDF提取的文本内容
//
// 返回值:
//
//	*types.KnowledgeUploadRes: 上传结果，包含成功消息和处理的知识块数量
//	error: 处理过程中的任何错误，包括分块失败、向量化失败、存储失败等
func (l *KnowledgeUploadLogic) KnowledgeUpload(req *types.KnowledgeUploadReq) (*types.KnowledgeUploadRes, error) {
	// 步骤1: 执行文档分块策略
	// 根据VectorDB.Knowledge.MaxChunkSize配置将长文档分割为合适大小的片段
	// 分块大小直接影响后续的向量化效果和检索精度
	chunks := utils.SplitText(req.Content, l.svcCtx.Config.VectorDB.Knowledge.MaxChunkSize)

	// 调试日志：标记分块处理开始
	logx.Infof("准备分块！！：\n")

	// 步骤2: 批量存储知识块到向量数据库
	// 遍历每个分块，逐一进行向量化和存储处理
	// 采用原子操作，要么全部成功，要么全部失败。禁止上传一半失败，数据库里留下僵尸数据的情况
	if err := l.svcCtx.VectorStore.SaveKnowledgeBatch(req.Title, chunks); err != nil {
		logx.Errorf("保存知识失败: %v", err)
		return nil, err
	}

	// 调试日志：标记存储流程完成)
	logx.Infof("分块保存结束！！：\n")

	// 步骤3: 构造并返回成功响应
	// 向上层返回处理结果，包含用户友好的消息和统计信息
	return &types.KnowledgeUploadRes{
		Msg:    "知识上传成功",    // 用户友好的成功提示信息
		Chunks: len(chunks), // 实际处理的知识块数量，便于用户了解处理效果
	}, nil
}
