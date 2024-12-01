package gate

import (
	"github.com/hwcer/cosgo/options"
	"github.com/hwcer/cosgo/registry"
	"github.com/hwcer/cosgo/session"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosrpc/xclient"
	"net/url"
	"strings"
)

func metadata(raw string) (req, res options.Metadata, err error) {
	var query url.Values
	query, err = url.ParseQuery(raw)
	if err != nil {
		return
	}

	req = make(options.Metadata)
	res = make(options.Metadata)
	for k, _ := range query {
		req[k] = query.Get(k)
	}
	return
}

// request rpc转发,返回实际转发的servicePath
func request(p *session.Data, path string, args []byte, req, res options.Metadata, reply any) (err error) {
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
