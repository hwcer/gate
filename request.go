package gate

import (
	"bytes"
	"github.com/hwcer/cosgo/registry"
	"github.com/hwcer/cosgo/session"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosrpc/xclient"
	"github.com/hwcer/cosrpc/xshare"
	"github.com/hwcer/gate/players"
	"github.com/hwcer/wower/options"
	"github.com/hwcer/wower/share"
	"net/url"
	"strings"
)

func metadata(raw string) (req, res xshare.Metadata, err error) {
	var query url.Values
	query, err = url.ParseQuery(raw)
	if err != nil {
		return
	}

	req = make(xshare.Metadata)
	res = make(xshare.Metadata)
	for k, _ := range query {
		req[k] = query.Get(k)
	}
	return
}

// request rpc转发,返回实际转发的servicePath
func request(p *session.Data, path string, args []byte, req, res xshare.Metadata, reply any) (err error) {
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	index := strings.Index(path, "/")
	if index < 0 {
		return values.Errorf(404, "page not found")
	}
	servicePath := path[0:index]
	serviceMethod := registry.Formatter(path[index:])
	if options.Gate.Prefix != "" {
		serviceMethod = registry.Join(options.Gate.Prefix, serviceMethod)
	}
	if p != nil {
		if serviceAddress := p.GetString(options.GetServiceSelectorAddress(servicePath)); serviceAddress != "" {
			req.SetAddress(serviceAddress)
		}
	}
	err = xclient.CallWithMetadata(req, res, servicePath, serviceMethod, args, reply)
	return
}

type RequestHandle interface {
	Data() (*session.Data, error)
	Token() string
	Query() (*url.URL, error)
	Login(guid string, cookie values.Values) error
	Buffer() (buf *bytes.Buffer, err error)
	Delete() error
}

func proxy(h RequestHandle) ([]byte, error) {
	uri, err := h.Query()
	if err != nil {
		return nil, err
	}
	req, res, err := metadata(uri.RawQuery)
	if err != nil {
		return nil, values.Parse(err)
	}

	var p *session.Data
	path := Formatter(uri.Path)
	limit := share.Authorizes.Get(path)
	if limit != share.AuthorizesTypeNone {
		if p, err = h.Data(); err != nil {
			return nil, values.Parse(err)
		} else if p == nil {
			return nil, values.Error("not login")
		}
		p.KeepAlive()
		req[options.ServicePlayerSession] = h.Token()
		if limit == share.AuthorizesTypeOAuth {
			req[options.ServiceMetadataGUID] = p.UUID()
			req[options.ServicePlayerGateway] = options.Gate.Address
		} else {
			req[options.ServiceMetadataUID] = p.GetString(options.ServiceMetadataUID)
		}
	}
	//if ct := c.Binder.String(); ct != binder.Json.String() {
	//	req[binder.ContentType] = ct
	//}
	buff, err := h.Buffer()
	if err != nil {
		return nil, values.Parse(err)
	}
	reply := make([]byte, 0)
	if err = request(p, path, buff.Bytes(), req, res, &reply); err != nil {
		return nil, values.Parse(err)
	}
	if len(res) == 0 {
		return reply, nil
	}
	if guid, ok := res[options.ServicePlayerOAuth]; ok {
		err = h.Login(guid, CookiesFilter(res))
	} else if _, ok = res[options.ServicePlayerLogout]; ok {
		if err = h.Delete(); err == nil && p != nil {
			players.Delete(p)
		}
	} else if p != nil {
		CookiesUpdate(res, p)
	}
	if err != nil {
		return nil, values.Parse(err)
	} else {
		return reply, nil
	}
}
