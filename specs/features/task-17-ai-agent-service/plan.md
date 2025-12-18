# Task 17: AI Agent Service - Implementation Plan

**Task ID**: task-17-ai-agent-service
**Phase**: 2B - New Microservices
**Priority**: P3 (Low - Advanced Feature)
**Estimated Duration**: 3-4 weeks
**Dependencies**: Task 16 (Analytics), External AI APIs

---

## 1. Technical Context

### Current State
- **AI Agent Service**: Not yet implemented (new service)
- **Analytics Service**: Complete with historical data aggregation
- **External APIs**: OpenAI API for NLU, AWS SageMaker/Azure ML for model hosting

### Technology Stack
- **Language**: Go 1.23.4 + Python 3.11 (for ML models)
- **Architecture**: Clean Architecture (Domain, Use Cases, Infrastructure)
- **Database**: PostgreSQL 16 for forecasts, recommendations, and model metadata
- **gRPC**: Protocol Buffers v3
- **ML Framework**: scikit-learn, pandas, numpy (Python microservice)
- **Model Serving**: gRPC bridge between Go service and Python ML service
- **External AI**: OpenAI API (GPT-4) for natural language processing
- **Testing**: testify, pytest

### Key Design Decisions
- **Hybrid Architecture**: Go main service + Python ML microservice
- **gRPC Communication**: Go ↔ Python for model inference
- **Model Versioning**: Track model versions and A/B test
- **Caching**: Cache predictions to reduce API costs
- **Batch Processing**: Process forecasts in batches for efficiency
- **Multi-tenancy**: organization_id filtering

---

## 2. Project Structure

### Files to Create

```
giia-core-engine/
├── services/ai-agent-service/           # Main Go service
│   ├── api/proto/ai_agent/v1/
│   │   ├── ai_agent.proto                [NEW]
│   │   ├── ai_agent.pb.go                [GENERATED]
│   │   └── ai_agent_grpc.pb.go           [GENERATED]
│   │
│   ├── internal/
│   │   ├── core/
│   │   │   ├── domain/
│   │   │   │   ├── seasonality_analysis.go       [NEW]
│   │   │   │   ├── cpd_adjustment_recommendation.go [NEW]
│   │   │   │   ├── new_product_forecast.go       [NEW]
│   │   │   │   ├── forecast.go                   [NEW]
│   │   │   │   ├── recommendation.go             [NEW]
│   │   │   │   └── anomaly.go                    [NEW]
│   │   │   │
│   │   │   ├── providers/
│   │   │   │   ├── ml_service_client.go          [NEW]
│   │   │   │   ├── openai_client.go              [NEW]
│   │   │   │   ├── analytics_client.go           [NEW]
│   │   │   │   └── ddmrp_client.go               [NEW]
│   │   │   │
│   │   │   └── usecases/
│   │   │       ├── seasonality/
│   │   │       │   ├── detect_seasonality.go     [NEW]
│   │   │       │   └── generate_fad_recommendation.go [NEW]
│   │   │       │
│   │   │       ├── cpd_adjustment/
│   │   │       │   ├── recommend_cpd_adjustment.go [NEW]
│   │   │       │   └── apply_recommendation.go    [NEW]
│   │   │       │
│   │   │       ├── new_product/
│   │   │       │   ├── estimate_cpd.go           [NEW]
│   │   │       │   └── find_similar_products.go  [NEW]
│   │   │       │
│   │   │       ├── forecast/
│   │   │       │   └── generate_forecast.go       [NEW]
│   │   │       │
│   │   │       └── anomaly/
│   │   │           └── detect_anomalies.go        [NEW]
│   │   │
│   │   └── infrastructure/
│   │       ├── repositories/
│   │       │   ├── seasonality_repository.go      [NEW]
│   │       │   ├── cpd_adjustment_repository.go   [NEW]
│   │       │   ├── forecast_repository.go         [NEW]
│   │       │   └── recommendation_repository.go   [NEW]
│   │       │
│   │       └── adapters/
│   │           ├── ml_grpc_client.go              [NEW]
│   │           ├── openai_client.go               [NEW]
│   │           └── analytics_grpc_client.go       [NEW]
│   │
│   ├── migrations/
│   │   ├── 000001_create_seasonality_analysis.up.sql  [NEW]
│   │   ├── 000002_create_cpd_adjustments.up.sql       [NEW]
│   │   ├── 000003_create_new_product_forecasts.up.sql [NEW]
│   │   └── 000004_create_forecasts.up.sql             [NEW]
│   │
│   └── cmd/main.go                        [NEW]
│
└── services/ml-service/                   # Python ML microservice
    ├── api/
    │   └── grpc/
    │       ├── ml_service.proto            [NEW]
    │       ├── ml_service_pb2.py           [GENERATED]
    │       ├── ml_service_pb2_grpc.py      [GENERATED]
    │       └── server.py                   [NEW]
    │
    ├── models/
    │   ├── seasonality_detector.py         [NEW]
    │   ├── demand_forecaster.py            [NEW]
    │   └── similarity_matcher.py           [NEW]
    │
    ├── requirements.txt                    [NEW]
    ├── Dockerfile                          [NEW]
    └── main.py                             [NEW]
```

