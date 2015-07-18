package lldbpi

import "github.com/maxymania/go-nntp-backends/groupdb"
import "github.com/cznic/exp/lldb"
import "io"
import "bytes"
import "os"
import "errors"

func openWAL(fn string) (lldb.Filer,io.Closer,bool,error) {
	first := false
	f,e := os.OpenFile(fn,os.O_RDWR,0660)
	if e!=nil {
		first = true
		f,e = os.OpenFile(fn,os.O_RDWR|os.O_CREATE,0660)
		if e!=nil { return nil,nil,false,e }
	}
	fl,e := os.OpenFile(fn+".wal",os.O_RDWR|os.O_CREATE,0660)
	if e!=nil { f.Close(); return nil,nil,false,e }
	filer,e := lldb.NewACIDFiler(lldb.NewSimpleFileFiler(f),fl)
	if e!=nil { f.Close(); fl.Close(); return nil,nil,false,e }
	return filer,fl,first,nil
}

func Open(opts *groupdb.Options) (groupdb.GroupDB,error) {
	var h1,h2,h3 int64
	f,c,first,e := openWAL(opts.FileName)
	if e!=nil { return nil,e }
	g := new(groupDB)
	g.dprov = opts.DayProvider
	g.filer = f
	g.closr = c
	g.alloc,e = lldb.NewAllocator(g.filer,&lldb.Options{})
	if e!=nil { f.Close(); c.Close(); return nil,e }
	if first {
		g.filer.BeginUpdate()
		defer g.filer.EndUpdate()
		i,e := g.alloc.Alloc([]byte("......"))
		if e!=nil { f.Close(); c.Close(); return nil,e }
		if i!=1 { f.Close(); c.Close(); return nil,errors.New("corrupted") }
		g.group,h1,e = lldb.CreateBTree(g.alloc,bytes.Compare)
		if e!=nil { f.Close(); c.Close(); return nil,e }
		g.grass,h2,e = lldb.CreateBTree(g.alloc,bytes.Compare)
		if e!=nil { f.Close(); c.Close(); return nil,e }
		g.tmlog,h3,e = lldb.CreateBTree(g.alloc,reverseComp)
		if e!=nil { f.Close(); c.Close(); return nil,e }
		v,e := lldb.EncodeScalars(h1,h2,h3)
		if e!=nil { f.Close(); c.Close(); return nil,e }
		e = g.alloc.Realloc(1,v)
		if e!=nil { f.Close(); c.Close(); return nil,e }
	} else {
		v,e := g.alloc.Get(nil,1)
		if e!=nil { f.Close(); c.Close(); return nil,e }
		s,e := lldb.DecodeScalars(v)
		if e!=nil { f.Close(); c.Close(); return nil,e }
		g.group,e = lldb.OpenBTree(g.alloc,bytes.Compare,s[0].(int64))
		if e!=nil { f.Close(); c.Close(); return nil,e }
		g.grass,e = lldb.OpenBTree(g.alloc,bytes.Compare,s[1].(int64))
		if e!=nil { f.Close(); c.Close(); return nil,e }
		g.tmlog,e = lldb.OpenBTree(g.alloc,reverseComp,s[2].(int64))
		if e!=nil { f.Close(); c.Close(); return nil,e }
	}
	return g,nil
}

func init() {
	groupdb.RegisterPlugin("lldb",Open)
}




