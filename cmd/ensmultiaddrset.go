// Copyright © 2017 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"fmt"
	"os"

	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wealdtech/ethereal/cli"
	"github.com/wealdtech/ethereal/ens"
)

var ensMultiaddrSetMultiaddrStr string

// ensMultiaddrSetCmd represents the ens multiaddr set command
var ensMultiaddrSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the multiaddr of an ENS domain",
	Long: `Set the multiaddr of a name registered with the Ethereum Name Service (ENS).  For example:

    ethereal ens multiaddr set --domain=enstest.eth --multiaddr=/ip4/1.2.3.4 --passphrase="my secret passphrase"

The keystore for the account that owns the name must be local (i.e. listed with 'get accounts list') and unlockable with the supplied passphrase.

In quiet mode this will return 0 if the transaction to set the multiaddr is sent successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		cli.Assert(!offline, quiet, "Offline mode not supported at current with this command")
		cli.Assert(ensDomain != "", quiet, "--domain is required")

		registryContract, err := ens.RegistryContract(client)
		cli.ErrCheck(err, quiet, "cannot obtain ENS registry contract")

		// Fetch the owner of the name
		owner, err := registryContract.Owner(nil, ens.NameHash(ensDomain))
		cli.ErrCheck(err, quiet, "cannot obtain owner")
		cli.Assert(bytes.Compare(owner.Bytes(), ens.UnknownAddress.Bytes()) != 0, quiet, fmt.Sprintf("owner of %s is not set", ensDomain))

		// Obtain the multiaddr
		multiaddr, err := ma.NewMultiaddr(ensMultiaddrSetMultiaddrStr)
		cli.ErrCheck(err, quiet, fmt.Sprintf("invalid multiaddr %s", ensMultiaddrSetMultiaddrStr))

		// Obtain the resolver for this name
		resolver, err := ens.ResolverContract(client, ensDomain)
		cli.ErrCheck(err, quiet, "No resolver for that name")

		opts, err := generateTxOpts(owner)
		cli.ErrCheck(err, quiet, "failed to generate transaction options")

		signedTx, err := resolver.SetMultiaddr(opts, ens.NameHash(ensDomain), multiaddr.Bytes())
		cli.ErrCheck(err, quiet, "failed to send transaction")

		setupLogging()
		log.WithFields(log.Fields{
			"group":         "ens/multiaddr",
			"command":       "set",
			"domain":        ensDomain,
			"multiaddr":     multiaddr.String(),
			"networkid":     chainID,
			"gas":           signedTx.Gas(),
			"gasprice":      signedTx.GasPrice().String(),
			"transactionid": signedTx.Hash().Hex(),
		}).Info("success")

		if quiet {
			os.Exit(0)
		}

		fmt.Println(signedTx.Hash().Hex())
	},
}

func init() {
	ensMultiaddrCmd.AddCommand(ensMultiaddrSetCmd)
	ensMultiaddrFlags(ensMultiaddrSetCmd)
	ensMultiaddrSetCmd.Flags().StringVar(&ensMultiaddrSetMultiaddrStr, "multiaddr", "", "The multiaddr to set")
	addTransactionFlags(ensMultiaddrSetCmd, "passphrase for the account that owns the domain")
}
