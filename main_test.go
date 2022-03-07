package main

import (
	"archive/zip"
	"io/ioutil"
	"os"
	"reflect"
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

func Test_filePathWalkDir(t *testing.T) {
	type args struct {
		root string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{name: "test walkdir", args: args{root: "tests"}, want: []string{
			"tests/he_module-email-log-fix.zip",
			"tests/he_module-email-log-fix_wrong_format.zip",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filePathWalkDir(tt.args.root)
			if (err != nil) != tt.wantErr {
				t.Errorf("filePathWalkDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filePathWalkDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validModuleComponents(t *testing.T) {
	type args struct {
		info *ModuleInfo
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "validModuleComponenets", args: args{info: &ModuleInfo{
			ModuleName:        "looney-tunes",
			VendorName:        "acme",
			ModuleVersion:     "1.0.0",
			ComposerJSONPath:  "composer.json",
			TempUnzipPath:     "",
			RegistrationPHP:   "registration.php",
			ModuleXML:         "etc/module.xml",
			OutputZipFileName: "",
			ComposerParentDir: "",
		}}, want: true},
		{name: "validModuleComponenets", args: args{info: &ModuleInfo{
			ModuleName:        "",
			VendorName:        "",
			ModuleVersion:     "",
			ComposerJSONPath:  "",
			TempUnzipPath:     "",
			RegistrationPHP:   "",
			ModuleXML:         "",
			OutputZipFileName: "",
			ComposerParentDir: "",
		}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validModuleComponents(tt.args.info); got != tt.want {
				t.Errorf("validModuleComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_zipWriter(t *testing.T) {
	type args struct {
		ComposerParentDir string
		outputZipFileName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test zipWriter", args: args{
			ComposerParentDir: "tests/",
			outputZipFileName: "tests.zip",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := zipWriter(tt.args.ComposerParentDir, tt.args.outputZipFileName); (err != nil) != tt.wantErr {
				t.Errorf("zipWriter() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Cleanup(func() {
				if err := os.Remove("tests.zip"); err != nil {
					t.Errorf("couldnt remove dir: %v", err)
				}
			})
		})
	}
}

func TestModuleInfo_printModuleInfo(t *testing.T) {
	type fields struct {
		ModuleName        string
		VendorName        string
		ModuleVersion     string
		ComposerJSONPath  string
		TempUnzipPath     string
		RegistrationPHP   string
		ModuleXML         string
		OutputZipFileName string
		ComposerParentDir string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test print",
			fields: fields{
				ModuleName:        "lonney-tunes",
				VendorName:        "acme",
				ModuleVersion:     "1.2.3",
				ComposerJSONPath:  "",
				TempUnzipPath:     "",
				RegistrationPHP:   "",
				ModuleXML:         "",
				OutputZipFileName: "test.zip",
				ComposerParentDir: "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ModuleInfo{
				ModuleName:        tt.fields.ModuleName,
				VendorName:        tt.fields.VendorName,
				ModuleVersion:     tt.fields.ModuleVersion,
				ComposerJSONPath:  tt.fields.ComposerJSONPath,
				TempUnzipPath:     tt.fields.TempUnzipPath,
				RegistrationPHP:   tt.fields.RegistrationPHP,
				ModuleXML:         tt.fields.ModuleXML,
				OutputZipFileName: tt.fields.OutputZipFileName,
				ComposerParentDir: tt.fields.ComposerParentDir,
			}
			if err := info.printModuleInfo(); (err != nil) != tt.wantErr {
				t.Errorf("ModuleInfo.printModuleInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_unzipFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "temp_extension")
	if err != nil {
		t.Errorf("%s", err)
	}

	reader, err := zip.OpenReader("tests/he_module-email-log-fix_wrong_format.zip")
	if err != nil {
		t.Errorf("%s", err)
	}
	zipFile := &zip.File{}
	for _, i := range reader.File {
		zipFile = i
	}
	defer reader.Close()
	type args struct {
		f           *zip.File
		destination string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test unzipFile", args: args{
			f:           zipFile,
			destination: tmpDir,
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err = unzipFile(tt.args.f, tt.args.destination); (err != nil) != tt.wantErr {
				t.Errorf("unzipFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		t.Cleanup(func() {
			if err = os.RemoveAll(tmpDir); err != nil {
				t.Errorf("couldnt remove dir: %v", err)
			}
		})
	}
}

func Test_createArtifact(t *testing.T) {
	type args struct {
		info *ModuleInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createArtifact(tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("createArtifact() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_addFiles(t *testing.T) {
	type args struct {
		w         *zip.Writer
		basePath  string
		baseInZip string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := addFiles(tt.args.w, tt.args.basePath, tt.args.baseInZip); (err != nil) != tt.wantErr {
				t.Errorf("addFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
