package model

import "time"

// RuleType 规则类型
type RuleType string

const (
	RuleTypeDomain    RuleType = "DOMAIN"
	RuleTypeDomainSuffix RuleType = "DOMAIN-SUFFIX"
	RuleTypeDomainKeyword RuleType = "DOMAIN-KEYWORD"
	RuleTypeIP        RuleType = "IP-CIDR"
	RuleTypeSrcIP     RuleType = "SRC-IP-CIDR"
	RuleTypeGeoIP     RuleType = "GEOIP"
	RuleTypeGeoSite   RuleType = "GEOSITE"
	RuleTypeProcess   RuleType = "PROCESS-NAME"
	RuleTypeFinal     RuleType = "FINAL"
)

// RulePolicy 规则策略
type RulePolicy string

const (
	PolicyProxy    RulePolicy = "PROXY"
	PolicyDirect   RulePolicy = "DIRECT"
	PolicyReject   RulePolicy = "REJECT"
	PolicyBlock    RulePolicy = "BLOCK"
)

// Rule 代理规则模型
type Rule struct {
	ID          string     `json:"id"`
	Type        RuleType   `json:"type"`
	Value       string     `json:"value"`
	Policy      RulePolicy `json:"policy"`
	Description string     `json:"description"`
	Enabled     bool       `json:"enabled"`
	Priority    int        `json:"priority"` // 优先级，数字越大优先级越高
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// RuleCreateRequest 创建规则请求
type RuleCreateRequest struct {
	Type        RuleType   `json:"type" binding:"required"`
	Value       string     `json:"value" binding:"required"`
	Policy      RulePolicy `json:"policy" binding:"required"`
	Description string     `json:"description"`
	Enabled     bool       `json:"enabled"`
	Priority    int        `json:"priority"`
}

// RuleUpdateRequest 更新规则请求
type RuleUpdateRequest struct {
	Type        RuleType   `json:"type"`
	Value       string     `json:"value"`
	Policy      RulePolicy `json:"policy"`
	Description string     `json:"description"`
	Enabled     bool       `json:"enabled"`
	Priority    int        `json:"priority"`
}
