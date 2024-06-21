package gate

import (
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosrpc/xclient"
	"github.com/hwcer/cosrpc/xshare"
	"github.com/hwcer/gate/options"
	"github.com/hwcer/registry"
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

type player interface {
	GetString(any) string
}

// request rpc转发,返回实际转发的servicePath
func request(p player, path string, args []byte, req, res xshare.Metadata, reply any) (err error) {
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	index := strings.Index(path, "/")
	if index < 0 {
		return values.Errorf(404, "page not found")
	}
	servicePath := path[0:index]
	serviceMethod := registry.Formatter(path[index:])
	if options.Options.Route.Prefix != "" {
		serviceMethod = registry.Join(options.Options.Route.Prefix, serviceMethod)
	}
	if p != nil {
		if serviceAddress := p.GetString(options.GetServiceAddress(servicePath)); serviceAddress != "" {
			req.SetAddress(serviceAddress)
		}
	}
	err = xclient.CallWithMetadata(req, res, servicePath, serviceMethod, args, reply)
	return
}
