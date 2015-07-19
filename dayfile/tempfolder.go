package dayfile

import "github.com/cznic/fileutil"
import "os"

type tempBlobPool struct {
	blobs chan *ArticleBlob
	dir, pre, post string
}

func TempfileArticleBlobPool(directory, prefix, postfix string,size int) ArticleBlobPool {
	return &tempBlobPool{make(chan *ArticleBlob,size),directory, prefix, postfix}
}

func(m tempBlobPool) Borrow() *ArticleBlob {
	select {
	case a := <- m.blobs:
		return a
	default:
		f,e := fileutil.TempFile(m.dir,m.pre,m.post)
		if e!=nil { return nil }
		return &ArticleBlob{Temp:f}
	}
}
func(m tempBlobPool) Release(o *ArticleBlob) {
	select {
	case m.blobs <- o:
		return
	default:
		if o.Temp!=nil {
			n := o.Temp.Name()
			defer os.Remove(n)
		}
		o.Dispose()
		return
	}
}


