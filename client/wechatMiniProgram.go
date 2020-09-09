package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/webx-top/gopay/common"
	"github.com/webx-top/gopay/util"
)

var defaultWechatMiniProgramClient *WechatMiniProgramClient

func InitWxMiniProgramClient(env *common.WechatClientData) {
	c := &WechatMiniProgramClient{Env: env}
	if len(c.Env.PrivateKey) != 0 && len(c.Env.PublicKey) != 0 {
		c.httpsClient = NewHTTPSClient(c.Env.PublicKey, c.Env.PrivateKey)
	}

	defaultWechatMiniProgramClient = c
}

func DefaultWechatMiniProgramClient() *WechatMiniProgramClient {
	return defaultWechatMiniProgramClient
}

// WechatMiniProgramClient 微信扫码支付
type WechatMiniProgramClient struct {
	Env         *common.WechatClientData
	httpsClient *HTTPSClient // 双向证书链接
}

// Pay 支付
func (this *WechatMiniProgramClient) Pay(charge *common.Charge) (map[string]string, error) {
	var m = make(map[string]string)
	appId := this.AppID
	if charge.APPID != "" {
		appId = charge.APPID
	}
	m["appid"] = appId
	m["mch_id"] = this.MchID
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
	c["appId"] = appId
	c["timeStamp"] = fmt.Sprintf("%d", time.Now().Unix())
	c["nonceStr"] = util.RandomStr()
	c["package"] = fmt.Sprintf("prepay_id=%s", xmlRe.PrepayID)
	c["signType"] = "MD5"
	sign2, err := WechatGenSign(this.Env.Key, c)
	if err != nil {
		return map[string]string{}, errors.New("WechatWeb: " + err.Error())
	}
	c["paySign"] = sign2
	delete(c, "appId")
	return c, nil
}

// 关闭订单
func (this *WechatMiniProgramClient) CloseOrder(outTradeNo string) (common.WeChatQueryResult, error) {
	return WachatCloseOrder(this.AppID, this.MchID, this.Key, outTradeNo)
}

// 支付到用户的微信账号
func (this *WechatMiniProgramClient) PayToClient(charge *common.Charge) (map[string]string, error) {
	return WachatCompanyChange(this.Env.AppID, this.Env.MchID, this.Env.Key, this.httpsClient, charge)
}

// QueryOrder 查询订单
func (this *WechatMiniProgramClient) QueryOrder(tradeNum string) (common.WeChatQueryResult, error) {
	return WachatQueryOrder(this.AppID, this.MchID, this.Key, tradeNum)
}
