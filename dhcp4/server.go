package dhcp4

import (
	"context"
	"fmt"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"go.universe.tf/netboot/dhcp4"
)

type Server struct {
	Handler Handler
	Conn    *dhcp4.Conn
}

func (s *Server) Serve(ctx context.Context) error {
	env := cmd.NewEnvironment(ctx)
	for {
		pkt, intf, err := s.Conn.RecvDHCP()
		if err != nil {
			return fmt.Errorf("Receiving DHCP packet: %s", err)
		}
		if intf == nil {
			return fmt.Errorf("Received DHCP packet with no interface information (this is a violation of dhcp4.Conn's contract, please file a bug)")
		}

		env.Go(func(ctx context.Context) error {
			resp, err := s.Handler.handleDHCP(ctx, pkt, intf)
			if err != nil {
				log.Error("handler returns an error", map[string]interface{}{
					log.FnError: err.Error(),
				})
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
