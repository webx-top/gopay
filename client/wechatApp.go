package client

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fengzie/gopay/common"
	"github.com/fengzie/gopay/util"
)

var defaultWechatAppClient *WechatAppClient

func InitWxAppClient(env *common.WechatClientData) {
	c := &WechatAppClient{Env: env}
	if len(c.Env.PrivateKey) != 0 && len(c.Env.PublicKey) != 0 {
		c.httpsClient = NewHTTPSClient(c.Env.PublicKey, c.Env.PrivateKey)
	}

	defaultWechatAppClient = c
}

// DefaultWechatAppClient 默认微信app客户端
func DefaultWechatAppClient() *WechatAppClient {
	return defaultWechatAppClient
}

// WechatAppClient 微信扫码支付
type WechatAppClient struct {
	Env         *common.WechatClientData
	httpsClient *HTTPSClient // 双向证书链接
}

// Pay 支付
func (this *WechatAppClient) Pay(charge *common.Charge) (map[string]string, error) {
	var m = make(map[string]string)
	m["appid"] = this.Env.AppID
	m["mch_id"] = this.Env.MchID
	m["nonce_str"] = util.RandomStr()
	m["body"] = TruncatedText(charge.Describe, 32)
	m["out_trade_no"] = charge.TradeNum
	m["total_fee"] = WechatMoneyFeeToString(charge.MoneyFee)
	m["spbill_create_ip"] = util.LocalIP()
	m["notify_url"] = charge.CallbackURL
	m["trade_type"] = "APP"
	m["sign_type"] = "MD5"

	sign, err := WechatGenSign(this.Env.Key, m)
	if err != nil {
		return map[string]string{}, errors.New("WechatApp.sign: " + err.Error())
	}

	m["sign"] = sign

	xmlRe, err := PostWechat("https://api.mch.weixin.qq.com/pay/unifiedorder", m, nil)
	if err != nil {
		return map[string]string{}, err
	}

	var c = make(map[string]string)
	c["appid"] = this.Env.AppID
	c["partnerid"] = this.Env.MchID
	c["prepayid"] = xmlRe.PrepayID
	c["package"] = "Sign=WXPay"
	c["noncestr"] = util.RandomStr()
	c["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())

	sign2, err := WechatGenSign(this.Env.Key, c)
	if err != nil {
		return map[string]string{}, errors.New("WechatApp.paySign: " + err.Error())
	}
	c["paySign"] = strings.ToUpper(sign2)

	return c, nil
}

// 支付到用户的微信账号
func (this *WechatAppClient) PayToClient(charge *common.Charge) (map[string]string, error) {
	return WachatCompanyChange(this.Env.AppID, this.Env.MchID, this.Env.Key, this.httpsClient, charge)
}

// QueryOrder 查询订单
func (this *WechatAppClient) QueryOrder(tradeNum string) (common.WeChatQueryResult, error) {
	var m = make(map[string]string)
	m["appid"] = this.Env.AppID
	m["mch_id"] = this.Env.MchID
	m["out_trade_no"] = tradeNum
	m["nonce_str"] = util.RandomStr()

	sign, err := WechatGenSign(this.Env.Key, m)
	if err != nil {
		return common.WeChatQueryResult{}, err
	}

	m["sign"] = sign

	return PostWechat("https://api.mch.weixin.qq.com/pay/orderquery", m, nil)
}
