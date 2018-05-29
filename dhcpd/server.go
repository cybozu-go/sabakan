package dhcpd

import (
	"context"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"go.universe.tf/netboot/dhcp4"
)

// Server is DHCP4 server.
type Server struct {
	Handler Handler
	Conn    *dhcp4.Conn
}

// Serve runs until context is canceled.
//
// Once ctx is canceled, s.Conn will be closed.
func (s Server) Serve(ctx context.Context) error {
	env := cmd.NewEnvironment(ctx)
	env.Go(func(ctx context.Context) error {
		<-ctx.Done()
		return s.Conn.Close()
	})

	for {
		pkt, intf, err := s.Conn.RecvDHCP()
		if err != nil {
			log.Error("RecvDHCP returns an error, exiting", map[string]interface{}{
				log.FnError: err.Error(),
			})
			break
		}
		if intf == nil {
			log.Error("received DHCP packet with no interface information (this is a violation of dhcp4.Conn's contract, please file a bug)", nil)
			continue
		}
		log.Info("dhcp: received", getPacketLog(pkt, intf))
		log.Debug("dhcp: options", getOptionsLog(pkt))

		wrappedIntf := nativeInterface{intf}

		env.Go(func(ctx context.Context) error {
			resp, err := s.Handler.ServeDHCP(ctx, pkt, wrappedIntf)
			switch err {
			case errNotChosen, errNoRecord, errNoAction:
				// do nothing
				return nil
			case errUnknownMsgType:
				// already logged
				return nil
			case nil:
				// continue to SendDHCP
			default:
				log.Error("handler returns an error", map[string]interface{}{
					log.FnError: err.Error(),
				})
				return nil
			}

			log.Info("dhcp: sending", getPacketLog(resp, intf))
			log.Debug("dhcp: options", getOptionsLog(resp))
			err = s.Conn.SendDHCP(resp, intf)
			if err != nil {
				log.Error("SendDHCP returns an error", map[string]interface{}{
					log.FnError: err.Error(),
				})
			}
			return nil
		})
	}

	env.Stop()
	return env.Wait()
}