---

## 3. Implementation Steps

### Phase 1: Database Schema & Domain Entities (Week 1)

#### Migrations

**File**: `services/ai-agent-service/migrations/000001_create_seasonality_analysis.up.sql`

```sql
-- Seasonality Analysis table
CREATE TABLE IF NOT EXISTS seasonality_analysis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    seasonal_pattern VARCHAR(20) NOT NULL,
    detected_at TIMESTAMP NOT NULL,
    confidence DECIMAL(3,2) NOT NULL,
    peak_months INTEGER[],
    low_months INTEGER[],
    peak_multiplier DECIMAL(5,2) NOT NULL,
    low_multiplier DECIMAL(5,2) NOT NULL,
    baseline_value DECIMAL(15,2) NOT NULL,
    seasonal_indices JSONB NOT NULL,
    year_over_year_growth DECIMAL(5,2),
    last_updated TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_seasonality_product UNIQUE (product_id, organization_id),
    CONSTRAINT chk_seasonal_pattern CHECK (seasonal_pattern IN ('monthly', 'quarterly', 'yearly', 'none')),
    CONSTRAINT chk_confidence CHECK (confidence >= 0 AND confidence <= 1)
);

CREATE INDEX idx_seasonality_product ON seasonality_analysis(product_id, organization_id);
CREATE INDEX idx_seasonality_pattern ON seasonality_analysis(seasonal_pattern);
```

**File**: `services/ai-agent-service/migrations/000002_create_cpd_adjustments.up.sql`

```sql
-- CPD Adjustment Recommendations table
CREATE TABLE IF NOT EXISTS cpd_adjustment_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    current_cpd DECIMAL(15,2) NOT NULL,
    recommended_cpd DECIMAL(15,2) NOT NULL,
    adjustment_type VARCHAR(30) NOT NULL,
    reasoning TEXT NOT NULL,
    effective_from DATE NOT NULL,
    effective_to DATE NOT NULL,
    fad_factor DECIMAL(5,2) NOT NULL,
    confidence DECIMAL(3,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_cpd_adjustment_type CHECK (adjustment_type IN (
        'seasonal', 'trend', 'new_product', 'discontinue', 'promotion', 'market_shift'
    )),
    CONSTRAINT chk_cpd_status CHECK (status IN ('pending', 'accepted', 'rejected', 'applied')),
    CONSTRAINT chk_cpd_confidence CHECK (confidence >= 0 AND confidence <= 1)
);

CREATE INDEX idx_cpd_adj_product ON cpd_adjustment_recommendations(product_id, organization_id);
CREATE INDEX idx_cpd_adj_status ON cpd_adjustment_recommendations(status);
CREATE INDEX idx_cpd_adj_dates ON cpd_adjustment_recommendations(effective_from, effective_to);
```

**File**: `services/ai-agent-service/migrations/000003_create_new_product_forecasts.up.sql`

