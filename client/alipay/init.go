package alipay

import (
	"github.com/webx-top/gopay/client"
	"github.com/webx-top/gopay/common"
)

func init() {
	client.Register(`alipay.web`, func() common.PayClient {
		return DefaultWeb()
	})
	client.Register(`alipay.app`, func() common.PayClient {
		return DefaultApp()
	})
}
