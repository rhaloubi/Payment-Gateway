package client

import (
	"context"
	"math/rand"
	"time"

	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"go.uber.org/zap"
)

// FraudClient communicates with Fraud Detection Service
// TODO: Replace with actual gRPC client when fraud service is built
type FraudClient struct {
	enabled bool
}

func NewFraudClient() *FraudClient {
	return &FraudClient{
		enabled: true,
	}
}

// FraudCheckRequest represents fraud check request
type FraudCheckRequest struct {
	MerchantID        string
	Amount            int64
	Currency          string
	CardToken         string
	CardBrand         string
	CardLast4         string
	CustomerEmail     string
	CustomerIP        string
	DeviceFingerprint string
}

// FraudCheckResponse represents fraud check result
type FraudCheckResponse struct {
	RiskScore      int    // 0-100
	Decision       string // "approve", "review", "decline"
	RulesTriggered []string
	Reason         string
}

// CheckFraud performs fraud analysis
func (c *FraudClient) CheckFraud(ctx context.Context, req *FraudCheckRequest) (*FraudCheckResponse, error) {
	logger.Log.Info("Running fraud check (mock)",
		zap.String("merchant_id", req.MerchantID),
		zap.Int64("amount", req.Amount),
		zap.String("card_last4", req.CardLast4),
	)

	// Simulate fraud check processing time
	time.Sleep(50 * time.Millisecond)

	// Mock fraud scoring logic
	riskScore := calculateMockRiskScore(req)
	decision := determineDecision(riskScore)
	rulesTriggered := []string{}

	// Add rules based on risk factors
	if req.Amount > 100000 { // > $1000
		rulesTriggered = append(rulesTriggered, "high_amount")
		riskScore += 10
	}

	if riskScore > 70 {
		rulesTriggered = append(rulesTriggered, "high_risk_score")
	}

	response := &FraudCheckResponse{
		RiskScore:      riskScore,
		Decision:       decision,
		RulesTriggered: rulesTriggered,
		Reason:         getDecisionReason(decision, riskScore),
	}

	logger.Log.Info("Fraud check completed",
		zap.Int("risk_score", riskScore),
		zap.String("decision", decision),
	)

	return response, nil
}

// calculateMockRiskScore generates a realistic risk score
func calculateMockRiskScore(req *FraudCheckRequest) int {
	rand.Seed(time.Now().UnixNano())

	// Base risk: 10-30 (most transactions are low risk)
	baseRisk := rand.Intn(21) + 10

	// Amount-based risk
	if req.Amount > 500000 { // > $5000
		baseRisk += 20
	} else if req.Amount > 100000 { // > $1000
		baseRisk += 10
	}

	// Random high-risk scenario (5% chance)
	if rand.Float64() < 0.05 {
		baseRisk += 50
	}

	// Cap at 100
	if baseRisk > 100 {
		baseRisk = 100
	}

	return baseRisk
}

// determineDecision maps risk score to decision
func determineDecision(riskScore int) string {
	if riskScore < 30 {
		return "approve"
	} else if riskScore < 70 {
		return "review"
	} else {
		return "decline"
	}
}

// getDecisionReason provides human-readable reason
func getDecisionReason(decision string, score int) string {
	switch decision {
	case "approve":
		return "Transaction approved - low risk"
	case "review":
		return "Transaction requires manual review - medium risk"
	case "decline":
		return "Transaction declined - high risk indicators detected"
	default:
		return "Unknown decision"
	}
}

// Close closes the client connection (no-op for mock)
func (c *FraudClient) Close() error {
	return nil
}
