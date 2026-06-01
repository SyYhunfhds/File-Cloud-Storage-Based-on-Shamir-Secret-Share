package business

import (
	"fmt"
	"simulator/internal/algorithm"
	"time"
)

type ProposalStatus int

const (
	ProposalStatusPending ProposalStatus = iota
	ProposalStatusApproved
	ProposalStatusRejected
)

func (s ProposalStatus) String() string {
	switch s {
	case ProposalStatusPending:
		return "待审批"
	case ProposalStatusApproved:
		return "已通过"
	case ProposalStatusRejected:
		return "已拒绝"
	default:
		return "未知"
	}
}

type Proposal struct {
	ID          string
	Title       string
	Description string
	Amount      int64
	Status      ProposalStatus
	Required    int
	Submitted   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProposalManager struct {
	proposals map[string]*Proposal
	nextID    int
	shamir    *algorithm.Shamir
}

func NewProposalManager(shamir *algorithm.Shamir) *ProposalManager {
	return &ProposalManager{
		proposals: make(map[string]*Proposal),
		nextID:    1,
		shamir:    shamir,
	}
}

func (pm *ProposalManager) CreateProposal(title, description string, amount int64) (*Proposal, error) {
	id := fmt.Sprintf("proposal_%d", pm.nextID)
	pm.nextID++

	proposal := &Proposal{
		ID:          id,
		Title:       title,
		Description: description,
		Amount:      amount,
		Status:      ProposalStatusPending,
		Required:    pm.shamir.GetThreshold(),
		Submitted:   0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	pm.proposals[id] = proposal
	return proposal, nil
}

func (pm *ProposalManager) GetProposal(id string) (*Proposal, error) {
	proposal, ok := pm.proposals[id]
	if !ok {
		return nil, fmt.Errorf("proposal not found: %s", id)
	}
	return proposal, nil
}

func (pm *ProposalManager) ListProposals() []*Proposal {
	proposals := make([]*Proposal, 0, len(pm.proposals))
	for _, proposal := range pm.proposals {
		proposals = append(proposals, proposal)
	}
	return proposals
}

func (pm *ProposalManager) ListPendingProposals() []*Proposal {
	proposals := make([]*Proposal, 0)
	for _, proposal := range pm.proposals {
		if proposal.Status == ProposalStatusPending {
			proposals = append(proposals, proposal)
		}
	}
	return proposals
}

func (pm *ProposalManager) SubmitShare(proposalID string) error {
	proposal, ok := pm.proposals[proposalID]
	if !ok {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalStatusPending {
		return fmt.Errorf("proposal is not pending")
	}

	proposal.Submitted++
	proposal.UpdatedAt = time.Now()

	return nil
}

func (pm *ProposalManager) ApproveProposal(proposalID string) error {
	proposal, ok := pm.proposals[proposalID]
	if !ok {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Submitted < proposal.Required {
		return fmt.Errorf("not enough shares submitted")
	}

	proposal.Status = ProposalStatusApproved
	proposal.UpdatedAt = time.Now()

	return nil
}

func (pm *ProposalManager) RejectProposal(proposalID string) error {
	proposal, ok := pm.proposals[proposalID]
	if !ok {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	proposal.Status = ProposalStatusRejected
	proposal.UpdatedAt = time.Now()

	return nil
}

func (pm *ProposalManager) CanApprove(proposalID string) bool {
	proposal, ok := pm.proposals[proposalID]
	if !ok {
		return false
	}
	return proposal.Submitted >= proposal.Required
}

func (pm *ProposalManager) Count() int {
	return len(pm.proposals)
}
