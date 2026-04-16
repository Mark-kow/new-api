package setting

import "strings"

var WechatPayEnabled = false
var WechatPayAppID = ""
var WechatPayMchID = ""
var WechatPaySerialNo = ""
var WechatPayAPIv3Key = ""
var WechatPayPrivateKey = ""
var WechatPayPlatformCert = ""
var WechatPayNotifyURL = ""
var WechatPayH5Domain = ""

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
