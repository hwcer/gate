package gate

import (
	"errors"
	"github.com/hwcer/cosgo/logger"
	"github.com/hwcer/cosgo/options"
	"github.com/hwcer/cosgo/session"
	"github.com/hwcer/cosgo/utils"
	"github.com/hwcer/cosnet"
	"github.com/hwcer/cosnet/tcp"
	"github.com/hwcer/gate/players"
	"net"
	"net/url"
	"strconv"
	"time"
)

func init() {
	srv := &socket{}
	srv.Server = cosnet.New(nil)
	cosnet.Options.SocketConnectTime = 1000 * 3600
	service := srv.Server.Service("")
	//_ = service.Register(srv.ping, "ping")
	//_ = Sockets.Register(Socket.login)
	_ = service.Register(srv.proxy, "/*")

	srv.Server.On(cosnet.EventTypeError, srv.Errorf)
	srv.Server.On(cosnet.EventTypeConnected, srv.Connected)
	srv.Server.On(cosnet.EventTypeDisconnect, srv.Disconnect)
	srv.Server.On(cosnet.EventTypeDestroyed, srv.Destroyed)

	mod.Socket = srv
}

type socket struct {
	*cosnet.Server
}

//var socketSerialize cosnet.HandlerSerialize = func(c *cosnet.Context, reply interface{}) (any, error) {
//	if v, ok := reply.([]byte); ok && string(v) == "null" {
//		return nil, nil
//	}
//	return reply, nil
//}

func (this *socket) Start(address string) error {
	addr := utils.NewAddress(address)
	if addr.Scheme == "" {
		addr.Scheme = "tcp"
	}
	ln, err := net.Listen(addr.Scheme, addr.String())
	if err == nil {
		err = this.Listen(ln)
	}
	return err
}
func (this *socket) Listen(ln net.Listener) error {
	this.Server.Accept(&tcp.Listener{Listener: ln})
	logger.Trace("网关长连接启动：%v", options.Gate.Address)
	return nil
}

func (this *socket) Errorf(socket *cosnet.Socket, err interface{}) bool {
	logger.Alert(err)
	return false
}

func (this *socket) ping(c *cosnet.Context) interface{} {
	var s string
	_ = c.Bind(&s)
	return []byte(strconv.Itoa(int(time.Now().Unix())))
}

func (this *socket) proxy(c *cosnet.Context) interface{} {
	urlPath, err := url.Parse(c.Path())
	if err != nil {
		return c.Errorf(0, err)
	}

	//logger.Trace("socket request,PATH:%v   BODY:%v", urlPath.String(), string(c.Message.Body()))
	req, res, err := metadata(urlPath.RawQuery)
	if err != nil {
		return c.Errorf(0, err)
	}
	p, _ := c.Data.Get().(*session.Data)
	path := Formatter(urlPath.Path)
	limit := options.Apis.Get(path)
	if limit != options.ApisTypeNone {
		if p == nil {
			return c.Errorf(0, "not login")
		}
		if limit == options.ApisTypeOAuth {
			req[options.ServiceMetadataGUID] = p.GetString(options.ServiceMetadataGUID)
		} else {
			req[options.ServiceMetadataUID] = p.GetString(options.ServiceMetadataUID)
		}
	}

	reply := make([]byte, 0)
	if p == nil {
		err = request(nil, path, c.Message.Body(), req, res, &reply)
	} else {
		err = request(p, path, c.Message.Body(), req, res, &reply)
	}
	if err != nil {
		//logger.Trace("socket response error:%v,PATH:%v   Error:%v", path, err)
		return err
	}
	//logger.Trace("socket response,PATH:%v   BODY:%v", path, string(reply))
	if err = this.setCookie(c, res); err != nil {
		return c.Error(err)
	}
	if len(reply) == 0 {
		return nil
	}
	return reply
}

func (this *socket) setCookie(c *cosnet.Context, cookie options.Metadata) (err error) {
	if len(cookie) == 0 {
		return
	}
	var s *session.Data
	if i := c.Socket.Get(); i != nil {
		s = i.(*session.Data)
	}
	//账号登录
	if guid, ok := cookie[options.ServicePlayerOAuth]; ok {
		_, err = players.Binding(c.Socket, guid, CookiesFilter(cookie))
	} else if _, ok = cookie[options.ServicePlayerLogout]; ok {
		players.Delete(s)
		c.Socket.Close()
	} else if s != nil {
		CookiesUpdate(cookie, s)
	} else {
		return errors.New("not login")
	}

	return
}

func (this *socket) Connected(sock *cosnet.Socket, _ interface{}) bool {
	logger.Debug("Connected:%v", sock.Id())
	return true
}

func (this *socket) Disconnect(sock *cosnet.Socket, _ interface{}) bool {
	logger.Debug("Disconnect:%v", sock.Id())
	return true
}

func (this *socket) Destroyed(sock *cosnet.Socket, _ interface{}) bool {
	logger.Debug("Destroyed:%v", sock.Id())
	p, _ := sock.Get().(*session.Data)
	if p == nil {
		return true
	}
	s := players.Players.Socket(p)
	if s != nil && s.Id() == sock.Id() {
		players.Players.Delete(p)
		//_ = share.Notify.Publish(share.NotifyChannelSocketDestroyed, uid)
	}
	return true
}
