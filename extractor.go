package d2protocol

import (
	"fmt"
	"strings"

	"errors"

	"github.com/kelvyne/as3"
	"github.com/kelvyne/as3/bytecode"
)

func (b *builder) ExtractClass(class as3.Class) (Class, error) {
	trait, found := findMethodWithPrefix(class, "serializeAs_")
	if !found {
		return Class{}, fmt.Errorf("serialize method not found in class %v", class.Name)
	}

	m := b.abcFile.Methods[trait.Method]
	if err := m.BodyInfo.Disassemble(); err != nil {
		return Class{}, fmt.Errorf("failed to disassemble %v", class.Name)
	}

	fields, err := b.extractMessageFields(class)
	if err != nil {
		return Class{}, fmt.Errorf("failed retrieve %v's fields", class.Name)
	}

	fieldMap := map[string]*Field{}
	for i, f := range fields {
		fieldMap[f.Name] = &fields[i]
	}

	if err := b.extractSerializeMethods(class, m, fieldMap); err != nil {
		return Class{}, err
	}

	for i := range fields {
		reduceType(&fields[i])
	}

	return Class{class.Name, class.SuperName, fields}, nil
}

func (b *builder) extractMessageFields(class as3.Class) (f []Field, err error) {
	createField := func(name string, typeId uint32) Field {
		t := b.abcFile.Source.ConstantPool.MultinameString(typeId)
		var isVector bool
		if strings.HasPrefix(t, "Vector<") {
			typename := b.abcFile.Source.ConstantPool.Multinames[typeId]
			param := b.abcFile.Source.ConstantPool.MultinameString(typename.Params[0])
			t = param
			isVector = true
		} else if t == "ByteArray" {
			isVector = true
		}
		return Field{Name: name, Type: t, IsVector: isVector}
	}

	for _, slot := range class.InstanceTraits.Slots {
		name := b.abcFile.Source.ConstantPool.Multinames[slot.Source.Name]
		if !isPublicNamespace(b.abcFile, name.Namespace) {
			continue
		}
		field := createField(slot.Name, slot.Source.Typename)
		f = append(f, field)
	}

	// NetworkDataContainerMessage uses a pair of setter/getter to store content
	// It seems to be useless and the only packet that does so we need to
	// also check for pairs of getter/setter
	type getSetter struct {
		getter     bool
		getterType uint32
		setter     bool
	}
	getSetters := map[string]*getSetter{}

	for _, m := range class.InstanceTraits.Methods {
		isGetter := m.Source.Kind == bytecode.TraitsInfoGetter
		isSetter := m.Source.Kind == bytecode.TraitsInfoSetter
		name := b.abcFile.Source.ConstantPool.Multinames[m.Source.Name]
		if !(isGetter || isSetter) || !isPublicNamespace(b.abcFile, name.Namespace) {
			continue
		}
		v, ok := getSetters[m.Name]
		if !ok {
			v = &getSetter{}
			getSetters[m.Name] = v
		}
		v.getter = v.getter || isGetter
		v.setter = v.setter || isSetter
		if isGetter {
			info := b.abcFile.Source.Methods[m.Source.Method]
			v.getterType = info.ReturnType
		}
	}

	for name, gs := range getSetters {
		if !(gs.getter && gs.setter) {
			continue
		}
		field := createField(name, gs.getterType)
		f = append(f, field)
	}
	return
}

func handleSimpleProp(b *builder, class as3.Class, fields map[string]*Field, instrs []bytecode.Instr, last *Field) (*Field, error) {
	get := instrs[0]
	call := instrs[1]
	getMultiname := b.abcFile.Source.ConstantPool.Multinames[get.Operands[0]]
	callMultiname := b.abcFile.Source.ConstantPool.Multinames[call.Operands[0]]
	if !isPublicQName(b.abcFile, getMultiname) {
		return nil, nil
	}

	prop := b.abcFile.Source.ConstantPool.Strings[getMultiname.Name]
	writeMethod := b.abcFile.Source.ConstantPool.Strings[callMultiname.Name]

	if !strings.HasPrefix(writeMethod, "write") {
		return nil, nil
	}

	field, ok := fields[prop]
	if !ok {
		return nil, fmt.Errorf("%v.%v.%v field not found", class.Namespace, class.Name, prop)
	}

	field.WriteMethod = writeMethod
	return field, nil
}

