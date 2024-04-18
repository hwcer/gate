package gate

import (
	"errors"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosrpc/xclient"
	"github.com/hwcer/cosrpc/xserver"
	"github.com/hwcer/coswss"
	"github.com/hwcer/gate/options"
	"github.com/hwcer/scc"
	"github.com/soheilhy/cmux"
	"net"
	"strings"
	"time"
)

var mod = &Module{Module: cosgo.Module{Id: options.Name}}
var opt = options.Options

func New() cosgo.IModule {
	return mod
}

type Module struct {
	cosgo.Module
	mux       cmux.CMux
	Server    *server
	Socket    *socket
	WebSocket *coswss.Server
}

func (this *Module) Init() (err error) {
	if err = cosgo.Config.Unmarshal(opt); err != nil {
		return
	}
	if opt.Gate.Address == "" {
		return errors.New("网关地址没有配置")
	}
	if i := strings.Index(opt.Gate.Address, ":"); i < 0 {
		return errors.New("网关地址配置错误,格式: ip:port")
	} else if opt.Gate.Address[0:i] == "" {
		opt.Gate.Address = "0.0.0.0" + opt.Gate.Address
	}
	opt.Rpcx.BasePath = opt.Appid

	if opt.Gate.Broadcast == 1 {
		s := xclient.Service(options.Name)
		if err = s.Merge(Service()); err != nil {
			return err
		}
	} else if opt.Gate.Broadcast == 2 {
		service := xserver.Service(options.Name)
		if err = service.Merge(Service()); err != nil {
			return
		}
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
	if opt.Gate.Protocol.CMux() {
		var ln net.Listener
		if ln, err = net.Listen("tcp", opt.Gate.Address); err != nil {
			return err
		}
		this.mux = cmux.New(ln)
	}
	p := opt.Gate.Protocol
	// websocket
	if p.Has(options.ProtocolTypeWSS) {
		if p.Has(options.ProtocolTypeTCP) {
			this.WebSocket, err = coswss.New(this.Socket.Server)
		} else {
			this.WebSocket, err = coswss.New(nil)
		}
		if p.Has(options.ProtocolTypeHTTP) {
			this.WebSocket.Binding(this.Server.Server, options.Options.Gate.Websocket)
		} else {
			err = this.WebSocket.Start(opt.Gate.Address)
		}
	}
	//SOCKET
	if p.Has(options.ProtocolTypeTCP) {
		if this.mux != nil {
			so := this.mux.Match(this.Socket.Matcher())
			err = this.Socket.Listen(so)
		} else {
			err = this.Socket.Start(opt.Gate.Address)
		}
		if err != nil {
			return err
		}
	}

	if p.Has(options.ProtocolTypeHTTP) {
		if this.mux != nil {
			so := this.mux.Match(cmux.HTTP1Fast())
			err = this.Server.Listen(so)
		} else {
			err = this.Server.Start(opt.Gate.Address)
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
		_ = this.Server.Close()
		this.mux.Close()
	} else if opt.Gate.Protocol.Has(options.ProtocolTypeTCP) {
		err = this.Socket.Close()
	} else if opt.Gate.Protocol.Has(options.ProtocolTypeHTTP) {
		err = this.Server.Close()
	}
	if err != nil {
		return err
	}
	if err = xclient.Close(); err != nil {
		return
	}
	if opt.Gate.Broadcast == 2 {
		err = xserver.Close()
	}
	return
}
