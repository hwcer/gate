package gate

import (
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosnet"
	"github.com/hwcer/cosweb/session"
	"net/http"
)

func WSVerify(w http.ResponseWriter, r *http.Request) (uid string, err error) {
	//logger.Trace("Sec-Websocket-Extensions:%v", r.Header.Get("Sec-Websocket-Extensions"))
	//logger.Trace("Sec-Websocket-Key:%v", r.Header.Get("Sec-Websocket-Key"))
	//logger.Trace("Sec-Websocket-Protocol:%v", r.Header.Get("Sec-Websocket-Protocol"))
	//logger.Trace("Sec-Websocket-Version:%v", r.Header.Get("Sec-Websocket-Version"))
	token := r.Header.Get("Sec-Websocket-Protocol")
	if token == "" {
		return "", values.Error("token empty")
	}
	sess := session.New()
	if err = sess.Start(token, session.StartTypeAuth); err != nil {
		return "", values.Parse(err)
	}
	uid = sess.GetString(opt.Metadata.UID)
	if uid == "" {
		return "", values.Error("请登录")
	}
	return
}
func WSAccept(s *cosnet.Socket, uid string) {
	_, _ = Players.Binding(uid, s)
	v := values.Values{}
	v[opt.Metadata.UID] = uid
	s.Set(v)
	return
}
