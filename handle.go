package gate

import (
	"strings"
	"time"
)

const elapsedMillisecond = 100 * time.Millisecond

type ApiLevel int8

const (
	ApiLevelNone   ApiLevel = iota //不需要登录
	ApiLevelLogin                  //需要登录
	ApiLevelSelect                 //需要选择角色
)

var paths = map[string]ApiLevel{}

func init() {
	paths["game/login"] = ApiLevelNone        //登录什么都不需要
	paths["game/role/select"] = ApiLevelLogin //选择角色 需要guid
	paths["game/role/create"] = ApiLevelLogin //创建角色 需要guid
}

func limits(s string) ApiLevel {
	if strings.HasPrefix(s, "/") {
		s = s[1:]
	}
	if l, ok := paths[s]; ok {
		return l
	} else {
		return ApiLevelSelect
	}
}
