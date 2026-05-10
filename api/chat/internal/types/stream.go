package types

// ChatRes 是 chat SSE 链路内部使用的流式分片结构体，
// 不代表最终 HTTP JSON 响应体。
type ChatRes struct {
	Content  string `json:"content"`
	IsLatest bool   `json:"isLatest"`
	Event    string `json:"event,optional"`
}