```sql
-- New Product Forecasts table
CREATE TABLE IF NOT EXISTS new_product_forecasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    similar_products_ids UUID[],
    estimated_cpd DECIMAL(15,2) NOT NULL,
    confidence_level DECIMAL(3,2) NOT NULL,
    recommended_fad DECIMAL(5,2) NOT NULL,
    seasonality_factors JSONB,
    estimation_method VARCHAR(50) NOT NULL,
    assumptions TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_new_product_forecast UNIQUE (product_id, organization_id),
    CONSTRAINT chk_npf_confidence CHECK (confidence_level >= 0 AND confidence_level <= 1)
);

CREATE INDEX idx_npf_product ON new_product_forecasts(product_id, organization_id);

-- Daily Demand Estimates for New Products
CREATE TABLE IF NOT EXISTS daily_demand_estimates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    forecast_id UUID NOT NULL REFERENCES new_product_forecasts(id) ON DELETE CASCADE,
    estimate_date DATE NOT NULL,
    estimated_demand DECIMAL(15,2) NOT NULL,
    lower_bound DECIMAL(15,2) NOT NULL,
    upper_bound DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_daily_estimates_forecast ON daily_demand_estimates(forecast_id);
CREATE INDEX idx_daily_estimates_date ON daily_demand_estimates(estimate_date);
```

---

### Phase 2: Python ML Service (Week 1 Day 4 - Week 2)

#### ML Service gRPC Server

**File**: `services/ml-service/api/grpc/ml_service.proto`

```protobuf
syntax = "proto3";

package ml.v1;

option go_package = "giia-core-engine/services/ml-service/api/proto/ml/v1;mlv1";

service MLService {
  rpc DetectSeasonality(DetectSeasonalityRequest) returns (DetectSeasonalityResponse);
  rpc FindSimilarProducts(FindSimilarProductsRequest) returns (FindSimilarProductsResponse);
  rpc ForecastDemand(ForecastDemandRequest) returns (ForecastDemandResponse);
  rpc DetectAnomalies(DetectAnomaliesRequest) returns (DetectAnomaliesResponse);
}

message DetectSeasonalityRequest {
  string product_id = 1;
  repeated DemandDataPoint demand_history = 2;
}

message DemandDataPoint {
  string date = 1;
  double demand = 2;
}

message DetectSeasonalityResponse {
  string seasonal_pattern = 1;
  double confidence = 2;
  repeated int32 peak_months = 3;
  repeated int32 low_months = 4;
  double peak_multiplier = 5;
  double low_multiplier = 6;
  double baseline_value = 7;
  map<string, double> seasonal_indices = 8;
  double year_over_year_growth = 9;
}

// ... (other messages)
```

**File**: `services/ml-service/models/seasonality_detector.py`

```python
import numpy as np
import pandas as pd
from statsmodels.tsa.seasonal import seasonal_decompose
from scipy import stats

class SeasonalityDetector:
    def __init__(self):
        self.min_data_points = 24  # Minimum 2 years of monthly data

    def detect(self, demand_history: pd.DataFrame):
        """
        Detect seasonality pattern in demand history.

        Args:
            demand_history: DataFrame with 'date' and 'demand' columns

        Returns:
            dict with seasonality analysis
        """
        if len(demand_history) < self.min_data_points:
            return {
                'seasonal_pattern': 'none',
                'confidence': 0.0,
                'reason': 'insufficient data'
            }

        # Convert to time series
        ts = demand_history.set_index('date')['demand']
        ts.index = pd.to_datetime(ts.index)
        ts = ts.asfreq('MS')  # Monthly start frequency

        # Decompose time series
        try:
            decomposition = seasonal_decompose(ts, model='multiplicative', period=12)
        except:
            return {
                'seasonal_pattern': 'none',
                'confidence': 0.0,
                'reason': 'decomposition failed'
            }

        seasonal_component = decomposition.seasonal

        # Calculate seasonal indices
        seasonal_indices = {}
        for month in range(1, 13):
            month_name = pd.Timestamp(2000, month, 1).strftime('%b')
            month_data = seasonal_component[seasonal_component.index.month == month]
            seasonal_indices[month_name] = float(month_data.mean())

        # Detect peak and low months
        peak_months = []
        low_months = []

        for month, index in seasonal_indices.items():
            if index > 1.2:  # 20% above baseline
                peak_months.append(pd.to_datetime(month, format='%b').month)
            elif index < 0.8:  # 20% below baseline
                low_months.append(pd.to_datetime(month, format='%b').month)

        # Calculate multipliers
        peak_multiplier = max(seasonal_indices.values()) if seasonal_indices else 1.0
        low_multiplier = min(seasonal_indices.values()) if seasonal_indices else 1.0

        # Determine pattern type and confidence
        if len(peak_months) > 0:
            seasonal_pattern = 'monthly'

            # Calculate confidence based on variance explained
            variance_seasonal = seasonal_component.var()
            variance_total = ts.var()
            confidence = min(variance_seasonal / variance_total, 1.0)
        else:
            seasonal_pattern = 'none'
            confidence = 0.0

        # Calculate trend (year-over-year growth)
        trend_component = decomposition.trend.dropna()
        if len(trend_component) > 1:
            yoy_growth = (trend_component.iloc[-1] / trend_component.iloc[0]) - 1
        else:
            yoy_growth = 0.0

        return {
            'seasonal_pattern': seasonal_pattern,
            'confidence': float(confidence),
            'peak_months': peak_months,
            'low_months': low_months,
            'peak_multiplier': float(peak_multiplier),
            'low_multiplier': float(low_multiplier),
            'baseline_value': float(decomposition.trend.mean()),
            'seasonal_indices': seasonal_indices,
            'year_over_year_growth': float(yoy_growth)
        }
```

