// oc adm release info quay.io/openshift-release-dev/ocp-release:4.2.10 --size  -a ~/redhat/crc/crc-pull-secret  -o json
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/openshift/oc/pkg/cli/admin/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func printReleaseInfo(releaseInfo *release.ReleaseInfo) {
	fmt.Printf("Image: %s\n", releaseInfo.Image)
	fmt.Printf("Digest: %s\n", releaseInfo.Digest)
	fmt.Printf("ContentDigest: %s\n", releaseInfo.ContentDigest)

	fmt.Printf("%v\n", releaseInfo.Images)
	for k, v := range releaseInfo.Images {
		fmt.Printf("\t%s -> %v\n", k, v)
	}
}

func main() {
	var releaseInfo release.ReleaseInfo
	data, err := ioutil.ReadFile("images.json")
	// can file be opened?
	if err != nil {
		fmt.Print(err)
		return
	}
	err = json.Unmarshal(data, &releaseInfo)
	if err != nil {
		fmt.Print(err)
		return
	}
	printReleaseInfo(&releaseInfo)

	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	infoOptions := release.NewInfoOptions(ioStreams)

	infoOptions.Images = append(infoOptions.Images, "quay.io/openshift-release-dev/ocp-release:4.2.10")
	infoOptions.Output = "json"
	infoOptions.ShowSize = true
	infoOptions.SecurityOptions.RegistryConfig = "/home/teuf/redhat/crc/crc-pull-secret"
	releaseInfoPtr, err := infoOptions.LoadReleaseInfo("quay.io/openshift-release-dev/ocp-release:4.2.10", true)
	if err != nil {
		fmt.Print(err)
		return
	}
	printReleaseInfo(releaseInfoPtr)

}
