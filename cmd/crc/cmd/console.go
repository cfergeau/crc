package cmd

import (
	gocontext "context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/crc-org/crc/pkg/crc/constants"
	"github.com/crc-org/crc/pkg/crc/machine"
	"github.com/crc-org/crc/pkg/crc/machine/types"
	"github.com/openshift/oc/pkg/helpers/tokencmd"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/third_party/forked/golang/netutil"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	consolePrintURL         bool
	consolePrintCredentials bool
)

func init() {
	addOutputFormatFlag(consoleCmd)
	consoleCmd.Flags().BoolVar(&consolePrintURL, "url", false, "Print the URL for the OpenShift Web Console")
	consoleCmd.Flags().BoolVar(&consolePrintCredentials, "credentials", false, "Print the credentials for the OpenShift Web Console")
	rootCmd.AddCommand(consoleCmd)
}

// consoleCmd represents the console command
var consoleCmd = &cobra.Command{
	Use:     "console",
	Aliases: []string{"dashboard"},
	Short:   "Open the OpenShift Web Console in the default browser",
	Long:    `Open the OpenShift Web Console in the default browser or print its URL or credentials`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConsole(os.Stdout, newMachine(), consolePrintURL, consolePrintCredentials, outputFormat)
	},
}

func showConsole(client machine.Client) (*types.ConsoleResult, error) {
	if err := checkIfMachineMissing(client); err != nil {
		// In case of machine doesn't exist then consoleResult error
		// should be updated so that when rendering the result it have
		// error details also.
		return nil, err
	}
	return client.GetConsoleURL()
}

func runConsole(writer io.Writer, client machine.Client, consolePrintURL, consolePrintCredentials bool, outputFormat string) error {
	consoleResult, err := showConsole(client)
	if err != nil {
		return err
	}

	return writeKubeconfig("127.0.0.1", &consoleResult.ClusterConfig)
}

func addContext(cfg *api.Config, ip string, clusterConfig *types.ClusterConfig, ca []byte, context, username, password string) error {
	host, err := hostname(clusterConfig.ClusterAPI)
	if err != nil {
		return err
	}
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(ca)
	if !ok {
		return fmt.Errorf("failed to parse root certificate")
	}
	token, err := tokencmd.RequestToken(&restclient.Config{
		Proxy: clusterConfig.ProxyConfig.ProxyFunc(),
		Host:  clusterConfig.ClusterAPI,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    roots,
				MinVersion: tls.VersionTLS12,
			},
			DialContext: func(ctx gocontext.Context, network, address string) (net.Conn, error) {
				port := strings.SplitN(address, ":", 2)[1]
				dialer := net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}
				return dialer.Dial(network, fmt.Sprintf("%s:%s", ip, port))
			},
		},
	}, nil, username, password)
	if err != nil {
		return err
	}
	cfg.AuthInfos[username] = &api.AuthInfo{
		Token: token,
	}
	cfg.Contexts[context] = &api.Context{
		Cluster:   host,
		AuthInfo:  username,
		Namespace: "default",
	}
	return nil
}

const (
	adminContext     = "crc-admin"
	developerContext = "crc-developer"
)

func writeKubeconfig(ip string, clusterConfig *types.ClusterConfig) error {
	kubeconfig := getGlobalKubeConfigPath()
	dir := filepath.Dir(kubeconfig)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// Make sure .kube/config exist if not then this will create
	_, _ = os.OpenFile(kubeconfig, os.O_RDONLY|os.O_CREATE, 0600)

	ca, err := certificateAuthority(clusterConfig.KubeConfig)
	if err != nil {
		return err
	}
	host, err := hostname(clusterConfig.ClusterAPI)
	if err != nil {
		return err
	}

	cfg, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return err
	}
	cfg.Clusters[host] = &api.Cluster{
		Server:                   clusterConfig.ClusterAPI,
		CertificateAuthorityData: ca,
	}

	if err := addContext(cfg, ip, clusterConfig, ca, adminContext, "kubeadmin", clusterConfig.KubeAdminPass); err != nil {
		return err
	}
	if err := addContext(cfg, ip, clusterConfig, ca, developerContext, "developer", "developer"); err != nil {
		return err
	}
	return nil

	/*
		if cfg.CurrentContext == "" {
			cfg.CurrentContext = adminContext
		}

		return clientcmd.WriteToFile(*cfg, kubeconfig)
	*/
}

func certificateAuthority(kubeconfigFile string) ([]byte, error) {
	builtin, err := clientcmd.LoadFromFile(kubeconfigFile)
	if err != nil {
		return nil, err
	}
	cluster, ok := builtin.Clusters["crc"]
	if !ok {
		return nil, fmt.Errorf("crc cluster not found in kubeconfig %s", kubeconfigFile)
	}
	return cluster.CertificateAuthorityData, nil
}

// https://github.com/openshift/oc/blob/f94afb52dc8a3185b3b9eacaf92ec34d80f8708d/pkg/helpers/kubeconfig/smart_merge.go#L21
func hostname(clusterAPI string) (string, error) {
	p, err := url.Parse(clusterAPI)
	if err != nil {
		return "", err
	}
	h := netutil.CanonicalAddr(p)
	return strings.ReplaceAll(h, ".", "-"), nil
}

// getGlobalKubeConfigPath returns the path to the first entry in the KUBECONFIG environment variable
// or if KUBECONFIG is not set then $HOME/.kube/config
func getGlobalKubeConfigPath() string {
	pathList := filepath.SplitList(os.Getenv("KUBECONFIG"))
	if len(pathList) > 0 {
		// Tools should write to the last entry in the KUBECONFIG file instead of the first one.
		// oc cluster up also does the same.
		return pathList[len(pathList)-1]
	}
	return filepath.Join(constants.GetHomeDir(), ".kube", "config")
}