**File**: `services/ml-service/models/similarity_matcher.py`

```python
import numpy as np
import pandas as pd
from sklearn.metrics.pairwise import cosine_similarity

class SimilarityMatcher:
    def find_similar_products(self, new_product_features, all_products_features, top_k=5):
        """
        Find products similar to a new product based on features.

        Args:
            new_product_features: Feature vector for new product
            all_products_features: Feature vectors for all existing products
            top_k: Number of similar products to return

        Returns:
            list of similar product IDs with similarity scores
        """
        # Features: [category_encoded, price_range, supplier_reliability, lead_time_category]

        # Calculate cosine similarity
        similarities = cosine_similarity(
            new_product_features.reshape(1, -1),
            all_products_features
        )[0]

        # Get top K similar products
        top_indices = np.argsort(similarities)[-top_k:][::-1]

        similar_products = []
        for idx in top_indices:
            similar_products.append({
                'product_id': all_products_features.index[idx],
                'similarity_score': float(similarities[idx])
            })

        return similar_products
```

---

### Phase 3: Go Use Cases (Week 2)

#### Seasonality Detection Use Case

**File**: `services/ai-agent-service/internal/core/usecases/seasonality/detect_seasonality.go`

```go
package seasonality

import (
	"context"

	"github.com/google/uuid"
	"giia-core-engine/services/ai-agent-service/internal/core/domain"
	"giia-core-engine/services/ai-agent-service/internal/core/providers"
)

type DetectSeasonalityUseCase struct {
	seasonalityRepo  providers.SeasonalityRepository
	analyticsClient  providers.AnalyticsServiceClient
	mlClient         providers.MLServiceClient
}

func (uc *DetectSeasonalityUseCase) Execute(ctx context.Context, productID, orgID uuid.UUID) (*domain.SeasonalityAnalysis, error) {
	// 1. Get historical demand data from Analytics service
	demandHistory, err := uc.analyticsClient.GetDemandHistory(ctx, productID, orgID, 24) // 24 months
	if err != nil {
		return nil, err
	}

	// 2. Call ML service to detect seasonality
	mlResult, err := uc.mlClient.DetectSeasonality(ctx, productID.String(), demandHistory)
	if err != nil {
		return nil, err
	}

	// 3. Create SeasonalityAnalysis entity
	analysis := &domain.SeasonalityAnalysis{
		ID:                 uuid.New(),
		ProductID:          productID,
		OrganizationID:     orgID,
		SeasonalPattern:    domain.SeasonalPattern(mlResult.SeasonalPattern),
		Confidence:         mlResult.Confidence,
		PeakMonths:         mlResult.PeakMonths,
		LowMonths:          mlResult.LowMonths,
		PeakMultiplier:     mlResult.PeakMultiplier,
		LowMultiplier:      mlResult.LowMultiplier,
		BaselineValue:      mlResult.BaselineValue,
		SeasonalIndices:    mlResult.SeasonalIndices,
		YearOverYearGrowth: mlResult.YearOverYearGrowth,
	}

	// 4. Save analysis
	if err := uc.seasonalityRepo.Save(ctx, analysis); err != nil {
		return nil, err
	}

	return analysis, nil
}
```

#### CPD Adjustment Recommendation

**File**: `services/ai-agent-service/internal/core/usecases/cpd_adjustment/recommend_cpd_adjustment.go`

