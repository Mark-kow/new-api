package setting

import "strings"

var AlipayEnabled = false
var AlipayAppID = ""
var AlipayPrivateKey = ""
var AlipayPublicKey = ""
var AlipayNotifyURL = ""
var AlipayReturnURL = ""

func IsAlipayDirectEnabled() bool {
	if !AlipayEnabled {
		return false
	}
	return strings.TrimSpace(AlipayAppID) != "" &&
		strings.TrimSpace(AlipayPrivateKey) != "" &&
		strings.TrimSpace(AlipayPublicKey) != ""
}
