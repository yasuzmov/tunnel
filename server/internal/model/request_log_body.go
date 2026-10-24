package model

// RequestLogBody 保存截断后的请求/响应体。
// 文本类最多保留配置的字节数并标记 truncated；二进制只记类型与大小，body 为空。
type RequestLogBody struct {
	ID                   uint   `json:"id" gorm:"primaryKey"`
	RequestLogID         uint   `json:"request_log_id" gorm:"uniqueIndex;not null"`
	RequestBody          []byte `json:"request_body" gorm:"type:mediumblob"`
	RequestTruncated     bool   `json:"request_truncated"`
	RequestContentType   string `json:"request_content_type" gorm:"size:256"`
	ResponseBody         []byte `json:"response_body" gorm:"type:mediumblob"`
	ResponseTruncated    bool   `json:"response_truncated"`
	ResponseContentType  string `json:"response_content_type" gorm:"size:256"`
}

// TableName 指定表名。
func (RequestLogBody) TableName() string {
	return "request_log_bodies"
}
