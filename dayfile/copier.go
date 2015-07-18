package dayfile

import "github.com/cznic/bufs"
import "io"

func writeTo(dest io.WriterAt, src io.Reader, pos, lng int64) error{
	buf := bufs.GCache.Get(1024)
	defer bufs.GCache.Put(buf)
	c := pos
	e := pos + lng
	for c<e{
		n := len(buf)
		if int64(n)>(e-c) { n = int(e-c) }
		n,err := src.Read(buf[:n])
		if err!=nil { return err }
		_,err = dest.WriteAt(buf[:n],c)
		if err!=nil { return err }
		c+=int64(n)
	}
	return nil
}


