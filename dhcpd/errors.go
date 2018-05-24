package dhcpd

import "errors"

var (
	errUnknownMsgType = errors.New("unknown message type")
	errNotChosen      = errors.New("not chosen")
	errNoRecord       = errors.New("no record of the client")
	errNoAction       = errors.New("no need to reply")
)
