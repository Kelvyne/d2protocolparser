package d2protocolparser

import (
	"strings"
)

// WriteMethodTypesMap maps an as3 write method to a Golang type
var WriteMethodTypesMap = map[string]string{
	"writeVarInt":      "int32",
	"writeVarShort":    "int16",
	"writeVarLong":     "int64",
	"writeBoolean":     "bool",
	"writeByte":        "int8",
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
		// Sometimes, unsigned variables are written with signed functions
		if f.Type == "uint" && strings.HasPrefix(reduced, "int") {
			reduced = "u" + reduced // dirty but works for intX types
		}
		f.Type = reduced
		return
	}
	return
}
