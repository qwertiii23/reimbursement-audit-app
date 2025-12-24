// rule_request.go 规则管理请求结构体和参数校验
// 功能点：
// 1. 定义规则创建请求结构体
// 2. 定义规则更新请求结构体
// 3. 定义规则查询请求结构体
// 4. 定义规则测试请求结构体
// 5. 实现参数校验规则
// 6. 提供参数绑定和校验方法

package request

// CreateRuleRequest 创建规则请求
type CreateRuleRequest struct {
	Name        string   `json:"name"`        // 规则名称
	Description string   `json:"description"` // 规则描述
	Type        string   `json:"type"`        // 规则类型(金额/频次/发票/合规等)
	Category    string   `json:"category"`    // 规则分类
	Definition  string   `json:"definition"`  // 规则定义(Grule语法)
	Priority    int      `json:"priority"`    // 优先级(数字越大优先级越高)
	Enabled     bool     `json:"enabled"`     // 是否启用
	CreatedBy   string   `json:"created_by"`  // 创建人
	UpdatedBy   string   `json:"updated_by"`  // 更新人
	Version     int      `json:"version"`     // 版本号
	Tags        []string `json:"tags"`        // 标签
}

// UpdateRuleRequest 更新规则请求
type UpdateRuleRequest struct {
	ID          string   `json:"id"`          // 规则ID
	RuleCode    string   `json:"rule_code"`   // 规则编码(唯一)
	Name        string   `json:"name"`        // 规则名称
	Description string   `json:"description"` // 规则描述
	Type        string   `json:"type"`        // 规则类型(金额/频次/发票/合规等)
	Category    string   `json:"category"`    // 规则分类
	Status      string   `json:"status"`      // 规则状态(启用/禁用/草稿)
	Definition  string   `json:"definition"`  // 规则定义(Grule语法)
	Priority    int      `json:"priority"`    // 优先级(数字越大优先级越高)
	Enabled     bool     `json:"enabled"`     // 是否启用
	CreatedBy   string   `json:"created_by"`  // 创建人
	UpdatedBy   string   `json:"updated_by"`  // 更新人
	Version     int      `json:"version"`     // 版本号
	Tags        []string `json:"tags"`        // 标签
}
