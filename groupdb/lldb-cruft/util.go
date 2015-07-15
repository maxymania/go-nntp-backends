package lldbpi

import "bytes"
import "encoding/binary"

func r2b(v interface{}) []byte {
	var b bytes.Buffer
	binary.Write(&b,binary.BigEndian,v)
	return b.Bytes()
}
func b2r(r []byte, v interface{}) bool {
	return binary.Read(bytes.NewReader(r),binary.BigEndian,v)==nil
}
