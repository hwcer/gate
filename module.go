package gate

import (
	"errors"
	"github.com/hwcer/cosgo/options"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/cosrpc/xclient"
	"github.com/hwcer/cosrpc/xserver"
	"github.com/hwcer/coswss"
	"github.com/soheilhy/cmux"
	"net"
	"strings"
	"time"
)

var mod = &Module{}

func New() *Module {
	return mod
}

type Module struct {
	mux       cmux.CMux
	Server    *server
	Socket    *socket
	WebSocket *coswss.Server
}

func (this *Module) Id() string {
	return options.ServiceTypeGate
}

func (this *Module) Init() (err error) {
	if err = options.Initialize(); err != nil {
		return
	}
	if options.Gate.Address == "" {
		return errors.New("网关地址没有配置")
	}
	if i := strings.Index(options.Gate.Address, ":"); i < 0 {
		return errors.New("网关地址配置错误,格式: ip:port")
	} else if options.Gate.Address[0:i] == "" {
		options.Gate.Address = "0.0.0.0" + options.Gate.Address
	}
	if options.Rpcx.Redis != "" {
		if err = xserver.Start(); err != nil {
			return err
		}
	}
	if err = xclient.Start(); err != nil {
		return err
	}

	return nil
}

func (this *Module) Start() (err error) {
	if options.Gate.Protocol.CMux() {
		var ln net.Listener
		if ln, err = net.Listen("tcp", options.Gate.Address); err != nil {
			return err
		}
		this.mux = cmux.New(ln)
	}
	p := options.Gate.Protocol
	// websocket
	if p.Has(options.ProtocolTypeWSS) {
		this.WebSocket, err = coswss.New(this.Socket.Server)
		this.WebSocket.Binding(this.Server.Server, options.Options.Gate.Websocket)
		//this.Server.Server.Register(options.Options.Gate.Websocket, this.WebSocket.Handle)
		this.WebSocket.Verify = WSVerify
		this.WebSocket.Accept = WSAccept
	}
	//SOCKET
	if p.Has(options.ProtocolTypeTCP) {
		if this.mux != nil {
			so := this.mux.Match(this.Socket.Matcher())
			err = this.Socket.Listen(so)
		} else {
			err = this.Socket.Start(options.Gate.Address)
		}
		if err != nil {
			return err
		}
	}

	if p.Has(options.ProtocolTypeHTTP) || p.Has(options.ProtocolTypeWSS) {
		if this.mux != nil {
			so := this.mux.Match(cmux.HTTP1Fast())
			err = this.Server.Listen(so)
		} else {
			err = this.Server.Start(options.Gate.Address)
		}
		if err != nil {
			return err
		}
	}

	if this.mux != nil {
		if err = scc.Timeout(time.Second, func() error { return this.mux.Serve() }); errors.Is(err, scc.ErrorTimeout) {
			err = nil
		}
	}

	return err
}

func (this *Module) Close() (err error) {
	if this.mux != nil {
		_ = this.Socket.Close()
		this.mux.Close()
	} else if options.Gate.Protocol.Has(options.ProtocolTypeTCP) {
		err = this.Socket.Close()
	} else if options.Gate.Protocol.Has(options.ProtocolTypeHTTP) {
		//err = this.Server.Close()
	}
	if err != nil {
		return err
	}
	if err = xclient.Close(); err != nil {
		return
	}
	if options.Rpcx.Redis != "" {
		err = xserver.Close()
	}
	return
}
