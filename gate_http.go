package gate

import (
	"encoding/json"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosrpc/xshare"
	"github.com/hwcer/cosweb"
	"github.com/hwcer/cosweb/middleware"
	"github.com/hwcer/cosweb/session"
	"github.com/hwcer/logger"
	"net"
	"net/http"
	"strings"
	"time"
)

var Method = []string{"POST", "GET", "OPTIONS"}

//var ServiceProxyRoute = map[string]string{}

func init() {
	mod.Server = &server{}
	mod.Server.Server = cosweb.New(nil)
	session.Options.Name = "_cookie_vars"
}

type server struct {
	*cosweb.Server
}

func (this *server) init() error {
	if err := session.Start(nil); err != nil {
		return err
	}
	//跨域
	access := middleware.NewAccessControlAllow()
	access.Origin("*")
	access.Methods(Method...)
	headers := []string{session.Options.Name, "Content-Type", "Set-Cookie", "*"}
	access.Headers(strings.Join(headers, ","))
	this.Server.Use(access.Handle)
	this.Server.Register("/*", this.proxy, Method...)
	return nil
}

func (this *server) Start(address string) (err error) {
	if err = this.init(); err != nil {
		return
	}
	if err = this.Server.Start(address); err == nil {
		logger.Trace("网关短连接启动：%v", opt.Gate.Address)
	}
	return
}
func (this *server) Listen(ln net.Listener) (err error) {
	if err = this.init(); err != nil {
		return
	}
	if err = this.Server.Listen(ln); err == nil {
		logger.Trace("网关短连接启动：%v", opt.Gate.Address)
	}
	return
}

// Login 登录
func (this *server) login(c *cosweb.Context, guid string) (cookie *http.Cookie, err error) {
	value := values.Values{}
	value["time"] = time.Now().Unix()
	cookie = &http.Cookie{Name: session.Options.Name, Path: "/"}
	if cookie.Value, err = c.Session.Create(guid, value); err == nil {
		c.Cookie.SetCookie(cookie)
	}
	//c.Header().Set(cookie.Name, cookie.Value)
	return
}

func (this *server) proxy(c *cosweb.Context, next cosweb.Next) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = values.Errorf(0, r)
		}
	}()

	startTime := time.Now()
	defer func() {
		if elapsed := time.Since(startTime); elapsed > elapsedMillisecond {
			logger.Debug("发现高延时请求,TIME:%v,PATH:%v,BODY:%v", elapsed, c.Request.URL.Path, string(c.Body.Bytes()))
		}
	}()

	req, res, err := metadata(c.Request.URL.RawQuery)
	if err != nil {
		return c.JSON(values.Parse(err))
	}

	var p player
	limit := limits(c.Request.URL.Path)
	if limit != ApiLevelNone {
		token := c.GetString(session.Options.Name, cosweb.RequestDataTypeCookie, cosweb.RequestDataTypeQuery, cosweb.RequestDataTypeHeader)
		if token == "" {
			return c.JSON(values.Error("token empty"))
		}
		if err = c.Session.Start(token, session.StartTypeAuth); err != nil {
			return c.JSON(values.Parse(err))
		}
		p = c.Session
		if limit == ApiLevelLogin {
			req[opt.Metadata.GUID] = c.Session.UUID()
		} else {
			req[opt.Metadata.UID] = c.Session.GetString(opt.Metadata.UID)
		}
	}

	reply := make([]byte, 0)
	if err = request(p, c.Request.URL.Path, c.Body.Bytes(), req, res, &reply); err != nil {
		return c.JSON(values.Parse(err))
	}
	var cookie *http.Cookie
	if cookie, err = this.setCookie(c, res); err != nil {
		return c.JSON(values.Parse(err))
	}
	if cookie != nil {
		r := map[string]any{}
		if err = json.Unmarshal(reply, &r); err != nil {
			return c.JSON(values.Parse(err))
		}
		r["cookie"] = map[string]string{"Name": cookie.Name, "Value": cookie.Value}
		return c.JSON(r)
	} else {
		return c.Bytes(cosweb.ContentTypeApplicationJSON, reply)
	}
}

func (this *server) setCookie(c *cosweb.Context, cookie xshare.Metadata) (r *http.Cookie, err error) {
	if len(cookie) == 0 {
		return
	}
	if guid := cookie[opt.Metadata.GUID]; guid != "" {
		if r, err = this.login(c, guid); err != nil {
			return
		}
	}
	for key, val := range cookie {
		if key != opt.Metadata.GUID {
			c.Session.Set(key, val)
		}
	}
	return
}
