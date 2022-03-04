package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_unzipSource(t *testing.T) {
	type args struct {
		source      string
		destination string
	}
	tmpDir, err := ioutil.TempDir("", "temp_extension")
	if err != nil {
		t.Errorf("%s", err)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "unzipSource", args: args{
			source:      "tests/he_module-email-log-fix.zip",
			destination: tmpDir,
		}, wantErr: false},
		{name: "unzipSource", args: args{
			source:      "tests/he_module-email-log-fix_wrong_format.zip",
			destination: tmpDir,
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := unzipSource(tt.args.source, tt.args.destination); (err != nil) != tt.wantErr {
				t.Errorf("unzipSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		t.Cleanup(func() {
			if err = os.RemoveAll(tmpDir); err != nil {
				t.Errorf("couldnt remove dir: %v", err)
			}
		})
	}
}

func TestSanitizeArchivePath(t *testing.T) {
	type args struct {
		d string
		t string
	}
	tests := []struct {
		name    string
		args    args
		wantV   string
		wantErr bool
	}{
		{name: "SanitizePath", args: args{
			d: "example/",
			t: "tests/he_module-email-log-fix.zip",
		}, wantV: "example/tests/he_module-email-log-fix.zip", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := SanitizeArchivePath(tt.args.d, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeArchivePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("SanitizeArchivePath() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func Test_parseZip(t *testing.T) {
	type args struct {
		zipFilePath string
	}
	tests := []struct {
		name    string
		args    args
		want    *ModuleInfo
		wantErr bool
	}{
		{name: "ParseZip", args: args{zipFilePath: "tests/he_module-email-log-fix.zip"}, want: &ModuleInfo{}, wantErr: false},
		{name: "ParseZip", args: args{zipFilePath: "tests/he_module-email-log-fix_wrong_format.zip"}, want: &ModuleInfo{}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseZip(tt.args.zipFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseZip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if ok := assert.IsType(t, &ModuleInfo{}, got); !ok {
				t.Errorf("getLatestReleasesFromGithubRepo() = %v, want %v", got, tt.want)
			}
			t.Cleanup(func() {
				if err = os.RemoveAll(got.TempUnzipPath); err != nil {
					t.Errorf("couldnt remove dir: %v", err)
				}
			})
		})
	}
}
