package types

// KnowledgeUploadReq 新增：知识上传请求结构体
// 定义知识库上传接口的请求参数格式
type KnowledgeUploadReq struct {
	Title   string `form:"title"`   // 知识的标题
	Content string `form:"content"` // 知识的内容
}

// KnowledgeUploadRes 新增：知识上传响应结构体
type KnowledgeUploadRes struct {
	Msg    string `json:"msg"` // 返回给前端的提示信息
	Chunks int    `json:"chunks"` // 用来提示前端你的文件已经被切分为多少块
}
