package alipay

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/webx-top/gopay/client"
)

var defaultPayURL = `https://mapi.alipay.com/gateway.do`
var defaultOpenAPIURL = `https://openapi.alipay.com/gateway.do`

// Get 对支付宝者查订单
func Get(url string) (WebQueryResult, error) {
	var xmlRe WebQueryResult

	re, err := client.HTTPSC.GetData(url)
	if err != nil {
		return xmlRe, errors.New("HTTPSC.PostData: " + err.Error())
	}
	err = xml.Unmarshal(re, &xmlRe)
	if err != nil {
		return xmlRe, errors.New("xml.Unmarshal: " + err.Error())
	}
	return xmlRe, nil
}

// GetApp 对支付宝者查订单
func GetApp(urls string) (AppQueryResult, error) {
	var aliPay AppQueryResult

	re, err := client.HTTPSC.GetData(urls)
	if err != nil {
		return aliPay, errors.New("HTTPSC.PostData: " + err.Error())
	}

	err = json.Unmarshal(re, &aliPay)
	if err != nil {
		panic(fmt.Sprintf("re is %v, err is %v", re, err))
	}

	return aliPay, nil
}

// MoneyFeeToString 支付宝金额转字符串
func MoneyFeeToString(moneyFee float64) string {
	return decimal.NewFromFloat(moneyFee).Truncate(2).String()
}
