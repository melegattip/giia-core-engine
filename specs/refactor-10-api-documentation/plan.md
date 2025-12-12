# Implementation Plan: Document API with OpenAPI/Swagger

**Date**: 2025-12-10
**Spec**: [spec.md](./spec.md)

## Summary

Generate OpenAPI 3.0 specification for all HTTP endpoints using swaggo/swag. Add Swagger UI for interactive API exploration. Document all request/response schemas, error codes, and authentication requirements.

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: swaggo/swag, swaggo/gin-swagger, swaggo/files
**API Framework**: Gin (HTTP), protocol buffers (gRPC - not covered in this plan)
**Testing**: Manual testing via Swagger UI
**Target Platform**: Web browser (Swagger UI)
**Performance Goals**: Swagger UI loads <2 seconds
**Constraints**: Must auto-generate from code (no manual YAML editing)
**Scale/Scope**: ~15-20 HTTP endpoints across auth, user, role handlers

## Project Structure

### Files to Create/Modify

```text
services/auth-service/
├── cmd/
│   └── api/
│       └── main.go                         # MODIFY - add Swagger UI integration
├── internal/
│   └── infrastructure/
│       └── entrypoints/
│           └── http/
│               └── handlers/
│                   ├── auth_handler.go      # MODIFY - add swagger annotations
│                   ├── user_handler.go      # MODIFY - add swagger annotations
│                   └── role_handler.go      # MODIFY - add swagger annotations
├── docs/
│   ├── swagger.yaml                        # GENERATED - OpenAPI spec
│   ├── swagger.json                        # GENERATED - OpenAPI spec (JSON)
│   └── docs.go                             # GENERATED - embedded docs for Swagger UI
├── Makefile                                # UPDATE - add swagger target
├── README.md                               # UPDATE - add API docs section
└── .gitignore                              # UPDATE - don't ignore generated docs
```

---

## Phase 1: Install Swaggo & Dependencies

**Purpose**: Set up swaggo tooling

- [ ] T001 Install swag CLI: `go install github.com/swaggo/swag/cmd/swag@latest`
- [ ] T002 Verify installation: `swag --version`
- [ ] T003 Add swag dependencies to go.mod: `go get -u github.com/swaggo/gin-swagger github.com/swaggo/files`
- [ ] T004 Create docs/ directory: `mkdir -p docs`
- [ ] T005 Run initial swag init to verify setup: `swag init -g cmd/api/main.go`

**Checkpoint**: Swaggo installed, can generate docs

---

## Phase 2: Add General API Info to main.go

**Goal**: Add top-level API documentation

- [ ] T006 Open cmd/api/main.go
- [ ] T007 Add swag general info comments at top of file
- [ ] T008 Add `@title` annotation: "GIIA Auth Service API"
- [ ] T009 Add `@version` annotation: "1.0"
- [ ] T010 Add `@description` annotation: "Authentication and authorization service for GIIA platform"
- [ ] T011 Add `@contact.name` annotation: "API Support"
- [ ] T012 Add `@contact.email` annotation: "support@giia.io"
- [ ] T013 Add `@host` annotation: "localhost:8080"
- [ ] T014 Add `@BasePath` annotation: "/api/v1"
- [ ] T015 Add `@securityDefinitions.apikey BearerAuth` for JWT authentication
- [ ] T016 Test: run `swag init -g cmd/api/main.go`

**Checkpoint**: General API info configured

---

## Phase 3: Add Swagger Annotations - Auth Endpoints

**Goal**: Document authentication endpoints

- [ ] T017 [P] [US1] Open internal/infrastructure/entrypoints/http/handlers/auth_handler.go
- [ ] T018 [P] [US1] Add Login endpoint annotations:
  - `@Summary User login`
  - `@Description Authenticate user with email and password`
  - `@Tags auth`
  - `@Accept json`
  - `@Produce json`
  - `@Param request body LoginRequest true "Login credentials"`
  - `@Success 200 {object} LoginResponse "Access and refresh tokens"`
  - `@Failure 400 {object} ErrorResponse "Invalid request"`
  - `@Failure 401 {object} ErrorResponse "Invalid credentials"`
  - `@Failure 500 {object} ErrorResponse "Internal server error"`
  - `@Router /auth/login [post]`
