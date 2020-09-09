package wechat

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/webx-top/gopay/client"
	"github.com/webx-top/gopay/common"
	"github.com/webx-top/gopay/util"
)

var defaultApp *App

func InitApp(env *ClientData) {
	defaultApp = NewApp(env)
}

// DefaultApp 默认微信app客户端
func DefaultApp() *App {
	return defaultApp
}

func NewApp(env *ClientData) *App {
	if len(env.PayURL) == 0 {
		env.PayURL = defaultPayURL
	}
	c := &App{ClientData: env}
	if len(c.PrivateKey) != 0 && len(c.PublicKey) != 0 {
		c.client = client.NewHTTPSClient(c.PublicKey, c.PrivateKey)
	}
	return c
}

// App 微信扫码支付
type App struct {
	*ClientData
	client *client.HTTPSClient // 双向证书链接
}

// Pay 支付
func (a *App) Pay(charge *common.Charge) (map[string]string, error) {
	var m = make(map[string]string)
	m["appid"] = a.AppID
	m["mch_id"] = a.MchID
	m["nonce_str"] = util.RandomStr()
	m["body"] = client.TruncatedText(charge.Describe, 32)
	m["out_trade_no"] = charge.TradeNum
	if charge.Attach != "" {
		m["attach"] = charge.Attach
	}
	m["total_fee"] = MoneyFeeToString(charge.MoneyFee)
	m["spbill_create_ip"] = util.LocalIP()
	m["notify_url"] = charge.CallbackURL
	m["trade_type"] = "APP"
	m["sign_type"] = "MD5"

	sign, err := GenSign(a.Key, m)
	if err != nil {
		return map[string]string{}, errors.New("WechatApp.sign: " + err.Error())
	}

	m["sign"] = sign

	xmlRe, err := Post(a.PayURL, m, nil)
	if err != nil {
		return map[string]string{}, err
	}

	var c = make(map[string]string)
	c["appid"] = a.AppID
	c["partnerid"] = a.MchID
	c["prepayid"] = xmlRe.PrepayID
	c["package"] = "Sign=WXPay"
	c["noncestr"] = util.RandomStr()
	c["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())

	sign2, err := GenSign(a.Key, c)
	if err != nil {
		return map[string]string{}, errors.New("WechatApp.paySign: " + err.Error())
	}
	c["paySign"] = strings.ToUpper(sign2)

	return c, nil
}

// CloseOrder 关闭订单
func (a *App) CloseOrder(outTradeNo string) (QueryResult, error) {
	return CloseOrder(a.AppID, a.MchID, a.Key, outTradeNo)
}

// PayToClient 支付到用户的微信账号
func (a *App) PayToClient(charge *common.Charge) (map[string]string, error) {

	if a.client == nil {
		a.client = client.NewHTTPSClient(a.PublicKey, a.PrivateKey)
	}

	return CompanyChange(a.AppID, a.MchID, a.Key, a.client, charge)
}

// QueryOrder 查询订单
func (a *App) QueryOrder(tradeNum string) (QueryResult, error) {
	return QueryOrder(a.AppID, a.MchID, a.Key, tradeNum)
}
