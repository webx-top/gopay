package alipay

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/webx-top/gopay/client"
	"github.com/webx-top/gopay/common"
)

var defaultApp *App

type App struct {
	*PayClient
	SellerID string //合作者ID
}

func InitApp(c *App) {
	defaultApp = c
}

// DefaultApp 得到默认支付宝app客户端
func DefaultApp() *App {
	return defaultApp
}

func NewApp() *App {
	return &App{PayClient: New()}
}

func (a *App) Pay(charge *common.Charge) (map[string]string, error) {
	var m = make(map[string]string)
	var bizContent = make(map[string]string)
	m["app_id"] = a.AppID
	m["method"] = "alipay.trade.app.pay"
	m["format"] = "JSON"
	m["charset"] = "utf-8"
	m["timestamp"] = time.Now().Format("2006-01-02 15:04:05")
	m["version"] = "1.0"
	m["notify_url"] = charge.CallbackURL
	m["sign_type"] = a.RSAType
	bizContent["subject"] = client.TruncatedText(charge.Describe, 32)
	bizContent["out_trade_no"] = charge.TradeNum
	bizContent["product_code"] = "QUICK_MSECURITY_PAY"
	bizContent["total_amount"] = MoneyFeeToString(charge.MoneyFee)

	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return map[string]string{}, errors.New("json.Marshal: " + err.Error())
	}
	m["biz_content"] = string(bizContentJSON)

	m["sign"] = a.GenSign(m)

	return map[string]string{"orderString": a.ToURL(m)}, nil
}

func (a *App) CloseOrder(charge *common.Charge) (map[string]string, error) {
	return map[string]string{}, errors.New("暂未开发该功能")
}

func (a *App) PayToClient(charge *common.Charge) (map[string]string, error) {
	return map[string]string{}, errors.New("暂未开发该功能")
}

// QueryOrder 订单查询
func (a *App) QueryOrder(outTradeNo string) (AppQueryResult, error) {
	var m = make(map[string]string)
	m["method"] = "alipay.trade.query"
	m["app_id"] = a.AppID
	m["format"] = "JSON"
	m["charset"] = "utf-8"
	m["timestamp"] = time.Now().Format("2006-01-02 15:04:05")
	m["version"] = "1.0"
	m["sign_type"] = a.RSAType
	bizContent := map[string]string{"out_trade_no": outTradeNo}
	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return AppQueryResult{}, errors.New("json.Marshal: " + err.Error())
	}
	m["biz_content"] = string(bizContentJSON)
	sign := a.GenSign(m)
	m["sign"] = sign

	url := fmt.Sprintf("%s?%s", a.OpenAPIURL, a.ToURL(m))

	return GetApp(url)
}

//// GenSign 产生签名
//func (a *App) GenSign(m map[string]string) string {
//	var data []string
//
//	for k, v := range m {
//		if v != "" && k != "sign" {
//			data = append(data, fmt.Sprintf(`%s=%s`, k, v))
//		}
//	}
//	sort.Strings(data)
//	signData := strings.Join(data, "&")
//
//	s := sha1.New()
//	_, err := s.Write([]byte(signData))
//	if err != nil {
//		panic(err)
//	}
//	hashByte := s.Sum(nil)
//	signByte, err := a.PrivateKey.Sign(rand.Reader, hashByte, crypto.SHA1)
//	if err != nil {
//		panic(err)
//	}
//
//	return base64.StdEncoding.EncodeToString(signByte)
//}
//
////CheckSign 检测签名
//func (a *App) CheckSign(signData, sign string) {
//	signByte, err := base64.StdEncoding.DecodeString(sign)
//	if err != nil {
//		panic(err)
//	}
//	s := sha1.New()
//	_, err = s.Write([]byte(signData))
//	if err != nil {
//		panic(err)
//	}
//	hash := s.Sum(nil)
//	err = rsa.VerifyPKCS1v15(a.PublicKey, crypto.SHA1, hash, signByte)
//	if err != nil {
//		panic(err)
//	}
//}
//
//// ToURL
//func (a *App) ToURL(m map[string]string) string {
//	var buf []string
//	for k, v := range m {
//		buf = append(buf, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
//	}
//	return strings.Join(buf, "&")
//}
