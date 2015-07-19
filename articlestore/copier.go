package articlestore

import "io"
import "io/ioutil"
import "errors"
import "fmt"

var TooBig = errors.New("Too-Big error");

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Modified!
func ioCopy(dst io.Writer, src io.Reader, max int) (count int, lines int, err error){
	count = 0
	lines = 0
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			for i:=0; i<nr; i++ {
				if buf[i]=='\n' { lines++ }
			}
			if nw > 0 {
				count += nw
				if count>max {
					err = TooBig
					err = errors.New(fmt.Sprint("Too-big: ",count," max=",max))
					io.Copy(ioutil.Discard,src)
					break
				}
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return
}

