package lldbpi

import "github.com/maxymania/go-nntp-backends/groupdb"
import "github.com/cznic/exp/lldb"
import "github.com/cznic/bufs"
import "bytes"
import "io"
import "sync"

type groupnum struct{
	G,N int64
}
type groupdescr struct{
	Begin int64
	End   int64
	DayID groupdb.DayID
}

type groupday struct{
	DayID groupdb.DayID
	Grp   int64
}

func getGroup(ge []byte,ptr *groupdb.Group, i *int64) bool {
	s,e := lldb.DecodeScalars(ge)
	if e!=nil { return false }
	if len(s)<4 { return false }
	var ok bool
	ptr.Name,ok = s[0].(string)
	if !ok { return false }
	ptr.Desc,ok = s[1].(string)
	if !ok { return false }
	state,ok := s[2].(uint64)
	ptr.State = byte(state)
	if !ok { return false }
	*i,ok = s[3].(int64)
	if !ok { return false }
	return true
}

type innerGroupDB struct{
	mutex sync.RWMutex
	dprov groupdb.DayProvider
	filer lldb.Filer
	closr io.Closer
	alloc *lldb.Allocator
	group *lldb.BTree
	grass *lldb.BTree
	tmlog *lldb.BTree
}

func (i *innerGroupDB) gget(grp, num int64) string {
	var gd groupdescr
	buf := bufs.GCache.Get(20)
	v,_ := i.alloc.Get(buf,grp)
	if len(v)!=20 { return "" }
	if !b2r(v,&gd) { return "" }
	bufs.GCache.Put(buf)
	if gd.Begin>num { return "" }
	if gd.End<=num { return "" }
	v,_ = i.grass.Get(nil, r2b(groupnum{grp,num}) )
	return string(v)
}
func (i *innerGroupDB) gassign(grp int64, id string) error {
	var gd groupdescr
	buf := bufs.GCache.Get(20)
	v,_ := i.alloc.Get(buf,grp)
	if len(v)!=20 { return nil }
	b2r(v,&gd)
	bufs.GCache.Put(buf)

	// Delete old entries
	{
		begin := r2b(grp)
		end := r2b(groupnum{grp,gd.Begin})
		enum,_,e := i.grass.Seek(begin)
		if e!=nil {
			lst := make([][]byte,0,10)
			for i := 0 ; i<10 ; i++ {
				k,_,e := enum.Next()
				if e!=nil { break }
				if bytes.Compare(k,end)>=0 { break }
				lst = append(lst,k)
			}
			for _,k := range lst {
				i.grass.Delete(k)
			}
		}
	}
	i.tmlog.Set(r2b(groupday{gd.DayID,grp}),r2b(gd.End))
	gd.DayID = i.dprov()

	// Add new entries
	e := i.grass.Set(r2b(groupnum{grp,gd.End}),[]byte(id))
	if e!=nil { return e }
	gd.End++
	return i.alloc.Realloc(grp,r2b(gd))
}
func (i *innerGroupDB) groupdescr(grp int64) (gd groupdescr) {
	buf := bufs.GCache.Get(20)
	v,_ := i.alloc.Get(buf,grp)
	if len(v)!=20 { return }
	b2r(v,&gd)
	bufs.GCache.Put(buf)
	return
}
func (i *innerGroupDB) add(group, descr string, state byte) error{
	R,_ := i.group.Get(nil,[]byte(group))
	if R!=nil { return nil }
	h,e := i.alloc.Alloc(r2b(groupdescr{1,1,i.dprov()}))
	if e!=nil { return e }
	v,e := lldb.EncodeScalars(group,descr,uint64(state),h)
	if e!=nil { return e }
	return i.group.Set([]byte(group),v)
}
func (i *innerGroupDB) erase(upto groupdb.DayID) {
	var ga groupday
	var gd groupdescr
	var in int64
	buf := bufs.GCache.Get(20)
	defer bufs.GCache.Put(buf)
	pf := r2b(groupday{upto,-1})
	enum,_,_ := i.tmlog.Seek(pf)
	if enum==nil { return }
	for {
		k,v,_ := enum.Next()
		if !b2r(k,&ga) { return }
		if ga.DayID!=upto { return }
		if !b2r(v,&in) { continue }
		vv,_ := i.alloc.Get(buf,ga.Grp)
		if len(vv)!=20 { continue }
		if !b2r(vv,&gd) { continue }
		if gd.Begin >= in { continue }
		gd.Begin = in
		i.alloc.Realloc(ga.Grp,r2b(gd))
	}
}



