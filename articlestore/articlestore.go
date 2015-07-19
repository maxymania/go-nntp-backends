package articlestore

import "io"
import "io/ioutil"
import "github.com/maxymania/go-nntp"
import "github.com/maxymania/go-nntp-backends/dayfile"
import "net/textproto"
// import "compress/flate"
import "encoding/binary"
import "encoding/gob"
import "bytes"
// compress int

type ArticleRecord struct{
	Header textproto.MIMEHeader
	Bytes int
	Lines int
}

func encode(v interface{}) ([]byte,error){
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(),err
}
func decode(s []byte,v interface{}) error{
	return gob.NewDecoder(bytes.NewReader(s)).Decode(v)
}

const max = 1<<20;

func SerializeRAW(dest io.Writer, art *nntp.Article, pool dayfile.ArticleBlobPool) error{
	var err error
	ar := ArticleRecord{art.Header,0,0}
	// temp := dayfile.NewArticleBlob()
	temp := pool.Borrow()
	defer pool.Release(temp)
	temp.Truncate()
	
	ar.Bytes,ar.Lines,err = ioCopy(temp,art.Body,max)
	if err!=nil { io.Copy(ioutil.Discard,art.Body); return err }
	
	data,err := encode(ar)
	if err!=nil { return err }
	
	err = binary.Write(dest,binary.BigEndian,int32(len(data)))
	if err!=nil { return err }
	
	_,err = dest.Write(data)
	if err!=nil { return err }
	
	_,err = temp.WriteTo(dest)
	return err
}

func DeSerializeRAW(src io.Reader) (*nntp.Article,error){
	var dl int32
	var ar ArticleRecord
	err := binary.Read(src,binary.BigEndian,&dl)
	if err!=nil { return nil,err }
	data := make([]byte,int(dl))
	_,err = io.ReadFull(src,data)
	if err!=nil { return nil,err }
	err = decode(data,&ar)
	if err!=nil { return nil,err }
	return &nntp.Article{
		Header:ar.Header,
		Bytes:ar.Bytes,
		Lines:ar.Lines,
		Body:src,
	},nil
}

