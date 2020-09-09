package alipay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/webx-top/gopay/client"
	"github.com/webx-top/gopay/common"
)

var defaultWeb *Web

// Web 支付宝网页支付
type Web struct {
	*PayClient
	PartnerID   string // 支付宝合作身份ID
	SellerID    string // 卖家支付宝用户号
	CallbackURL string // 回调接口
}

func InitWeb(c *Web) {
	defaultWeb = c
}

// DefaultWeb 默认支付宝网页支付客户端
func DefaultWeb() *Web {
	return defaultWeb
}

func NewWeb() *Web {
	return &Web{PayClient: New()}
}

// Pay 实现支付接口
func (a *Web) Pay(charge *common.Charge) (map[string]string, error) {
	var m = make(map[string]string)
	m["service"] = "create_direct_pay_by_user"
	m["partner"] = a.PartnerID
	m["_input_charset"] = "UTF-8"
	m["notify_url"] = charge.CallbackURL
	m["return_url"] = charge.ReturnURL // 注意链接不能有&符号，否则会签名错误
	m["out_trade_no"] = charge.TradeNum
	m["subject"] = client.TruncatedText(charge.Describe, 32)
	m["total_fee"] = MoneyFeeToString(charge.MoneyFee)
	m["seller_id"] = a.SellerID

	sign := a.GenSign(m)

	m["sign"] = sign
	m["sign_type"] = "RSA"
	return map[string]string{"url": client.ToURL(a.PayURL, m)}, nil
}

func (a *Web) CloseOrder(charge *common.Charge) (map[string]string, error) {
	return map[string]string{}, errors.New("暂未开发该功能")
}

func (a *Web) PayToClient(charge *common.Charge) (map[string]string, error) {
	return map[string]string{}, errors.New("暂未开发该功能")
}

// QueryOrder 订单查询
func (a *Web) QueryOrder(outTradeNo string) (WebQueryResult, error) {
	var m = make(map[string]string)
	m["service"] = "single_trade_query"
	m["partner"] = a.PartnerID
	m["_input_charset"] = "utf-8"
	m["out_trade_no"] = outTradeNo

	sign := a.GenSign(m)

	m["sign"] = sign
	m["sign_type"] = "RSA"
	return Get(client.ToURL(a.PayURL, m))
}

// GenSign 产生签名
func (a *Web) GenSign(m map[string]string) string {
	var data []string
	for k, v := range m {
		if v != "" && k != "sign" && k != "sign_type" {
			data = append(data, fmt.Sprintf(`%s=%s`, k, v))
		}
	}
	sort.Strings(data)
	signData := strings.Join(data, "&")
	s := sha1.New()
	_, err := s.Write([]byte(signData))
	if err != nil {
		panic(err)
	}
	hashByte := s.Sum(nil)
	signByte, err := a.PrivateKey.Sign(rand.Reader, hashByte, crypto.SHA1)
	if err != nil {
		panic(err)
	}
	return url.QueryEscape(base64.StdEncoding.EncodeToString(signByte))
}

// CheckSign 检测签名
func (a *Web) CheckSign(signData, sign string) {
	signByte, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		panic(err)
	}
	s := sha1.New()
	_, err = s.Write([]byte(signData))
	if err != nil {
		panic(err)
	}
	hash := s.Sum(nil)
	err = rsa.VerifyPKCS1v15(a.PublicKey, crypto.SHA1, hash, signByte)
	if err != nil {
		panic(err)
	}
}
