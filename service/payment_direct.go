package service

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
)

const (
	ClientScenePC     = "pc"
	ClientSceneMobile = "mobile"

	PaymentActionQR       = "qr"
	PaymentActionRedirect = "redirect"
)

type DirectPaymentRequest struct {
	TradeNo     string
	Subject     string
	Description string
	Money       float64
	ClientScene string
	NotifyURL   string
	ReturnURL   string
	ClientIP    string
}

type DirectPaymentResponse struct {
	Provider        string `json:"provider"`
	TradeNo         string `json:"trade_no"`
	Action          string `json:"action"`
	QRCodeURL       string `json:"qr_code_url,omitempty"`
	RedirectURL     string `json:"redirect_url,omitempty"`
	ExpiresAt       int64  `json:"expires_at,omitempty"`
	ProviderTradeNo string `json:"provider_trade_no,omitempty"`
}

type DirectPaymentNotification struct {
	TradeNo         string
	ProviderTradeNo string
	Status          string
	Money           float64
	RawPayload      string
}

func NormalizeClientScene(scene string) string {
	switch strings.ToLower(strings.TrimSpace(scene)) {
	case ClientSceneMobile:
		return ClientSceneMobile
	default:
		return ClientScenePC
	}
}

func PaymentMoneyMatches(expected float64, actual float64) bool {
	return math.Abs(expected-actual) < 0.01
}

