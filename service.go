package gate

import (
	"github.com/hwcer/cosnet"
	"github.com/hwcer/cosrpc/xclient"
	"github.com/hwcer/cosrpc/xshare"
	"github.com/hwcer/logger"
)

var Service = xclient.Service(Name)

func init() {
	RegisterPrivateService(send)
	RegisterPrivateService(broadcast)
}

func RegisterPrivateService(i any, prefix ...string) {
	if err := Service.Register(i, prefix...); err != nil {
		logger.Fatal("%v", err)
	}
}

func send(c *xshare.Context) any {
	uid := c.GetMetadata(Options.Metadata.UID)
	//logger.Debug("推送消息:%v  %v  %v", c.GetMetadata(rpcx.MetadataMessagePath), uid, string(c.Payload()))
	player := mod.Socket.Players.Get(uid)
	//sock := Sockets.Socket(uid)
	if player == nil {
		logger.Debug("用户不在线,消息丢弃:%v", uid)
		return nil
	}
	sock := player.Socket()
	if sock == nil {
		logger.Debug("用户不在线,消息丢弃:%v", uid)
		return nil
	}
	path := c.GetMetadata(Options.Gate.Name)
	if err := sock.Send(path, c.Bytes()); err != nil {
		logger.Debug("socket send error:%v", err)
	}
	return nil
}

func broadcast(c *xshare.Context) any {
	sid := c.GetMetadata(xshare.ServicesMetadataServerId)
	//logger.Debug("推送消息:%v  %v  %v", c.GetMetadata(rpcx.MetadataMessagePath), uid, string(c.Payload()))
	path := c.GetMetadata(Options.Gate.Name)

	mod.Socket.Broadcast(path, c.Bytes(), func(s *cosnet.Socket) bool {
		if p := s.Player(); p != nil && sid != "" && p.GetString("sid") == sid {
			return true
		} else {
			return false
		}
	})
	return nil
}
