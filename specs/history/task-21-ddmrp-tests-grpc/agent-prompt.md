# Agent Prompt: Task 21 - DDMRP Engine Unit Tests & gRPC Registration

## 🤖 Agent Identity

You are an **Expert Go Test Engineer** specialized in writing comprehensive unit tests and configuring gRPC services. You have deep expertise in:
- Go (Golang) 1.21+ testing patterns
- Table-driven tests and mocking
- gRPC service registration and Protocol Buffers
- GORM repository testing with sqlmock
- Achieving 85%+ code coverage

---

## 📋 Mission

Complete the **DDMRP Engine Service** by:
1. Writing comprehensive unit tests for all repositories (5 files)
2. Writing unit tests for remaining use cases (NFP and FAD)
3. Writing unit tests for all gRPC handlers (3 handlers)
4. Registering all gRPC handlers in main.go

---

## 🏗️ Project Context

### Current Service State (65% Complete)
```
services/ddmrp-engine-service/
├── cmd/api/main.go                    # ⚠️ Needs gRPC handler registration
├── internal/
│   ├── domain/
│   │   ├── entities/                  # ✅ Complete
│   │   └── repositories/              # ✅ Interfaces defined
│   ├── usecases/
│   │   ├── calculate_buffer.go        # ✅ Has tests
│   │   ├── calculate_buffer_test.go   # ✅ Exists
│   │   ├── get_buffer.go              # ⚠️ Needs tests
│   │   ├── list_buffers.go            # ⚠️ Needs tests
│   │   ├── create_fad.go              # ⚠️ Needs tests
│   │   ├── update_fad.go              # ⚠️ Needs tests
│   │   ├── delete_fad.go              # ⚠️ Needs tests
│   │   ├── list_fads.go               # ⚠️ Needs tests
│   │   ├── update_nfp.go              # ⚠️ Needs tests
│   │   └── check_replenishment.go     # ⚠️ Needs tests
│   ├── handlers/grpc/
│   │   ├── buffer_handler.go          # ⚠️ Needs tests + registration
│   │   ├── fad_handler.go             # ⚠️ Needs tests + registration
│   │   └── nfp_handler.go             # ⚠️ Needs tests + registration
│   └── repository/postgres/
│       ├── buffer_repository.go       # ⚠️ Needs tests
│       ├── adjustment_repository.go   # ⚠️ Needs tests
│       ├── history_repository.go      # ⚠️ Needs tests
│       ├── adu_repository.go          # ⚠️ Needs tests
│       └── demand_repository.go       # ⚠️ Needs tests
└── go.mod
```

---

## 📂 Files to Create

### 1. Repository Tests
```
internal/repository/postgres/
├── buffer_repository_test.go
├── adjustment_repository_test.go
├── history_repository_test.go
├── adu_repository_test.go
└── demand_repository_test.go
```

### 2. Use Case Tests
```
internal/usecases/
├── get_buffer_test.go
├── list_buffers_test.go
├── create_fad_test.go
├── update_fad_test.go
├── delete_fad_test.go
├── list_fads_test.go
├── update_nfp_test.go
└── check_replenishment_test.go
```

### 3. Handler Tests
```
internal/handlers/grpc/
├── buffer_handler_test.go
├── fad_handler_test.go
└── nfp_handler_test.go
```

---

## 🔧 Implementation Requirements

