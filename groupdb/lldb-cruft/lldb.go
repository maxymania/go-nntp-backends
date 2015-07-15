package lldbpi

import "github.com/maxymania/go-nntp-backends/groupdb"
import "github.com/cznic/exp/lldb"
import "io"
import "bytes"
import "sync"

type groupnum struct{
	G,N int64
}

func reverseComp(a, b []byte) int {
	return -bytes.Compare(a,b)
}

type groupDB struct{
	mutex sync.RWMutex
	dprov groupdb.DayProvider
	filer lldb.Filer
	closr io.Closer
	alloc *lldb.Allocator
	group *lldb.BTree
	grass *lldb.BTree
}

/*
	g.mutex.Lock(); defer g.mutex.Unlock()
	e := g.filer.BeginUpdate()
	if e!=nil { return }
	defer g.filer.EndUpdate()
*/

func getGroupId(ge []byte) int64 {
	s,e := lldb.DecodeScalars(ge)
	if e!=nil { return 0 }
	if len(s)<4 { return 0 }
	n,_ := s[3].(int64)
	return n
}
func serializeGrp(ptr *groupdb.Group) ([]byte,error) {
	return EncodeScalars(ptr.Low,ptr.High,ptr.Count)
}


func (g *groupDB) getGroup(ge []byte,ptr *groupdb.Group) bool {
	s,e := lldb.DecodeScalars(ge)
	if e!=nil { return false }
	if len(s)<4 { return false }
	var ok bool
	ptr.Name,ok = s[0].(string)
	if !ok { return false }
	ptr.Desc,ok = s[1].(string)
	if !ok { return false }
	ptr.State,ok = s[2].(uint8)
	if !ok { return false }
	n,ok := s[3].(int64)
	if !ok { return false }
	buf,e := g.alloc.Get(nil,n)
	if e!=nil { return false }
	s,e = lldb.DecodeScalars(buf)
	if e!=nil { return false }
	if len(s)<3 { return false }
	ptr.Low,ok = s[0].(int64)
	if !ok { return false }
	ptr.High,ok = s[1].(int64)
	if !ok { return false }
	ptr.Count,ok = s[2].(int64)
	if !ok { return false }
	return true
}

func (g *groupDB) AddGroup(group, descr string,state byte) {
	
}
func (g *groupDB) Groups(prefix string,ptr *groupdb.Group,cb func()) {
	g.mutex.RLock(); defer g.mutex.RUnlock()
	x := []byte(prefix)
	en,_,_ := g.group.Seek(x)
	if en==nil { return }
	for {
		k,v,e := en.Next()
		if e!=nil { return }
		if !bytes.HasPrefix(k,x) { return }
		if g.getGroup(v,ptr) { cb() }
	}
}
func (g *groupDB) Group(group string, ptr *groupdb.Group) bool {
	g.mutex.RLock(); defer g.mutex.RUnlock()
	v,_ := g.group.Get(nil,[]byte(group))
	if len(v)==0 { return false }
	return g.getGroup(v,ptr)
}
func (g *groupDB) Numberate(groups []string, id string) {
	var grp groupdb.Group
	g.mutex.Lock(); defer g.mutex.Unlock()
	e := g.filer.BeginUpdate()
	if e!=nil { return }
	defer g.filer.EndUpdate()
	for _,group := range groups {
		v,_ := g.group.Get(nil,[]byte(group))
		if len(v)==0 { continue }
		if !g.getGroup(v,&grp) { continue }
		if grp.Low==grp.High {
			if _,e = g.grass.Get(nil,r2b(groupnum{gnum,num})); e!=nil {
				grp.High++
			}
		}else{
			grp.High++
		}
		grp.Count++
		g.grass.Set(r2b(groupnum{gnum,grp.High}),[]byte(id))
		g.alloc.Realloc(gnum,serializeGrp(&grp))
	}
}
func (g *groupDB) GetArticleID(group string, num int64) string {
	g.mutex.RLock(); defer g.mutex.RUnlock()
	v,_ := g.group.Get(nil,[]byte(group))
	gnum := getGroupId(v)
	if gnum<1 { return "" }
	k := r2b(groupnum{gnum,num})
	v,_ = g.grass.Get(nil,k)
	return string(v)
}
func (g *groupDB) Erase(upto groupdb.DayID) {
	// TODO: Implement me
}



