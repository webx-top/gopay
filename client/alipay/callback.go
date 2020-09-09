package alipay

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"
)

func Callback(body *[]byte) (*PayResult, string, error) {
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

	DefaultApp().CheckSign(signData, m["sign"])

	mByte, err := json.Marshal(m)
	if err != nil {
		return nil, "error", errors.New("error")
	}

	var aliPay PayResult
	err = json.Unmarshal(mByte, &aliPay)
	if err != nil {
		return nil, "error", fmt.Errorf("m is %v, err is %v", m, err)
	}
	return &aliPay, "SUCCESS", nil
}
