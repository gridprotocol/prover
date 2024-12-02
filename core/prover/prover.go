package prover

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"grid-prover/core/client"
	"grid-prover/core/types"
	"grid-prover/logs"

	"github.com/ethereum/go-ethereum/crypto"
)

var logger = logs.Logger("grid prover")

type GRIDProver struct {
	nodeID types.NodeID

	last            int64
	prepareInterval time.Duration
	proverInterval  time.Duration
	waitInterval    time.Duration

	diffcult int

	done  chan struct{}
	doned bool

	client.Client
}

func NewGRIDProver(chain string, validatorUrl string, sk *ecdsa.PrivateKey, id int64) (*GRIDProver, error) {
	// get time information from validator
	prepareInterval := 10 * time.Second
	proveInterval := 10 * time.Second
	waitInterval := 2*time.Minute - prepareInterval - proveInterval

	return &GRIDProver{
		nodeID: types.NodeID{
			Address: crypto.PubkeyToAddress(sk.PublicKey).Hex(),
			ID:      id,
		},

		last:            0,
		prepareInterval: prepareInterval,
		proverInterval:  proveInterval,
		waitInterval:    waitInterval,

		diffcult: 10,

		done:  make(chan struct{}),
		doned: false,

		// new gridClient to communicate with validator
		Client: *client.NewClient(validatorUrl),
	}, nil
}

func (p *GRIDProver) Start(ctx context.Context) {
	for {
		// 1 sec for each loop
		time.Sleep(1 * time.Second)

		wait, nextTime := p.CalculateWatingTime()
		select {
		case <-ctx.Done():
			p.doned = true
			return
		case <-p.done:
			p.doned = true
			return
		case <-time.After(wait):
		}

		cnt, err := p.GetV1OrderCount(ctx, "0xEf95c72C836605203F7f66788E450Af2a4141957")
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		fmt.Println("order count for provider: ", cnt)

		// get rnd from Validator/Contract
		rnd, err := p.GetRND(ctx)
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		// generate proof
		res, err := p.GenerateProof(ctx, rnd)
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		// submit proof to validator
		success, err := p.ProveToValidator(ctx, res)
		if err != nil {
			logger.Error(err.Error())
			// continue
		}

		if success {
			logger.Infof("Proof of Work Generation Successfully result[%d]", res)
		} else {
			logger.Info("Proof of Work Generation Falied")
		}

		p.last = nextTime
	}
}

func (p *GRIDProver) Stop() {
	close(p.done)

	for !p.doned {
		time.Sleep(200 * time.Millisecond)
	}
}

func (p *GRIDProver) CalculateWatingTime() (time.Duration, int64) {
	challengeCycleSeconds := int64((p.prepareInterval + p.proverInterval + p.waitInterval).Seconds())
	now := time.Now().Unix()
	duration := now - p.last
	over := duration % challengeCycleSeconds

	var waitingSeconds int64 = 0
	if over < int64(p.prepareInterval.Seconds()) {
		waitingSeconds = int64(p.prepareInterval.Seconds()) - over
		p.last = now - over
	} else if over > int64((p.prepareInterval + p.proverInterval).Seconds()) {
		waitingSeconds = challengeCycleSeconds + int64(p.prepareInterval.Seconds()) - over
		p.last = now - over + challengeCycleSeconds
	}

	next := p.last + challengeCycleSeconds

	return time.Duration(waitingSeconds) * time.Second, next
}

// generate a proof with a random value
func (p *GRIDProver) GenerateProof(ctx context.Context, rnd [32]byte) (int64, error) {
	return GeneratePOW(p.nodeID, rnd[:], p.diffcult)
}

// submit proof to validator
func (p *GRIDProver) ProveToValidator(ctx context.Context, result int64) (bool, error) {
	err := p.SubmitProof(ctx, types.Proof{
		NodeID: p.nodeID,
		Nonce:  result,
	})
	if err != nil {
		return false, err
	}

	return true, nil
}
