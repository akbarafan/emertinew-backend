package model

import (
	"testing"
)

func TestRoleHierarchyAndPermissions(t *testing.T) {
	// 1. RoleRank checks
	if RoleRank(RoleSuperadmin) <= RoleRank(RoleAdmin) {
		t.Errorf("Superadmin should rank higher than Admin")
	}
	if RoleRank(RoleAdmin) <= RoleRank(RoleCompanyAdmin) {
		t.Errorf("Admin should rank higher than Company Admin")
	}
	if RoleRank(RoleCompanyAdmin) <= RoleRank(RoleOperator) {
		t.Errorf("Company Admin should rank higher than Operator")
	}

	// 2. CanAssignRole checks
	// Superadmin can assign anything
	if !CanAssignRole(RoleSuperadmin, RoleSuperadmin) {
		t.Errorf("Superadmin should be able to assign Superadmin")
	}
	if !CanAssignRole(RoleSuperadmin, RoleAdmin) {
		t.Errorf("Superadmin should be able to assign Admin")
	}

	// Admin CANNOT create or assign Superadmin
	if CanAssignRole(RoleAdmin, RoleSuperadmin) {
		t.Errorf("Admin should NOT be able to create or assign Superadmin")
	}
	// Admin CANNOT create or assign TechAdmin (higher privilege)
	if CanAssignRole(RoleAdmin, RoleTechAdmin) {
		t.Errorf("Admin should NOT be able to create or assign TechAdmin")
	}
	// Admin CAN create Admin or roles below
	if !CanAssignRole(RoleAdmin, RoleAdmin) {
		t.Errorf("Admin should be able to assign Admin")
	}
	if !CanAssignRole(RoleAdmin, RoleCompanyAdmin) {
		t.Errorf("Admin should be able to assign CompanyAdmin")
	}
	if !CanAssignRole(RoleAdmin, RoleOperator) {
		t.Errorf("Admin should be able to assign Operator")
	}

	// CompanyAdmin CANNOT create Superadmin, TechAdmin, Admin, or CompanyAdmin
	if CanAssignRole(RoleCompanyAdmin, RoleSuperadmin) {
		t.Errorf("CompanyAdmin should NOT be able to assign Superadmin")
	}
	if CanAssignRole(RoleCompanyAdmin, RoleAdmin) {
		t.Errorf("CompanyAdmin should NOT be able to assign Admin")
	}
	if CanAssignRole(RoleCompanyAdmin, RoleCompanyAdmin) {
		t.Errorf("CompanyAdmin should NOT be able to assign CompanyAdmin")
	}
	if !CanAssignRole(RoleCompanyAdmin, RoleOperator) {
		t.Errorf("CompanyAdmin should be able to assign Operator")
	}

	// 3. CanDeleteUser checks
	// Superadmin can delete any role
	if !CanDeleteUser(RoleSuperadmin, RoleSuperadmin) {
		t.Errorf("Superadmin can delete other accounts")
	}
	if !CanDeleteUser(RoleSuperadmin, RoleAdmin) {
		t.Errorf("Superadmin can delete Admin")
	}

	// Admin CANNOT delete Superadmin
	if CanDeleteUser(RoleAdmin, RoleSuperadmin) {
		t.Errorf("Admin should NOT be able to delete Superadmin")
	}
	// Admin CANNOT delete TechAdmin
	if CanDeleteUser(RoleAdmin, RoleTechAdmin) {
		t.Errorf("Admin should NOT be able to delete TechAdmin")
	}
	// Admin CAN delete lower roles
	if !CanDeleteUser(RoleAdmin, RoleCompanyAdmin) {
		t.Errorf("Admin should be able to delete CompanyAdmin")
	}

	// 4. CanUpdateUser checks
	// Admin CANNOT update Superadmin
	if CanUpdateUser(RoleAdmin, RoleSuperadmin) {
		t.Errorf("Admin should NOT be able to update Superadmin")
	}
	// Superadmin CAN update Superadmin
	if !CanUpdateUser(RoleSuperadmin, RoleSuperadmin) {
		t.Errorf("Superadmin should be able to update Superadmin")
	}
}