func formatFen(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

func randomNonce(length int) string {
	if length <= 0 {
		length = 24
	}
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	buf := make([]byte, length)
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return common.GetRandomString(length)
	}
	for i := range buf {
		buf[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(buf)
}

func sortedSignContent(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if value == "" || key == "sign" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for idx, key := range keys {
		if idx > 0 {
			builder.WriteByte('&')
		}
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(params[key])
	}
	return builder.String()
}

func ensurePEMBlock(raw string, blockType string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if strings.Contains(trimmed, "BEGIN ") {
		return trimmed
	}
	var builder strings.Builder
	builder.WriteString("-----BEGIN ")
	builder.WriteString(blockType)
	builder.WriteString("-----\n")
	for len(trimmed) > 64 {
		builder.WriteString(trimmed[:64])
		builder.WriteByte('\n')
		trimmed = trimmed[64:]
	}
	builder.WriteString(trimmed)
	builder.WriteByte('\n')
	builder.WriteString("-----END ")
	builder.WriteString(blockType)
	builder.WriteString("-----")
	return builder.String()
}

func parseRSAPrivateKey(raw string) (*rsa.PrivateKey, error) {
	pemText := ensurePEMBlock(raw, "PRIVATE KEY")
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, fmt.Errorf("invalid private key")
	}
	if parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		privateKey, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
		return privateKey, nil
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func parseRSAPublicKey(raw string) (*rsa.PublicKey, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("public key is empty")
	}
	candidates := []string{
		trimmed,
		ensurePEMBlock(trimmed, "PUBLIC KEY"),
		ensurePEMBlock(trimmed, "CERTIFICATE"),
	}
	for _, candidate := range candidates {
		block, _ := pem.Decode([]byte(candidate))
		if block == nil {
			continue
		}
		switch block.Type {
		case "CERTIFICATE":
			cert, err := x509.ParseCertificate(block.Bytes)
			if err == nil {
				if publicKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
					return publicKey, nil
				}
			}
		default:
			parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err == nil {
				if publicKey, ok := parsed.(*rsa.PublicKey); ok {
					return publicKey, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("invalid public key")
}

func signRSA256(content string, privateKey string) (string, error) {
	key, err := parseRSAPrivateKey(privateKey)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(content))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func verifyRSA256(content string, signature string, publicKey string) error {
	key, err := parseRSAPublicKey(publicKey)
	if err != nil {
		return err
	}
	signBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}
	sum := sha256.Sum256([]byte(content))
	return rsa.VerifyPKCS1v15(key, crypto.SHA256, sum[:], signBytes)
}

func defaultAlipayNotifyURL() string {
	if strings.TrimSpace(setting.AlipayNotifyURL) != "" {
		return strings.TrimSpace(setting.AlipayNotifyURL)
	}
	return strings.TrimRight(GetCallbackAddress(), "/") + "/api/alipay/notify"
}

func defaultAlipayReturnURL() string {
	if strings.TrimSpace(setting.AlipayReturnURL) != "" {
		return strings.TrimSpace(setting.AlipayReturnURL)
	}
	return strings.TrimRight(GetCallbackAddress(), "/") + "/api/alipay/return"
}

func defaultWechatNotifyURL() string {
	if strings.TrimSpace(setting.WechatPayNotifyURL) != "" {
		return strings.TrimSpace(setting.WechatPayNotifyURL)
	}
	return strings.TrimRight(GetCallbackAddress(), "/") + "/api/wechat/notify"
}

type alipayGatewayResponse struct {
	AlipayTradePrecreateResponse struct {
		Code       string `json:"code"`
		Msg        string `json:"msg"`
		SubCode    string `json:"sub_code"`
		SubMsg     string `json:"sub_msg"`
		OutTradeNo string `json:"out_trade_no"`
		QRCode     string `json:"qr_code"`
	} `json:"alipay_trade_precreate_response"`
	Sign string `json:"sign"`
}

type AlipayService struct{}

func (s *AlipayService) gatewayURL() string {
	return "https://openapi.alipay.com/gateway.do"
}

func (s *AlipayService) CreatePayment(req *DirectPaymentRequest) (*DirectPaymentResponse, error) {
	if !setting.IsAlipayDirectEnabled() {
		return nil, fmt.Errorf("alipay direct payment is not configured")
	}
	req.ClientScene = NormalizeClientScene(req.ClientScene)
	if req.NotifyURL == "" {
		req.NotifyURL = defaultAlipayNotifyURL()
	}
	if req.ReturnURL == "" {
		req.ReturnURL = defaultAlipayReturnURL()
	}
	if req.ClientScene == ClientSceneMobile {
		return s.createWapPayment(req)
	}
	return s.createPCPayment(req)
}

func (s *AlipayService) createPCPayment(req *DirectPaymentRequest) (*DirectPaymentResponse, error) {
	jsonBytes, err := common.Marshal(map[string]any{
		"out_trade_no": req.TradeNo,
		"total_amount": formatAmount(req.Money),
		"subject":      req.Subject,
		"body":         req.Description,
	})
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"app_id":      setting.AlipayAppID,
		"method":      "alipay.trade.precreate",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  req.NotifyURL,
		"return_url":  req.ReturnURL,
		"biz_content": string(jsonBytes),
	}
	signature, err := signRSA256(sortedSignContent(params), setting.AlipayPrivateKey)
	if err != nil {
		return nil, err
	}
	params["sign"] = signature
	form := url.Values{}
	for key, value := range params {
		form.Set(key, value)
	}
	httpReq, err := http.NewRequest(http.MethodPost, s.gatewayURL(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := GetHttpClient().Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("alipay precreate http status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	var payload alipayGatewayResponse
	if err := common.DecodeJson(resp.Body, &payload); err != nil {
		return nil, err
	}
	result := payload.AlipayTradePrecreateResponse
	if result.Code != "10000" || strings.TrimSpace(result.QRCode) == "" {
		msg := strings.TrimSpace(result.SubMsg)
		if msg == "" {
			msg = strings.TrimSpace(result.Msg)
		}
		if msg == "" {
			msg = "alipay precreate failed"
		}
		return nil, fmt.Errorf(msg)
	}
	return &DirectPaymentResponse{
		Provider:  "alipay",
		TradeNo:   req.TradeNo,
		Action:    PaymentActionQR,
		QRCodeURL: result.QRCode,
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}, nil
}

func (s *AlipayService) createWapPayment(req *DirectPaymentRequest) (*DirectPaymentResponse, error) {
	jsonBytes, err := common.Marshal(map[string]any{
		"out_trade_no": req.TradeNo,
		"total_amount": formatAmount(req.Money),
		"subject":      req.Subject,
		"body":         req.Description,
		"product_code": "QUICK_WAP_WAY",
		"quit_url":     req.ReturnURL,
	})
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"app_id":      setting.AlipayAppID,
		"method":      "alipay.trade.wap.pay",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  req.NotifyURL,
		"return_url":  req.ReturnURL,
		"biz_content": string(jsonBytes),
	}
	signature, err := signRSA256(sortedSignContent(params), setting.AlipayPrivateKey)
	if err != nil {
		return nil, err
	}
	params["sign"] = signature
	form := url.Values{}
	for key, value := range params {
		form.Set(key, value)
	}
	return &DirectPaymentResponse{
		Provider:    "alipay",
		TradeNo:     req.TradeNo,
		Action:      PaymentActionRedirect,
		RedirectURL: s.gatewayURL() + "?" + form.Encode(),
		ExpiresAt:   time.Now().Add(15 * time.Minute).Unix(),
	}, nil
}

func (s *AlipayService) VerifyCallback(params map[string]string) (*DirectPaymentNotification, error) {
	signature := strings.TrimSpace(params["sign"])
	if signature == "" {
		return nil, fmt.Errorf("missing sign")
	}
	content := sortedSignContent(params)
	if err := verifyRSA256(content, signature, setting.AlipayPublicKey); err != nil {
		return nil, err
	}
	amount := 0.0
	if total := strings.TrimSpace(params["total_amount"]); total != "" {
		fmt.Sscanf(total, "%f", &amount)
	}
	raw, _ := common.Marshal(params)
	return &DirectPaymentNotification{
		TradeNo:         strings.TrimSpace(params["out_trade_no"]),
		ProviderTradeNo: strings.TrimSpace(params["trade_no"]),
		Status:          strings.TrimSpace(params["trade_status"]),
		Money:           amount,
		RawPayload:      string(raw),
	}, nil
}

type wechatCreateResponse struct {
	CodeURL string `json:"code_url"`
	H5URL   string `json:"h5_url"`
	Prepay  string `json:"prepay_id"`
}

type wechatNotifyEnvelope struct {
	ID           string `json:"id"`
	CreateTime   string `json:"create_time"`
	EventType    string `json:"event_type"`
	ResourceType string `json:"resource_type"`
	Summary      string `json:"summary"`
	Resource     struct {
		OriginalType   string `json:"original_type"`
		Algorithm      string `json:"algorithm"`
		Ciphertext     string `json:"ciphertext"`
		AssociatedData string `json:"associated_data"`
		Nonce          string `json:"nonce"`
	} `json:"resource"`
}

type wechatNotifyPayload struct {
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeState    string `json:"trade_state"`
	Amount        struct {
		Total int64 `json:"total"`
	} `json:"amount"`
}

type WechatPayService struct{}

func (s *WechatPayService) apiBase() string {
	return "https://api.mch.weixin.qq.com"
}

func (s *WechatPayService) CreatePayment(req *DirectPaymentRequest) (*DirectPaymentResponse, error) {
	if !setting.IsWechatPayDirectEnabled() {
		return nil, fmt.Errorf("wechat pay direct payment is not configured")
	}
	req.ClientScene = NormalizeClientScene(req.ClientScene)
	if req.NotifyURL == "" {
		req.NotifyURL = defaultWechatNotifyURL()
	}
	if req.ClientScene == ClientSceneMobile {
		return s.createH5Payment(req)
	}
	return s.createNativePayment(req)
}

func (s *WechatPayService) createNativePayment(req *DirectPaymentRequest) (*DirectPaymentResponse, error) {
	resp := &wechatCreateResponse{}
	if err := s.createOrder("/v3/pay/transactions/native", map[string]any{
		"appid":        setting.WechatPayAppID,
		"mchid":        setting.WechatPayMchID,
		"description":  req.Subject,
		"out_trade_no": req.TradeNo,
		"notify_url":   req.NotifyURL,
		"time_expire":  time.Now().Add(15 * time.Minute).Format(time.RFC3339),
		"amount": map[string]any{
			"total":    formatFen(req.Money),
			"currency": "CNY",
		},
	}, resp); err != nil {
		return nil, err
	}
	if strings.TrimSpace(resp.CodeURL) == "" {
		return nil, fmt.Errorf("wechat native order missing code_url")
	}
	return &DirectPaymentResponse{
		Provider:  "wechat",
		TradeNo:   req.TradeNo,
		Action:    PaymentActionQR,
		QRCodeURL: resp.CodeURL,
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}, nil
}

func (s *WechatPayService) createH5Payment(req *DirectPaymentRequest) (*DirectPaymentResponse, error) {
	clientIP := strings.TrimSpace(req.ClientIP)
	if clientIP == "" {
		clientIP = "127.0.0.1"
	}
	resp := &wechatCreateResponse{}
	body := map[string]any{
		"appid":        setting.WechatPayAppID,
		"mchid":        setting.WechatPayMchID,
		"description":  req.Subject,
		"out_trade_no": req.TradeNo,
		"notify_url":   req.NotifyURL,
		"time_expire":  time.Now().Add(15 * time.Minute).Format(time.RFC3339),
		"amount": map[string]any{
			"total":    formatFen(req.Money),
			"currency": "CNY",
		},
		"scene_info": map[string]any{
			"payer_client_ip": clientIP,
			"h5_info": map[string]any{
				"type": "Wap",
			},
		},
	}
	if domain := strings.TrimSpace(setting.WechatPayH5Domain); domain != "" {
		body["scene_info"].(map[string]any)["app_name"] = "new-api"
		body["scene_info"].(map[string]any)["app_url"] = domain
	}
	if err := s.createOrder("/v3/pay/transactions/h5", body, resp); err != nil {
		return nil, err
	}
	if strings.TrimSpace(resp.H5URL) == "" {
		return nil, fmt.Errorf("wechat h5 order missing h5_url")
	}
	return &DirectPaymentResponse{
		Provider:    "wechat",
		TradeNo:     req.TradeNo,
		Action:      PaymentActionRedirect,
		RedirectURL: resp.H5URL,
		ExpiresAt:   time.Now().Add(15 * time.Minute).Unix(),
	}, nil
}

func (s *WechatPayService) createOrder(path string, payload map[string]any, out any) error {
	jsonBytes, err := common.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, s.apiBase()+path, bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	auth, err := s.authorization(http.MethodPost, path, string(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Accept", "application/json")
	resp, err := GetHttpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("wechat pay http status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	if err := common.Unmarshal(bodyBytes, out); err != nil {
		return err
	}
	return nil
}

func (s *WechatPayService) authorization(method string, canonicalURL string, body string) (string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := randomNonce(24)
	message := method + "\n" + canonicalURL + "\n" + timestamp + "\n" + nonce + "\n" + body + "\n"
	signature, err := signRSA256(message, setting.WechatPayPrivateKey)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`,
		setting.WechatPayMchID,
		nonce,
		timestamp,
		setting.WechatPaySerialNo,
		signature,
	), nil
}

func (s *WechatPayService) VerifyAndDecryptNotification(headers http.Header, body []byte) (*DirectPaymentNotification, error) {
	timestamp := strings.TrimSpace(headers.Get("Wechatpay-Timestamp"))
	nonce := strings.TrimSpace(headers.Get("Wechatpay-Nonce"))
	signature := strings.TrimSpace(headers.Get("Wechatpay-Signature"))
	if timestamp == "" || nonce == "" || signature == "" {
		return nil, fmt.Errorf("missing wechat pay signature headers")
	}
	message := timestamp + "\n" + nonce + "\n" + string(body) + "\n"
	if err := verifyRSA256(message, signature, setting.WechatPayPlatformCert); err != nil {
		return nil, err
	}
	var envelope wechatNotifyEnvelope
	if err := common.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	plaintext, err := decryptWechatResource(envelope.Resource.Ciphertext, envelope.Resource.Nonce, envelope.Resource.AssociatedData, setting.WechatPayAPIv3Key)
	if err != nil {
		return nil, err
	}
	var payload wechatNotifyPayload
	if err := common.Unmarshal(plaintext, &payload); err != nil {
		return nil, err
	}
	rawPayload := string(plaintext)
	return &DirectPaymentNotification{
		TradeNo:         strings.TrimSpace(payload.OutTradeNo),
		ProviderTradeNo: strings.TrimSpace(payload.TransactionID),
		Status:          strings.TrimSpace(payload.TradeState),
		Money:           float64(payload.Amount.Total) / 100.0,
		RawPayload:      rawPayload,
	}, nil
}

func decryptWechatResource(ciphertext string, nonce string, associatedData string, apiV3Key string) ([]byte, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher([]byte(apiV3Key))
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plainText, err := aead.Open(nil, []byte(nonce), cipherBytes, []byte(associatedData))
	if err != nil {
		return nil, err
	}
	return plainText, nil
}
