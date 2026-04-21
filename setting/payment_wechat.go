// payment_wechat.go 微信支付官方直连支付配置。
// 管理员在后台配置微信支付商户号、API v3 密钥、商户私钥等参数后，
// 用户即可通过微信扫码（Native）或手机浏览器（H5）完成充值/订阅购买。
package setting

import "strings"

// WechatPayEnabled 总开关，控制是否启用微信支付
var WechatPayEnabled = false

// WechatPayAppID 微信开放平台/公众号 AppID
var WechatPayAppID = ""

// WechatPayMchID 微信支付商户号
var WechatPayMchID = ""

// WechatPaySerialNo 商户 API 证书序列号
var WechatPaySerialNo = ""

// WechatPayAPIv3Key API v3 密钥（32 字节），用于解密回调通知中的 AEAD_AES_256_GCM 密文
var WechatPayAPIv3Key = ""

// WechatPayPrivateKey 商户 API 私钥（PKCS8 格式），用于对请求签名
var WechatPayPrivateKey = ""

// WechatPayPlatformCert 微信支付平台证书（PEM 格式），用于验证回调签名
var WechatPayPlatformCert = ""

// WechatPayNotifyURL 异步通知回调地址（可选，为空时自动拼接 ServerAddress）
var WechatPayNotifyURL = ""

// WechatPayH5Domain H5 支付场景的域名（可选），用于 scene_info.app_url
var WechatPayH5Domain = ""

// IsWechatPayDirectEnabled 判断微信支付直连是否已正确配置并启用
func IsWechatPayDirectEnabled() bool {
	if !WechatPayEnabled {
		return false
	}
	return strings.TrimSpace(WechatPayAppID) != "" &&
		strings.TrimSpace(WechatPayMchID) != "" &&
		strings.TrimSpace(WechatPaySerialNo) != "" &&
		strings.TrimSpace(WechatPayAPIv3Key) != "" &&
		strings.TrimSpace(WechatPayPrivateKey) != "" &&
		strings.TrimSpace(WechatPayPlatformCert) != ""
}
