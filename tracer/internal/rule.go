package internal

// Rule represents a sampling rule which contains rule properties and reservoir which keeps tracks of sampling statistics of a rule.
type Rule struct {
	samplingStatistics *samplingStatistics

	// reservoir has equivalent fields to store what we receive from service API getSamplingTargets.
	// https://docs.aws.amazon.com/xray/latest/api/API_GetSamplingTargets.html
	reservoir *reservoir

	// ruleProperty is equivalent to what we receive from service API getSamplingRules.
	// https://docs.aws.amazon.com/cli/latest/reference/xray/get-sampling-rules.html
	ruleProperties ruleProperties
}

type samplingStatistics struct {
	// matchedRequests is the number of requests matched against specific rule.
	matchedRequests int64

	// sampledRequests is the number of requests sampled using specific rule.
	sampledRequests int64

	// borrowedRequests is the number of requests borrowed using specific rule.
	borrowedRequests int64
}
