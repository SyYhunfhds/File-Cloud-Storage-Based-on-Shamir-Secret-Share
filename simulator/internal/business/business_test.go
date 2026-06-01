package business

import (
	"testing"

	"simulator/internal/algorithm"
)

func TestRoleManager(t *testing.T) {
	config := &algorithm.Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     257,
	}

	shamir, err := algorithm.NewShamir(config)
	if err != nil {
		t.Fatalf("NewShamir() failed: %v", err)
	}

	roleManager := NewRoleManager(shamir)

	// 测试添加角色
	role1, err := roleManager.AddRole("Role1", 1)
	if err != nil {
		t.Fatalf("AddRole() failed: %v", err)
	}

	role2, err := roleManager.AddRole("Role2", 1)
	if err != nil {
		t.Fatalf("AddRole() failed: %v", err)
	}

	role3, err := roleManager.AddRole("Role3", 1)
	if err != nil {
		t.Fatalf("AddRole() failed: %v", err)
	}

	roles := roleManager.ListRoles()
	if len(roles) != 3 {
		t.Errorf("Expected 3 roles, got %d", len(roles))
	}

	// 测试分发份额
	secret := int64(42)
	err = roleManager.DistributeShares(secret)
	if err != nil {
		t.Fatalf("DistributeShares() failed: %v", err)
	}

	// 测试获取角色份额
	if len(role1.Shares) == 0 {
		t.Error("Expected shares for Role1")
	}

	if len(role2.Shares) == 0 {
		t.Error("Expected shares for Role2")
	}

	if len(role3.Shares) == 0 {
		t.Error("Expected shares for Role3")
	}

	// 测试获取角色
	retrievedRole, err := roleManager.GetRole(role1.ID)
	if err != nil {
		t.Fatalf("GetRole() failed: %v", err)
	}

	if retrievedRole.Name != "Role1" {
		t.Errorf("Expected role name 'Role1', got '%s'", retrievedRole.Name)
	}

	// 测试删除角色
	err = roleManager.RemoveRole(role1.ID)
	if err != nil {
		t.Fatalf("RemoveRole() failed: %v", err)
	}

	roles = roleManager.ListRoles()
	if len(roles) != 2 {
		t.Errorf("Expected 2 roles after removal, got %d", len(roles))
	}
}

func TestProposalManager(t *testing.T) {
	config := &algorithm.Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     257,
	}

	shamir, err := algorithm.NewShamir(config)
	if err != nil {
		t.Fatalf("NewShamir() failed: %v", err)
	}

	proposalManager := NewProposalManager(shamir)

	// 测试创建提案
	proposal, err := proposalManager.CreateProposal("Test Proposal", "Test description", int64(10000))
	if err != nil {
		t.Fatalf("CreateProposal() failed: %v", err)
	}

	// 测试获取提案
	retrievedProposal, err := proposalManager.GetProposal(proposal.ID)
	if err != nil {
		t.Fatalf("GetProposal() failed: %v", err)
	}

	if retrievedProposal.Title != "Test Proposal" {
		t.Errorf("Expected title 'Test Proposal', got '%s'", retrievedProposal.Title)
	}

	if retrievedProposal.Status != ProposalStatusPending {
		t.Errorf("Expected status ProposalStatusPending, got %v", retrievedProposal.Status)
	}

	// 测试提交份额
	err = proposalManager.SubmitShare(proposal.ID)
	if err != nil {
		t.Fatalf("SubmitShare() failed: %v", err)
	}

	// 测试获取待审批提案
	pendingProposals := proposalManager.ListPendingProposals()
	if len(pendingProposals) != 1 {
		t.Errorf("Expected 1 pending proposal, got %d", len(pendingProposals))
	}

	// 测试获取所有提案
	allProposals := proposalManager.ListProposals()
	if len(allProposals) != 1 {
		t.Errorf("Expected 1 proposal, got %d", len(allProposals))
	}

	// 测试提案不存在的情况
	_, err = proposalManager.GetProposal("NonExistentProposal")
	if err == nil {
		t.Error("Expected error for non-existent proposal")
	}

	// 测试审批提案
	for i := 0; i < 2; i++ {
		err = proposalManager.SubmitShare(proposal.ID)
		if err != nil {
			t.Fatalf("SubmitShare() failed: %v", err)
		}
	}

	err = proposalManager.ApproveProposal(proposal.ID)
	if err != nil {
		t.Fatalf("ApproveProposal() failed: %v", err)
	}

	retrievedProposal, err = proposalManager.GetProposal(proposal.ID)
	if err != nil {
		t.Fatalf("GetProposal() failed: %v", err)
	}

	if retrievedProposal.Status != ProposalStatusApproved {
		t.Errorf("Expected status ProposalStatusApproved, got %v", retrievedProposal.Status)
	}
}

