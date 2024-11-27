package gate

import (
	"github.com/hwcer/cosgo/options"
	"github.com/hwcer/cosnet"
	"github.com/hwcer/cosrpc/xshare"
	"github.com/hwcer/cosweb/session"
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
	return rs.Service(options.ServiceTypeGate)
}

// Register 注册协议，用于服务器推送消息
func Register(i any, prefix ...string) {
	s := Service()
	if err := s.Register(i, prefix...); err != nil {
		logger.Fatal("%v", err)
	}
}

func send(c *xshare.Context) any {
	uid := c.GetMetadata(options.ServiceMetadataUID)
	guid := c.GetMetadata(options.ServiceMetadataGUID)
	//logger.Debug("推送消息:%v  %v  %v", c.GetMetadata(rpcx.MetadataMessagePath), uid, string(c.Payload()))
	p := Players.Get(guid)
	//sock := Sockets.Socket(uid)
	if p == nil {
		logger.Debug("用户不在线,消息丢弃,UID:%v GUID:%v", uid, guid)
		return nil
	}
	if id := p.GetString(options.ServiceMetadataUID); id != uid {
		return nil
	}
	//注册消息
	for k, v := range c.Metadata() {
		if strings.HasPrefix(k, options.ServiceSelectorPrefix) || strings.HasPrefix(k, options.PlayerMessageChannel) {
			p.Set(k, v)
		}
	}

	path := c.GetMetadata(options.ServiceMetadataApi)
	if len(path) == 0 {
		return nil //仅仅设置信息，不需要发送
	}
	sock := Players.Socket(p)
	if sock == nil {
		logger.Debug("用户不在线,消息丢弃:%v", uid)
		return nil
	}
	if err := sock.Send(path, c.Bytes()); err != nil {
		logger.Debug("socket send error:%v", err)
	}
	return nil
}

func broadcast(c *xshare.Context) any {
	sid := c.GetMetadata(options.ServiceMetadataServerId)
	//logger.Debug("推送消息:%v  %v  %v", c.GetMetadata(rpcx.MetadataMessagePath), uid, string(c.Payload()))
	path := c.GetMetadata(options.ServiceMetadataApi)

	mod.Socket.Broadcast(path, c.Bytes(), func(s *cosnet.Socket) bool {
		p, _ := s.Data.Get().(*session.Player)
		if p != nil && sid != "" && p.GetString(options.ServiceMetadataServerId) == sid {
			return false
		} else {
			return false
		}
	})
	return nil
}
