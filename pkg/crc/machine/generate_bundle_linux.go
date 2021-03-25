package machine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	crcos "github.com/code-ready/crc/pkg/os"
)

func copyDiskImage(destDir string, crcBundleMetadata *bundle.CrcBundleInfo) error {
	tmpDir := filepath.Join(constants.MachineCacheDir, "image-rebase")
	_ = os.RemoveAll(tmpDir) // clean up before using it
	if err := os.Mkdir(tmpDir, 0700); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpDir) // clean up after using it
	}()

	// Copy the disk image from the machine/crc folder
	diskName := filepath.Base(crcBundleMetadata.GetDiskImagePath())
	diskPath := filepath.Join(constants.MachineInstanceDir, constants.DefaultName, diskName)
	if err := crcos.CopyFileContents(diskPath, filepath.Join(tmpDir, fmt.Sprintf("%s_back", diskName)), 0644); err != nil {
		return err
	}

	// Copy the original image from cache folder
	if err := crcos.CopyFileContents(crcBundleMetadata.GetDiskImagePath(),
		filepath.Join(tmpDir, diskName), 0644); err != nil {
		return err
	}

	// Use qemu-img commands to rebase and commit it.
	if _, _, err := crcos.RunWithDefaultLocale("qemu-img", "rebase", "-F", "qcow2", "-b",
		filepath.Join(tmpDir, diskName), filepath.Join(tmpDir, fmt.Sprintf("%s_back", diskName))); err != nil {
		return err
	}
	if _, _, err := crcos.RunWithDefaultLocale("qemu-img", "commit", filepath.Join(tmpDir, fmt.Sprintf("%s_back", diskName))); err != nil {
		return err
	}

	// Copy the final image to destination dir
	if err := crcos.CopyFileContents(filepath.Join(tmpDir, diskName), filepath.Join(destDir, diskName), 0644); err != nil {
		return err
	}
	return nil
}
