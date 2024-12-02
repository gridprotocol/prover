package cmd

import (
	"context"
	"grid-prover/core/prover"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

var ProverCmd = &cli.Command{
	Name:  "prover",
	Usage: "grid prover node",
	Subcommands: []*cli.Command{
		runCmd,
		// challengerNodeStopCmd,
		// queryProfitsCmd,
	},
}

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "run grid prover node",
	Flags: []cli.Flag{
		// &cli.StringFlag{
		// 	Name:    "endpoint",
		// 	Aliases: []string{"e"},
		// 	Usage:   "input your endpoint",
		// 	Value:   ":8082",
		// },
		&cli.StringFlag{
			Name:  "sk",
			Usage: "input your private key",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "chain",
			Usage: "input chain name, e.g.(dev)",
			Value: "dev",
		},
		&cli.StringFlag{
			Name:  "ip",
			Usage: "input validator node's ip address",
			Value: "http://localhost:8081",
		},
	},
	Action: func(ctx *cli.Context) error {
		// endPoint := ctx.String("endpoint")
		sk := ctx.String("sk")
		chain := ctx.String("chain")
		ip := ctx.String("ip")

		privateKey, err := crypto.HexToECDSA(sk)
		if err != nil {
			privateKey, err = crypto.GenerateKey()
			if err != nil {
				return err
			}
		}

		cctx, cancel := context.WithCancel(ctx.Context)
		defer cancel()

		// err = database.InitDatabase("~/.meeda-prover")
		// if err != nil {
		// 	return err
		// }

		// dumper, err := core.NewGRIDDumper(chain, common.Address{})
		// if err != nil {
		// 	return err
		// }

		// err = dumper.
		// if err != nil {
		// 	return err
		// }
		// go dumper.SubscribeFileProof(cctx)

		prover, err := prover.NewGRIDProver(chain, ip, privateKey, 1)
		if err != nil {
			log.Fatalf("new light node prover: %s\n", err)
		}
		go prover.Start(cctx)

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")

		prover.Stop()
		log.Println("Server exiting")

		return nil
	},
}
