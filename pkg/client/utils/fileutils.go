package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func IsArchive(url string) bool {
	return strings.HasSuffix(url, ".zip") ||
		strings.HasSuffix(url, ".tar.gz") ||
		strings.HasSuffix(url, ".tgz")
}

func ExtractExecutable(archivePath, targetPath, toolName, archiveName string) error {
	switch {
	case strings.HasSuffix(archiveName, ".zip"):
		return ExtractFromZip(archivePath, targetPath, toolName)
	case strings.HasSuffix(archiveName, ".tar.gz") || strings.HasSuffix(archiveName, ".tgz"):
		return ExtractFromTarGz(archivePath, targetPath, toolName)
	default:
		return fmt.Errorf("unsupported archive format for %s", archiveName)
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

			targetFile, err := CreateFileWithElevatedPermissions(targetPath)
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
			targetFile, err := CreateFileWithElevatedPermissions(targetPath)
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

	destFile, err := CreateFileWithElevatedPermissions(dst)
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

func CreateFileWithElevatedPermissions(path string) (*os.File, error) {
	file, err := os.Create(path)
	if err == nil {
		return file, nil
	}

	if os.IsPermission(err) {
		fmt.Printf("Permission denied. Attempting to create %s with sudo...\n", path)

		dirCmd := exec.Command("sudo", "mkdir", "-p", filepath.Dir(path))
		if err := dirCmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to create directory with sudo: %w", err)
		}

		touchCmd := exec.Command("sudo", "touch", path)
		if err := touchCmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to create file with sudo: %w", err)
		}

		chownCmd := exec.Command("sudo", "chown", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), path)
		if err := chownCmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to change file ownership with sudo: %w", err)
		}

		return os.OpenFile(path, os.O_RDWR, 0644)
	}

	return nil, err
}
