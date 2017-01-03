package d2protocolparser

import (
	"strings"

	"github.com/kelvyne/as3"
	"github.com/kelvyne/as3/bytecode"
)

func findMethodWithPrefix(c as3.Class, prefix string) (bytecode.TraitsInfo, bool) {
	for _, t := range c.InstanceTraits.Methods {
		if strings.HasPrefix(t.Name, prefix) {
			return t.Source, true
		}
	}
	return bytecode.TraitsInfo{}, false
}

func isPublicQName(abc *as3.AbcFile, m bytecode.MultinameInfo) bool {
	if m.Kind != bytecode.MultinameKindQName {
		return false
	}
	return isPublicNamespace(abc, m.Namespace)
}

func isPublicNamespace(abc *as3.AbcFile, nsID uint32) bool {
	ns := abc.Source.ConstantPool.Namespaces[nsID]
	return ns.Kind == bytecode.NamespaceKindPackageNamespace || ns.Kind == bytecode.NamespaceKindNamespace
}

func isAs3ScalarType(t string) bool {
	scalarTypes := []string{"int", "uint", "float", "bool", "byte"}
	for _, s := range scalarTypes {
		if strings.HasPrefix(t, s) {
			return true
		}
	}
	return false
}