- [ ] T019 [P] [US1] Add Register endpoint annotations (similar structure)
- [ ] T020 [P] [US1] Add Refresh Token endpoint annotations
- [ ] T021 [P] [US1] Add Logout endpoint annotations (if exists)
- [ ] T022 [P] [US1] Add example values to request/response structs using `example:"value"` tags

**Checkpoint**: Auth endpoints documented

---

## Phase 4: Add Swagger Annotations - User Endpoints

**Goal**: Document user management endpoints

- [ ] T023 [US1] Open internal/infrastructure/entrypoints/http/handlers/user_handler.go
- [ ] T024 [US1] Add GetCurrentUser endpoint annotations:
  - Include `@Security BearerAuth` for protected endpoints
  - Document request headers (Authorization: Bearer token)
- [ ] T025 [US1] Add UpdateProfile endpoint annotations
- [ ] T026 [US1] Add ChangePassword endpoint annotations
- [ ] T027 [US1] Add ListUsers endpoint annotations (admin only)
- [ ] T028 [US1] Add example values to User model struct

**Checkpoint**: User endpoints documented

---

## Phase 5: Add Swagger Annotations - Role/Permission Endpoints

**Goal**: Document RBAC endpoints

- [ ] T029 [US1] Open internal/infrastructure/entrypoints/http/handlers/role_handler.go
- [ ] T030 [US1] Add CreateRole endpoint annotations
- [ ] T031 [US1] Add AssignPermission endpoint annotations
- [ ] T032 [US1] Add GetUserPermissions endpoint annotations
- [ ] T033 [US1] Add CheckPermission endpoint annotations
- [ ] T034 [US1] Document Permission model with example permission codes
- [ ] T035 [US1] Document Role model with example role structure

**Checkpoint**: RBAC endpoints documented

---

## Phase 6: Document Error Responses

**Goal**: Standardize error response documentation

- [ ] T036 [US4] Create ErrorResponse model with swagger annotations
- [ ] T037 [US4] Add fields: error_code, message, http_status
- [ ] T038 [US4] Add example values for each error type
- [ ] T039 [US4] Document common error codes: BAD_REQUEST, UNAUTHORIZED_REQUEST, FORBIDDEN, RESOURCE_NOT_FOUND, INTERNAL_SERVER_ERROR
- [ ] T040 [US4] Create error code reference table in docs/api-error-codes.md
- [ ] T041 [US4] Verify all endpoints document appropriate error responses

**Checkpoint**: Error responses standardized and documented

---

## Phase 7: Generate OpenAPI Spec & Integrate Swagger UI

**Goal**: Generate docs and serve Swagger UI

- [ ] T042 Run swag init: `swag init -g cmd/api/main.go --output docs/`
- [ ] T043 Verify docs generated: docs/swagger.yaml, docs/swagger.json, docs/docs.go
- [ ] T044 Open cmd/api/main.go
- [ ] T045 Import swagger dependencies:
  ```go
  import (
      swaggerFiles "github.com/swaggo/files"
      ginSwagger "github.com/swaggo/gin-swagger"
      _ "github.com/giia/giia-core-engine/services/auth-service/docs" // import generated docs
  )
  ```
- [ ] T046 Add Swagger UI route:
  ```go
  router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
  ```
- [ ] T047 Build and run application: `go run cmd/api/main.go`
- [ ] T048 Access Swagger UI: http://localhost:8080/swagger/index.html
- [ ] T049 Test "Try it out" functionality for each endpoint
- [ ] T050 Test authentication: use /auth/login to get token, then "Authorize" button

**Checkpoint**: Swagger UI accessible and functional

---

