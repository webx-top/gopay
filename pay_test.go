package gopay

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/webx-top/gopay/client/alipay"
	"github.com/webx-top/gopay/client/wechat"
	"github.com/webx-top/gopay/common"
)

func TestPay(t *testing.T) {
	initClient()
	initHandle()
	charge := new(common.Charge)
	charge.PayMethod = `alipay.web`
	charge.MoneyFee = 1
	charge.Describe = "test pay"
	charge.TradeNum = "11111111122"
	charge.CallbackURL = "http://127.0.0.1/callback/aliappcallback"

	fdata, err := Pay(charge)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v", fdata)
}

func initClient() {

	// 私钥
	block, _ := pem.Decode([]byte(`-----BEGIN PRIVATE KEY-----
xxxxxxx
-----END PRIVATE KEY-----`))

	if block == nil {
		panic("Sign private key decode error")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	// 公钥
	block, _ = pem.Decode([]byte(`-----BEGIN PUBLIC KEY-----
xxxxxxxx
-----END PUBLIC KEY-----`))

	if block == nil {
		panic("Sign public key decode error")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	web := alipay.NewWeb()
	web.PartnerID = "xxxxxxxxxxxx"
	web.SellerID = "xxxxxxxxxxxx"
	web.AppID = "xxxxxxxxxxxx"
	web.PrivateKey = privateKey.(*rsa.PrivateKey)
	web.PublicKey = publicKey.(*rsa.PublicKey)
	alipay.InitWeb(web)
}

func TestWechatPay(t *testing.T) {
	initWechatClient()
	initHandle()
	charge := new(common.Charge)
	charge.PayMethod = `wechat.web`
	charge.MoneyFee = 1
	charge.Describe = "test pay"
	charge.TradeNum = "11111111122"
	charge.CallbackURL = "http://127.0.0.1/callback/aliappcallback"
	charge.OpenID = "123"

	fdata, err := Pay(charge)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v", fdata)
}

func initWechatClient() {
	wechat.InitWeb(&wechat.ClientData{
		AppID:  "xxxxxxxxxxxx",
		MchID:  "xxxxxxxxxxxx",
		Key:    "xxxxxxxxxxxx",
		PayURL: "https://api.mch.weixin.qq.com/pay/unifiedorder",
	})
}

func initHandle() {
	http.HandleFunc("callback/aliappcallback", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		result, echo, err := alipay.Callback(&b)
		if err != nil {
			fmt.Println(err)
			return
		}
		selfHandler(result)
		w.Write([]byte(echo))
	})
}

func selfHandler(i interface{}) {
}