func TestShareManager(t *testing.T) {
	config := &algorithm.Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     257,
	}

	shamir, err := algorithm.NewShamir(config)
	if err != nil {
		t.Fatalf("NewShamir() failed: %v", err)
	}

	shareManager := NewShareManager(shamir)

	// 测试提交份额
	proposalID := "test-proposal"
	share1 := algorithm.Share{X: 1, Y: 100}
	share2 := algorithm.Share{X: 2, Y: 200}
	share3 := algorithm.Share{X: 3, Y: 300}

	err = shareManager.SubmitShare(proposalID, share1)
	if err != nil {
		t.Fatalf("SubmitShare() failed: %v", err)
	}

	err = shareManager.SubmitShare(proposalID, share2)
	if err != nil {
		t.Fatalf("SubmitShare() failed: %v", err)
	}

	err = shareManager.SubmitShare(proposalID, share3)
	if err != nil {
		t.Fatalf("SubmitShare() failed: %v", err)
	}

	// 测试获取提交的份额
	shares := shareManager.GetSubmittedShares(proposalID)
	if len(shares) != 3 {
		t.Errorf("Expected 3 shares, got %d", len(shares))
	}

	// 测试是否可以恢复
	if !shareManager.CanRecover(proposalID) {
		t.Error("Expected can recover with 3 shares")
	}

	// 测试份额计数
	count := shareManager.GetShareCount(proposalID)
	if count != 3 {
		t.Errorf("Expected 3 shares, got %d", count)
	}

	// 测试清除份额
	shareManager.ClearShares(proposalID)
	shares = shareManager.GetSubmittedShares(proposalID)
	if len(shares) != 0 {
		t.Errorf("Expected 0 shares after clear, got %d", len(shares))
	}

	// 测试份额不足的情况
	shareManager.SubmitShare(proposalID, share1)
	if shareManager.CanRecover(proposalID) {
		t.Error("Expected cannot recover with 1 share")
	}
}

func TestCompleteFlow(t *testing.T) {
	config := &algorithm.Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     257,
	}

	shamir, err := algorithm.NewShamir(config)
	if err != nil {
		t.Fatalf("NewShamir() failed: %v", err)
	}

	roleManager := NewRoleManager(shamir)
	proposalManager := NewProposalManager(shamir)
	shareManager := NewShareManager(shamir)

	// 添加角色
	_, err = roleManager.AddRole("Role1", 1)
	if err != nil {
		t.Fatalf("AddRole() failed: %v", err)
	}

	_, err = roleManager.AddRole("Role2", 1)
	if err != nil {
		t.Fatalf("AddRole() failed: %v", err)
	}

	_, err = roleManager.AddRole("Role3", 1)
	if err != nil {
		t.Fatalf("AddRole() failed: %v", err)
	}

	// 创建议案
	proposal, err := proposalManager.CreateProposal("Fund Transfer", "Test transfer", int64(50000))
	if err != nil {
		t.Fatalf("CreateProposal() failed: %v", err)
	}

	// 提交份额
	shares := []algorithm.Share{
		{X: 1, Y: 150},
		{X: 2, Y: 200},
		{X: 3, Y: 250},
	}

	for _, share := range shares {
		err = shareManager.SubmitShare(proposal.ID, share)
		if err != nil {
			t.Fatalf("SubmitShare() failed: %v", err)
		}
		err = proposalManager.SubmitShare(proposal.ID)
		if err != nil {
			t.Fatalf("SubmitShare() failed: %v", err)
		}
	}

	// 测试是否可以恢复
	if !shareManager.CanRecover(proposal.ID) {
		t.Error("Expected can recover with 3 shares")
	}

	// 测试是否可以审批
	if !proposalManager.CanApprove(proposal.ID) {
		t.Error("Expected can approve with 3 shares")
	}

	// 测试审批提案
	err = proposalManager.ApproveProposal(proposal.ID)
	if err != nil {
		t.Fatalf("ApproveProposal() failed: %v", err)
	}
}
