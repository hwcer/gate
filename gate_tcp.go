package gate

import (
	"errors"
	"github.com/hwcer/cosgo/utils"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosnet"
	"github.com/hwcer/cosnet/tcp"
	"github.com/hwcer/cosrpc/xshare"
	"github.com/hwcer/logger"
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
	_ = service.Register(srv.ping, "ping")
	//_ = Sockets.Register(Socket.login)
	_ = service.Register(srv.proxy, "/*")

	srv.Server.On(cosnet.EventTypeError, srv.Errorf)
	srv.Server.On(cosnet.EventTypeVerified, srv.Connected)
	srv.Server.On(cosnet.EventTypeDisconnect, srv.Disconnect)
	srv.Server.On(cosnet.EventTypeDestroyed, srv.Destroyed)

	mod.Socket = srv
}

type socket struct {
	*cosnet.Server
}

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
	logger.Trace("网关长连接启动：%v", opt.Gate.Address)
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
	path, err := url.Parse(c.Path())
	if err != nil {
		return c.Errorf(0, err)
	}

	//logger.Trace("socket request,PATH:%v   BODY:%v", path, string(c.Message.Body()))
	req, res, err := metadata(path.Path)
	if err != nil {
		return c.Errorf(0, err)
	}
	p := c.Values()
	limit := limits(path.Path)
	if limit != ApiLevelNone {
		if p == nil {
			return c.Errorf(0, "not login")
		}
		if limit == ApiLevelLogin {
			req[opt.Metadata.GUID] = p.GetString(opt.Metadata.GUID)
		} else {
			req[opt.Metadata.UID] = p.GetString(opt.Metadata.UID)
		}
	}

	reply := make([]byte, 0)
	if p == nil {
		err = request(nil, path.Path, c.Message.Body(), req, res, &reply)
	} else {
		err = request(p, path.Path, c.Message.Body(), req, res, &reply)
	}
	if err != nil {
		//logger.Trace("socket response error:%v,PATH:%v   Error:%v", path, err)
		return err
	}
	//logger.Trace("socket response,PATH:%v   BODY:%v", path, string(reply))
	if err = this.setCookie(c, res); err != nil {
		return err
	}
	return reply
}

func (this *socket) setCookie(c *cosnet.Context, cookie xshare.Metadata) (err error) {
	if len(cookie) == 0 {
		return
	}
	var v values.Values
	//账号登录
	if i := c.Socket.Get(); i == nil {
		if guid := cookie[opt.Metadata.GUID]; guid != "" {
			v = values.Values{}
			c.Socket.Set(v)
		} else {
			return errors.New("not login")
		}
	} else {
		v = i.(values.Values)
	}
	//角色绑定
	if uid := cookie[opt.Metadata.UID]; uid != "" {
		if _, err = Players.Binding(uid, c.Socket); err != nil {
			return
		}
	}
	//同步数据
	for key, val := range cookie {
		v.Set(key, val)
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
	vs := sock.Values()
	if vs == nil {
		return true
	}
	if uid := vs.GetString(opt.Metadata.UID); uid != "" {
		Players.Delete(uid)
		//_ = share.Notify.Publish(share.NotifyChannelSocketDestroyed, uid)
	}
	return true
}