func handleVecPropLength(b *builder, class as3.Class, fields map[string]*Field, instrs []bytecode.Instr, last *Field) (*Field, error) {
	get := instrs[0]
	getLen := instrs[1]
	call := instrs[2]

	getMultiname := b.abcFile.Source.ConstantPool.Multinames[get.Operands[0]]
	getLenMultiname := b.abcFile.Source.ConstantPool.Multinames[getLen.Operands[0]]
	callMultiname := b.abcFile.Source.ConstantPool.Multinames[call.Operands[0]]
	if !isPublicQName(b.abcFile, getMultiname) || !isPublicQName(b.abcFile, getLenMultiname) {
		return nil, nil
	}

	if b.abcFile.Source.ConstantPool.Strings[getLenMultiname.Name] != "length" {
		return nil, nil
	}
	prop := b.abcFile.Source.ConstantPool.Strings[getMultiname.Name]

	field, ok := fields[prop]
	if !ok || !field.IsVector {
		return nil, fmt.Errorf("%v.%v: write length on non-vector %v", class.Namespace, class.Name, prop)
	}
	writeMethod := b.abcFile.Source.ConstantPool.Strings[callMultiname.Name]

	if !strings.HasPrefix(writeMethod, "write") {
		return nil, nil
	}

	field.IsDynamicLength = true
	field.WriteLengthMethod = writeMethod
	return field, nil
}

func handleTypeManagerProp(b *builder, class as3.Class, fields map[string]*Field, instrs []bytecode.Instr, last *Field) (*Field, error) {
	get := instrs[0]
	getType := instrs[1]
	call := instrs[2]

	getMultiname := b.abcFile.Source.ConstantPool.Multinames[get.Operands[0]]
	getTypeMultiname := b.abcFile.Source.ConstantPool.Multinames[getType.Operands[0]]
	callMultiname := b.abcFile.Source.ConstantPool.Multinames[call.Operands[0]]

	if !isPublicQName(b.abcFile, getMultiname) || !isPublicQName(b.abcFile, getTypeMultiname) {
		return nil, nil
	}

	if b.abcFile.Source.ConstantPool.Strings[getTypeMultiname.Name] != "getTypeId" {
		return nil, nil
	}

	prop := b.abcFile.Source.ConstantPool.Strings[getMultiname.Name]
	field, ok := fields[prop]
	if !ok {
		return nil, fmt.Errorf("%v.%v: getTypeId on %v field", class.Namespace, class.Name, prop)
	}

	writeMethod := b.abcFile.Source.ConstantPool.Strings[callMultiname.Name]
	if writeMethod != "writeShort" {
		return nil, fmt.Errorf("%v.%v: invalid %v for getTypeId", class.Namespace, class.Name, writeMethod)
	}

	field.UseTypeManager = true
	return field, nil
}

func handleVecScalarProp(b *builder, class as3.Class, fields map[string]*Field, instrs []bytecode.Instr, last *Field) (*Field, error) {
	get := instrs[0]
	getIndex := instrs[2]
	getMultiname := b.abcFile.Source.ConstantPool.Multinames[get.Operands[0]]
	getIndexMultiname := b.abcFile.Source.ConstantPool.Multinames[getIndex.Operands[0]]
	if !isPublicQName(b.abcFile, getMultiname) || getIndexMultiname.Kind != bytecode.MultinameKindMultinameL {
		return nil, nil
	}

	call := instrs[3]
	callMultiname := b.abcFile.Source.ConstantPool.Multinames[call.Operands[0]]
	if callMultiname.Kind != bytecode.MultinameKindQName {
		return nil, nil
	}

	writeMethod := b.abcFile.Source.ConstantPool.Strings[callMultiname.Name]
	if !strings.HasPrefix(writeMethod, "write") {
		return nil, fmt.Errorf("%v.%v: %v method for vector of scalar types", class.Namespace, class.Name, writeMethod)
	}

	prop := b.abcFile.Source.ConstantPool.Strings[getMultiname.Name]
	field, ok := fields[prop]
	if !ok || !field.IsVector {
		return nil, fmt.Errorf("%v.%v: vector of scalar write on %v field", class.Namespace, class.Name, prop)
	}
	field.WriteMethod = writeMethod
	return field, nil
}

