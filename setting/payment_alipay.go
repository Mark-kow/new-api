// payment_alipay.go 支付宝官方直连支付配置。
// 管理员在后台配置支付宝应用 ID、商户私钥、支付宝公钥等参数后，
// 用户即可通过支付宝扫码或手机网页完成充值/订阅购买。
package setting

import "strings"

// AlipayEnabled 总开关，控制是否启用支付宝官方支付
var AlipayEnabled = false

// AlipayAppID 支付宝开放平台应用 ID
var AlipayAppID = ""

// AlipayPrivateKey 商户 RSA2 私钥（PKCS8 格式），用于对请求签名
var AlipayPrivateKey = ""

// AlipayPublicKey 支付宝公钥，用于验证支付宝回调签名
var AlipayPublicKey = ""

// AlipayNotifyURL 异步通知回调地址（可选，为空时自动拼接 ServerAddress）
var AlipayNotifyURL = ""

// AlipayReturnURL 同步跳转回调地址（可选，为空时自动拼接 ServerAddress）
var AlipayReturnURL = ""

// IsAlipayDirectEnabled 判断支付宝直连支付是否已正确配置并启用
func IsAlipayDirectEnabled() bool {
	if !AlipayEnabled {
		return false
	}
	return strings.TrimSpace(AlipayAppID) != "" &&
		strings.TrimSpace(AlipayPrivateKey) != "" &&
		strings.TrimSpace(AlipayPublicKey) != ""
}
