package lldbpi

import "github.com/maxymania/go-nntp-backends/groupdb"
import "bytes"

type groupDB struct{
	innerGroupDB
}

func (g *groupDB) AddGroups(src <- chan groupdb.GroupMeta) {
	g.mutex.Lock(); defer g.mutex.Unlock()
	e := g.filer.BeginUpdate()
	if e!=nil { return }
	for grpo := range src {
		v,_ := g.group.Get(nil,[]byte(grpo.Name))
		if len(v)>0 { continue }
		if g.add(grpo.Name,grpo.Desc,grpo.State)!=nil {
			g.filer.Rollback()
			for _ = range src { }
			return
		}
	}
	g.filer.EndUpdate()
}
func (g *groupDB) AddGroup(group, descr string,state byte) {
	g.mutex.Lock(); defer g.mutex.Unlock()
	e := g.filer.BeginUpdate()
	if e!=nil { return }
	if g.add(group,descr,state)==nil {
		g.filer.EndUpdate()
	}else{
		g.filer.Rollback()
	}
}
func (g *groupDB) Groups(prefix string,ptr *groupdb.Group,cb func()) {
	var idx int64
	g.mutex.RLock(); defer g.mutex.RUnlock()
	pfx := []byte(prefix)
	enum,_,_ := g.group.Seek(pfx)
	if enum==nil { return }
	for {
		k,v,_ := enum.Next()
		if len(k)==0 { break }
		if !bytes.HasPrefix(k,pfx) { break }
		if len(v)==0 { continue }
		if getGroup(v,ptr,&idx) {
			gd := g.groupdescr(idx)
			if gd.Begin<gd.End {
				ptr.Low = gd.Begin
				ptr.High = gd.End-1
				ptr.Count = gd.End-gd.Begin
			} else {
				ptr.Low = 0
				ptr.High = 0
				ptr.Count = 0
			}
			cb()
		}
	}
}
func (g *groupDB) Group(group string, ptr *groupdb.Group) bool {
	var idx int64
	g.mutex.RLock(); defer g.mutex.RUnlock()
	v,_ := g.group.Get(nil,[]byte(group))
	if len(v)==0 { return false }
	if getGroup(v,ptr,&idx) {
		gd := g.groupdescr(idx)
		if gd.Begin<gd.End {
			ptr.Low = gd.Begin
			ptr.High = gd.End-1
			ptr.Count = gd.End-gd.Begin
		} else {
			ptr.Low = 0
			ptr.High = 0
			ptr.Count = 0
		}
		return true
	}
	return false
}
func (g *groupDB) Numberate(groups []string, id string) {
	var grpobj groupdb.Group
	var idx int64
	g.mutex.Lock(); defer g.mutex.Unlock()
	e := g.filer.BeginUpdate()
	if e!=nil { return }
	defer g.filer.EndUpdate()
	for _,grpn := range groups {
		v,_ := g.group.Get(nil,[]byte(grpn))
		if len(v)==0 { continue }
		if !getGroup(v,&grpobj,&idx) { continue }
		g.gassign(idx,id)
	}
}
func (g *groupDB) GetArticleID(group string, num int64) string {
	var grpobj groupdb.Group
	var idx int64
	g.mutex.RLock(); defer g.mutex.RUnlock()
	v,_ := g.group.Get(nil,[]byte(group))
	if len(v)==0 { return "" }
	if !getGroup(v,&grpobj,&idx) { return "" }
	return g.gget(idx,num)
}
func (g *groupDB) Erase(upto groupdb.DayID) {
	g.mutex.Lock(); defer g.mutex.Unlock()
	e := g.filer.BeginUpdate()
	if e!=nil { return }
	defer g.filer.EndUpdate()
	g.erase(upto)
}



