package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/repositories"
)

func main() {
	ctx := context.Background()

	dsn := "host=localhost user=postgres password=postgres dbname=giia_auth port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	permRepo := repositories.NewPermissionRepository(db)
	roleRepo := repositories.NewRoleRepository(db)

	permissions := getSystemPermissions()

	log.Printf("Seeding %d permissions...", len(permissions))

	if err := permRepo.BatchCreate(ctx, permissions); err != nil {
		log.Fatalf("Failed to seed permissions: %v", err)
	}

	log.Println("Permissions seeded successfully!")

	if err := assignDefaultPermissions(ctx, roleRepo, permRepo); err != nil {
		log.Fatalf("Failed to assign default permissions: %v", err)
	}

	log.Println("Default permissions assigned to system roles successfully!")
	log.Println("✅ Permission seeding completed!")
}

func getSystemPermissions() []*domain.Permission {
	permissions := []*domain.Permission{}

	authPermissions := [][]string{
		{"auth:users:read", "View user information"},
		{"auth:users:write", "Create and update users"},
		{"auth:users:delete", "Delete users"},
		{"auth:roles:read", "View roles and permissions"},
		{"auth:roles:write", "Create and update roles"},
		{"auth:roles:delete", "Delete roles"},
		{"auth:permissions:read", "View permissions"},
		{"auth:permissions:write", "Create and update permissions"},
	}

	catalogPermissions := [][]string{
		{"catalog:products:read", "View products"},
		{"catalog:products:write", "Create and update products"},
		{"catalog:products:delete", "Delete products"},
		{"catalog:suppliers:read", "View suppliers"},
		{"catalog:suppliers:write", "Create and update suppliers"},
		{"catalog:suppliers:delete", "Delete suppliers"},
		{"catalog:profiles:read", "View product profiles"},
		{"catalog:profiles:write", "Create and update product profiles"},
		{"catalog:profiles:delete", "Delete product profiles"},
	}

	ddmrpPermissions := [][]string{
		{"ddmrp:buffers:read", "View DDMRP buffers"},
		{"ddmrp:buffers:write", "Create and update DDMRP buffers"},
		{"ddmrp:buffers:delete", "Delete DDMRP buffers"},
		{"ddmrp:calculations:read", "View DDMRP calculations"},
		{"ddmrp:calculations:execute", "Execute DDMRP calculations"},
		{"ddmrp:zones:read", "View buffer zones"},
		{"ddmrp:zones:write", "Update buffer zones"},
	}

	executionPermissions := [][]string{
		{"execution:orders:read", "View execution orders"},
		{"execution:orders:write", "Create and update execution orders"},
		{"execution:orders:delete", "Delete execution orders"},
		{"execution:schedules:read", "View schedules"},
		{"execution:schedules:write", "Create and update schedules"},
	}

	analyticsPermissions := [][]string{
		{"analytics:reports:read", "View analytics reports"},
		{"analytics:reports:write", "Create and update reports"},
		{"analytics:dashboards:read", "View dashboards"},
		{"analytics:dashboards:write", "Create and update dashboards"},
	}

	aiAgentPermissions := [][]string{
		{"ai_agent:queries:read", "View AI agent queries"},
		{"ai_agent:queries:write", "Execute AI agent queries"},
		{"ai_agent:models:read", "View AI models"},
		{"ai_agent:models:write", "Update AI models"},
	}

	allServicePermissions := [][]string{}
	allServicePermissions = append(allServicePermissions, authPermissions...)
	allServicePermissions = append(allServicePermissions, catalogPermissions...)
	allServicePermissions = append(allServicePermissions, ddmrpPermissions...)
	allServicePermissions = append(allServicePermissions, executionPermissions...)
	allServicePermissions = append(allServicePermissions, analyticsPermissions...)
	allServicePermissions = append(allServicePermissions, aiAgentPermissions...)

	for _, perm := range allServicePermissions {
		code := perm[0]
		description := perm[1]
		parts := parsePermissionCode(code)

		permissions = append(permissions, &domain.Permission{
			Code:        code,
			Description: description,
			Service:     parts[0],
			Resource:    parts[1],
			Action:      parts[2],
		})
	}

	return permissions
}

func parsePermissionCode(code string) []string {
	parts := []string{}
	current := ""
	for _, char := range code {
		if char == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	parts = append(parts, current)
	return parts
}

func assignDefaultPermissions(ctx context.Context, roleRepo providers.RoleRepository, permRepo providers.PermissionRepository) error {
	allPermissions, err := permRepo.List(ctx)
	if err != nil {
		return pkgErrors.NewInternalServerError("failed to get all permissions")
	}

	viewerID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	analystID := uuid.MustParse("00000000-0000-0000-0000-000000000020")
	managerID := uuid.MustParse("00000000-0000-0000-0000-000000000030")

	viewerPerms := []uuid.UUID{}
	analystPerms := []uuid.UUID{}
	managerPerms := []uuid.UUID{}

	for _, perm := range allPermissions {
		if perm.Code == "*:*:*" {
			continue
		}

		if perm.Action == "read" {
			viewerPerms = append(viewerPerms, perm.ID)
		}

		if perm.Service == "analytics" {
			analystPerms = append(analystPerms, perm.ID)
		}

		if perm.Service == "catalog" || perm.Service == "ddmrp" || perm.Service == "execution" {
			managerPerms = append(managerPerms, perm.ID)
		}
	}

	log.Printf("Assigning %d permissions to Viewer role", len(viewerPerms))
	if err := permRepo.AssignPermissionsToRole(ctx, viewerID, viewerPerms); err != nil {
		return pkgErrors.NewInternalServerError("failed to assign permissions to Viewer role")
	}

	log.Printf("Assigning %d permissions to Analyst role", len(analystPerms))
	if err := permRepo.AssignPermissionsToRole(ctx, analystID, analystPerms); err != nil {
		return pkgErrors.NewInternalServerError("failed to assign permissions to Analyst role")
	}

	log.Printf("Assigning %d permissions to Manager role", len(managerPerms))
	if err := permRepo.AssignPermissionsToRole(ctx, managerID, managerPerms); err != nil {
		return pkgErrors.NewInternalServerError("failed to assign permissions to Manager role")
	}

	return nil
}
