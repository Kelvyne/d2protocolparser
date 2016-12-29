package d2protocol

import "fmt"

var reduceMap = map[string]string{
	"writeVarInt":      "",
	"writeVarShort":    "",
	"writeVarLong":     "",
	"writeBoolean":     "",
	"writeByte":        "",
	"writeShort":       "",
	"writeInt":         "",
	"writeUnsignedInt": "",
	"writeFloat":       "",
	"writeDouble":      "",
}

func reduceType(f *Field) {
	if f.WriteMethod == "" {
		return
	}
	reduced, canReduce := reduceMap[f.WriteMethod]
	if !canReduce {
		fmt.Printf("%v\n", f.WriteMethod)
		return
	}
	f.Type = reduced
	return
}