## Phase 8: Makefile & Automation

**Goal**: Automate swagger generation

- [ ] T051 Add Makefile target: `make swagger`
  ```makefile
  swagger:
      swag init -g cmd/api/main.go --output docs/
      @echo "Swagger docs generated at docs/"
  ```
- [ ] T052 Add Makefile target: `make swagger-serve`
  ```makefile
  swagger-serve: swagger
      @echo "Swagger UI available at http://localhost:8080/swagger/index.html"
      go run cmd/api/main.go
  ```
- [ ] T053 Add swagger generation to `make build` (or create `make build-with-docs`)
- [ ] T054 Update .gitignore: commit generated docs (don't ignore docs/)
- [ ] T055 Test targets: `make swagger && make swagger-serve`

**Checkpoint**: Swagger generation automated

---

## Phase 9: Documentation & Team Rollout

**Purpose**: Document API and educate team

- [ ] T056 Create docs/api-documentation.md
- [ ] T057 Document how to access Swagger UI: http://localhost:8080/swagger/index.html
- [ ] T058 Document how to regenerate docs: `make swagger`
- [ ] T059 Document how to add annotations to new endpoints
- [ ] T060 Document authentication flow: login → copy token → Authorize button
- [ ] T061 Update README.md: add "API Documentation" section with Swagger UI link
- [ ] T062 Add API documentation link to developer onboarding
- [ ] T063 Create team announcement: API docs now available
- [ ] T064 Demo to frontend team: show how to use Swagger UI for testing
- [ ] T065 Demo to QA team: show how to test endpoints interactively
- [ ] T066 Add API docs review to PR checklist: new endpoints must have swagger annotations

**Checkpoint**: API documentation complete and team trained

---

## Phase 10: Advanced Features (Optional)

**Purpose**: Enhance API documentation

- [ ] T067 Add request/response examples with realistic data
- [ ] T068 Add schema descriptions for all model fields
- [ ] T069 Document query parameters for list endpoints (pagination, filtering)
- [ ] T070 Document multipart/form-data endpoints if any (file uploads)
- [ ] T071 Add tags for grouping: @Tags auth, @Tags user, @Tags rbac
- [ ] T072 Configure Swagger UI theme/styling if needed
- [ ] T073 Add postman collection export (Swagger can generate)
- [ ] T074 Consider API versioning strategy (v1, v2 paths)

**Checkpoint**: Enhanced API documentation with advanced features

---

## Dependencies & Execution Order

### Phase Dependencies

- **Install (Phase 1)**: No dependencies - must complete first
- **General Info (Phase 2)**: Depends on Phase 1
- **Auth Endpoints (Phase 3)**: Depends on Phase 2
- **User Endpoints (Phase 4)**: Depends on Phase 2, can run parallel with Phase 3
- **RBAC Endpoints (Phase 5)**: Depends on Phase 2, can run parallel with Phase 3-4
- **Error Responses (Phase 6)**: Depends on Phase 3-5 (needs all endpoints documented)
- **Generate & Integrate (Phase 7)**: Depends on Phase 3-6 (all annotations must exist)
- **Makefile (Phase 8)**: Depends on Phase 7
- **Documentation (Phase 9)**: Depends on all previous phases
- **Advanced (Phase 10)**: Optional, depends on Phase 9

### Parallel Execution

Phases 3-5 (endpoint annotations) can run in parallel.

## Notes

- Swaggo automatically generates OpenAPI 3.0 spec from code annotations
- Regenerate docs after any endpoint changes: `make swagger`
- Swagger UI is interactive - allows testing endpoints without Postman
- Example values in annotations improve API consumer experience
- Keep swagger annotations close to handler code (single source of truth)
- Frontend team can import swagger.yaml into code generation tools
- Consider hosting Swagger UI in staging/production for external API consumers
- Swagger supports both JSON and YAML formats
- Estimated effort: 2-3 days for comprehensive API documentation
- Regenerate docs in CI/CD to catch annotation errors early
