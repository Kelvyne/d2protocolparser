package d2protocolparser

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
	unsigned, _ := abc.GetClassByName("CharacterLevelUpMessage")
	typeClass, _ := abc.GetClassByName("KrosmasterFigure")
	bbw, _ := abc.GetClassByName("IdentificationMessage")
	typeManagerVector, _ := abc.GetClassByName("BasicCharactersListMessage")
	typeManager, _ := abc.GetClassByName("GameContextActorInformations")
	longInt, _ := abc.GetClassByName("AllianceInvitationMessage")

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
					Field{Name: "fightId", Type: "uint16", WriteMethod: "writeShort", Method: "UInt16"},
					Field{Name: "teamId", Type: "uint8", WriteMethod: "writeByte", Method: "UInt8"},
					Field{Name: "option", Type: "uint8", WriteMethod: "writeByte", Method: "UInt8"},
					Field{Name: "state", Type: "bool", WriteMethod: "writeBoolean", Method: "Boolean"},
				},
				5927,
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
						Name: "content", Type: "uint8", WriteMethod: "writeByte", Method: "UInt8",
						IsVector: true, IsDynamicLength: true, WriteLengthMethod: "writeVarInt",
					},
				},
				6253,
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
					Field{Name: "loginToken", Type: "string", WriteMethod: "writeUTF", Method: "String"},
				},
				6209,
			},
			false,
		},
		{
			"unsigned",
			args{unsigned},
			Class{
				"CharacterLevelUpMessage",
				"NetworkMessage",
				[]Field{
					Field{Name: "newLevel", Type: "uint8", WriteMethod: "writeByte", Method: "UInt8"},
				},
				5670,
			},
			false,
		},
		{
			"type",
			args{typeClass},
			Class{
				"KrosmasterFigure",
				"",
				[]Field{
					Field{Name: "uid", Type: "string", WriteMethod: "writeUTF", Method: "String"},
					Field{Name: "figure", Type: "uint16", WriteMethod: "writeVarShort", Method: "VarUInt16"},
					Field{Name: "pedestal", Type: "uint16", WriteMethod: "writeVarShort", Method: "VarUInt16"},
					Field{Name: "bound", Type: "bool", WriteMethod: "writeBoolean", Method: "Boolean"},
				},
				397,
			},
			false,
		},
		{
			"BooleanByteWrapper",
			args{bbw},
			Class{
				"IdentificationMessage",
				"NetworkMessage",
				[]Field{
					Field{Name: "version", Type: "VersionExtended"},
					Field{Name: "lang", Type: "string", WriteMethod: "writeUTF", Method: "String"},
					Field{Name: "credentials", Type: "int8", WriteMethod: "writeByte", Method: "Int8", IsVector: true, IsDynamicLength: true, WriteLengthMethod: "writeVarInt"},
					Field{Name: "serverId", Type: "int16", WriteMethod: "writeShort", Method: "Int16"},
					Field{Name: "autoconnect", Type: "Boolean", UseBBW: true, BBWPosition: 0},
					Field{Name: "useCertificate", Type: "Boolean", UseBBW: true, BBWPosition: 1},
					Field{Name: "useLoginToken", Type: "Boolean", UseBBW: true, BBWPosition: 2},
					Field{Name: "sessionOptionalSalt", Type: "int64", WriteMethod: "writeVarLong", Method: "VarInt64"},
					Field{Name: "failedAttempts", Type: "uint16", WriteMethod: "writeVarShort", Method: "VarUInt16", IsVector: true, IsDynamicLength: true, WriteLengthMethod: "writeShort"},
				},
				4,
			},
			false,
		},
		{
			"typeManagerVector",
			args{typeManagerVector},
			Class{
				"BasicCharactersListMessage",
				"NetworkMessage",
				[]Field{
					Field{Name: "characters", Type: "CharacterBaseInformations", IsVector: true, IsDynamicLength: true, WriteLengthMethod: "writeShort", UseTypeManager: true},
				},
				6475,
			},
			false,
		},
		{
			"typeManager",
			args{typeManager},
			Class{
				"GameContextActorInformations",
				"",
				[]Field{
					Field{Name: "contextualId", Type: "float64", WriteMethod: "writeDouble", Method: "Double"},
					Field{Name: "look", Type: "EntityLook"},
					Field{Name: "disposition", Type: "EntityDispositionInformations", UseTypeManager: true},
				},
				150,
			},
			false,
		},
		{
			"longInt",
			args{longInt},
			Class{
				"AllianceInvitationMessage",
				"NetworkMessage",
				[]Field{
					Field{Name: "targetId", Type: "int64", WriteMethod: "writeVarLong", Method: "VarInt64"},
				},
				6395,
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
