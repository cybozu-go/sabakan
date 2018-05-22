package dhcpd

import (
	"context"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"go.universe.tf/netboot/dhcp4"
)

// Server is dhcp server
type Server struct {
	Handler Handler
	Conn    *dhcp4.Conn
}

// Serve runs until context is canceled
func (s *Server) Serve(ctx context.Context) error {
	env := cmd.NewEnvironment(ctx)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		pkt, intf, err := s.Conn.RecvDHCP()
		if err != nil {
			log.Error("receiving malformed DHCP packet", map[string]interface{}{
				log.FnError: err.Error(),
			})
			continue
		}
		if intf == nil {
			log.Error("received DHCP packet with no interface information (this is a violation of dhcp4.Conn's contract, please file a bug)", nil)
			continue
		}

		env.Go(func(ctx context.Context) error {
			resp, err := s.Handler.ServeDHCP(ctx, pkt, intf)
			if err != nil {
				log.Error("handler returns an error", map[string]interface{}{
					log.FnError: err.Error(),
				})
				return nil
			}
			err = s.Conn.SendDHCP(resp, intf)
			if err != nil {
				log.Error("handler returns an error", map[string]interface{}{
					log.FnError: err.Error(),
				})
			}
			return nil
		})
	}
}
