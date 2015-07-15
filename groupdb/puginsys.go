package groupdb

import "errors"

type Options struct{
	DayProvider DayProvider
	FileName    string
}

type GroupDBPlugin func(opts *Options) (GroupDB,error)

var pluginmap = make(map[string]GroupDBPlugin)

func RegisterPlugin(s string,g GroupDBPlugin) {
	pluginmap[s] = g
}

func Open(s string,opts *Options) (GroupDB,error) {
	g,ok := pluginmap[s]
	if !ok { return nil,errors.New("plugin not found: "+s) }
	return g(opts)
}

