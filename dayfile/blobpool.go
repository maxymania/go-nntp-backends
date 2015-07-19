package dayfile

// ArticleBlob

type ArticleBlobPool interface{
	Borrow() *ArticleBlob
	Release(o *ArticleBlob)
}

func MemoryArticleBlobPool(n int) ArticleBlobPool {
	return make(memoryBlobPool,n)
}

type memoryBlobPool chan *ArticleBlob
func(m memoryBlobPool) Borrow() *ArticleBlob {
	select {
	case a := <- m:
		return a
	default:
		return NewArticleBlob()
	}
}
func(m memoryBlobPool) Release(o *ArticleBlob) {
	select {
	case m <- o:
		return
	default:
		o.Dispose()
		return
	}
}

type FixedBlobPool chan *ArticleBlob
func (f FixedBlobPool) Borrow() *ArticleBlob { return <- f }
func (f FixedBlobPool) Release(o *ArticleBlob) { f <- o }

