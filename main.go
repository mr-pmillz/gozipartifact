package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	figure "github.com/common-nighthawk/go-figure"
	"github.com/mkideal/cli"
	"github.com/olekukonko/tablewriter"
)

type argT struct {
	cli.Helper
	ZipFilePath string `cli:"z,zipfile" usage:"zipfile to create artifact from"`
}

type ModuleInfo struct {
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

type ComposerJSON struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Require     Require `json:"require"`
	Suggest     Suggest `json:"suggest"`
	Type        string  `json:"type"`
	Version     string  `json:"version"`
}
type Require struct {
}
type Suggest struct {
}

func unzipSource(source, destination string) error {
	// Open the zip file
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Get the absolute destination path
	destination, err = filepath.Abs(destination)
	if err != nil {
		return err
	}

	// Iterate over zip files inside the archive and unzip each of them
	for _, f := range reader.File {
		err = unzipFile(f, destination)
		if err != nil {
			return err
		}
	}

	return nil
}

// SanitizeArchivePath Sanitizes an archive file pathing from "G305: Zip Slip vulnerability"
func SanitizeArchivePath(d, t string) (v string, err error) {
	v = filepath.Join(d, t)
	if strings.HasPrefix(v, filepath.Clean(d)) {
		return v, nil
	}

	return "", fmt.Errorf("%s: %s", "content filepath is tainted", t)
}

func unzipFile(f *zip.File, destination string) error {
	// Check if file paths are not vulnerable to Zip Slip
	filePath, err := SanitizeArchivePath(destination, f.Name)
	if err != nil {
		return err
	}

	// Create directory tree
	if f.FileInfo().IsDir() {
		if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// Create a destination file for unzipped content
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Unzip the content of a file and copy it to the destination file
	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()

	for {
		_, err = io.CopyN(destinationFile, zippedFile, 1024)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}

// parseZip Unzips archive to tmp dir, parses composer.json and returns relevant module info
// returns *ModuleInfo and an error
func parseZip(zipFilePath string) (*ModuleInfo, error) {
	m := &ModuleInfo{}

	// create tmp dir to unzip to
	tmpDir, err := ioutil.TempDir("", "temp_extension")
	if err != nil {
		return nil, err
	}
	m.TempUnzipPath = tmpDir

	// unzip to tmp dir
	err = unzipSource(zipFilePath, tmpDir)
	if err != nil {
		return nil, err
	}

	// get all file paths in array list
	allSourceFilesList, err := filePathWalkDir(tmpDir)
	if err != nil {
		return nil, err
	}

	for _, fPath := range allSourceFilesList {
		parentDir, baseFileName := filepath.Split(fPath)
		switch baseFileName {
		case "composer.json":
			m.ComposerParentDir = parentDir

			// read composer.json
			composerFile, err := ioutil.ReadFile(fPath)
			if err != nil {
				return nil, err
			}

			data := ComposerJSON{}
			err = json.Unmarshal(composerFile, &data)
			if err != nil {
				return nil, err
			}

			m.ComposerJSONPath = fPath
			m.ModuleVersion = data.Version
			vendorParts := strings.Split(data.Name, "/")
			m.VendorName = vendorParts[0]
			m.ModuleName = vendorParts[1]
		case "module.xml":
			m.ModuleXML = fPath
		case "registration.php":
			m.RegistrationPHP = fPath
		default:
			//None
		}
	}

	m.OutputZipFileName = strings.ToLower(fmt.Sprintf("%s-%s-%s.zip", m.VendorName, m.ModuleName, m.ModuleVersion))
	return m, nil
}

// filePathWalkDir returns a list of all the relative file paths in a given dir
func filePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// validModuleComponents checks the bare minimum valid magento 2 module requirements
func validModuleComponents(info *ModuleInfo) bool {
	if len(info.ModuleXML) >= 1 && len(info.ComposerJSONPath) >= 1 && len(info.RegistrationPHP) >= 1 {
		return true
	}
	return false
}

// createArtifact is a helper func around zipWriter
func createArtifact(info *ModuleInfo) error {
	return zipWriter(info.ComposerParentDir, info.OutputZipFileName)
}

// zipWriter creates a valid m2 artifact zip archive
func zipWriter(ComposerParentDir, outputZipFileName string) error {
	// Get a Buffer to Write To
	outFile, err := os.Create(outputZipFileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(outFile)

	// Add some files to the archive.
	if err = addFiles(w, ComposerParentDir, ""); err != nil {
		return err
	}

	if err != nil {
		return err
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}

// addFiles adds files to a zip archive
func addFiles(w *zip.Writer, basePath, baseInZip string) error {
	// Open the Directory
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			dat, err := ioutil.ReadFile(basePath + file.Name())
			if err != nil {
				return err
			}

			// Add some files to the archive.
			f, err := w.Create(baseInZip + file.Name())
			if err != nil {
				return err
			}

			_, err = f.Write(dat)
			if err != nil {
				return err
			}

		} else if file.IsDir() {
			// Recurse
			newBase := basePath + file.Name() + "/"
			if err = addFiles(w, newBase, baseInZip+file.Name()+"/"); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	os.Exit(cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		if len(argv.ZipFilePath) >= 1 {
			absZipPath, err := filepath.Abs(argv.ZipFilePath)
			if err != nil {
				return err
			}
			fmt.Println()
			figure.NewColorFigure("M2 Artifact", "colossal", "green", true).Print()
			fmt.Println()
			fmt.Printf("[+] Creating Artifact from %s\n", absZipPath)
			fmt.Println()

			info, err := parseZip(absZipPath)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Vendor", "Module", "Version", "FilePath"})
			table.SetBorder(false)
			table.SetHeaderColor(
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiMagentaColor, tablewriter.BgBlackColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiMagentaColor, tablewriter.BgBlackColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiMagentaColor, tablewriter.BgBlackColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiMagentaColor, tablewriter.BgBlackColor},
			)

			table.SetColumnColor(
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor, tablewriter.BgBlackColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor, tablewriter.BgBlackColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor, tablewriter.BgBlackColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor, tablewriter.BgBlackColor},
			)

			absOutputFilePath, err := filepath.Abs(info.OutputZipFileName)
			if err != nil {
				return err
			}
			colorData := []string{info.VendorName, info.ModuleName, info.ModuleVersion, absOutputFilePath}
			table.Rich(colorData, []tablewriter.Colors{
				{tablewriter.Normal, tablewriter.FgHiGreenColor},
				{tablewriter.Normal, tablewriter.FgHiYellowColor},
				{tablewriter.Normal, tablewriter.FgHiCyanColor},
				{tablewriter.Normal, tablewriter.FgHiRedColor},
			})

			if validModuleComponents(info) {
				if err = createArtifact(info); err != nil {
					return err
				}
				table.SetAutoMergeCells(true)
				table.Render()
			} else {
				fmt.Printf("[!] %s isn't a valid magento 2 module\n", argv.ZipFilePath)
			}

			if err = os.RemoveAll(info.TempUnzipPath); err != nil {
				return err
			}
		}

		return nil
	}))
}
