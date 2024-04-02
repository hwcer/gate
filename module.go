package gate

import (
	"errors"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosnet/message"
	"github.com/hwcer/cosrpc/xclient"
	"github.com/hwcer/cosrpc/xserver"
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
	mux    cmux.CMux
	Server *server
	Socket *socket
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
func (this *Module) Start() error {
	ln, err := net.Listen("tcp", opt.Gate.Address)
	if err != nil {
		return err
	}
	if opt.Gate.Protocol.Has(options.ProtocolTypeALL) {
		this.mux = cmux.New(ln)
		sv := this.mux.Match(cmux.HTTP1Fast())
		if err = this.Server.Start(sv); err != nil {
			return err
		}
		so := this.mux.Match(message.Matcher())
		if err = this.Socket.Start(so); err != nil {
			return err
		}

		err = scc.Timeout(time.Second, func() error {
			return this.mux.Serve()
		})
		if errors.Is(err, scc.ErrorTimeout) {
			err = nil
		}
	} else if opt.Gate.Protocol.Has(options.ProtocolTypeTCP) {
		err = this.Socket.Start(ln)
	} else if opt.Gate.Protocol.Has(options.ProtocolTypeHTTP) {
		err = this.Server.Start(ln)
	}

	return err
}

func (this *Module) Close() (err error) {
	if opt.Gate.Protocol.Has(options.ProtocolTypeALL) {
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
