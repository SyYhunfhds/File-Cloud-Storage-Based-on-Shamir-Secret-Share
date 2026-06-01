package business

import (
	"fmt"
	"simulator/internal/algorithm"
)

type ShareManager struct {
	shamir         *algorithm.Shamir
	submittedShares map[string][]algorithm.Share
}

func NewShareManager(shamir *algorithm.Shamir) *ShareManager {
	return &ShareManager{
		shamir:         shamir,
		submittedShares: make(map[string][]algorithm.Share),
	}
}

func (sm *ShareManager) SubmitShare(proposalID string, share algorithm.Share) error {
	if _, ok := sm.submittedShares[proposalID]; !ok {
		sm.submittedShares[proposalID] = make([]algorithm.Share, 0)
	}

	sm.submittedShares[proposalID] = append(sm.submittedShares[proposalID], share)
	return nil
}

func (sm *ShareManager) GetSubmittedShares(proposalID string) []algorithm.Share {
	shares, ok := sm.submittedShares[proposalID]
	if !ok {
		return make([]algorithm.Share, 0)
	}
	return shares
}

func (sm *ShareManager) CanRecover(proposalID string) bool {
	shares := sm.GetSubmittedShares(proposalID)
	return len(shares) >= sm.shamir.GetThreshold()
}

func (sm *ShareManager) RecoverSecret(proposalID string) (int64, error) {
	shares := sm.GetSubmittedShares(proposalID)
	if len(shares) < sm.shamir.GetThreshold() {
		return 0, fmt.Errorf("not enough shares to recover secret")
	}

	return sm.shamir.Recover(shares[:sm.shamir.GetThreshold()])
}

func (sm *ShareManager) ClearShares(proposalID string) {
	delete(sm.submittedShares, proposalID)
}

func (sm *ShareManager) GetShareCount(proposalID string) int {
	return len(sm.GetSubmittedShares(proposalID))
}

func (sm *ShareManager) Reshare(proposalID string) error {
	shares := sm.GetSubmittedShares(proposalID)
	if len(shares) < sm.shamir.GetThreshold() {
		return fmt.Errorf("not enough shares to reshare")
	}

	newShares, err := sm.shamir.Reshare(shares[:sm.shamir.GetThreshold()])
	if err != nil {
		return err
	}

	sm.submittedShares[proposalID] = newShares
	return nil
}
