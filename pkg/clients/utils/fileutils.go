package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func IsArchive(url string) bool {
	return strings.HasSuffix(url, ".zip") ||
		strings.HasSuffix(url, ".tar.gz") ||
		strings.HasSuffix(url, ".tgz")
}

func ExtractExecutable(archivePath, targetPath, toolName string) error {
	switch {
	case strings.HasSuffix(archivePath, ".zip"):
		return ExtractFromZip(archivePath, targetPath, toolName)
	case strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz"):
		return ExtractFromTarGz(archivePath, targetPath, toolName)
	default:
		return fmt.Errorf("unsupported archive format")
	}
}

func ExtractFromZip(archivePath, targetPath, toolName string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if strings.Contains(filepath.Base(file.Name), toolName) || file.Name == toolName {
			fileReader, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open file in archive: %w", err)
			}
			defer fileReader.Close()

			targetFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("failed to create target file: %w", err)
			}
			defer targetFile.Close()

			_, err = io.Copy(targetFile, fileReader)
			if err != nil {
				return fmt.Errorf("failed to copy file content: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("executable not found in zip archive")
}

func ExtractFromTarGz(archivePath, targetPath, toolName string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		if header.Typeflag == tar.TypeReg &&
			(strings.Contains(filepath.Base(header.Name), toolName) || header.Name == toolName) {
			targetFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("failed to create target file: %w", err)
			}
			defer targetFile.Close()

			_, err = io.Copy(targetFile, tr)
			if err != nil {
				return fmt.Errorf("failed to copy file content: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("executable not found in tar.gz archive")
}

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
