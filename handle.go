package gate

import (
	"github.com/hwcer/cosgo/apis"
	"github.com/hwcer/cosweb"
	"net/http"
	"strings"
	"time"
)

const elapsedMillisecond = 100 * time.Millisecond

func init() {
	apis.Set("game/login", apis.None)
	apis.Set("game/role/select", apis.OAuth)
	apis.Set("game/role/create", apis.OAuth)
}

// Formatter 格式化路径
var Formatter = func(s string) string {
	return strings.ToLower(s)
}

var Writer = func(c *cosweb.Context, reply []byte, cookie *http.Cookie) error {
	return c.Bytes(cosweb.ContentTypeApplicationJSON, reply)
}

/*
	if cookie != nil {
		r := map[string]any{}
		if err = json.Unmarshal(reply, &r); err != nil {
			return c.JSON(values.Parse(err))
		}

		//r["cookie"] = map[string]string{"Name": cookie.Name, "Value": cookie.Value}
		return c.JSON(r)
	} else {
		return c.Bytes(cosweb.ContentTypeApplicationJSON, reply)
	}
*/
