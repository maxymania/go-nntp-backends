package dayfiledb

import "github.com/maxymania/go-nntp-backends/groupdb"

type ArtPos struct{
	Dfid groupdb.DayID
	Pos int64
	Length int64
}

type DayfileDB interface{
	GetArticlePos(id string) (ArtPos,bool)
	SetArticlePos(id string, pos ArtPos)
	Erase(upto groupdb.DayID)
}


