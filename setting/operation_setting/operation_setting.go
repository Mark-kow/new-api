package operation_setting

import "strings"

var DemoSiteEnabled = false
var SelfUseModeEnabled = false

// AutomaticDisableKeywords 是自动禁用渠道的默认关键词列表（大小写不敏感）。
// 除涉及 HTTP 状态码的情况外，以下错误消息关键词用于补充覆盖无法单纯由状态码检测的情况：
// - Arrearage：华为云等指费用相关错误
// - billing_not_active / account_deactivated：账户计费/禁用状态
// - insufficient_quota / quota 相关：额度耗尽
var AutomaticDisableKeywords = []string{
	"Your credit balance is too low",
	"This organization has been disabled.",
	"You exceeded your current quota",
	"Permission denied",
	"The security token included in the request is invalid",
	"Operation not allowed",
	"Your account is not authorized",
	// 以下关键词弥补移除硬编码 error code/type 后的视野缺口
	"Arrearage",
	"billing_not_active",
	"account_deactivated",
	"insufficient_quota",
	"insufficient_user_quota",
}

func AutomaticDisableKeywordsToString() string {
	return strings.Join(AutomaticDisableKeywords, "\n")
}

func AutomaticDisableKeywordsFromString(s string) {
	AutomaticDisableKeywords = []string{}
	ak := strings.Split(s, "\n")
	for _, k := range ak {
		k = strings.TrimSpace(k)
		k = strings.ToLower(k)
		if k != "" {
			AutomaticDisableKeywords = append(AutomaticDisableKeywords, k)
		}
	}
}
