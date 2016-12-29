package d2protocol

import (
	"errors"
	"fmt"
)

// ErrVerifyNoStaticLength means that a vector field that has a static length
// does not have a static length
var ErrVerifyNoStaticLength = errors.New("vector field has no length")

// ErrVerifyScalarNoWrite means that an as3 scalar type (int, uint, Number) field
// has no write method set
var ErrVerifyScalarNoWrite = errors.New("scalar type has no write method")

type verifyError struct {
	err error
	c   Class
	f   Field
}

func (e verifyError) Error() string {
	return fmt.Sprintf("%v:%v : %v", e.c.Name, e.f.Name, e.err)
}

// Verify checks that a Protocol is well-formed and that it is complete
func Verify(p *Protocol) error {
	for _, t := range p.Types {
		if err := verifyClass(t); err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func verifyClass(c Class) error {
	for _, f := range c.Fields {
		if err := verifyField(f); err != nil {
			return verifyError{err, c, f}
		}
	}
	return nil
}

func verifyField(f Field) error {
	// scalar type but no write method
	if isAs3ScalarType(f.Type) && f.WriteMethod == "" {
		return ErrVerifyScalarNoWrite
	}
	// vector with static type but no length
	if f.IsVector && !f.IsDynamicLength && f.Length == 0 && f.Type != "ByteArray" {
		return ErrVerifyNoStaticLength
	}
	return nil
}
