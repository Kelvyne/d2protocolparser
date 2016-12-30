package d2protocolbuilder

import (
	"os"
	"reflect"
	"testing"

	"github.com/kelvyne/as3"
	"github.com/kelvyne/swf"
)

func open(t *testing.T) *as3.AbcFile {
	f, err := os.Open("./fixtures/DofusInvoker.swf")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if cErr := f.Close(); cErr != nil {
			t.Fatal(cErr)
		}
	}()

	s, err := swf.Parse(f)
	if err != nil {
		t.Error(err)
	}
	abc, err := parseAbc(&s)
	if err != nil {
		t.Error(err)
	}
	return abc
}

func Test_builder_ExtractClass(t *testing.T) {
	abc := open(t)
	simple, _ := abc.GetClassByName("GameFightOptionStateUpdateMessage")
	byteArray, _ := abc.GetClassByName("RawDataMessage")
	child, _ := abc.GetClassByName("IdentificationSuccessWithLoginTokenMessage")

	type args struct {
		class as3.Class
	}
	tests := []struct {
		name    string
		args    args
		want    Class
		wantErr bool
	}{
		{
			"simple",
			args{simple},
			Class{
				"GameFightOptionStateUpdateMessage",
				"NetworkMessage",
				[]Field{
					Field{Name: "fightId", Type: "int16", WriteMethod: "writeShort"},
					Field{Name: "teamId", Type: "byte", WriteMethod: "writeByte"},
					Field{Name: "option", Type: "byte", WriteMethod: "writeByte"},
					Field{Name: "state", Type: "bool", WriteMethod: "writeBoolean"},
				},
			},
			false,
		},
		{
			"ByteArray",
			args{byteArray},
			Class{
				"RawDataMessage",
				"NetworkMessage",
				[]Field{
					Field{
						Name: "content", Type: "byte", WriteMethod: "writeByte",
						IsVector: true, IsDynamicLength: true, WriteLengthMethod: "writeVarInt",
					},
				},
			},
			false,
		},
		{
			"child",
			args{child},
			Class{
				"IdentificationSuccessWithLoginTokenMessage",
				"IdentificationSuccessMessage",
				[]Field{
					Field{Name: "loginToken", Type: "String", WriteMethod: "writeUTF"},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &builder{
				abcFile: abc,
			}
			got, err := b.ExtractClass(tt.args.class)
			if (err != nil) != tt.wantErr {
				t.Errorf("builder.ExtractClass() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("builder.ExtractClass() = %v, want %v", got, tt.want)
			}
		})
	}
}
