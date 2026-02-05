package deimos

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// Testing marschain reward halving at different block heights
func TestRewardHalving(t *testing.T) {
	baseRewardPerBlockString := "7750496031750000000000"
	halvingBlock := int64(28800 * 448)
	halvingOffset := int64(77420)

	testCases := []struct {
		height int64
	}{
		{1},
		{28800},
		{28800*448 - halvingOffset - 1},
		{28800*448 - halvingOffset},
		{28800*448 - halvingOffset + 1},
		{28800*448*2 - halvingOffset - 1},
		{28800*448*2 - halvingOffset},
		{28800*448*2 - halvingOffset + 1},
		{28800*448*3 - halvingOffset},
		{28800*448*4 - halvingOffset},
		{28800*448*5 - halvingOffset},
	}

	for _, tc := range testCases {
		halvingCount := (tc.height + halvingOffset) / halvingBlock
		baseRewardPerBlock, _ := new(big.Int).SetString(baseRewardPerBlockString, 10)
		miningReward := new(big.Int).Div(baseRewardPerBlock, new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(halvingCount)), nil))
		fmt.Println("height", tc.height, "halvingCount:", halvingCount, "miningReward:", miningReward.String())
	}
}

func TestImpactOfValidatorOutOfService(t *testing.T) {
	testCases := []struct {
		totalValidators int
		downValidators  int
	}{
		{3, 1},
		{5, 2},
		{10, 1},
		{10, 4},
		{21, 1},
		{21, 3},
		{21, 5},
		{21, 10},
	}
	for _, tc := range testCases {
		simulateValidatorOutOfService(tc.totalValidators, tc.downValidators)
	}
}

func simulateValidatorOutOfService(totalValidators int, downValidators int) {
	downBlocks := 10000
	recoverBlocks := 10000
	recents := make(map[uint64]int)

	validators := make(map[int]bool, totalValidators)
	down := make([]int, totalValidators)
	for i := 0; i < totalValidators; i++ {
		validators[i] = true
		down[i] = i
	}
	rand.Shuffle(totalValidators, func(i, j int) {
		down[i], down[j] = down[j], down[i]
	})
	for i := 0; i < downValidators; i++ {
		delete(validators, down[i])
	}
	isRecentSign := func(idx int) bool {
		for _, signIdx := range recents {
			if signIdx == idx {
				return true
			}
		}
		return false
	}
	isInService := func(idx int) bool {
		return validators[idx]
	}

	downDelay := uint64(0)
	for h := 1; h <= downBlocks; h++ {
		if limit := uint64(totalValidators/2 + 1); uint64(h) >= limit {
			delete(recents, uint64(h)-limit)
		}
		proposer := h % totalValidators
		if !isInService(proposer) || isRecentSign(proposer) {
			candidates := make(map[int]bool, totalValidators/2)
			for v := range validators {
				if !isRecentSign(v) {
					candidates[v] = true
				}
			}
			if len(candidates) == 0 {
				panic("can not test such case")
			}
			idx, delay := producerBlockDelay(candidates, h, totalValidators)
			downDelay = downDelay + delay
			recents[uint64(h)] = idx
		} else {
			recents[uint64(h)] = proposer
		}
	}
	fmt.Printf("average delay is %v  when there is %d validators and %d is down \n",
		downDelay/uint64(downBlocks), totalValidators, downValidators)

	for i := 0; i < downValidators; i++ {
		validators[down[i]] = true
	}

	recoverDelay := uint64(0)
	lastseen := downBlocks
	for h := downBlocks + 1; h <= downBlocks+recoverBlocks; h++ {
		if limit := uint64(totalValidators/2 + 1); uint64(h) >= limit {
			delete(recents, uint64(h)-limit)
		}
		proposer := h % totalValidators
		if !isInService(proposer) || isRecentSign(proposer) {
			lastseen = h
			candidates := make(map[int]bool, totalValidators/2)
			for v := range validators {
				if !isRecentSign(v) {
					candidates[v] = true
				}
			}
			if len(candidates) == 0 {
				panic("can not test such case")
			}
			idx, delay := producerBlockDelay(candidates, h, totalValidators)
			recoverDelay = recoverDelay + delay
			recents[uint64(h)] = idx
		} else {
			recents[uint64(h)] = proposer
		}
	}
	fmt.Printf("total delay is %v after recover when there is %d validators down ever, last seen not proposer at height %d\n",
		recoverDelay, downValidators, lastseen)
}

func producerBlockDelay(candidates map[int]bool, height, numOfValidators int) (int, uint64) {

	s := rand.NewSource(int64(height))
	r := rand.New(s)
	n := numOfValidators
	backOffSteps := make([]int, 0, n)
	for idx := 0; idx < n; idx++ {
		backOffSteps = append(backOffSteps, idx)
	}
	r.Shuffle(n, func(i, j int) {
		backOffSteps[i], backOffSteps[j] = backOffSteps[j], backOffSteps[i]
	})
	minDelay := numOfValidators
	minCandidate := 0
	for c := range candidates {
		if minDelay > backOffSteps[c] {
			minDelay = backOffSteps[c]
			minCandidate = c
		}
	}
	delay := initialBackOffTime + uint64(minDelay)*wiggleTime
	return minCandidate, delay
}

func randomAddress() common.Address {
	addrBytes := make([]byte, 20)
	rand.Read(addrBytes)
	return common.BytesToAddress(addrBytes)
}
