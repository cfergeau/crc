package compress

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/logging"

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

	return filepath.Walk(src, func(file string, fi os.FileInfo, err1 error) error {
		logging.Debugf("adding %s", file)
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}
		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		header.Name = filepath.ToSlash(file)
		if filepath.IsAbs(header.Name) {
			if len(header.Name) <= 1 {
				// similar to what the 'tar' command does
				header.Name = "./"
			}
			// the 'tar' command strips the leading '/' from absolute paths
			header.Name = header.Name[1:]
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			defer data.Close()
			if _, err := io.Copy(tarWriter, data); err != nil {
				return err
			}
		}
		return nil
	})
}
