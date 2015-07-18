package lldbpi

import "github.com/maxymania/go-nntp-backends/groupdb"
import "github.com/maxymania/go-nntp-backends/dayfiledb"

type dfDB struct{
	innerDfDB
}

func (g *dfDB) GetArticlePos(id string) (dayfiledb.ArtPos,bool) {
	g.mutex.RLock(); defer g.mutex.RUnlock()
	return g.getAp(id)
}
func (g *dfDB) SetArticlePos(id string, pos dayfiledb.ArtPos) {
	g.mutex.Lock(); defer g.mutex.Unlock()
	e := g.filer.BeginUpdate()
	if e!=nil { return }
	if g.setAp(id,pos)==nil {
		g.filer.EndUpdate()
	} else {
		g.filer.Rollback()
	}
}
func (g *dfDB) Erase(upto groupdb.DayID) {
	g.mutex.Lock(); defer g.mutex.Unlock()
	e := g.filer.BeginUpdate()
	if e!=nil { return }
	if g.erase(upto)==nil {
		g.filer.EndUpdate()
	} else {
		g.filer.Rollback()
	}
}


