package db

import (
	"testing"
)

func Test_getDataSource(t *testing.T) {
	type args struct {
		dataBaseUrl string
	}
	tests := []struct {
		name                 string
		args                 args
		wantDriverName       string
		wantConnectionString string
	}{
		{
			name: "get postgresql datasource",
			args: args{
				dataBaseUrl: "postgresql://postgres:postgres@127.0.0.1:5432/db?sslmode=disable",
			},
			wantDriverName:       "postgresql",
			wantConnectionString: "postgres:postgres@127.0.0.1:5432/db?sslmode=disable",
		},

		{
			name: "get mysql datasource",
			args: args{
				dataBaseUrl: "username:password@tcp(localhost:3306)/dbname",
			},
			wantDriverName:       "mysql",
			wantConnectionString: "username:password@tcp(localhost:3306)/dbname",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDriverName, gotConnectionString := getDataSource(tt.args.dataBaseUrl)
			if gotDriverName != tt.wantDriverName {
				t.Errorf("getDataSource() got = %v, want %v", gotDriverName, tt.wantDriverName)
			}
			if gotConnectionString != tt.wantConnectionString {
				t.Errorf("getDataSource() got1 = %v, want %v", gotConnectionString, tt.wantConnectionString)
			}
		})
	}
}
