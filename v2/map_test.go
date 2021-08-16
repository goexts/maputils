package extmap

import (
	"testing"
)

func Test_innerMap_Bind(t *testing.T) {
	type fields struct {
		data   map[interface{}]interface{}
		option *Option
	}
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				data:   nil,
				option: nil,
			},
			args: args{
				v: struct {
				}{},
			},
			wantErr: false,
		},
		{
			name: "",
			fields: fields{
				data:   nil,
				option: nil,
			},
			args: args{
				v: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			if err := m.Bind(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Bind() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
