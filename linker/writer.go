package linker

import (
	"encoding/binary"
	"io"
)

type binaryWriter struct {
	w   io.WriteSeeker
	bo  binary.ByteOrder
	err error
}

func (bw *binaryWriter) write(data interface{}) {
	if bw.err != nil {
		return
	}
	bw.err = binary.Write(bw.w, bw.bo, data)
}

func (bw *binaryWriter) writeAt(data interface{}, offset int64) {
	if bw.err != nil {
		return
	}
	_, bw.err = bw.w.Seek(offset, io.SeekStart)
	bw.write(data)
}