```go
package cpd_adjustment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"giia-core-engine/services/ai-agent-service/internal/core/domain"
	"giia-core-engine/services/ai-agent-service/internal/core/providers"
)

type RecommendCPDAdjustmentUseCase struct {
	cpdAdjRepo       providers.CPDAdjustmentRepository
	seasonalityRepo  providers.SeasonalityRepository
	ddmrpClient      providers.DDMRPServiceClient
	openAIClient     providers.OpenAIClient
}

func (uc *RecommendCPDAdjustmentUseCase) Execute(ctx context.Context, productID, orgID uuid.UUID) (*domain.CPDAdjustmentRecommendation, error) {
	// 1. Get current buffer and CPD
	buffer, err := uc.ddmrpClient.GetBuffer(ctx, productID, orgID)
	if err != nil {
		return nil, err
	}

	currentCPD := buffer.CPD

	// 2. Get seasonality analysis
	seasonality, err := uc.seasonalityRepo.GetByProduct(ctx, productID, orgID)
	if err != nil || seasonality.SeasonalPattern == domain.SeasonalPatternNone {
		return nil, domain.NewValidationError("no seasonality detected for product")
	}

	// 3. Determine current month and calculate recommended CPD
	currentMonth := time.Now().Month()
	seasonalIndex := seasonality.SeasonalIndices[currentMonth.String()]

	recommendedCPD := seasonality.BaselineValue * seasonalIndex
	fadFactor := recommendedCPD / currentCPD

	// 4. Generate AI reasoning using OpenAI
	reasoning, err := uc.openAIClient.GenerateCPDAdjustmentReasoning(ctx, productID.String(), currentCPD, recommendedCPD, seasonality)
	if err != nil {
		reasoning = fmt.Sprintf("Seasonal adjustment based on detected %s pattern", seasonality.SeasonalPattern)
	}

	// 5. Determine effective date range (current month)
	effectiveFrom := time.Date(time.Now().Year(), currentMonth, 1, 0, 0, 0, 0, time.UTC)
	effectiveTo := effectiveFrom.AddDate(0, 1, -1)

	// 6. Create recommendation
	recommendation := &domain.CPDAdjustmentRecommendation{
		ID:             uuid.New(),
		ProductID:      productID,
		OrganizationID: orgID,
		CurrentCPD:     currentCPD,
		RecommendedCPD: recommendedCPD,
		AdjustmentType: domain.CPDAdjustmentSeasonal,
		Reasoning:      reasoning,
		EffectiveFrom:  effectiveFrom,
		EffectiveTo:    effectiveTo,
		FADFactor:      fadFactor,
		Confidence:     seasonality.Confidence,
		Status:         domain.RecommendationStatusPending,
		CreatedAt:      time.Now(),
	}

	// 7. Save recommendation
	if err := uc.cpdAdjRepo.Save(ctx, recommendation); err != nil {
		return nil, err
	}

	return recommendation, nil
}
```

#### New Product CPD Estimation

**File**: `services/ai-agent-service/internal/core/usecases/new_product/estimate_cpd.go`

```go
package new_product

import (
	"context"

	"github.com/google/uuid"
	"giia-core-engine/services/ai-agent-service/internal/core/domain"
	"giia-core-engine/services/ai-agent-service/internal/core/providers"
)

type EstimateCPDUseCase struct {
	forecastRepo     providers.ForecastRepository
	catalogClient    providers.CatalogServiceClient
	analyticsClient  providers.AnalyticsServiceClient
	mlClient         providers.MLServiceClient
}

func (uc *EstimateCPDUseCase) Execute(ctx context.Context, productID, orgID uuid.UUID) (*domain.NewProductForecast, error) {
	// 1. Get new product details
	product, err := uc.catalogClient.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	// 2. Find similar products using ML service
	productFeatures := extractProductFeatures(product)
	similarProducts, err := uc.mlClient.FindSimilarProducts(ctx, productFeatures)
	if err != nil {
		return nil, err
	}

	// 3. Get CPD of similar products
	similarCPDs := []float64{}
	for _, simProduct := range similarProducts {
		buffer, err := uc.catalogClient.GetBuffer(ctx, simProduct.ProductID, orgID)
		if err == nil {
			similarCPDs = append(similarCPDs, buffer.CPD)
		}
	}

	// 4. Calculate estimated CPD (average of similar products)
	estimatedCPD := calculateAverage(similarCPDs)

	// 5. Get seasonality from similar products
	seasonalityFactors := uc.estimateSeasonalityFactors(ctx, similarProducts, orgID)

	// 6. Create forecast
	forecast := &domain.NewProductForecast{
		ID:                 uuid.New(),
		ProductID:          productID,
		OrganizationID:     orgID,
		SimilarProductsIDs: extractProductIDs(similarProducts),
		EstimatedCPD:       estimatedCPD,
		ConfidenceLevel:    calculateConfidence(similarProducts),
		RecommendedFAD:     1.0, // Conservative initial FAD
		SeasonalityFactors: seasonalityFactors,
		EstimationMethod:   "similar_products",
		Assumptions:        generateAssumptions(product, similarProducts),
		CreatedAt:          time.Now(),
	}

	// 7. Save forecast
	if err := uc.forecastRepo.SaveNewProductForecast(ctx, forecast); err != nil {
		return nil, err
	}

	return forecast, nil
}
```

