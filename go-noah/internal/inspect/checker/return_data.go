package checker

// ReturnData 返回数据（用于规则引擎兼容）
type ReturnData struct {
	Summary      []string `json:"summary"`       // 摘要
	Level        string   `json:"level"`         // 级别,INFO/WARN/ERROR/PASS
	AffectedRows int      `json:"affected_rows"` // 影响行数
	Type         string   `json:"type"`          // SQL类型
	FingerId     string   `json:"finger_id"`     // 指纹
	Query        string   `json:"query"`         // 原始SQL
}

// ToAuditResult 转换为 AuditResult
func (r *ReturnData) ToAuditResult() *AuditResult {
	result := &AuditResult{
		SQL:      r.Query,
		Type:     r.Type,
		Level:    AuditLevel(r.Level), // 直接转换，不进行任何映射（与老服务保持一致）
		Messages: make([]string, 0),
		Summary:  r.Summary,
	}
	
	// 将 Summary 同步到 Messages
	if len(r.Summary) > 0 {
		result.Messages = append(result.Messages, r.Summary...)
	} else {
		result.Messages = append(result.Messages, "审核通过")
	}
	
	result.AffectedRows = int64(r.AffectedRows)
	return result
}

