package wechat

import (
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/webx-top/gopay/util"
)

func (c *ClientData) Callback(body *[]byte) (*PayResult, string, error) {
	var returnCode = "FAIL"
	var returnMsg string
	var reXML PayResult

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

	key := c.Key

	mySign, err := GenSign(key, m)
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
