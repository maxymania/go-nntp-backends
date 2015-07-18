package dayfiledb

import "errors"

type Options struct{
	FileName    string
}

type DayfileDBPlugin func(opts *Options) (DayfileDB,error)

var pluginmap = make(map[string]DayfileDBPlugin)

func RegisterPlugin(s string,g DayfileDBPlugin) {
	pluginmap[s] = g
}

func Open(s string,opts *Options) (DayfileDB,error) {
	g,ok := pluginmap[s]
	if !ok { return nil,errors.New("plugin not found: "+s) }
	return g(opts)
}

