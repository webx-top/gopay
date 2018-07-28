package client

import (
	"fmt"
	"time"

	"github.com/fengzie/gopay/common"
	"github.com/fengzie/gopay/util"
)

var defaultWechatNativeClient *WechatNativeClient

func InitWxNativeClient(env *common.WechatClientData) {
	c := &WechatNativeClient{Env: env}
	if len(c.Env.PrivateKey) != 0 && len(c.Env.PublicKey) != 0 {
		c.httpsClient = NewHTTPSClient(c.Env.PublicKey, c.Env.PrivateKey)
	}

	defaultWechatNativeClient = c
}

func DefaultWechatNativeClient() *WechatNativeClient {
	return defaultWechatNativeClient
}

// WechatNativeClient 微信扫码支付
type WechatNativeClient struct {
	Env         *common.WechatClientData
	httpsClient *HTTPSClient // 双向证书链接
}

// Pay 支付
func (this *WechatNativeClient) Pay(charge *common.Charge) (map[string]string, error) {
	var m = make(map[string]string)
	m["appid"] = this.Env.AppID
	m["mch_id"] = this.Env.MchID
	m["nonce_str"] = util.RandomStr()
	m["body"] = TruncatedText(charge.Describe, 32)
	m["out_trade_no"] = charge.TradeNum
	m["total_fee"] = WechatMoneyFeeToString(charge.MoneyFee)
	m["spbill_create_ip"] = util.LocalIP()
	m["notify_url"] = charge.CallbackURL
	m["trade_type"] = "NATIVE"
	m["product_id"] = charge.ProductID
	m["sign_type"] = "MD5"

	sign, err := WechatGenSign(this.Env.Key, m)
	if err != nil {
		return map[string]string{}, err
	}
	m["sign"] = sign

	// 转出xml结构
	xmlRe, err := PostWechat("https://api.mch.weixin.qq.com/pay/unifiedorder", m, nil)
	if err != nil {
		return map[string]string{}, err
	}

	var c = make(map[string]string)
	c["timeStamp"] = fmt.Sprintf("%d", time.Now().Unix())
	c["code_url"] = xmlRe.CodeURL

	return c, nil
}

// 支付到用户的微信账号
func (this *WechatNativeClient) PayToClient(charge *common.Charge) (map[string]string, error) {
	return WachatCompanyChange(this.Env.AppID, this.Env.MchID, this.Env.Key, this.httpsClient, charge)
}

// QueryOrder 查询订单
func (this *WechatNativeClient) QueryOrder(tradeNum string) (common.WeChatQueryResult, error) {
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
