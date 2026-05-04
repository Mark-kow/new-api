package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

const (
	PaymentMethodAlipayDirect = "alipay_direct"
	PaymentMethodWechatDirect = "wechat_direct"

	// wechatNotifyMaxBodySize 微信支付回调请求体最大限制（1MB），防止恶意超大请求导致 DoS
	wechatNotifyMaxBodySize = 1 << 20
)

// ErrOrderNotFound 订单不存在时的统一错误
var ErrOrderNotFound = errors.New("payment order not found")

type DirectTopUpPayRequest struct {
	Amount      int64  `json:"amount"`
	ClientScene string `json:"client_scene"`
}

type DirectSubscriptionPayRequest struct {
	PlanId      int    `json:"plan_id"`
	ClientScene string `json:"client_scene"`
}

func normalizeStoredTopUpAmount(amount int64) int64 {
	if operation_setting.GetQuotaDisplayType() != operation_setting.QuotaDisplayTypeTokens {
		return amount
	}
	dAmount := decimal.NewFromInt(amount)
	dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
	return dAmount.Div(dQuotaPerUnit).IntPart()
}

func buildDirectPaymentResponse(result *service.DirectPaymentResponse) gin.H {
	if result == nil {
		return gin.H{}
	}
	return gin.H{
		"provider":          result.Provider,
		"trade_no":          result.TradeNo,
		"action":            result.Action,
		"qr_code_url":       result.QRCodeURL,
		"redirect_url":      result.RedirectURL,
		"expires_at":        result.ExpiresAt,
		"provider_trade_no": result.ProviderTradeNo,
	}
}

func createTopUpDirectOrder(c *gin.Context, paymentMethod string, provider string) {
	var req DirectTopUpPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if req.Amount < getMinTopup() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "充值数量不能小于 " + strconv.FormatInt(getMinTopup(), 10)})
		return
	}
	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	payMoney := getPayMoney(req.Amount, group)
	if payMoney < 0.01 {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}

	tradeNo := "DIRUSR" + strconv.Itoa(id) + "NO" + common.GetRandomString(6) + strconv.FormatInt(time.Now().Unix(), 10)
	topUp := &model.TopUp{
		UserId:        id,
		Amount:        normalizeStoredTopUpAmount(req.Amount),
		Money:         payMoney,
		TradeNo:       tradeNo,
		PaymentMethod: paymentMethod,
		CreateTime:    time.Now().Unix(),
		Status:        common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	directReq := &service.DirectPaymentRequest{
		TradeNo:     tradeNo,
		Subject:     "账户充值",
		Description: "new-api 账户充值",
		Money:       payMoney,
		ClientScene: req.ClientScene,
		ClientIP:    c.ClientIP(),
		ReturnURL:   strings.TrimRight(system_setting.ServerAddress, "/") + "/console/topup?show_history=true",
	}

	var result *service.DirectPaymentResponse
	switch provider {
	case "alipay":
		result, err = (&service.AlipayService{}).CreatePayment(directReq)
	case "wechat":
		result, err = (&service.WechatPayService{}).CreatePayment(directReq)
	default:
		err = fmt.Errorf("unsupported provider")
	}
	if err != nil {
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "data": buildDirectPaymentResponse(result)})
}

