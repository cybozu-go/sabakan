package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

const (
	defaultSabakanURL = "http://localhost:10080"
	defaultTPMDev     = "/dev/tpm0"
	defaultCipher     = "aes-xts-plain64"
	defaultKeySize    = 512
	defaultCACert     = "/etc/sabakan/sabakan-tls-ca.crt"
)

var opts struct {
	sabakanURL string
	cipher     string
	tpmdev     string
	keySize    int
	excludes   []string
	caCert     string
}

var rootCmd = &cobra.Command{
	Use:   "sabakan-cryptsetup",
	Short: "Automatic disk encrypt utility",
	Long: `A utility to help automatic full disk encryption.

It generates disk encryption key and setup encrypted disks by
using cryptsetup, a front-end tool of dm-crypt kernel module.
The generated encryption key is encrypted with another key and
sent to sabakan server.  At the next boot, sabakan-cryptsetup
will download the encrypted disk encryption key from sabakan,
decrypt it, and setup encrypted disks automatically.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		if opts.keySize%8 != 0 {
			return errors.New("key size must be multiple of 8")
		}

		InitModules()
		disks, err := FindDisks(opts.excludes)
		if err != nil {
			return err
		}
		driver, err := NewDriver(opts.sabakanURL, opts.cipher, opts.keySize/8, opts.tpmdev, disks)
		if err != nil {
			return err
		}
		well.Go(driver.Setup)
		well.Stop()
		return well.Wait()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	sabaURL := os.Getenv("SABAKAN_URL")
	if sabaURL == "" {
		sabaURL = defaultSabakanURL
	}
	rootCmd.Flags().StringVar(&opts.sabakanURL, "server", sabaURL, "URL of sabakan server")
	rootCmd.Flags().StringVar(&opts.tpmdev, "tpmdev", defaultTPMDev, "device file path of tpm")
	rootCmd.Flags().StringVar(&opts.cipher, "cipher", defaultCipher, "cipher specification")
	rootCmd.Flags().IntVar(&opts.keySize, "keysize", defaultKeySize, "key size in bits")
	rootCmd.Flags().StringArrayVar(&opts.excludes, "excludes", nil, `disk name patterns to be excluded, e.g. "nvme*"`)
	rootCmd.Flags().StringVar(&opts.caCert, "cert", defaultCACert, "location of sabakan CA certificate")
}
