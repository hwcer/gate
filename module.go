package gate

import (
	"errors"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosnet/message"
	"github.com/hwcer/cosrpc/xclient"
	"github.com/hwcer/cosrpc/xserver"
	"github.com/hwcer/cosrpc/xshare"
	"github.com/hwcer/scc"
	"github.com/soheilhy/cmux"
	"net"
	"strings"
	"time"
)

var mod = &module{}

func init() {
	mod.Id = "gate"
}

func New() cosgo.IModule {
	return mod
}

type module struct {
	cosgo.Module
	mux    cmux.CMux
	Server *server
	Socket *socket
}

func (this *module) Init() (err error) {
	if err = cosgo.Config.Unmarshal(Options); err != nil {
		return
	}
	if Options.Gate.Address == "" {
		return errors.New("网关地址没有配置")
	}
	if i := strings.Index(Options.Gate.Address, ":"); i < 0 {
		return errors.New("网关地址配置错误,格式: ip:port")
	} else if Options.Gate.Address[0:i] == "" {
		Options.Gate.Address = "0.0.0.0" + Options.Gate.Address
	}
	//加载RPCX配置
	if err = cosgo.Config.Unmarshal(xshare.Options); err != nil {
		return
	}
	if Options.Gate.Broadcast == 1 {
		s := xclient.Service(Options.Gate.Name)
		if err = s.Merge(Service()); err != nil {
			return err
		}
	} else if Options.Gate.Broadcast == 2 {
		service := xserver.Service(Options.Gate.Name)
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
func (this *module) Start() error {
	ln, err := net.Listen("tcp", Options.Gate.Address)
	if err != nil {
		return err
	}
	if Options.Gate.Protocol.Has(protocolTypeALL) {
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
	} else if Options.Gate.Protocol.Has(protocolTypeTCP) {
		err = this.Socket.Start(ln)
	} else if Options.Gate.Protocol.Has(protocolTypeHTTP) {
		err = this.Server.Start(ln)
	}

	return err
}

func (this *module) Close() (err error) {
	if Options.Gate.Protocol.Has(protocolTypeALL) {
		_ = this.Socket.Close()
		_ = this.Server.Close()
		this.mux.Close()
	} else if Options.Gate.Protocol.Has(protocolTypeTCP) {
		err = this.Socket.Close()
	} else if Options.Gate.Protocol.Has(protocolTypeHTTP) {
		err = this.Server.Close()
	}
	if err != nil {
		return err
	}
	if err = xclient.Close(); err != nil {
		return
	}
	if Options.Gate.Broadcast == 2 {
		err = xserver.Close()
	}
	return
}
