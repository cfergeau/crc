package machine

import (
	"fmt"
	"time"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
)

/* Returned when a certificate is past its expiry date */
type certsExpiredError struct {
	expiryTime time.Time
}

const (
	certsExpiryUrl = "For more details, read https://code-ready.github.io/crc/#troubleshooting-expired-certificates_gsg"
)

func (err certsExpiredError) Error() string {
	return fmt.Sprintf("Instance certificates have expired, they were valid until %s\n%s", err.expiryTime.Format(time.RFC822), certsExpiryUrl)
}

/* Returned when a certificate will expire in less than 7 days */
type certsExpiringError struct {
	daysLeft int
}

func (err certsExpiringError) Error() string {
	return fmt.Sprintf("Instance certificates will expire in %d days, you should consider upgrading to a newer release\n%s", err.daysLeft, certsExpiryUrl)
}

func daysTillExpiry(expiryTime time.Time) int {
	durationTillExpiry := expiryTime.Sub(time.Now())
	return int(durationTillExpiry.Hours() / 24)
}

func checkCertsExpiry(expiryTime time.Time) error {
	daysLeft := daysTillExpiry(expiryTime)
	if daysLeft < 0 {
		return certsExpiredError{expiryTime}
	} else if daysLeft < 7 {
		return certsExpiringError{daysLeft}
	}

	return nil
}

func checkVMCertsExpiry(sshRunner *crcssh.SSHRunner) error {
	expiryTime, err := cluster.GetCertExpiryDateFromVM(sshRunner)
	if err != nil {
		return err
	}

	return checkCertsExpiry(expiryTime)
}

func checkBundleCertsExpiry(bundleMetadata *bundle.CrcBundleInfo) error {
	buildTime, err := bundleMetadata.GetBundleBuildTime()
	if err != nil {
		return err
	}

	return checkCertsExpiry(buildTime.AddDate(0, 1, 0))
}
