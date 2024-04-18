package gate

import (
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosrpc/xclient"
	"github.com/hwcer/cosrpc/xshare"
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
func request(path string, args []byte, req, res xshare.Metadata, reply any) (err error) {
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	index := strings.Index(path, "/")
	if index < 0 {
		return values.Errorf(404, "page not found")
	}
	servicePath := path[0:index]
	service := Service()
	serviceMethod := service.Formatter(path[index:])
	err = xclient.CallWithMetadata(req, res, servicePath, serviceMethod, args, reply)
	return
}