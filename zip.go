package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UnzipBytes(zipData []byte, dest string) error {
	bytesReader := bytes.NewReader(zipData)

	r, err := zip.NewReader(bytesReader, int64(len(zipData)))
	if err != nil {
		return err
	}

	destAbs, err := filepath.Abs(filepath.Clean(dest))
	if err != nil {
		return err
	}

	for _, f := range r.File {
		fpath := filepath.Join(destAbs, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(destAbs)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path / directory traversal attempt: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		srcFile, err := f.Open()
		if err != nil {
			dstFile.Close()
			return err
		}

		_, err = io.Copy(dstFile, srcFile)

		srcFile.Close()
		dstFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func HasBepInExFolder(zipBytes []byte) (bool, error) {
	reader := bytes.NewReader(zipBytes)

	zipReader, err := zip.NewReader(reader, int64(len(zipBytes)))
	if err != nil {
		return false, err
	}

	for _, file := range zipReader.File {
		if file.Name == "BepInEx" || strings.HasPrefix(file.Name, "BepInEx/") {
			return true, nil
		}
	}

	return false, nil
}

func GenerateManifest(zipBytes []byte, extractDir string) (*ModManifest, error) {
	bytesReader := bytes.NewReader(zipBytes)
	r, err := zip.NewReader(bytesReader, int64(bytesReader.Len()))
	if err != nil {
		return nil, err
	}

	manifest := &ModManifest{
		Folders: []string{},
		Files:   []string{},
		Version: "",
		GUID:    "",
	}

	for _, f := range r.File {
		// no malware allowed
		cleanedPath := filepath.Clean(f.Name)
		if strings.HasPrefix(cleanedPath, "..") || strings.HasPrefix(cleanedPath, "/") {
			continue
		}

		if f.FileInfo().IsDir() {
			if cleanedPath == "BepInEx" || cleanedPath == "BepInEx/plugins" {
				continue
			}

			manifest.Folders = append(manifest.Folders, filepath.Join(extractDir, cleanedPath))
		} else {
			manifest.Files = append(manifest.Files, filepath.Join(extractDir, cleanedPath))
		}
	}

	return manifest, nil
}