func createSubscriptionDirectOrder(c *gin.Context, paymentMethod string, provider string) {
	var req DirectSubscriptionPayRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	plan, err := model.GetSubscriptionPlanById(req.PlanId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !plan.Enabled {
		common.ApiErrorMsg(c, "套餐未启用")
		return
	}
	if plan.PriceAmount < 0.01 {
		common.ApiErrorMsg(c, "套餐金额过低")
		return
	}
	userId := c.GetInt("id")
	if plan.MaxPurchasePerUser > 0 {
		count, err := model.CountUserSubscriptionsByPlan(userId, plan.Id)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		if count >= int64(plan.MaxPurchasePerUser) {
			common.ApiErrorMsg(c, "已达到该套餐购买上限")
			return
		}
	}

	tradeNo := "DIRSUBUSR" + strconv.Itoa(userId) + "NO" + common.GetRandomString(6) + strconv.FormatInt(time.Now().Unix(), 10)
	order := &model.SubscriptionOrder{
		UserId:        userId,
		PlanId:        plan.Id,
		Money:         plan.PriceAmount,
		TradeNo:       tradeNo,
		PaymentMethod: paymentMethod,
		CreateTime:    time.Now().Unix(),
		Status:        common.TopUpStatusPending,
	}
	if err := order.Insert(); err != nil {
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}

	directReq := &service.DirectPaymentRequest{
		TradeNo:     tradeNo,
		Subject:     "订阅套餐: " + plan.Title,
		Description: "new-api 订阅套餐购买",
		Money:       plan.PriceAmount,
		ClientScene: req.ClientScene,
		ClientIP:    c.ClientIP(),
		ReturnURL:   strings.TrimRight(system_setting.ServerAddress, "/") + "/console/topup?show_history=true",
	}

	var result *service.DirectPaymentResponse
	switch provider {
	case "alipay":
		result, err = (&service.AlipayService{}).CreatePayment(directReq)
	case "wechat":
		result, err = (&service.WechatPayService{}).CreatePayment(directReq)
	default:
		err = fmt.Errorf("unsupported provider")
	}
	if err != nil {
		_ = model.ExpireSubscriptionOrder(tradeNo, "")
		common.ApiErrorMsg(c, "拉起支付失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "data": buildDirectPaymentResponse(result)})
}

func finalizeDirectPayment(notification *service.DirectPaymentNotification, paymentMethod string) error {
	if notification == nil || notification.TradeNo == "" {
		return nil
	}
	LockOrder(notification.TradeNo)
	defer UnlockOrder(notification.TradeNo)

	if topUp := model.GetTopUpByTradeNo(notification.TradeNo); topUp != nil {
		if topUp.PaymentMethod != paymentMethod {
			return model.ErrPaymentMethodMismatch
		}
		if !service.PaymentMoneyMatches(topUp.Money, notification.Money) {
			return fmt.Errorf("payment amount mismatch")
		}
		return model.RechargeDirect(notification.TradeNo, paymentMethod, notification.ProviderTradeNo, notification.RawPayload)
	}

	if order := model.GetSubscriptionOrderByTradeNo(notification.TradeNo); order != nil {
		if order.PaymentMethod != paymentMethod {
			return model.ErrPaymentMethodMismatch
		}
		if !service.PaymentMoneyMatches(order.Money, notification.Money) {
			return fmt.Errorf("payment amount mismatch")
		}
		return model.CompleteSubscriptionOrderWithProvider(notification.TradeNo, notification.ProviderTradeNo, notification.RawPayload)
	}
	return fmt.Errorf("payment order not found")
}

func RequestAlipayDirectPay(c *gin.Context) {
	if !setting.IsAlipayDirectEnabled() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "当前管理员未配置支付信息"})
		return
	}
	createTopUpDirectOrder(c, PaymentMethodAlipayDirect, "alipay")
}

func RequestWechatDirectPay(c *gin.Context) {
	if !setting.IsWechatPayDirectEnabled() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "当前管理员未配置支付信息"})
		return
	}
	createTopUpDirectOrder(c, PaymentMethodWechatDirect, "wechat")
}

func SubscriptionRequestAlipayDirectPay(c *gin.Context) {
	if !setting.IsAlipayDirectEnabled() {
		common.ApiErrorMsg(c, "当前管理员未配置支付信息")
		return
	}
	createSubscriptionDirectOrder(c, PaymentMethodAlipayDirect, "alipay")
}

func SubscriptionRequestWechatDirectPay(c *gin.Context) {
	if !setting.IsWechatPayDirectEnabled() {
		common.ApiErrorMsg(c, "当前管理员未配置支付信息")
		return
	}
	createSubscriptionDirectOrder(c, PaymentMethodWechatDirect, "wechat")
}

