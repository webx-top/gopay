package gopay

import (
	"errors"
	"fmt"

	"github.com/webx-top/gopay/client"
	"github.com/webx-top/gopay/common"
)

var (
	// ErrUnsupported 不支持的付款方式
	ErrUnsupported = errors.New("unsupported")
	// ErrInstanceIsNotInitialized 实例未初始化
	ErrInstanceIsNotInitialized = errors.New("instance is not initialized")
)

// IsUnsupported 是否无法开启支付
func IsUnsupported(err error) bool {
	return errors.Is(err, ErrUnsupported) || errors.Is(err, ErrInstanceIsNotInitialized)
}

// Pay 用户下单支付接口
func Pay(charge *common.Charge) (map[string]string, error) {
	engine := client.Get(charge.PayMethod)
	if engine == nil {
		return nil, fmt.Errorf("[gopay]%w: %s", ErrUnsupported, charge.PayMethod)
	}
	obj := engine()
	if obj == nil {
		return nil, fmt.Errorf("[gopay]%w: %s", ErrInstanceIsNotInitialized, charge.PayMethod)
	}
	return obj.Pay(charge)
}

// PayToClient 付款给用户接口
func PayToClient(charge *common.Charge) (map[string]string, error) {
	engine := client.Get(charge.PayMethod)
	if engine == nil {
		return nil, fmt.Errorf("[gopay]%w: %s", ErrUnsupported, charge.PayMethod)
	}
	obj := engine()
	if obj == nil {
		return nil, fmt.Errorf("[gopay]%w: %s", ErrInstanceIsNotInitialized, charge.PayMethod)
	}
	return obj.PayToClient(charge)
}
