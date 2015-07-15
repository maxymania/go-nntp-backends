package groupdb

import "errors"

type GroupDBPlugin func(fn string) (GroupDB,error)

var pluginmap = make(map[string]GroupDBPlugin)

func RegisterPlugin(s string,g GroupDBPlugin) {
	pluginmap[s] = g
}

func Open(s,fn string) (GroupDB,error) {
	g,ok := pluginmap[s]
	if !ok { return nil,errors.New("plugin not found: "+s) }
	return g(fn)
}

