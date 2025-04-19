package db

import (
	"errors"
	"flag"
	"testing"

	"github.com/streamdp/ccd/config"
)

func Test_getDatabaseUrl(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		envs    map[string]string
		wantErr error
	}{
		{
			name:    "empty url",
			want:    "",
			envs:    nil,
			wantErr: errDatabaseUrl,
		},
		{
			name: "get database url",
			want: "postgresql://postgres:postgres@127.0.0.1:5432/db?sslmode=disable",
			envs: map[string]string{
				"CCDC_DATABASEURL": "postgresql://postgres:postgres@127.0.0.1:5432/db?sslmode=disable",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}
			flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
			config.ParseFlags()

			got, err := getDatabaseUrl()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("getDatabaseUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getDatabaseUrl() got = %v, want %v", got, tt.want)
			}
		})
	}
}

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