func AlipayNotify(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	params := make(map[string]string, len(c.Request.PostForm))
	for key := range c.Request.PostForm {
		params[key] = c.Request.PostForm.Get(key)
	}
	notification, err := (&service.AlipayService{}).VerifyCallback(params)
	if err != nil {
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	if notification.Status != "TRADE_SUCCESS" && notification.Status != "TRADE_FINISHED" {
		_, _ = c.Writer.Write([]byte("success"))
		return
	}
	paymentMethod := PaymentMethodAlipayDirect
	if err := finalizeDirectPayment(notification, paymentMethod); err != nil {
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	_, _ = c.Writer.Write([]byte("success"))
}

// AlipayReturn 处理支付宝同步跳转回调。
// 注意：此端点仅用于用户体验跳转，不执行订单完成逻辑。
// 订单完成统一由异步通知 (AlipayNotify) 处理，避免同步 return URL 参数被篡改的风险。
func AlipayReturn(c *gin.Context) {
	params := make(map[string]string, len(c.Request.URL.Query()))
	for key := range c.Request.URL.Query() {
		params[key] = c.Request.URL.Query().Get(key)
	}
	notification, err := (&service.AlipayService{}).VerifyCallback(params)
	if err != nil {
		c.Redirect(http.StatusFound, system_setting.ServerAddress+"/console/topup?pay=fail")
		return
	}
	// 仅根据验签后的状态做页面展示跳转，不做 finalizeDirectPayment
	if notification.Status == "TRADE_SUCCESS" || notification.Status == "TRADE_FINISHED" {
		c.Redirect(http.StatusFound, system_setting.ServerAddress+"/console/topup?pay=success&show_history=true")
		return
	}
	c.Redirect(http.StatusFound, system_setting.ServerAddress+"/console/topup?pay=pending&show_history=true")
}

func WechatNotify(c *gin.Context) {
	// 限制请求体大小，防止恶意超大请求导致内存耗尽
	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, wechatNotifyMaxBodySize))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "读取回调失败"})
		return
	}
	notification, err := (&service.WechatPayService{}).VerifyAndDecryptNotification(c.Request.Header, bodyBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "验签失败"})
		return
	}
	if notification.Status != "SUCCESS" {
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
		return
	}
	paymentMethod, finalizeErr := detectWechatPaymentMethod(notification.TradeNo)
	if finalizeErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "订单不存在"})
		return
	}
	if err := finalizeDirectPayment(notification, paymentMethod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "处理失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}

// detectWechatPaymentMethod 根据订单号从充值订单或订阅订单中查找支付方式
func detectWechatPaymentMethod(tradeNo string) (string, error) {
	if topUp := model.GetTopUpByTradeNo(tradeNo); topUp != nil {
		return topUp.PaymentMethod, nil
	}
	if order := model.GetSubscriptionOrderByTradeNo(tradeNo); order != nil {
		return order.PaymentMethod, nil
	}
	return "", ErrOrderNotFound
}

func GetTopUpStatus(c *gin.Context) {
	tradeNo := strings.TrimSpace(c.Query("trade_no"))
	if tradeNo == "" {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil || topUp.UserId != c.GetInt("id") {
		common.ApiErrorMsg(c, "充值订单不存在")
		return
	}
	common.ApiSuccess(c, gin.H{
		"trade_no":          topUp.TradeNo,
		"payment_method":    topUp.PaymentMethod,
		"status":            topUp.Status,
		"provider_trade_no": topUp.ProviderTradeNo,
	})
}

func GetSubscriptionPaymentStatus(c *gin.Context) {
	tradeNo := strings.TrimSpace(c.Query("trade_no"))
	if tradeNo == "" {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	order := model.GetSubscriptionOrderByTradeNo(tradeNo)
	if order == nil || order.UserId != c.GetInt("id") {
		common.ApiErrorMsg(c, "订阅订单不存在")
		return
	}
	common.ApiSuccess(c, gin.H{
		"trade_no":          order.TradeNo,
		"payment_method":    order.PaymentMethod,
		"status":            order.Status,
		"provider_trade_no": order.ProviderTradeNo,
	})
}
