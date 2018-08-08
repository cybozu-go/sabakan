package etcd

//
// import (
// 	"errors"
//
// 	"github.com/coreos/etcd/clientv3"
// )
//
// func (d *driver) getKernelParams() (string, error) {
// 	v := d.kernelParams.Load()
// 	if v == nil {
// 		return nil, errors.New("kernelParams is not set")
// 	}
//
// 	return v.(string), nil
// }
//
// func (d *driver) handleKernelParams(ev *clientv3.Event) error {
// 	return nil
// }
