package dayfile

import "github.com/maxymania/go-nntp-backends/groupdb"
import "sync/atomic"
import "sync"
import "bytes"
import "os"
import "io"
import "errors"
import "fmt"

type ArticleBlob struct {
	Buf  *bytes.Buffer
	Temp *os.File
}
func NewArticleBlob() *ArticleBlob{
	return &ArticleBlob{Buf:new(bytes.Buffer)}
}
func NewArticleBlobFromBuffer(buf *bytes.Buffer) *ArticleBlob{
	return &ArticleBlob{Buf:buf}
}
func NewArticleBlobFromArray(arr []byte) *ArticleBlob{
	return &ArticleBlob{Buf:bytes.NewBuffer(arr)}
}
func (ab *ArticleBlob) Write(p []byte) (n int, err error){
	if ab.Buf!=nil { return ab.Buf.Write(p) }
	return ab.Temp.Write(p)
}
func (ab *ArticleBlob) WriteTo(dst io.Writer) (n int64, err error){
	if ab.Buf!=nil { return ab.Buf.WriteTo(dst) }
	ab.Temp.Seek(0,0) // (re)set to start
	return io.Copy(dst,ab.Temp)
}
func (ab *ArticleBlob) Truncate() error {
	if ab.Buf!=nil { ab.Buf.Truncate(0); return nil }
	err := ab.Temp.Truncate(0)
	return err
}
func (ab *ArticleBlob) Dispose() error {
	if ab.Temp==nil { return nil }
	return ab.Temp.Close()
}
func (ab *ArticleBlob) Size() int64 {
	if ab.Temp!=nil {
		s,e := ab.Temp.Stat()
		if e==nil { return s.Size() }
	} else {
		return int64(ab.Buf.Len())
	}
	return 0
}

var NoDataToWrite = errors.New("no data to write")

type StorageFile struct{
	length int64
	fobj *os.File
}
func (s *StorageFile) ReadBlob(offset, length int64) io.Reader {
	return io.NewSectionReader(s.fobj,offset,length)
}
func (s *StorageFile) StoreConcurrently(a *ArticleBlob) (offset, length int64, err error){
	size := a.Size()
	if size==0 { return 0,0,NoDataToWrite }
	pos := atomic.AddInt64(&s.length,size)-size
	if a.Temp!=nil {
		a.Temp.Seek(0,0) // (re)set to start
		e := writeTo(s.fobj,a.Temp,pos,size)
		return pos,size,e
	} else {
		_,e := s.fobj.WriteAt(a.Buf.Bytes(),size)
		return pos,size,e
	}
	return 0,0,NoDataToWrite
}



type StorageFilePool struct{
	mutex sync.RWMutex
	files map[groupdb.DayID]*StorageFile
	basepath string
}
func (s *StorageFilePool) Dayfile(db groupdb.DayID, create bool) *StorageFile {
	sf,ok := s.getDayfile(db,create)
	if !ok {
		sf = s.createDayfile(db,create)
	}
	return sf
}
func (s *StorageFilePool) getDayfile(db groupdb.DayID, create bool) (*StorageFile,bool) {
	s.mutex.RLock(); defer s.mutex.RUnlock()
	sf,ok := s.files[db]
	return sf,ok
}
func (s *StorageFilePool) createDayfile(db groupdb.DayID, create bool) *StorageFile {
	s.mutex.Lock(); defer s.mutex.Unlock()
	sf,ok := s.files[db]
	if !ok {
		sf = s.create(db,create)
		s.files[db] = sf
	} else if create && sf==nil {
		sf = s.create(db,create)
		s.files[db] = sf
	}
	return sf
}
func (s *StorageFilePool) create(db groupdb.DayID, create bool) *StorageFile {
	y,m,d := db.Date()
	sr := s.basepath+fmt.Sprintf("df%04x%02x%02x.df",y,m,d)
	sf := new(StorageFile)
	var err error
	fm := os.O_RDWR
	if create { fm |= os.O_CREATE }
	var e error
	sf.fobj,e = os.OpenFile(sr,fm,0600)
	if e!=nil { return nil }
	sf.length,_ = sf.fobj.Seek(0,2)
	return sf
}


