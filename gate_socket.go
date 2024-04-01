package gate

import (
	"errors"
	"github.com/hwcer/cosnet"
	"github.com/hwcer/logger"
	"net"
	"net/url"
	"strconv"
	"time"
)

//var Socket = &socket{}
//var Sockets = cosnet.New()

func init() {
	srv := &socket{}
	srv.Server = cosnet.New(nil)
	cosnet.Options.SocketConnectTime = 1000 * 3600
	service := srv.Server.Service("")
	_ = service.Register(srv.ping, "ping")
	//_ = Sockets.Register(Socket.login)
	_ = service.Register(srv.proxy, "/*")
	mod.Socket = srv
}

type socket struct {
	*cosnet.Server
}

func (this *socket) Start(ln net.Listener) error {
	this.Server.Accept(ln)
	this.Server.On(cosnet.EventTypeError, this.Errorf)
	this.Server.On(cosnet.EventTypeVerified, this.Connected)
	this.Server.On(cosnet.EventTypeDisconnect, this.Disconnect)
	this.Server.On(cosnet.EventTypeDestroyed, this.Destroyed)
	logger.Trace("网关长连接启动：%v", Options.Gate.Address)
	return nil
}

func (this *socket) Errorf(socket *cosnet.Socket, err interface{}) bool {
	logger.Debug(err)
	return false
}

func (this *socket) ping(c *cosnet.Context) interface{} {
	var s string
	_ = c.Bind(&s)
	return []byte(strconv.Itoa(int(time.Now().Unix())))
}

func (this *socket) proxy(c *cosnet.Context) interface{} {
	path, err := url.Parse(c.Path())
	if err != nil {
		return c.Errorf(0, err)
	}

	//logger.Trace("socket request,PATH:%v   BODY:%v", path, string(c.Message.Body()))
	req, res, err := metadata(path.RawQuery)
	if err != nil {
		return c.Errorf(0, err)
	}

	limit := limits(path.RawPath)
	if limit != ApiLevelNone {
		p := c.Player()
		if p == nil {
			return c.Errorf(0, "not login")
		}
		if limit == ApiLevelLogin {
			req[Options.Metadata.GUID] = p.UUID()
		} else {
			req[Options.Metadata.UID] = p.GetString(Options.Metadata.UID)
		}

	}
	reply := make([]byte, 0)
	if err = request(path.Path, c.Message.Body(), req, res, &reply); err != nil {
		//logger.Trace("socket response error:%v,PATH:%v   Error:%v", path, err)
		return err
	}
	//logger.Trace("socket response,PATH:%v   BODY:%v", path, string(reply))
	if err = this.setCookie(c, res); err != nil {
		return err
	}
	return reply
}

func (this *socket) setCookie(c *cosnet.Context, cookie map[string]string) (err error) {
	if len(cookie) == 0 {
		return
	}
	p := c.Player()
	if p == nil {
		if guid := cookie[Options.Metadata.GUID]; guid != "" {
			if p, err = this.Server.Players.Verify(guid, c.Socket, nil); err != nil {
				return err
			}
		} else {
			return errors.New("not login")
		}
	}
	for key, val := range cookie {
		if key != Options.Metadata.GUID {
			p.Set(key, val)
		}
	}
	return
}

func (this *socket) Connected(sock *cosnet.Socket, _ interface{}) bool {
	logger.Debug("Connected:%v", sock.Id())
	if uid, err := this.GetUid(sock); err == nil && uid != "" {
		//_ = share.Notify.Publish(share.NotifyChannelSocketConnected, uid)
	}
	return true
}

func (this *socket) Disconnect(sock *cosnet.Socket, _ interface{}) bool {
	logger.Debug("Disconnect:%v", sock.Id())
	if uid, err := this.GetUid(sock); err == nil && uid != "" {
		//_ = share.Notify.Publish(share.NotifyChannelSocketDisconnect, uid)
	}
	return true
}

func (this *socket) Destroyed(sock *cosnet.Socket, _ interface{}) bool {
	logger.Debug("Destroyed:%v", sock.Id())
	if uid, err := this.GetUid(sock); err == nil && uid != "" {
		//_ = share.Notify.Publish(share.NotifyChannelSocketDestroyed, uid)
	}
	return true
}

func (this *socket) GetUid(sock *cosnet.Socket) (uid string, err error) {
	if player := sock.Player(); player != nil {
		uid = player.UUID()
	} else {
		err = errors.New("not login")
	}
	return
}
