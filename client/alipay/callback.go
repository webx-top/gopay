package alipay

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"
)

func (c *PayClient) Callback(body *[]byte) (interface{}, string, error) {
	var aliPay PayResult
	var m = make(map[string]string)
	if err := xml.Unmarshal(*body, &m); err != nil {
		return aliPay, "", err
	}
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
		return aliPay, "error", errors.New("签名类型未知")
	}

	c.CheckSign(signData, m["sign"])

	mByte, err := json.Marshal(m)
	if err != nil {
		return aliPay, "error", errors.New("error")
	}

	err = json.Unmarshal(mByte, &aliPay)
	if err != nil {
		return aliPay, "error", fmt.Errorf("m is %v, err is %v", m, err)
	}
	return aliPay, "SUCCESS", nil
}
