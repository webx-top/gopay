package wechat

import (
	"fmt"
	"time"

	"github.com/webx-top/gopay/client"
	"github.com/webx-top/gopay/common"
	"github.com/webx-top/gopay/util"
)

var defaultNative *Native

func InitWxNativeClient(env *ClientData) {
	defaultNative = NewNative(env)
}

func DefaultNative() *Native {
	return defaultNative
}

func NewNative(env *ClientData) *Native {
	if len(env.PayURL) == 0 {
		env.PayURL = defaultPayURL
	}
	c := &Native{ClientData: env}
	if len(c.PrivateKey) != 0 && len(c.PublicKey) != 0 {
		c.client = client.NewHTTPSClient(c.PublicKey, c.PrivateKey)
	}

	return c
}

// Native 微信扫码支付
type Native struct {
	*ClientData
	client *client.HTTPSClient // 双向证书链接
}

// Pay 支付
func (a *Native) Pay(charge *common.Charge) (map[string]string, error) {
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
	m["trade_type"] = "NATIVE"
	m["product_id"] = charge.ProductID
	m["sign_type"] = "MD5"

	sign, err := GenSign(a.Key, m)
	if err != nil {
		return map[string]string{}, err
	}
	m["sign"] = sign

	// 转出xml结构
	xmlRe, err := Post(a.PayURL, m, nil)
	if err != nil {
		return map[string]string{}, err
	}

	var c = make(map[string]string)
	c["timeStamp"] = fmt.Sprintf("%d", time.Now().Unix())
	c["code_url"] = xmlRe.CodeURL

	return c, nil
}

// PayToClient 支付到用户的微信账号
func (a *Native) PayToClient(charge *common.Charge) (map[string]string, error) {
	return CompanyChange(a.AppID, a.MchID, a.Key, a.client, charge)
}

// QueryOrder 查询订单
func (a *Native) QueryOrder(tradeNum string) (QueryResult, error) {
	return QueryOrder(a.AppID, a.MchID, a.Key, tradeNum)
}
