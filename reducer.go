package d2protocolparser

import (
	"strings"
)

var writeMethodTypesMap = map[string]string{
	"writeVarShort":    "int16",
	"writeVarInt":      "int32",
	"writeVarLong":     "int64",
	"writeBoolean":     "bool",
	"writeByte":        "int8",
	"writeShort":       "int16",
	"writeInt":         "int32",
	"writeUnsignedInt": "uint32",
	"writeFloat":       "float32",
	"writeDouble":      "float64",
	"writeUTF":         "string",
}

func reduceType(f *Field) {
	if f.Type == "Boolean" {
		f.Type = "bool"
	}
	if f.WriteMethod == "" {
		return
	} else if f.WriteMethod == "writeBytes" {
		// hack to get NetworkDataContainerMessage working
		f.IsVector = true
		f.IsDynamicLength = true
		f.WriteLengthMethod = "writeVarInt"
		f.WriteMethod = "writeByte"
	}
	reduced, canReduce := writeMethodTypesMap[f.WriteMethod]
	if canReduce {
		// Sometimes, unsigned variables are written with signed functions
		if f.Type == "uint" && strings.HasPrefix(reduced, "int") {
			reduced = "u" + reduced // dirty but works for intX types
		}
		f.Type = reduced
	}
	return
}

var typesToMethodMap = map[string]string{
	"int8":    "Int8",
	"int16":   "Int16",
	"int32":   "Int32",
	"int64":   "Int64",
	"uint8":   "UInt8",
	"uint16":  "UInt16",
	"uint32":  "UInt32",
	"uint64":  "UInt64",
	"float32": "Float",
	"float64": "Double",
	"string":  "String",
	"bool":    "Boolean",
}

func reduceMethod(f *Field) {
	m, ok := typesToMethodMap[f.Type]
	if !ok || f.WriteMethod == "" {
		return
	}
	if strings.Contains(f.WriteMethod, "Var") {
		m = "Var" + m
	}
	f.Method = m
}
