package business

import (
	"fmt"
	"simulator/internal/algorithm"
)

type Role struct {
	ID     string
	Name   string
	Weight int
	Shares []algorithm.Share
}

type RoleManager struct {
	roles    map[string]*Role
	nextID   int
	shamir   *algorithm.Shamir
}

func NewRoleManager(shamir *algorithm.Shamir) *RoleManager {
	return &RoleManager{
		roles:  make(map[string]*Role),
		nextID: 1,
		shamir: shamir,
	}
}

func (rm *RoleManager) AddRole(name string, weight int) (*Role, error) {
	id := fmt.Sprintf("role_%d", rm.nextID)
	rm.nextID++

	role := &Role{
		ID:     id,
		Name:   name,
		Weight: weight,
		Shares: make([]algorithm.Share, 0),
	}

	rm.roles[id] = role
	return role, nil
}

func (rm *RoleManager) RemoveRole(id string) error {
	if _, ok := rm.roles[id]; !ok {
		return fmt.Errorf("role not found: %s", id)
	}
	delete(rm.roles, id)
	return nil
}

func (rm *RoleManager) GetRole(id string) (*Role, error) {
	role, ok := rm.roles[id]
	if !ok {
		return nil, fmt.Errorf("role not found: %s", id)
	}
	return role, nil
}

func (rm *RoleManager) ListRoles() []*Role {
	roles := make([]*Role, 0, len(rm.roles))
	for _, role := range rm.roles {
		roles = append(roles, role)
	}
	return roles
}

func (rm *RoleManager) AssignShares(roleID string, shares []algorithm.Share) error {
	role, ok := rm.roles[roleID]
	if !ok {
		return fmt.Errorf("role not found: %s", roleID)
	}
	role.Shares = shares
	return nil
}

func (rm *RoleManager) DistributeShares(secret int64) error {
	shares, err := rm.shamir.Split(secret)
	if err != nil {
		return err
	}

	threshold := rm.shamir.GetThreshold()
	numShares := rm.shamir.GetNumShares()
	sharesPerRole := numShares / len(rm.roles)
	extraShares := numShares % len(rm.roles)

	currentShare := 0
	for _, role := range rm.roles {
		endIdx := currentShare + sharesPerRole
		if extraShares > 0 {
			endIdx++
			extraShares--
		}
		if endIdx > len(shares) {
			endIdx = len(shares)
		}

		roleShares := make([]algorithm.Share, endIdx-currentShare)
		copy(roleShares, shares[currentShare:endIdx])
		role.Shares = roleShares
		currentShare = endIdx
	}

	_ = threshold
	return nil
}

func (rm *RoleManager) Count() int {
	return len(rm.roles)
}
