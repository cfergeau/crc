package compress

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

func Compress(src, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	enc, err := zstd.NewWriter(out)
	if err != nil {
		return err
	}
	defer enc.Close()

	tarWriter := tar.NewWriter(enc)
	defer tarWriter.Close()

	// Just use top level directory as part of tarball
	// $ zstdcat crc_libvirt_4.7.1_custom.zstd  | tar t
	// crc_libvirt_4.7.1_custom
	// crc_libvirt_4.7.1_custom/test
	parent, child := filepath.Split(src)
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(parent); err != nil {
		return err
	}
	defer func() {
		_ = os.Chdir(currentDir)
	}()

	return filepath.Walk(child, func(file string, fi os.FileInfo, err1 error) error {
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}
		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		header.Name = filepath.ToSlash(file)

		// write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tarWriter, data); err != nil {
				return err
			}
		}
		return nil
	})
}
