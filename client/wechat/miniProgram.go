package wechat

import (
	"errors"
	"fmt"
	"time"

	"github.com/webx-top/gopay/client"
	"github.com/webx-top/gopay/common"
	"github.com/webx-top/gopay/util"
)

var defaultMiniProgram *MiniProgram

func InitMiniProgram(env *ClientData) {
	defaultMiniProgram = NewMiniProgram(env)
}

func DefaultMiniProgram() *MiniProgram {
	return defaultMiniProgram
}

func NewMiniProgram(env *ClientData) *MiniProgram {
	if len(env.PayURL) == 0 {
		env.PayURL = defaultPayURL
	}
	c := &MiniProgram{ClientData: env}
	if len(c.PrivateKey) != 0 && len(c.PublicKey) != 0 {
		c.client = client.NewHTTPSClient(c.PublicKey, c.PrivateKey)
	}

	return c
}

// MiniProgram 微信扫码支付
type MiniProgram struct {
	*ClientData
	client *client.HTTPSClient // 双向证书链接
}

// Pay 支付
func (a *MiniProgram) Pay(charge *common.Charge) (map[string]string, error) {
	var m = make(map[string]string)
	appId := a.AppID
	if charge.APPID != "" {
		appId = charge.APPID
	}
	m["appid"] = appId
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
	m["trade_type"] = "JSAPI"
	m["openid"] = charge.OpenID
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
	c["appId"] = appId
	c["timeStamp"] = fmt.Sprintf("%d", time.Now().Unix())
	c["nonceStr"] = util.RandomStr()
	c["package"] = fmt.Sprintf("prepay_id=%s", xmlRe.PrepayID)
	c["signType"] = "MD5"
	sign2, err := GenSign(a.Key, c)
	if err != nil {
		return map[string]string{}, errors.New("WechatWeb: " + err.Error())
	}
	c["paySign"] = sign2
	delete(c, "appId")
	return c, nil
}

// CloseOrder 关闭订单
func (a *MiniProgram) CloseOrder(outTradeNo string) (QueryResult, error) {
	return CloseOrder(a.AppID, a.MchID, a.Key, outTradeNo)
}

// PayToClient 支付到用户的微信账号
func (a *MiniProgram) PayToClient(charge *common.Charge) (map[string]string, error) {
	return CompanyChange(a.AppID, a.MchID, a.Key, a.client, charge)
}

// QueryOrder 查询订单
func (a *MiniProgram) QueryOrder(tradeNum string) (QueryResult, error) {
	return QueryOrder(a.AppID, a.MchID, a.Key, tradeNum)
}
