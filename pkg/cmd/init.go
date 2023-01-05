package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/spf13/cobra"
)

type initCmdFlagValues struct {
	account            string
	applianceURL       string
	authnType          string
	serviceID          string
	conjurrcFilePath   string
	certFilePath       string
	forceFileOverwrite bool
	selfSigned         bool
}

func getInitCmdFlagValues(cmd *cobra.Command) (initCmdFlagValues, error) {
	account, err := cmd.Flags().GetString("account")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	applianceURL, err := cmd.Flags().GetString("url")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	authnType, err := cmd.Flags().GetString("authn-type")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	serviceID, err := cmd.Flags().GetString("service-id")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	conjurrcFilePath, err := cmd.Flags().GetString("file")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	certFilePath, err := cmd.Flags().GetString("cert-file")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	selfSigned, err := cmd.Flags().GetBool("self-signed")
	if err != nil {
		return initCmdFlagValues{}, err
	}
	forceFileOverwrite, err := cmd.Flags().GetBool("force")
	if err != nil {
		return initCmdFlagValues{}, err
	}

	return initCmdFlagValues{
		account:            account,
		applianceURL:       applianceURL,
		authnType:          authnType,
		serviceID:          serviceID,
		conjurrcFilePath:   conjurrcFilePath,
		certFilePath:       certFilePath,
		selfSigned:         selfSigned,
		forceFileOverwrite: forceFileOverwrite,
	}, nil
}

func runInitCommand(cmd *cobra.Command, args []string) error {
	var err error

	cmdFlagVals, err := getInitCmdFlagValues(cmd)
	if err != nil {
		return err
	}

	setCommandStreamsOnPrompt := prompts.PromptDecoratorForCommand(cmd)

	account, applianceURL, err := prompts.MaybeAskForConnectionDetails(
		setCommandStreamsOnPrompt,
		cmdFlagVals.account,
		cmdFlagVals.applianceURL,
	)
	if err != nil {
		return err
	}

	config := conjurapi.Config{
		Account:      account,
		ApplianceURL: applianceURL,
		AuthnType:    cmdFlagVals.authnType,
		ServiceID:    cmdFlagVals.serviceID,
	}

	err = config.Validate()
	if err != nil {
		return err
	}

	err = fetchCertIfNeeded(&config, cmdFlagVals, setCommandStreamsOnPrompt)
	if err != nil {
		return err
	}
	if config.SSLCertPath != "" {
		cmd.Printf("Wrote certificate to %s", config.SSLCertPath)
	}

	err = writeConjurrc(
		config,
		cmdFlagVals,
		setCommandStreamsOnPrompt,
	)
	if err != nil {
		return err
	}

	cmd.Printf("Wrote configuration to %s\n", cmdFlagVals.conjurrcFilePath)
	return nil
}

func fetchCertIfNeeded(config *conjurapi.Config, cmdFlagVals initCmdFlagValues, setCommandStreamsOnPrompt prompts.DecoratePromptFunc) error {
	// Get TLS certificate from Conjur server
	url, err := url.Parse(config.ApplianceURL)
	if err != nil {
		return err
	}

	// Only fetch certificate if using HTTPS
	if url.Scheme != "https" {
		return nil
	}

	cert, err := utils.GetServerCert(url.Host, cmdFlagVals.selfSigned)
	if err != nil {
		return fmt.Errorf("Unable to retrieve certificate from %s: %s", url.Host, err)
	}

	// Prompt user to accept certificate
	err = prompts.AskToTrustCert(setCommandStreamsOnPrompt, cert.Fingerprint)
	if err != nil {
		return fmt.Errorf("You decided not to trust the certificate")
	}

	certPath := cmdFlagVals.certFilePath

	err = writeFile(certPath, []byte(cert.Cert), cmdFlagVals.forceFileOverwrite, setCommandStreamsOnPrompt)
	if err != nil {
		return err
	}

	config.SSLCert = cert.Cert
	config.SSLCertPath = certPath

	return nil
}

func writeConjurrc(config conjurapi.Config, cmdFlagVals initCmdFlagValues, setCommandStreamsOnPrompt prompts.DecoratePromptFunc) error {
	filePath := cmdFlagVals.conjurrcFilePath
	fileContents := config.Conjurrc()

	return writeFile(filePath, fileContents, cmdFlagVals.forceFileOverwrite, setCommandStreamsOnPrompt)
}

func writeFile(filePath string, fileContents []byte, forceFileOverwrite bool, setCommandStreamsOnPrompt prompts.DecoratePromptFunc) error {
	if !forceFileOverwrite {
		err := prompts.MaybeAskToOverwriteFile(setCommandStreamsOnPrompt, filePath)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(filePath, fileContents, 0644)
}

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Use the init command to initialize the Conjur CLI with a Conjur endpoint.",
		Long: `Use the init command to initialize the Conjur CLI with a Conjur endpoint.

The init command creates a configuration file (.conjurrc) that contains the details for connecting to Conjur. This file is located under the user's root directory.`,
		SilenceUsage: true,
		RunE:         runInitCommand,
	}

	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cmd.Flags().StringP("account", "a", "", "Conjur organization account name")
	cmd.Flags().StringP("url", "u", "", "URL of the Conjur service")
	cmd.Flags().StringP("certificate", "c", "", "Conjur SSL certificate (will be obtained from host unless provided by this option)")
	cmd.Flags().StringP("file", "f", filepath.Join(userHomeDir, ".conjurrc"), "File to write the configuration to")
	cmd.Flags().String("cert-file", filepath.Join(userHomeDir, "conjur-server.pem"), "File to write the server's certificate to")
	cmd.Flags().StringP("authn-type", "t", "", "Authentication type to use")
	cmd.Flags().String("service-id", "", "Service ID if using alternative authentication type")
	cmd.Flags().BoolP("self-signed", "s", false, "Allow self-signed certificates (insecure)")
	cmd.Flags().Bool("force", false, "Force overwrite of existing file")

	return cmd
}

func init() {
	initCmd := newInitCommand()
	rootCmd.AddCommand(initCmd)
}
