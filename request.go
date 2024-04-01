package gate

import (
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosrpc/xclient"
	"net/url"
	"strings"
)

//
//var emptyJsonByteMap = map[string]bool{
//	"\"\"": true,
//	"{}":   true,
//	"[]":   true,
//}
//
//func EmptyJsonByte(b []byte) bool {
//	if len(b) == 0 {
//		return true
//	}
//	return emptyJsonByteMap[string(b)]
//}

func metadata(raw string) (req, res map[string]string, err error) {
	var query url.Values
	query, err = url.ParseQuery(raw)
	if err != nil {
		return
	}

	req = make(map[string]string)
	res = make(map[string]string)
	for k, _ := range query {
		req[k] = query.Get(k)
	}
	return
}

// request rpc转发,返回实际转发的servicePath
func request(path string, args []byte, req, res map[string]string, reply any) (err error) {
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	index := strings.Index(path, "/")
	if index < 0 {
		return values.Errorf(404, "page not found")
	}
	servicePath := path[0:index]

	serviceMethod := strings.ToLower(HandleServiceMethod(path[index:]))
	err = xclient.CallWithMetadata(req, res, servicePath, serviceMethod, args, reply)
	return
}
