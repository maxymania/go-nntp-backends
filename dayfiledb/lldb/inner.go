package lldbpi

import "github.com/maxymania/go-nntp-backends/groupdb"
import "github.com/maxymania/go-nntp-backends/dayfiledb"
import "github.com/cznic/exp/lldb"
import "github.com/cznic/bufs"
import "io"
import "sync"

type innerDfDB struct{
	mutex sync.RWMutex
	filer lldb.Filer
	closr io.Closer
	alloc *lldb.Allocator
	ididx *lldb.BTree // ID-Index
	revin *lldb.BTree // Reverse-Index
	lastdf int64
}

func (i *innerDfDB) getAp(id string) (pos dayfiledb.ArtPos,ok bool){
	buf := bufs.GCache.Get(20)
	defer bufs.GCache.Put(buf)
	v,_ := i.ididx.Get(buf,[]byte(id))
	if len(v)!=20 { return }
	ok = b2r(v,&pos)
	return
}
func (i *innerDfDB) setAp(id string, pos dayfiledb.ArtPos) error {
	var dayid groupdb.DayID
	buf := bufs.GCache.Get(20)
	defer bufs.GCache.Put(buf)
	k := []byte(id)
	v := r2b(pos)
	e := i.ididx.Set(k,v)
	if e!=nil { return e }
	e = i.revin.Set(v,k)
	if e!=nil { return e }

	// Delete old article-id entries.
	v,_ = i.alloc.Get(buf,i.lastdf)
	if !b2r(v,&dayid) { return nil }
	ed,_,_ := i.revin.Seek(r2b(dayfiledb.ArtPos{dayid,-1,-1}))
	if ed==nil { return nil }
	for ii := 0 ; ii<3 ; ii++ {
		k,v,err := ed.Next()
		if err!=nil { return nil }
		i.revin.Delete(k)
		i.ididx.Delete(v)
	}
	return nil
}
func (i *innerDfDB) erase(df groupdb.DayID) error {
	var dayid groupdb.DayID
	buf := bufs.GCache.Get(20)
	defer bufs.GCache.Put(buf)
	v,_ := i.alloc.Get(buf,i.lastdf)
	if b2r(v,&dayid) && dayid>=df { return nil }
	return i.alloc.Realloc(i.lastdf,r2b(df))
}


