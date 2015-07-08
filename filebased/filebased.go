
/*
 A (slapdash coded) file based backend.
 */
package filebased

import "encoding/csv"
import "encoding/json"
import "github.com/maxymania/go-nntp"
import "github.com/maxymania/go-nntp/server"
import "os"
import "net/textproto"
import "bufio"
//import "strconv"
import "io"
import "io/ioutil"
import "fmt"

var Lock func(s string) (io.Closer, error)

func loadJson(n string) (r []int64){
	f,e := os.Open(n)
	if e!=nil { return []int64{0,0,0} }
	defer f.Close()
	nd := json.NewDecoder(f)
	e = nd.Decode(&r)
	if e!=nil { return []int64{0,0,0} }
	r = append(r,0,0,0)[:3]
	return
}

func NewDir(s string) nntpserver.Backend {
	return dirString(s)
}

type dirString string

func(d dirString) getFileName(g *nntp.Group,id string) (string,error) {
	if len(id)==0 { return "",nntpserver.ErrInvalidMessageID }
	if id[0]=='<' { return "",nntpserver.ErrInvalidMessageID }
	if !((id[0]<'0')||(id[0]>'9')) { return "",nntpserver.ErrInvalidMessageID }
	return string(d)+"/c"+g.Name+"/"+id,nil
}

func(d dirString) ListGroups(max int) (<- chan *nntp.Group, error){
	r := make(chan *nntp.Group)
	go func(){
		defer func(){close(r)}()
		f,e := os.Open(string(d)+"/groups-csv")
		if e!=nil { return }
		defer f.Close()
		cr := csv.NewReader(f)
		for {
			grp,e := cr.Read()
			if e!=nil { return }
			i := loadJson(string(d)+"/j"+grp[0])
			r <- &nntp.Group{grp[0],grp[1],i[0],i[1],i[2],nntp.PostingPermitted}
		}
	}()
	return r,nil
}
func(d dirString) GetGroup(name string) (*nntp.Group, error) {
	i := loadJson(string(d)+"/j"+name)
	return &nntp.Group{name,name,i[0],i[1],i[2],nntp.PostingPermitted},nil
}
func(d dirString) GetArticleWithNoGroup(id string) (*nntp.Article, error) {
	return nil, nntpserver.ErrNoGroupSelected
}
func(d dirString) GetArticle(group *nntp.Group, id string) (*nntp.Article, error) {
	fn,e := d.getFileName(group,id)
	if e!=nil { return nil,e }
	r,e := os.Open(fn)
	if e!=nil { return nil,nntpserver.ErrInvalidArticleNumber }
	tr := textproto.NewReader(bufio.NewReader(r))
	a := new(nntp.Article)
	a.Header,e = tr.ReadMIMEHeader()
	if e!=nil { return nil,nntpserver.ErrInvalidArticleNumber }
	a.Body = tr.R
	return a,nil
}
func(d dirString) GetArticles(group *nntp.Group, from, to int64) (<- chan nntpserver.NumberedArticle, error){
	r := make(chan nntpserver.NumberedArticle)
	go func(){
		defer func(){close(r)}()
		for i := from; i<=to; i++{
			ar,e := d.GetArticle(group,fmt.Sprint(i))
			if e!=nil { continue }
			r <- nntpserver.NumberedArticle{i,ar}
		}
	}()
	return r,nil
}
func(d dirString) Authorized() bool { return false }
func(d dirString) Authenticate(user, pass string) (nntpserver.Backend, error) { return nil,nil }
func(d dirString) AllowPost() bool { return true }
func(d dirString) Post(article *nntp.Article) error {
	ng := article.Header["Newsgroups"]
	if len(ng)==1 {
		name := string(d)+"/j"+ng[0]
		i := loadJson(name)
		c,e := Lock(name)
		if e==nil {
			i[0]++
			num := i[1]
			i[1]++
			fn,_ := d.getFileName(&nntp.Group{Name:ng[0]},fmt.Sprint(num))
			c.Close()
			fo,e := os.Create(fn)
			if e==nil{
				defer fo.Close()
				for k,vv := range article.Header {
					for _,v := range vv {
						fmt.Fprintf(fo,"%s: %s\r\n",k,v)
					}
				}
				fmt.Fprintf(fo,"\r\n")
				io.Copy(fo,article.Body)
				return nil
			}
		}
	}
	io.Copy(ioutil.Discard,article.Body)
	return nil
}