func handleVecTypeManagerProp(b *builder, class as3.Class, fields map[string]*Field, instrs []bytecode.Instr, last *Field) (*Field, error) {
	get := instrs[0]
	lex := instrs[3]
	call := instrs[5]
	getMultiname := b.abcFile.Source.ConstantPool.Multinames[get.Operands[0]]
	lexMultiname := b.abcFile.Source.ConstantPool.Multinames[lex.Operands[0]]
	callMultiname := b.abcFile.Source.ConstantPool.Multinames[call.Operands[0]]

	if !isPublicQName(b.abcFile, getMultiname) {
		return nil, nil
	}

	lexNs := b.abcFile.Source.ConstantPool.Namespaces[lexMultiname.Namespace]
	lexNsName := b.abcFile.Source.ConstantPool.Strings[lexNs.Name]
	if !strings.HasPrefix(lexNsName, "com.ankamagames.dofus.network.types") {
		return nil, nil
	}

	callName := b.abcFile.Source.ConstantPool.Strings[callMultiname.Name]
	if callName != "getTypeId" {
		return nil, nil
	}

	prop := b.abcFile.Source.ConstantPool.Strings[getMultiname.Name]
	f, ok := fields[prop]
	if !ok || !f.IsVector {
		return nil, fmt.Errorf("%v.%v: %v field is not a vector", class.Namespace, class.Name, prop)
	}

	f.UseTypeManager = true
	return f, nil
}

func handleVecPropDynamicLen(b *builder, class as3.Class, fields map[string]*Field, instrs []bytecode.Instr, last *Field) (*Field, error) {
	push := instrs[5]
	len := push.Operands[0]
	if last == nil || !last.IsVector || last.IsDynamicLength {
		return nil, errors.New("vector length found but no dynamic vector")
	}
	last.Length = len
	return last, nil
}

func handleGetProperty(b *builder, class as3.Class, fields map[string]*Field, instrs []bytecode.Instr, last *Field) (*Field, error) {
	get := instrs[0]
	multi := b.abcFile.Source.ConstantPool.Multinames[get.Operands[0]]
	if !isPublicQName(b.abcFile, multi) {
		return nil, nil
	}
	name := b.abcFile.Source.ConstantPool.Strings[multi.Name]
	field, ok := fields[name]
	if !ok {
		return nil, nil
	}
	return field, nil
}

func (b *builder) extractSerializeMethods(class as3.Class, m as3.Method, fields map[string]*Field) error {
	checkPattern := func(instrs []bytecode.Instr, pattern []string) bool {
		if len(pattern) > len(instrs) {
			return false
		}
		for i, str := range pattern {
			if !strings.HasPrefix(instrs[i].Model.Name, str) {
				return false
			}
		}
		return true
	}

	type pattern struct {
		Fn      func(*builder, as3.Class, map[string]*Field, []bytecode.Instr, *Field) (*Field, error)
		Pattern []string
	}

	// These must be sorted by pattern length to be sure to not miss any pattern
	patterns := []pattern{
		{handleVecPropDynamicLen, []string{"getlocal", "increment", "convert", "setlocal", "getlocal", "pushbyte", "iflt"}},
		{handleVecTypeManagerProp, []string{"getproperty", "getlocal", "getproperty", "getlex", "astypelate", "callproperty"}},
		{handleVecScalarProp, []string{"getproperty", "getlocal", "getproperty", "callpropvoid"}},
		{handleVecPropLength, []string{"getproperty", "getproperty", "callpropvoid"}},
		{handleSimpleProp, []string{"getproperty", "callpropvoid"}},
		{handleTypeManagerProp, []string{"getproperty", "callproperty", "callpropvoid"}},
		{handleGetProperty, []string{"getproperty"}},
	}

	instrs := m.BodyInfo.Instructions
	instrLen := len(m.BodyInfo.Instructions)
	var last *Field
	for i := 0; i < instrLen; {
		var f *Field
		var err error
		for _, p := range patterns {
			if checkPattern(instrs[i:], p.Pattern) {
				f, err = p.Fn(b, class, fields, instrs[i:], last)
				if err != nil {
					return err
				}
				i += len(p.Pattern)
			}
		}
		if f == nil {
			i++
		} else {
			last = f
		}
	}
	return nil
}
