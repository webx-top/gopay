package wechat

import (
	"github.com/webx-top/gopay/client"
	"github.com/webx-top/gopay/common"
)

func init() {
	client.Register(`wechat.web`, func() common.PayClient {
		return DefaultWeb()
	})
	client.Register(`wechat.app`, func() common.PayClient {
		return DefaultApp()
	})
	client.Register(`wechat.native`, func() common.PayClient {
		return DefaultNative()
	})
	client.Register(`wechat.miniProgram`, func() common.PayClient {
		return DefaultMiniProgram()
	})
}
