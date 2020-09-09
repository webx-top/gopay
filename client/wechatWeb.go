package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/webx-top/gopay/common"
	"github.com/webx-top/gopay/util"
)

var defaultWechatWebClient *WechatWebClient

func InitWxWebClient(env *common.WechatClientData) {
	c := &WechatWebClient{Env: env}
	if len(c.Env.PrivateKey) != 0 && len(c.Env.PublicKey) != 0 {
		c.httpsClient = NewHTTPSClient(c.Env.PublicKey, c.Env.PrivateKey)
	}

	defaultWechatWebClient = c
}

func DefaultWechatWebClient() *WechatWebClient {
	return defaultWechatWebClient
}

// WechatWebClient 微信公众号支付
type WechatWebClient struct {
	Env         *common.WechatClientData
	httpsClient *HTTPSClient // 双向证书链接
}

// Pay 支付
func (this *WechatWebClient) Pay(charge *common.Charge) (map[string]string, error) {
	var m = make(map[string]string)
	m["appid"] = this.Env.AppID
	m["mch_id"] = this.Env.MchID
	m["nonce_str"] = util.RandomStr()
	m["body"] = TruncatedText(charge.Describe, 32)
	m["out_trade_no"] = charge.TradeNum
	if charge.Attach != "" {
		m["attach"] = charge.Attach
	}
	m["total_fee"] = WechatMoneyFeeToString(charge.MoneyFee)
	m["spbill_create_ip"] = util.LocalIP()
	m["notify_url"] = charge.CallbackURL
	m["trade_type"] = "JSAPI"
	m["openid"] = charge.OpenID
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
	c["appId"] = this.Env.AppID
	c["timeStamp"] = fmt.Sprintf("%d", time.Now().Unix())
	c["nonceStr"] = util.RandomStr()
	c["package"] = fmt.Sprintf("prepay_id=%s", xmlRe.PrepayID)
	c["signType"] = "MD5"
	sign2, err := WechatGenSign(this.Env.Key, c)
	if err != nil {
		return map[string]string{}, errors.New("WechatWeb: " + err.Error())
	}
	c["paySign"] = sign2

	return c, nil
}

// 关闭订单
func (this *WechatWebClient) CloseOrder(outTradeNo string) (common.WeChatQueryResult, error) {
	return WachatCloseOrder(this.AppID, this.MchID, this.Key, outTradeNo)
}

// 支付到用户的微信账号
func (this *WechatWebClient) PayToClient(charge *common.Charge) (map[string]string, error) {
	return WachatCompanyChange(this.Env.AppID, this.Env.MchID, this.Env.Key, this.httpsClient, charge)
}

// QueryOrder 查询订单
func (this *WechatWebClient) QueryOrder(tradeNum string) (common.WeChatQueryResult, error) {
	return WachatQueryOrder(this.AppID, this.MchID, this.Key, tradeNum)
}
