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
