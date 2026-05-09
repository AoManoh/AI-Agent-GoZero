package types

// KnowledgeUploadInput 是 handler 提取 PDF 文本后的内部 DTO，
// 不代表外部 multipart/form-data 契约。
type KnowledgeUploadInput struct {
	Title   string
	Content string
	Source  string
	UserID  int64
}

// KnowledgeUploadOutput 是 logic 返回给 handler 的内部结果结构体。
type KnowledgeUploadOutput struct {
	Msg    string `json:"msg"`
	Chunks int    `json:"chunks"`
}
