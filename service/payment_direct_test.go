package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/stretchr/testify/require"
)

func TestAlipayVerifyCallback(t *testing.T) {
	privateKeyPEM, publicKeyPEM := mustGenerateRSAKeyPairPEM(t)
	previousPrivateKey := setting.AlipayPrivateKey
	previousPublicKey := setting.AlipayPublicKey
	setting.AlipayPrivateKey = privateKeyPEM
	setting.AlipayPublicKey = publicKeyPEM
	t.Cleanup(func() {
		setting.AlipayPrivateKey = previousPrivateKey
		setting.AlipayPublicKey = previousPublicKey
	})

	params := map[string]string{
		"app_id":       "2026000000000000",
		"out_trade_no": "DIRUSR1NOabc123",
		"trade_no":     "2026041600001",
		"trade_status": "TRADE_SUCCESS",
		"total_amount": "12.34",
		"charset":      "utf-8",
		"sign_type":    "RSA2",
	}
	signature, err := signRSA256(sortedSignContent(params), setting.AlipayPrivateKey)
	require.NoError(t, err)
	params["sign"] = signature

	notification, err := (&AlipayService{}).VerifyCallback(params)
	require.NoError(t, err)
	require.Equal(t, "DIRUSR1NOabc123", notification.TradeNo)
	require.Equal(t, "2026041600001", notification.ProviderTradeNo)
	require.Equal(t, "TRADE_SUCCESS", notification.Status)
	require.InDelta(t, 12.34, notification.Money, 0.0001)
	require.Contains(t, notification.RawPayload, "out_trade_no")
}

func TestWechatVerifyAndDecryptNotification(t *testing.T) {
	platformPrivateKey, platformCertPEM := mustGenerateSelfSignedCertPEM(t)
	previousAPIv3Key := setting.WechatPayAPIv3Key
	previousPlatformCert := setting.WechatPayPlatformCert
	setting.WechatPayAPIv3Key = "0123456789abcdef0123456789abcdef"
	setting.WechatPayPlatformCert = platformCertPEM
	t.Cleanup(func() {
		setting.WechatPayAPIv3Key = previousAPIv3Key
		setting.WechatPayPlatformCert = previousPlatformCert
	})

	plainPayload, err := common.Marshal(map[string]any{
		"out_trade_no":   "DIRSUBUSR9NOxyz789",
		"transaction_id": "4200000000001",
		"trade_state":    "SUCCESS",
		"amount": map[string]any{
			"total": 567,
		},
	})
	require.NoError(t, err)

	nonce := "123456789012"
	associatedData := "transaction"
	ciphertext := mustEncryptWechatPayload(t, plainPayload, nonce, associatedData, setting.WechatPayAPIv3Key)
	body, err := common.Marshal(map[string]any{
		"id":         "evt_1",
		"event_type": "TRANSACTION.SUCCESS",
		"resource": map[string]any{
			"algorithm":       "AEAD_AES_256_GCM",
			"ciphertext":      ciphertext,
			"nonce":           nonce,
			"associated_data": associatedData,
		},
	})
	require.NoError(t, err)

	timestamp := "1713200000"
	signingNonce := "notify-nonce"
	message := timestamp + "\n" + signingNonce + "\n" + string(body) + "\n"
	signature, err := signRSA256(message, platformPrivateKey)
	require.NoError(t, err)

	headers := http.Header{}
	headers.Set("Wechatpay-Timestamp", timestamp)
	headers.Set("Wechatpay-Nonce", signingNonce)
	headers.Set("Wechatpay-Signature", signature)

	notification, err := (&WechatPayService{}).VerifyAndDecryptNotification(headers, body)
	require.NoError(t, err)
	require.Equal(t, "DIRSUBUSR9NOxyz789", notification.TradeNo)
	require.Equal(t, "4200000000001", notification.ProviderTradeNo)
	require.Equal(t, "SUCCESS", notification.Status)
	require.InDelta(t, 5.67, notification.Money, 0.0001)
	require.Contains(t, notification.RawPayload, "transaction_id")
}

func mustGenerateRSAKeyPairPEM(t *testing.T) (string, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})

	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})
	return string(privatePEM), string(publicPEM)
}

func mustGenerateSelfSignedCertPEM(t *testing.T) (string, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "wechat-platform-test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	return string(privatePEM), string(certPEM)
}

func mustEncryptWechatPayload(t *testing.T, plainText []byte, nonce string, associatedData string, apiV3Key string) string {
	t.Helper()
	block, err := aes.NewCipher([]byte(apiV3Key))
	require.NoError(t, err)
	aead, err := cipher.NewGCM(block)
	require.NoError(t, err)
	cipherText := aead.Seal(nil, []byte(nonce), plainText, []byte(associatedData))
	return base64.StdEncoding.EncodeToString(cipherText)
}
