package gopay

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/webx-top/gopay/client"
	"github.com/webx-top/gopay/common"
	"github.com/webx-top/gopay/util"
)

func AlipayCallback(body *[]byte) (*common.AliPayResult, string, error) {
	var m = make(map[string]string)
	xml.Unmarshal(*body, &m)
	var signSlice []string

	for k, v := range m {
		if k == "sign" || k == "sign_type" {
			continue
		}
		signSlice = append(signSlice, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(signSlice)
	signData := strings.Join(signSlice, "&")
	if m["sign_type"] != "RSA2" {
		return nil, "error", errors.New("签名类型未知")
	}

	client.DefaultAliAppClient().CheckSign(signData, m["sign"])

	mByte, err := json.Marshal(m)
	if err != nil {
		return nil, "error", errors.New("error")
	}

	var aliPay common.AliPayResult
	err = json.Unmarshal(mByte, &aliPay)
	if err != nil {
		return nil, "error", fmt.Errorf("m is %v, err is %v", m, err)
	}
	return &aliPay, "SUCCESS", nil
}

func WeChatCallback(body *[]byte) (*common.WeChatPayResult, string, error) {
	var returnCode = "FAIL"
	var returnMsg = ""
	var reXML common.WeChatPayResult

	err := xml.Unmarshal(*body, &reXML)
	if err != nil {
		returnMsg = "参数错误"
		return nil, handleWechatCallbackResponse(returnCode, returnMsg), err
	}

	if reXML.ReturnCode != "SUCCESS" {
		returnMsg = "支付失败"
		return &reXML, handleWechatCallbackResponse(returnCode, returnMsg), errors.New(reXML.ReturnCode)
	}
	m := util.XmlToMap(*body)

	var signData []string
	for k, v := range m {
		if k == "sign" {
			continue
		}
		signData = append(signData, fmt.Sprintf("%v=%v", k, v))
	}

	key := client.DefaultWechatAppClient().Env.Key

	mySign, err := client.WechatGenSign(key, m)
	if err != nil {
		returnMsg = "签名失败"
		return &reXML, handleWechatCallbackResponse(returnCode, returnMsg), err
	}

	if mySign != m["sign"] {
		returnMsg = "签名失败"
		return &reXML, handleWechatCallbackResponse(returnCode, returnMsg), errors.New(returnMsg)
	}

	returnCode = "SUCCESS"
	return &reXML, handleWechatCallbackResponse(returnCode, returnMsg), nil
}

func handleWechatCallbackResponse(returnCode string, returnMsg string) string {
	formatStr := `<xml><return_code><![CDATA[%s]]></return_code>
                  <return_msg>![CDATA[%s]]</return_msg></xml>`
	returnBody := fmt.Sprintf(formatStr, returnCode, returnMsg)
	return returnBody
}
