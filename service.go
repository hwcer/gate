package gate

import (
	"github.com/hwcer/cosnet"
	"github.com/hwcer/cosrpc/xshare"
	"github.com/hwcer/cosweb/session"
	"github.com/hwcer/gate/options"
	"github.com/hwcer/logger"
	"github.com/hwcer/registry"
	"strings"
)

var rs = registry.New(nil)

func init() {
	Register(send)
	Register(broadcast)
}

func Service() *registry.Service {
	return rs.Service(options.Name)
}

// Register 注册协议，用于服务器推送消息
func Register(i any, prefix ...string) {
	s := Service()
	if err := s.Register(i, prefix...); err != nil {
		logger.Fatal("%v", err)
	}
}

func send(c *xshare.Context) any {
	uid := c.GetMetadata(opt.Metadata.UID)
	//logger.Debug("推送消息:%v  %v  %v", c.GetMetadata(rpcx.MetadataMessagePath), uid, string(c.Payload()))
	p := Players.Get(uid)
	//sock := Sockets.Socket(uid)
	if p == nil {
		logger.Debug("用户不在线,消息丢弃:%v", uid)
		return nil
	}
	sock := Players.Socket(p)
	if sock == nil {
		logger.Debug("用户不在线,消息丢弃:%v", uid)
		return nil
	}
	//注册消息
	for k, v := range c.Metadata() {
		if strings.HasPrefix(k, options.ServiceAddressPrefix) {
			p.Set(k, v)
		}
	}

	path := c.GetMetadata(opt.Metadata.API)
	if err := sock.Send(path, c.Bytes()); err != nil {
		logger.Debug("socket send error:%v", err)
	}
	return nil
}

func broadcast(c *xshare.Context) any {
	sid := c.GetMetadata(xshare.ServicesMetadataRpcServerId)
	//logger.Debug("推送消息:%v  %v  %v", c.GetMetadata(rpcx.MetadataMessagePath), uid, string(c.Payload()))
	path := c.GetMetadata(opt.Metadata.API)

	mod.Socket.Broadcast(path, c.Bytes(), func(s *cosnet.Socket) bool {
		p, _ := s.Data.Get().(*session.Player)
		if p != nil && sid != "" && p.GetString("sid") == sid {
			return false
		} else {
			return false
		}
	})
	return nil
}
