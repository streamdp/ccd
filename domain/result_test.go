package domain

import (
	"reflect"
	"testing"
)

func TestResult_UpdateDataField(t *testing.T) {
	type args struct {
		data interface{}
	}
	tests := []struct {
		name string
		res  *Result
		args args
		want *Result
	}{
		{
			name: "update data filed",
			res: &Result{
				Code:    200,
				Message: "test_message",
				Data:    "test_data",
			},
			args: args{
				data: "321",
			},
			want: &Result{
				Code:    200,
				Message: "test_message",
				Data:    "321",
			},
		},
		{
			name: "update data filed with nil value",
			res: &Result{
				Code:    200,
				Message: "test_message",
				Data:    "test_data",
			},
			args: args{},
			want: &Result{
				Code:    200,
				Message: "test_message",
				Data:    nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.res.UpdateDataField(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateDataField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_UpdateAllFields(t *testing.T) {
	type args struct {
		code int
		msg  string
		data interface{}
	}
	tests := []struct {
		name string
		res  *Result
		args args
		want *Result
	}{
		{
			name: "update empty result fields",
			res:  &Result{},
			args: args{
				code: 200,
				msg:  "test_message",
				data: "test_data",
			},
			want: &Result{
				Code:    200,
				Message: "test_message",
				Data:    "test_data",
			},
		},
		{
			name: "update not empty result fields",
			res: &Result{
				Code:    400,
				Message: "test_message",
				Data:    "test_data",
			},
			args: args{
				code: 200,
				msg:  "new_test_message",
				data: "new_test_data",
			},
			want: &Result{
				Code:    200,
				Message: "new_test_message",
				Data:    "new_test_data",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.res.UpdateAllFields(tt.args.code, tt.args.msg, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateAllFields() = %v, want %v", got, tt.want)
			}
		})
	}
}
