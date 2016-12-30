package d2protocol

// WriteMethodTypesMap maps an as3 write method to a Golang type
var WriteMethodTypesMap = map[string]string{
	"writeVarInt":      "int32",
	"writeVarShort":    "int16",
	"writeVarLong":     "int64",
	"writeBoolean":     "bool",
	"writeByte":        "byte",
	"writeShort":       "int16",
	"writeInt":         "int32",
	"writeUnsignedInt": "uint32",
	"writeFloat":       "float32",
	"writeDouble":      "float64",
}

func reduceType(f *Field) {
	if f.WriteMethod == "" {
		return
	}
	reduced, canReduce := WriteMethodTypesMap[f.WriteMethod]
	if canReduce {
		f.Type = reduced
		return
	}
	return
}
