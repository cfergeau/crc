package machine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/compress"
	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/pkg/errors"
)

func (client *client) GenerateBundle() error {
	libMachineAPIClient, cleanup := createLibMachineClient()
	defer cleanup()

	host, err := libMachineAPIClient.Load(client.name)
	if err != nil {
		return errors.Wrap(err, "Cannot load machine")
	}

	_, err = host.Driver.GetState()
	if err != nil {
		return errors.Wrap(err, "Cannot get machine state")
	}

	// Get crcBundleMetadata to update the ssh key and
	crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		return errors.Wrap(err, "Error loading bundle metadata")
	}

	// Remove the pull secret
	instanceIP, err := getIP(host, client.useVSock())
	if err != nil {
		return errors.Wrap(err, "Error getting the IP")
	}
	sshRunner, err := crcssh.CreateRunner(instanceIP, getSSHPort(client.useVSock()), crcBundleMetadata.GetSSHKeyPath(), constants.GetPrivateKeyPath(), constants.GetRsaPrivateKeyPath())
	if err != nil {
		return errors.Wrap(err, "Error creating the ssh client")
	}
	defer sshRunner.Close()
	if err := cluster.RemovePullSecretFromTheInstanceDisk(sshRunner); err != nil {
		return err
	}
	ocConfig := oc.UseOCWithSSH(sshRunner)
	if err := cluster.RemovePullSecretFromCluster(ocConfig); err != nil {
		return err
	}

	// Stop the cluster
	currentState, err := client.Stop()
	if err != nil {
		return err
	}
	if currentState != state.Stopped {
		return fmt.Errorf("VM is not stopped, current state is %s", currentState.String())
	}

	// Create a tmp directory with name of <bundle>_custom
	tmpDir, err := ioutil.TempDir("", "crc_custom_bundle")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Create the custom bundle directory which is used as top level directory for tarball during compression
	currentBundleName := strings.TrimSuffix(crcBundleMetadata.GetBundleName(), filepath.Ext(crcBundleMetadata.GetBundleName()))
	customBundleName := fmt.Sprintf("%s_custom", currentBundleName)
	customBundleDir := filepath.Join(tmpDir, customBundleName)
	if err := os.Mkdir(customBundleDir, 0775); err != nil {
		return err
	}

	// Copy the kubeconfig and kubeadmin-password from bundle to tmpDir
	kubeConfigFilePath := filepath.Join(customBundleDir, "kubeconfig")
	err = crcos.CopyFileContents(crcBundleMetadata.GetKubeConfigPath(),
		kubeConfigFilePath,
		0640)
	if err != nil {
		return fmt.Errorf("error copying kubeconfig file to %s: %v", kubeConfigFilePath, err)
	}

	kubeConfigAdminPasswordPath := filepath.Join(customBundleDir, "kubeadmin-password")
	kubePassword, err := crcBundleMetadata.GetKubeadminPassword()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(kubeConfigAdminPasswordPath, []byte(kubePassword), 0600)
	if err != nil {
		return fmt.Errorf("error writing to %s: %v", kubeConfigFilePath, err)
	}

	// Copy ssh keyfile
	sshKeyPath := filepath.Join(customBundleDir, "id_ecdsa_crc")
	generatedKey := filepath.Join(constants.MachineInstanceDir, constants.DefaultName, "id_ecdsa")
	if err := crcos.CopyFileContents(generatedKey, sshKeyPath, 0400); err != nil {
		return err
	}

	// Copy the oc binary
	ocFilePath := filepath.Join(constants.CrcOcBinDir, constants.OcExecutableName)
	if err := crcos.CopyFileContents(ocFilePath, filepath.Join(customBundleDir, constants.OcExecutableName), 0755); err != nil {
		return err
	}

	// Copy disk image
	logging.Infof("Copy the disk image to %s", customBundleDir)
	if err := copyDiskImage(customBundleDir, crcBundleMetadata); err != nil {
		return err
	}

	// Get size of newly created disk
	diskPath := filepath.Join(customBundleDir, filepath.Base(crcBundleMetadata.GetDiskImagePath()))
	stat, err := os.Stat(diskPath)
	if err != nil {
		return err
	}

	// update bundle info
	crcBundleMetadata.Type = "custom"
	crcBundleMetadata.BuildInfo.BuildTime = time.Now().String()
	crcBundleMetadata.Storage.DiskImages[0].Size = strconv.FormatInt(stat.Size(), 10)

	// Create the metadata json for custom bundle
	bundleContent, err := json.MarshalIndent(crcBundleMetadata, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(customBundleDir, "crc-bundle-info.json"), bundleContent, 0600)
	if err != nil {
		return fmt.Errorf("error copying bundle metadata  %v", err)
	}

	logging.Infof("Compressing %s", customBundleDir)
	if err := compress.Compress(customBundleDir, fmt.Sprintf("%s_custom.crcbundle", currentBundleName)); err != nil {
		return err
	}

	return nil
}
