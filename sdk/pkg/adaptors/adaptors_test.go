package adaptors

import "testing"

func TestGetFuncNameByTypeNames(t *testing.T) {
	type args struct {
		typeNameIn  string
		TypeNameOut string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				typeNameIn:  "int32",
				TypeNameOut: "string",
			},
			want: "IntegerToString",
		},
		{
			name: "test2",
			args: args{
				typeNameIn:  "string",
				TypeNameOut: "float64",
			},
			want: "StringToFloat",
		},
		{
			name: "test3",
			args: args{
				typeNameIn:  "bool",
				TypeNameOut: "int",
			},
			want: "BoolToInteger",
		},
		{
			name: "test4",
			args: args{
				typeNameIn:  "float32",
				TypeNameOut: "int64",
			},
			want: "FloatToInteger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFuncNameByTypeNames(tt.args.typeNameIn, tt.args.TypeNameOut); got != tt.want {
				t.Errorf("GetFuncNameByTypeNames() = %v, want %v", got, tt.want)
			}
		})
	}
}