### Repository Tests Pattern
```go
package postgres_test

import (
    "context"
    "testing"
    
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    
    dialector := postgres.New(postgres.Config{
        Conn:       db,
        DriverName: "postgres",
    })
    
    gormDB, err := gorm.Open(dialector, &gorm.Config{})
    require.NoError(t, err)
    
    return gormDB, mock
}

func TestBufferRepository_Create(t *testing.T) {
    tests := []struct {
        name    string
        buffer  *entities.Buffer
        mockFn  func(mock sqlmock.Sqlmock)
        wantErr bool
    }{
        {
            name: "successful creation",
            buffer: &entities.Buffer{
                ID:             uuid.New(),
                ProductID:      uuid.New(),
                OrganizationID: uuid.New(),
            },
            mockFn: func(mock sqlmock.Sqlmock) {
                mock.ExpectBegin()
                mock.ExpectExec("INSERT INTO").WillReturnResult(sqlmock.NewResult(1, 1))
                mock.ExpectCommit()
            },
            wantErr: false,
        },
        {
            name: "database error",
            buffer: &entities.Buffer{},
            mockFn: func(mock sqlmock.Sqlmock) {
                mock.ExpectBegin()
                mock.ExpectExec("INSERT INTO").WillReturnError(errors.New("db error"))
                mock.ExpectRollback()
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db, mock := setupTestDB(t)
            tt.mockFn(mock)
            
            repo := NewBufferRepository(db)
            err := repo.Create(context.Background(), tt.buffer)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
            assert.NoError(t, mock.ExpectationsWereMet())
        })
    }
}
```

### Use Case Tests Pattern
```go
package usecases_test

import (
    "context"
    "testing"
    
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockBufferRepository struct {
    mock.Mock
}

func (m *MockBufferRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Buffer, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entities.Buffer), args.Error(1)
}

func TestGetBuffer_Execute(t *testing.T) {
    tests := []struct {
        name      string
        bufferID  uuid.UUID
        setupMock func(m *MockBufferRepository)
        wantErr   bool
    }{
        {
            name:     "buffer found",
            bufferID: uuid.New(),
            setupMock: func(m *MockBufferRepository) {
                m.On("GetByID", mock.Anything, mock.Anything).Return(&entities.Buffer{
                    ID: uuid.New(),
                }, nil)
            },
            wantErr: false,
        },
        {
            name:     "buffer not found",
            bufferID: uuid.New(),
            setupMock: func(m *MockBufferRepository) {
                m.On("GetByID", mock.Anything, mock.Anything).Return(nil, ErrNotFound)
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockBufferRepository)
            tt.setupMock(mockRepo)
            
            uc := NewGetBufferUseCase(mockRepo)
            result, err := uc.Execute(context.Background(), tt.bufferID)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
            }
            mockRepo.AssertExpectations(t)
        })
    }
}
```

### gRPC Handler Registration in main.go
```go
// Register handlers
bufferHandler := grpchandlers.NewBufferHandler(calculateBufferUC, getBufferUC, listBuffersUC)
pb.RegisterBufferServiceServer(grpcServer, bufferHandler)

fadHandler := grpchandlers.NewFADHandler(createFADUC, updateFADUC, deleteFADUC, listFADsUC)
pb.RegisterFADServiceServer(grpcServer, fadHandler)

nfpHandler := grpchandlers.NewNFPHandler(updateNFPUC, checkReplenishmentUC)
pb.RegisterNFPServiceServer(grpcServer, nfpHandler)

// Enable reflection for grpcurl
reflection.Register(grpcServer)
```

---

## 📊 Test Coverage Requirements

| Package | Current | Target |
|---------|---------|--------|
| repository/postgres | ~10% | 85% |
| usecases | ~50% | 85% |
| handlers/grpc | ~0% | 85% |

---

## ✅ Success Criteria

- [ ] All 5 repository files have test files
- [ ] All NFP use case tests (UpdateNFP, CheckReplenishment)
- [ ] All FAD use case tests (Create, Update, Delete, List)
- [ ] All gRPC handler tests with mocked use cases
- [ ] gRPC handlers registered in main.go
- [ ] `grpcurl` can call all endpoints
- [ ] Overall coverage 85%+

---

## 🚀 Commands
```bash
cd services/ddmrp-engine-service
go get github.com/DATA-DOG/go-sqlmock
go get github.com/stretchr/testify
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
grpcurl -plaintext localhost:50053 list
```
