package main

import (
	"encoding/binary"
	"fmt"
)

func main() {
	var b byte = 1
	buf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(buf, uint64(b))
	fmt.Println(buf[:bytesWritten])
}
