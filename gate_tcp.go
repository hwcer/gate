package gate

import (
	"errors"
	"github.com/hwcer/cosgo/logger"
	"github.com/hwcer/cosgo/session"
	"github.com/hwcer/cosnet"
	"github.com/hwcer/cosnet/tcp"
	"github.com/hwcer/gate/players"
	"github.com/hwcer/wower/options"
	"github.com/hwcer/wower/share"
	"net"
	"net/url"
	"strconv"
	"time"
)

type Socket struct {
}

func (this *Socket) init() error {
	service := cosnet.Service("")
	_ = service.Register(this.proxy, "/*")
	cosnet.On(cosnet.EventTypeError, this.Errorf)
	cosnet.On(cosnet.EventTypeConnected, this.Connected)
	cosnet.On(cosnet.EventTypeDisconnect, this.Disconnect)
	cosnet.On(cosnet.EventTypeReleased, this.Destroyed)
	return nil
}

//var socketSerialize cosnet.HandlerSerialize = func(c *cosnet.Context, reply interface{}) (any, error) {
//	if v, ok := reply.([]byte); ok && string(v) == "null" {
//		return nil, nil
//	}
//	return reply, nil
//}

func (this *Socket) Listen(address string) error {
	_, err := cosnet.Listen(address)
	if err == nil {
		logger.Trace("网关长连接启动：%v", options.Gate.Address)
	}
	return err
}

func (this *Socket) Accept(ln net.Listener) error {
	cosnet.Accept(&tcp.Listener{Listener: ln})
	logger.Trace("网关长连接启动：%v", options.Gate.Address)
	return nil
}

func (this *Socket) Errorf(socket *cosnet.Socket, err interface{}) bool {
	logger.Alert(err)
	return false
}

func (this *Socket) ping(c *cosnet.Context) interface{} {
	var s string
	_ = c.Bind(&s)
	return []byte(strconv.Itoa(int(time.Now().Unix())))
}

func (this *Socket) proxy(c *cosnet.Context) interface{} {
	urlPath, err := url.Parse(c.Path())
	if err != nil {
		return c.Errorf(0, err)
	}

	//logger.Trace("Socket request,PATH:%v   BODY:%v", urlPath.String(), string(c.Message.Body()))
	req, res, err := metadata(urlPath.RawQuery)
	if err != nil {
		return c.Errorf(0, err)
	}
	var p *session.Data
	//p, _ := c.Socket.Get().(*session.Data)
	path := Formatter(urlPath.Path)
	limit := share.Authorizes.Get(path)
	if limit != share.AuthorizesTypeNone {
		if c.Socket.GetStatus() != cosnet.StatusTypeOAuth {
			return c.Errorf(0, "not login")
		}
		p = c.Socket.Data()
		if limit == share.AuthorizesTypeOAuth {
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
		//logger.Trace("Socket response error:%v,PATH:%v   Error:%v", path, err)
		return err
	}
	//logger.Trace("Socket response,PATH:%v   BODY:%v", path, string(reply))
	if err = this.setCookie(c, res); err != nil {
		return c.Error(err)
	}
	if len(reply) == 0 {
		return nil
	}
	return reply
}

func (this *Socket) setCookie(c *cosnet.Context, cookie options.Metadata) (err error) {
	if len(cookie) == 0 {
		return
	}
	s := c.Socket.Data()
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

func (this *Socket) Connected(sock *cosnet.Socket, _ interface{}) bool {
	logger.Debug("Connected:%v", sock.Id())
	return true
}

func (this *Socket) Disconnect(sock *cosnet.Socket, _ interface{}) bool {
	logger.Debug("Disconnect:%v", sock.Id())
	return true
}

func (this *Socket) Destroyed(sock *cosnet.Socket, _ interface{}) bool {
	logger.Debug("Destroyed:%v", sock.Id())
	p := sock.Data()
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
