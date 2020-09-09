package wechat

import (
	"bytes"
	"crypto/md5"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/webx-top/gopay/client"
	"github.com/webx-top/gopay/common"
	"github.com/webx-top/gopay/util"
)

var defaultPayURL = "https://api.mch.weixin.qq.com/pay/unifiedorder"

// MoneyFeeToString 微信金额浮点转字符串
func MoneyFeeToString(moneyFee float64) string {
	aDecimal := decimal.NewFromFloat(moneyFee)
	bDecimal := decimal.NewFromFloat(100)
	return aDecimal.Mul(bDecimal).Truncate(0).String()
}

// CompanyChange 微信企业付款到零钱
func CompanyChange(mchAppid, mchid, key string, conn *client.HTTPSClient, charge *common.Charge) (map[string]string, error) {
	var m = make(map[string]string)
	m["mch_appid"] = mchAppid
	m["mchid"] = mchid
	m["nonce_str"] = util.RandomStr()
	m["partner_trade_no"] = charge.TradeNum
	m["openid"] = charge.OpenID
	m["amount"] = MoneyFeeToString(charge.MoneyFee)
	m["spbill_create_ip"] = util.LocalIP()
	m["desc"] = client.TruncatedText(charge.Describe, 32)

	// 是否验证用户名称
	if charge.CheckName {
		m["check_name"] = "FORCE_CHECK"
		m["re_user_name"] = charge.ReUserName
	} else {
		m["check_name"] = "NO_CHECK"
	}

	sign, err := GenSign(key, m)
	if err != nil {
		return map[string]string{}, err
	}
	m["sign"] = sign

	// 转出xml结构
	result, err := Post("https://api.mch.weixin.qq.com/mmpaymkttransfers/promotion/transfers", m, conn)
	if err != nil {
		return map[string]string{}, err
	}

	return client.Struct2Map(result)
}

// CloseOrder 微信关闭订单
func CloseOrder(appid, mchid, key string, outTradeNo string) (QueryResult, error) {
	var m = make(map[string]string)
	m["appid"] = appid
	m["mch_id"] = mchid
	m["nonce_str"] = util.RandomStr()
	m["out_trade_no"] = outTradeNo
	m["sign_type"] = "MD5"

	sign, err := GenSign(key, m)
	if err != nil {
		return QueryResult{}, err
	}
	m["sign"] = sign

	// 转出xml结构
	result, err := Post("https://api.mch.weixin.qq.com/pay/closeorder", m, nil)
	if err != nil {
		return QueryResult{}, err
	}

	return result, err
}

// QueryOrder 微信订单查询
func QueryOrder(appID, mchID, key, tradeNum string) (QueryResult, error) {
	var m = make(map[string]string)
	m["appid"] = appID
	m["mch_id"] = mchID
	m["out_trade_no"] = tradeNum
	m["nonce_str"] = util.RandomStr()

	sign, err := GenSign(key, m)
	if err != nil {
		return QueryResult{}, err
	}

	m["sign"] = sign

	return Post("https://api.mch.weixin.qq.com/pay/orderquery", m, nil)
}

func GenSign(key string, m map[string]string) (string, error) {
	var signData []string
	for k, v := range m {
		if v != "" && k != "sign" && k != "key" {
			signData = append(signData, fmt.Sprintf("%s=%s", k, v))
		}
	}

	sort.Strings(signData)
	signStr := strings.Join(signData, "&")
	signStr = signStr + "&key=" + key

	c := md5.New()
	_, err := c.Write([]byte(signStr))
	if err != nil {
		return "", errors.New("WechatGenSign md5.Write: " + err.Error())
	}
	signByte := c.Sum(nil)
	if err != nil {
		return "", errors.New("WechatGenSign md5.Sum: " + err.Error())
	}
	return strings.ToUpper(fmt.Sprintf("%x", signByte)), nil
}

//Post 对微信下订单或者查订单
func Post(url string, data map[string]string, h *client.HTTPSClient) (QueryResult, error) {
	var xmlRe QueryResult
	buf := bytes.NewBufferString("")

	for k, v := range data {
		buf.WriteString(fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k))
	}
	xmlStr := fmt.Sprintf("<xml>%s</xml>", buf.String())

	hc := new(client.HTTPSClient)
	if h != nil {
		hc = h
	} else {
		hc = client.HTTPSC
	}

	re, err := hc.PostData(url, "text/xml:charset=UTF-8", xmlStr)
	if err != nil {
		return xmlRe, errors.New("HTTPSC.PostData: " + err.Error())
	}

	err = xml.Unmarshal(re, &xmlRe)
	if err != nil {
		return xmlRe, errors.New("xml.Unmarshal: " + err.Error())
	}

	if xmlRe.ReturnCode != "SUCCESS" {
		// 通信失败
		return xmlRe, errors.New("xmlRe.ReturnMsg: " + xmlRe.ReturnMsg)
	}

	if xmlRe.ResultCode != "SUCCESS" {
		// 业务结果失败
		return xmlRe, errors.New("xmlRe.ErrCodeDes: " + xmlRe.ErrCodeDes)
	}
	return xmlRe, nil
}

func SandBoxGetSign(url string, data map[string]string, h *client.HTTPSClient) (string, error) {
	var xmlRe struct {
		ReturnCode     string `xml:"return_code" json:"return_code,omitempty"`
		ReturnMsg      string `xml:"return_msg" json:"return_msg,omitempty"`
		SandboxSignkey string `xml:"sandbox_signkey" json:"sandbox_signkey"`
	}
	buf := bytes.NewBufferString("")
	var str string
	for k, v := range data {
		if k != "sign" {
			str = fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k)
		} else {
			str = fmt.Sprintf("<%s>%s</%s>", k, v, k)
		}
		buf.WriteString(str)
	}
	xmlStr := fmt.Sprintf("<xml>%s</xml>", buf.String())

	hc := new(client.HTTPSClient)
	if h != nil {
		hc = h
	} else {
		hc = client.HTTPSC
	}

	re, err := hc.PostData(url, "text/xml:charset=UTF-8", xmlStr)
	if err != nil {
		return "", errors.New("HTTPSC.PostData: " + err.Error())
	}

	err = xml.Unmarshal(re, &xmlRe)
	if err != nil {
		return "", errors.New("xml.Unmarshal: " + err.Error())
	}

	if xmlRe.ReturnCode != "SUCCESS" {
		// 通信失败
		return "", errors.New("xmlRe.ReturnMsg: " + xmlRe.ReturnMsg)
	}
	return xmlRe.SandboxSignkey, nil
}
