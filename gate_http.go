package gate

import (
	"github.com/hwcer/cosgo/logger"
	"github.com/hwcer/cosgo/session"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosweb"
	"github.com/hwcer/cosweb/middleware"
	"github.com/hwcer/gate/players"
	"github.com/hwcer/wower/options"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var Method = []string{"POST", "GET", "OPTIONS"}

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

func (this *Server) proxy(c *cosweb.Context, next cosweb.Next) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = values.Errorf(0, r)
		}
	}()
	startTime := time.Now()
	defer func() {
		if elapsed := time.Since(startTime); elapsed > elapsedMillisecond {
			buff, _ := c.Buffer()
			logger.Debug("发现高延时请求,TIME:%v,PATH:%v,BODY:%v", elapsed, c.Request.URL.Path, string(buff.Bytes()))
		}
	}()
	h := &httpProxy{Context: c}
	reply, err := proxy(h)
	if err != nil {
		return c.JSON(values.Parse(err))
	} else {
		return c.Bytes(cosweb.ContentTypeApplicationJSON, reply)
	}
}

type httpProxy struct {
	*cosweb.Context
}

func (this *httpProxy) Data() (*session.Data, error) {
	token := this.Context.GetString(session.Options.Name, cosweb.RequestDataTypeCookie, cosweb.RequestDataTypeQuery, cosweb.RequestDataTypeHeader)
	if token == "" {
		return nil, values.Error("token empty")
	}
	if err := this.Context.Session.Verify(token); err != nil {
		return nil, err
	}
	return this.Context.Session.Data, nil
}

func (this *httpProxy) Login(guid string, value values.Values) (err error) {
	cookie := &http.Cookie{Name: session.Options.Name, Path: "/"}
	cookie.Value, err = this.Context.Session.Create(guid, value)
	if err != nil {
		return err
	}
	err = players.Login(this.Context.Session.Data, nil)
	if err != nil {
		return err
	}
	http.SetCookie(this.Context.Response, cookie)
	header := this.Header()
	header.Set("X-Forwarded-Key", session.Options.Name)
	header.Set("X-Forwarded-Val", cookie.Value)
	return nil
}
func (this *httpProxy) Index() uint32 {
	q := this.Context.Request.URL.Query()
	s := q.Get(options.ServiceMetadataRequestId)
	if s == "" {
		return 0
	}
	i, _ := strconv.ParseUint(s, 10, 32)
	return uint32(i)
}
func (this *httpProxy) Delete() error {
	return this.Context.Session.Delete()
}
func (this *httpProxy) Query() (*url.URL, error) {
	return this.Context.Request.URL, nil
}
func (this *httpProxy) Session() string {
	return this.Context.Session.Token()
}
