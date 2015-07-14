package prefixing

import (
	"github.com/cznic/b"
	"strings"
	"regexp"
	"github.com/maxymania/go-nntp/server"
)

var wildMatPrefix = regexp.MustCompile(`^[^\*\?]*`)

type elem struct{
	S string
}

func compare (o1,o2 interface{}) int {
	e1 := o1.(*elem)
	e2 := o2.(*elem)
	l1 := len(e1.S)
	l2 := len(e2.S)
	n := l1
	if n>l2 { n=l2 }
	for i := 0; i<n ; i++ {
		if e1.S[i]<e2.S[i] { return -1 }
		if e1.S[i]>e2.S[i] { return  1 }
	}
	if l1<l2 { return -1 }
	if l1>l2 { return  1 }
	return 0
}

func findLower(tree *b.Tree, X *elem) *elem {
	e,_ := tree.Seek(X)
	k,_,_ := e.Prev()
	if k==nil {
		e.Close()
		e,_ = tree.SeekLast()
		if e==nil { return nil }
		k,_,_ = e.Prev()
	}
	if k==nil { e.Close(); return nil }
	if compare(k,X)>0 {
		k,_,_ = e.Prev()
	}
	e.Close()
	if k==nil { return nil }
	return k.(*elem)
}
func findHigher(tree *b.Tree, X *elem) *elem {
	e,_ := tree.Seek(X)
	k,_,_ := e.Next()
	e.Close()
	if k==nil { return nil }
	return k.(*elem)

}

func consumePrefixes(sc <- chan string) []string {
	tree := b.TreeNew(compare)
	defer tree.Close()
	for s := range sc {
		e := &elem{s}
		l := findLower(tree,e)
		if l!=nil {
			if strings.HasPrefix(s,l.S) {
				continue
			}
		}
		for {
			l = findHigher(tree,e)
			if l==nil { break }
			if !strings.HasPrefix(l.S,s) { break }
			tree.Delete(l)
		}
		tree.Set(e,e)
	}
	r := make([]string,0,100)
	cur,_ := tree.SeekFirst()
	if cur==nil { return nil }
	defer cur.Close()
	for {
		k,_,_ := cur.Next()
		if k==nil { break }
		r = append(r,k.(*elem).S)
	}
	return r
}

func walkWildMat(wm *nntpserver.WildMat, d chan <- string){
	for _,rs := range wm.RuleSets {
		for _,pat := range rs.Positive {
			d <- wildMatPrefix.FindString(pat)
		}
	}
	close (d)
}

// GetPrefixes gets a list of prefixes, that represents a superset of all
// newsgroups, matched by the given WildMat.
func GetPrefixes(wm *nntpserver.WildMat) []string {
	d := make(chan string,8)
	go walkWildMat(wm,d)
	return consumePrefixes(d)
}