---

### Phase 4: Testing & Integration (Week 3-4)

#### Unit Tests

**File**: `services/ai-agent-service/internal/core/usecases/seasonality/detect_seasonality_test.go`

```go
package seasonality_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"giia-core-engine/services/ai-agent-service/internal/core/domain"
	"giia-core-engine/services/ai-agent-service/internal/core/usecases/seasonality"
)

func TestDetectSeasonality_Success_MonthlyPattern(t *testing.T) {
	// Given
	mockSeasonalityRepo := new(MockSeasonalityRepository)
	mockAnalyticsClient := new(MockAnalyticsClient)
	mockMLClient := new(MockMLClient)

	useCase := seasonality.NewDetectSeasonalityUseCase(
		mockSeasonalityRepo,
		mockAnalyticsClient,
		mockMLClient,
	)

	givenProductID := uuid.New()
	givenOrgID := uuid.New()
	givenDemandHistory := []DemandDataPoint{
		{Date: "2023-01", Demand: 100},
		{Date: "2023-02", Demand: 95},
		// ... (24 months of data)
		{Date: "2024-12", Demand: 150}, // December peak
	}

	mockAnalyticsClient.On("GetDemandHistory", mock.Anything, givenProductID, givenOrgID, 24).
		Return(givenDemandHistory, nil)

	mockMLClient.On("DetectSeasonality", mock.Anything, givenProductID.String(), givenDemandHistory).
		Return(&MLSeasonalityResult{
			SeasonalPattern:    "monthly",
			Confidence:         0.85,
			PeakMonths:         []int{11, 12},
			LowMonths:          []int{2, 3},
			PeakMultiplier:     1.5,
			LowMultiplier:      0.7,
			BaselineValue:      100,
			YearOverYearGrowth: 0.1,
		}, nil)

	mockSeasonalityRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.SeasonalityAnalysis")).
		Return(nil)

	// When
	result, err := useCase.Execute(context.Background(), givenProductID, givenOrgID)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.SeasonalPatternMonthly, result.SeasonalPattern)
	assert.Equal(t, 0.85, result.Confidence)
	assert.Equal(t, []int{11, 12}, result.PeakMonths)
	mockSeasonalityRepo.AssertExpectations(t)
}
```

---

## 4. Success Criteria

### Mandatory
- ✅ Seasonality detection for products
- ✅ CPD adjustment recommendations based on seasonality
- ✅ New product CPD estimation
- ✅ Python ML microservice with gRPC
- ✅ Integration with Analytics service for historical data
- ✅ Integration with DDMRP Engine for CPD application
- ✅ OpenAI API integration for reasoning generation
- ✅ 80%+ test coverage (Go) + 80%+ (Python)
- ✅ Multi-tenancy support

### Optional (Nice to Have)
- ⚪ Deep learning models (LSTM, Transformers)
- ⚪ Real-time model retraining
- ⚪ Chatbot interface with NLU

---

## 5. Dependencies

- **Task 16**: Analytics service (for historical data)
- **Task 14**: DDMRP Engine (for applying recommendations)
- **External**: OpenAI API, scikit-learn, pandas, numpy
- **Shared packages**: pkg/events, pkg/database, pkg/logger

---

## 6. Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Model accuracy insufficient | Multiple models, ensemble methods, human review |
| External AI API costs | Caching, batching, cost monitoring |
| Python-Go integration complexity | Well-defined gRPC contracts, integration tests |
| Cold start (new products) | Conservative estimates, wide confidence intervals |

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Implementation
