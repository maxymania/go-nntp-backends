package dfbackend

import (
	"github.com/maxymania/go-nntp-backends/articlestore"
	"github.com/maxymania/go-nntp-backends/dayfile"
	"github.com/maxymania/go-nntp-backends/groupdb"
	"github.com/maxymania/go-nntp-backends/dayfiledb"
	"github.com/maxymania/go-nntp/server"
	"github.com/maxymania/go-nntp"
	"io"
	"io/ioutil"
	"time"
)

type backend struct{
	blobpool dayfile.ArticleBlobPool
	gdb groupdb.GroupDB
	ddb dayfiledb.DayfileDB
	spool dayfile.StorageFilePool
}

func getDay() groupdb.DayID {
	y,m,d := time.Now().Date()
	return groupdb.MakeDayID(y,int(m),d)
}

func toNGroup(gdb groupdb.Group) *nntp.Group{
	return &nntp.Group{
		gdb.Name,
		gdb.Desc,
		gdb.Count,
		gdb.High,
		gdb.Low,
		nntp.PostingStatus(gdb.State),
	}
}

func (b *backend) ListGroups() (<-chan *nntp.Group, error) {
	cng := make(chan *nntp.Group,1024)
	go func(){
		defer func(){close(cng)}()
		var gdb groupdb.Group
		b.gdb.Groups("",&gdb,func(){
			cng <- toNGroup(gdb)
		})
	}()
	return cng,nil
}

func (b *backend) GetGroup(name string) (*nntp.Group, error) {
	var gdb groupdb.Group
	if b.gdb.Group(name,&gdb) {
		return toNGroup(gdb),nil
	}
	return nil,nntpserver.ErrNoSuchGroup
}

func (b *backend) getArticle(id string) (*nntp.Article, error) {
	ap,ok := b.ddb.GetArticlePos(id)
	if !ok { return nil,nntpserver.ErrInvalidMessageID }
	sf := b.spool.Dayfile(ap.Dfid,false)
	if sf==nil { return nil,nntpserver.ErrInvalidMessageID }
	ar,err := articlestore.DeSerializeRAW(sf.ReadBlob(ap.Pos,ap.Length))
	if err!=nil { return nil,nntpserver.ErrInvalidMessageID }
	return ar,nil
}

// DONE: Add a way for Article Downloading without group select
// if not to implement DO: return nil, ErrNoGroupSelected
func (b *backend) GetArticleWithNoGroup(id string) (*nntp.Article, error) {
	if _,ok := nntpserver.ArticleIDOrNumber(id); ok {
		return nil,nntpserver.ErrNoGroupSelected
	}
	return b.getArticle(id)
}

func (b *backend) GetArticle(group *nntp.Group, id string) (*nntp.Article, error) {
	if num,ok := nntpserver.ArticleIDOrNumber(id); ok {
		id = b.gdb.GetArticleID(group.Name,num)
		if id=="" { return nil,nntpserver.ErrInvalidArticleNumber }
	}
	a,e := b.getArticle(id)
	if e!=nil { return nil,nntpserver.ErrInvalidArticleNumber }
	return a,nil
}

// old: GetArticles(group *nntp.Group, from, to int64) ([]NumberedArticle, error)
// channels are more suitable for large scale
func (b *backend) GetArticles(group *nntp.Group, from, to int64) (<-chan nntpserver.NumberedArticle, error) {
	if group==nil { return nil,nntpserver.ErrNoGroupSelected }
	a := make(chan nntpserver.NumberedArticle,1024)
	from2 := nntpserver.Downlimit(group.Low,from)
	to2 := nntpserver.Uplimit(group.High,to)
	go func(){
		defer func(){close(a)}()
		for i:=from2; i<=to2; i++ {
			id := b.gdb.GetArticleID(group.Name,i)
			if id=="" { continue }
			ar,err := b.getArticle(id)
			if err!=nil { continue }
			a <- nntpserver.NumberedArticle{i,ar}
		}
	}()
	return a,nil
}

func (b *backend) Authorized() bool {
	return true
}

// Authenticate and optionally swap out the backend for this session.
// You may return nil to continue using the same backend.
func (b *backend) Authenticate(user, pass string) (nntpserver.Backend, error) { return nil,nil }

func (b *backend) AllowPost() bool {
	return false
}

func (b *backend) Post(article *nntp.Article) error {
	ab := b.blobpool.Borrow()
	defer b.blobpool.Release(ab)
	defer io.Copy(ioutil.Discard,article.Body)
	articlestore.SerializeRAW(ab,article,b.blobpool)
	mid := article.MessageID()
	grps := nntpserver.GetGroups(article.Header)
	if len(grps)==0 {
		return nntpserver.ErrPostingFailed
	}
	d := getDay()
	sf := b.spool.Dayfile(d,true)
	o,l,e := sf.StoreConcurrently(ab)
	if e!=nil {
		return nntpserver.ErrPostingFailed
	}
	b.ddb.SetArticlePos(mid,dayfiledb.ArtPos{d,o,l})
	b.gdb.Numberate(grps,mid)
	return nil
}


