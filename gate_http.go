package gate

import (
	"github.com/hwcer/cosgo/binder"
	"github.com/hwcer/cosgo/logger"
	"github.com/hwcer/cosgo/options"
	"github.com/hwcer/cosgo/session"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosweb"
	"github.com/hwcer/cosweb/middleware"
	"github.com/hwcer/gate/players"
	"net"
	"net/http"
	"strings"
	"time"
)

var Method = []string{"POST", "GET", "OPTIONS"}

//var ServiceProxyRoute = map[string]string{}

func init() {
	session.Options.Name = "_cookie_vars"
}

type Server struct {
	*cosweb.Server
	redis any //是否使用redis存储session信息
}

func (this *Server) init() (err error) {
	this.Server = cosweb.New()
	if this.redis != nil {
		session.Options.Storage, err = session.NewRedis(this.redis)
	} else {
		session.Options.Storage = session.NewMemory()
	}
	if err != nil {
		return err
	}

	//跨域
	access := middleware.NewAccessControlAllow()
	access.Origin("*")
	access.Methods(Method...)
	headers := []string{session.Options.Name, "Content-Type", "Set-Cookie", "X-Forwarded-Key", "X-Forwarded-Val", "*"}
	access.Headers(strings.Join(headers, ","))
	this.Server.Use(access.Handle)
	this.Server.Register("/*", this.proxy, Method...)
	return nil
}

func (this *Server) Listen(address string) (err error) {
	if err = this.Server.Listen(address); err == nil {
		logger.Trace("网关短连接启动：%v", options.Gate.Address)
	}
	return
}
func (this *Server) Accept(ln net.Listener) (err error) {
	if err = this.Server.Accept(ln); err == nil {
		logger.Trace("网关短连接启动：%v", options.Gate.Address)
	}
	return
}

// Login 登录
func (this *Server) login(c *cosweb.Context, uuid string, data map[string]any) (cookie *http.Cookie, err error) {
	cookie = &http.Cookie{Name: session.Options.Name, Path: "/"}
	if cookie.Value, err = c.Session.Create(uuid, data); err == nil {
		http.SetCookie(c.Response, cookie)
	}
	c.Header().Set("X-Forwarded-Key", session.Options.Name)
	c.Header().Set("X-Forwarded-Val", cookie.Value)
	return
}

func (this *Server) proxy(c *cosweb.Context, next cosweb.Next) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = values.Errorf(0, r)
		}
	}()
	body, err := c.Buffer()
	if err != nil {
		return err
	}
	startTime := time.Now()
	defer func() {
		if elapsed := time.Since(startTime); elapsed > elapsedMillisecond {
			logger.Debug("发现高延时请求,TIME:%v,PATH:%v,BODY:%v", elapsed, c.Request.URL.Path, string(body.Bytes()))
		}
	}()

	req, res, err := metadata(c.Request.URL.RawQuery)
	if err != nil {
		return c.JSON(values.Parse(err))
	}

	var p *session.Data
	path := Formatter(c.Request.URL.Path)
	limit := options.Apis.Get(path)
	if limit != options.ApisTypeNone {
		token := c.GetString(session.Options.Name, cosweb.RequestDataTypeCookie, cosweb.RequestDataTypeQuery, cosweb.RequestDataTypeHeader)
		if token == "" {
			return c.JSON(values.Error("token empty"))
		}
		if err = c.Session.Verify(token); err != nil {
			return c.JSON(values.Parse(err))
		}
		p = c.Session.Data
		if p == nil {
			return c.JSON(values.Error("not login"))
		}
		if limit == options.ApisTypeOAuth {
			req[options.ServiceMetadataGUID] = p.UUID()
		} else {
			req[options.ServiceMetadataUID] = p.GetString(options.ServiceMetadataUID)
		}
	}
	if ct := c.Binder.String(); ct != binder.Json.String() {
		req[binder.ContentType] = ct
	}
	reply := make([]byte, 0)
	if err = request(p, path, body.Bytes(), req, res, &reply); err != nil {
		return c.JSON(values.Parse(err))
	}
	var cookie *http.Cookie
	if cookie, err = this.setCookie(c, res); err != nil {
		return c.JSON(values.Parse(err))
	}
	return Writer(c, reply, cookie)
}

func (this *Server) setCookie(c *cosweb.Context, cookie options.Metadata) (r *http.Cookie, err error) {
	if len(cookie) == 0 {
		return
	}
	if guid, ok := cookie[options.ServicePlayerOAuth]; ok {
		if r, err = this.login(c, guid, CookiesFilter(cookie)); err == nil {
			err = players.Login(c.Session.Data, nil)
		}
	} else if _, ok = cookie[options.ServicePlayerLogout]; ok {
		players.Delete(c.Session.Data)
		err = c.Session.Delete()
	} else if c.Session.Data != nil {
		CookiesUpdate(cookie, c.Session.Data)
	}
	return
}
