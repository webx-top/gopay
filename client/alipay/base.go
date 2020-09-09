package alipay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/webx-top/gopay/client"
	"github.com/webx-top/gopay/common"
)

type PayClient struct {
	AppID string // 应用ID

	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey

	RSAType    string // RSA or RSA2
	OpenAPIURL string
	PayURL     string
}

func New() *PayClient {
	return &PayClient{
		OpenAPIURL: defaultOpenAPIURL,
		PayURL:     defaultPayURL,
	}
}

func (a *PayClient) PayToClient(charge *common.Charge) (map[string]string, error) {
	var result = make(map[string]string)
	var m = make(map[string]string)
	var bizContent = make(map[string]string)
	m["app_id"] = a.AppID
	m["method"] = "alipay.fund.trans.toaccount.transfer"
	m["format"] = "JSON"
	m["charset"] = "utf-8"
	m["timestamp"] = time.Now().Format("2006-01-02 15:04:05")
	m["version"] = "1.0"
	m["sign_type"] = a.RSAType

	bizContent["out_biz_no"] = charge.TradeNum
	bizContent["amount"] = MoneyFeeToString(charge.MoneyFee)
	bizContent["payee_account"] = charge.AliAccount
	bizContent["payee_type"] = charge.AliAccountType

	bizContent["remark"] = client.TruncatedText(charge.Describe, 32)

	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return result, errors.New("json.Marshal: " + err.Error())
	}
	m["biz_content"] = string(bizContentJSON)

	m["sign"] = a.GenSign(m)

	requestURL := fmt.Sprintf("%s?%s", a.OpenAPIURL, a.ToURL(m))

	var resp map[string]interface{}
	bytes, err := client.HTTPSC.GetData(requestURL)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(bytes, &resp)
	if err != nil {
		return result, err
	}

	result, ok := resp["alipay_fund_trans_toaccount_transfer_response"].(map[string]string)
	if !ok {
		return result, fmt.Errorf("返回结果错误:%s", resp)
	}

	return result, nil
}

// GenSign 产生签名
func (a *PayClient) GenSign(m map[string]string) string {
	var data []string

	for k, v := range m {
		if v != "" && k != "sign" {
			data = append(data, fmt.Sprintf(`%s=%s`, k, v))
		}
	}
	sort.Strings(data)
	signData := strings.Join(data, "&")

	s := a.getHash(a.RSAType)

	_, err := s.Write([]byte(signData))
	if err != nil {
		panic(err)
	}
	hashByte := s.Sum(nil)
	signByte, err := a.PrivateKey.Sign(rand.Reader, hashByte, crypto.SHA256)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(signByte)
}

// CheckSign 检测签名
func (a *PayClient) CheckSign(signData, sign string) {
	signByte, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		panic(err)
	}
	s := a.getHash(a.RSAType)
	_, err = s.Write([]byte(signData))
	if err != nil {
		panic(err)
	}
	hashByte := s.Sum(nil)
	err = rsa.VerifyPKCS1v15(a.PublicKey, a.getCrypto(), hashByte, signByte)
	if err != nil {
		panic(err)
	}
}

// ToURL 构建网址参数
func (a *PayClient) ToURL(m map[string]string) string {
	var buf []string
	for k, v := range m {
		buf = append(buf, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
	}
	return strings.Join(buf, "&")
}

func (a *PayClient) getRsa() string {
	if a.RSAType == "" {
		a.RSAType = "RSA"
	}

	return a.RSAType
}

func (a *PayClient) getCrypto() crypto.Hash {
	if a.RSAType == "RSA2" {
		return crypto.SHA256
	}
	return crypto.SHA1
}

func (a *PayClient) getHash(rasType string) hash.Hash {
	if rasType == "RSA2" {
		return sha256.New()
	}
	return sha1.New()
}
