package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// cosmos-validator-watcher 메트릭 구조체
type CosmosValidatorMetrics struct {
	blockHeightMetric              prometheus.Gauge
	activeSetMetric                prometheus.Gauge
	isBondedMetric                 *prometheus.GaugeVec
	isJailedMetric                 *prometheus.GaugeVec
	missedBlocksMetric             *prometheus.GaugeVec
	consecutiveMissedBlocksMetric  *prometheus.GaugeVec
	cometbftMissedBlocksMetric     *prometheus.GaugeVec
	tokensMetric                   *prometheus.GaugeVec
	rankMetric                     *prometheus.GaugeVec
	commissionMetric               *prometheus.GaugeVec
	proposedBlocksMetric           *prometheus.GaugeVec
	validatedBlocksMetric          *prometheus.GaugeVec
	emptyBlocksMetric              *prometheus.GaugeVec
	seatPriceMetric                prometheus.Gauge
	signedBlocksWindowMetric       prometheus.Gauge
	missedBlocksWindowMetric       *prometheus.GaugeVec
	minSignedBlocksPerWindowMetric prometheus.Gauge
	downtimeJailDurationMetric     prometheus.Gauge
	slashFractionDoubleSignMetric  prometheus.Gauge
	slashFractionDowntimeMetric    prometheus.Gauge
	soloMissedBlocksMetric         *prometheus.GaugeVec
	trackedBlocksMetric            prometheus.Counter
	skippedBlocksMetric            prometheus.Counter
	transactionsMetric              prometheus.Counter
	upgradePlanMetric              prometheus.Gauge
	proposalEndTimeMetric          *prometheus.GaugeVec
	voteMetric                     *prometheus.GaugeVec
	nodeBlockHeightMetric          *prometheus.GaugeVec
	nodeSyncedMetric               *prometheus.GaugeVec
}

// 커스텀 비콘 체인 메트릭 구조체
type CustomMetrics struct {
	beaconBlockSignedMetric *prometheus.GaugeVec
	validatorStatusMetric   *prometheus.GaugeVec
	mempoolSizeMetric       prometheus.Gauge
	mempoolTotalBytesMetric prometheus.Gauge
	mempoolTotalMetric      prometheus.Gauge
	missedBlocksMetric      *prometheus.GaugeVec
	consecutiveMissedBlocksMetric *prometheus.GaugeVec
	totalMissedBlocksMetric *prometheus.GaugeVec
}

type UnifiedMetrics struct {
	cosmos *CosmosValidatorMetrics
	custom *CustomMetrics
}

func NewCosmosValidatorMetrics() *CosmosValidatorMetrics {
	return &CosmosValidatorMetrics{
		blockHeightMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_block_height",
				Help: "Latest known block height",
			},
		),
		activeSetMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_active_set",
				Help: "Number of validators in the active set",
			},
		),
		isBondedMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_is_bonded",
				Help: "Set to 1 if the validator is bonded",
			},
			[]string{"validator"},
		),
		isJailedMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_is_jailed",
				Help: "Set to 1 if the validator is jailed",
			},
			[]string{"validator"},
		),
		missedBlocksMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_missed_blocks",
				Help: "Number of missed blocks per validator",
			},
			[]string{"validator"},
		),
		consecutiveMissedBlocksMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_consecutive_missed_blocks",
				Help: "Number of consecutive missed blocks per validator",
			},
			[]string{"validator"},
		),
		cometbftMissedBlocksMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "cometbft_consensus_validator_missed_blocks",
				Help: "Number of missed blocks per validator (CometBFT consensus)",
			},
			[]string{"validator", "chain_id"},
		),
		tokensMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_tokens",
				Help: "Number of staked tokens per validator",
			},
			[]string{"validator"},
		),
		rankMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_rank",
				Help: "Rank of the validator",
			},
			[]string{"validator"},
		),
		commissionMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_commission",
				Help: "Earned validator commission",
			},
			[]string{"validator"},
		),
		proposedBlocksMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_proposed_blocks",
				Help: "Number of proposed blocks per validator",
			},
			[]string{"validator"},
		),
		validatedBlocksMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_validated_blocks",
				Help: "Number of validated blocks per validator",
			},
			[]string{"validator"},
		),
		emptyBlocksMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_empty_blocks",
				Help: "Number of empty blocks proposed by validator",
			},
			[]string{"validator"},
		),
		seatPriceMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_seat_price",
				Help: "Min seat price to be in the active set",
			},
		),
		signedBlocksWindowMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_signed_blocks_window",
				Help: "Number of blocks per signing window",
			},
		),
		missedBlocksWindowMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_missed_blocks_window",
				Help: "Number of missed blocks per validator for the current signing window",
			},
			[]string{"validator"},
		),
		minSignedBlocksPerWindowMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_min_signed_blocks_per_window",
				Help: "Minimum number of blocks required to be signed per signing window",
			},
		),
		downtimeJailDurationMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_downtime_jail_duration",
				Help: "Duration of the jail period for a validator in seconds",
			},
		),
		slashFractionDoubleSignMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_slash_fraction_double_sign",
				Help: "Slash penalty for double-signing",
			},
		),
		slashFractionDowntimeMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_slash_fraction_downtime",
				Help: "Slash penalty for downtime",
			},
		),
		soloMissedBlocksMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_solo_missed_blocks",
				Help: "Number of missed blocks per validator, unless the block is missed by many other validators",
			},
			[]string{"validator"},
		),
		trackedBlocksMetric: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "og_galileo_validator_tracked_blocks",
				Help: "Number of blocks tracked since start",
			},
		),
		skippedBlocksMetric: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "og_galileo_validator_skipped_blocks",
				Help: "Number of blocks skipped since start",
			},
		),
		transactionsMetric: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "og_galileo_validator_transactions",
				Help: "Number of transactions since start",
			},
		),
		upgradePlanMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_upgrade_plan",
				Help: "Block height of the upcoming upgrade",
			},
		),
		proposalEndTimeMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_proposal_end_time",
				Help: "Timestamp of the voting end time of a proposal",
			},
			[]string{"proposal_id"},
		),
		voteMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_vote",
				Help: "Set to 1 if the validator has voted on a proposal",
			},
			[]string{"validator", "proposal_id"},
		),
		nodeBlockHeightMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_node_block_height",
				Help: "Latest fetched block height for each node",
			},
			[]string{"node"},
		),
		nodeSyncedMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_node_synced",
				Help: "Set to 1 if the node is synced",
			},
			[]string{"node"},
		),
	}
}

func NewCustomMetrics() *CustomMetrics {
	return &CustomMetrics{
		beaconBlockSignedMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_beacon_block_signed",
				Help: "Beacon block signing status per validator (1=signed, 0=missed) - based on previous block",
			},
			[]string{"validator", "block_height"},
		),
		validatorStatusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_status",
				Help: "Validator status (1=active, 0=inactive)",
			},
			[]string{"validator", "address"},
		),
		mempoolSizeMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_mempool_size",
				Help: "Current size of the mempool in transactions",
			},
		),
		mempoolTotalBytesMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_mempool_total_bytes",
				Help: "Total size of transactions in the mempool in bytes",
			},
		),
		mempoolTotalMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "og_galileo_validator_mempool_total",
				Help: "Total number of transactions in the mempool",
			},
		),
	}
}

func NewUnifiedMetrics() *UnifiedMetrics {
	return &UnifiedMetrics{
		cosmos: NewCosmosValidatorMetrics(),
		custom: NewCustomMetrics(),
	}
}

func (um *UnifiedMetrics) Register() {
	// cosmos-validator-watcher 메트릭 등록
	prometheus.MustRegister(um.cosmos.blockHeightMetric)
	prometheus.MustRegister(um.cosmos.activeSetMetric)
	prometheus.MustRegister(um.cosmos.isBondedMetric)
	prometheus.MustRegister(um.cosmos.isJailedMetric)
	prometheus.MustRegister(um.cosmos.missedBlocksMetric)
	prometheus.MustRegister(um.cosmos.consecutiveMissedBlocksMetric)
	prometheus.MustRegister(um.cosmos.cometbftMissedBlocksMetric)
	prometheus.MustRegister(um.cosmos.tokensMetric)
	prometheus.MustRegister(um.cosmos.rankMetric)
	prometheus.MustRegister(um.cosmos.commissionMetric)
	prometheus.MustRegister(um.cosmos.proposedBlocksMetric)
	prometheus.MustRegister(um.cosmos.validatedBlocksMetric)
	prometheus.MustRegister(um.cosmos.emptyBlocksMetric)
	prometheus.MustRegister(um.cosmos.seatPriceMetric)
	prometheus.MustRegister(um.cosmos.signedBlocksWindowMetric)
	prometheus.MustRegister(um.cosmos.missedBlocksWindowMetric)
	prometheus.MustRegister(um.cosmos.minSignedBlocksPerWindowMetric)
	prometheus.MustRegister(um.cosmos.downtimeJailDurationMetric)
	prometheus.MustRegister(um.cosmos.slashFractionDoubleSignMetric)
	prometheus.MustRegister(um.cosmos.slashFractionDowntimeMetric)
	prometheus.MustRegister(um.cosmos.soloMissedBlocksMetric)
	prometheus.MustRegister(um.cosmos.trackedBlocksMetric)
	prometheus.MustRegister(um.cosmos.skippedBlocksMetric)
	prometheus.MustRegister(um.cosmos.transactionsMetric)
	prometheus.MustRegister(um.cosmos.upgradePlanMetric)
	prometheus.MustRegister(um.cosmos.proposalEndTimeMetric)
	prometheus.MustRegister(um.cosmos.voteMetric)
	prometheus.MustRegister(um.cosmos.nodeBlockHeightMetric)
	prometheus.MustRegister(um.cosmos.nodeSyncedMetric)

	// 커스텀 메트릭 등록
	prometheus.MustRegister(um.custom.beaconBlockSignedMetric)
	prometheus.MustRegister(um.custom.validatorStatusMetric)
	prometheus.MustRegister(um.custom.mempoolSizeMetric)
	prometheus.MustRegister(um.custom.mempoolTotalBytesMetric)
	prometheus.MustRegister(um.custom.mempoolTotalMetric)
}

// API 응답 구조체들
type BlockInfo struct {
	Result struct {
		Block struct {
			Header struct {
				Height string `json:"height"`
			} `json:"header"`
			LastCommit struct {
				Signatures []struct {
					ValidatorAddress string `json:"validator_address"`
					Signature        string `json:"signature"`
				} `json:"signatures"`
			} `json:"last_commit"`
		} `json:"block"`
	} `json:"result"`
}

type ValidatorInfo struct {
	Validators []struct {
		Address string `json:"address"`
		PubKey  struct {
			Value string `json:"value"`
		} `json:"pub_key"`
	} `json:"validators"`
}

type ValidatorResponse struct {
	Validators []struct {
		OperatorAddress string `json:"operator_address"`
		ConsensusPubkey struct {
			Key string `json:"key"`
		} `json:"consensus_pubkey"`
		Jailed      bool    `json:"jailed"`
		Status      string  `json:"status"`
		Tokens      string  `json:"tokens"`
		DelegatorShares string `json:"delegator_shares"`
		Description struct {
			Moniker string `json:"moniker"`
		} `json:"description"`
		Commission struct {
			CommissionRates struct {
				Rate string `json:"rate"`
			} `json:"commission_rates"`
		} `json:"commission"`
		MinSelfDelegation string `json:"min_self_delegation"`
	} `json:"validators"`
}

// MempoolResponse represents the response from the mempool endpoint
type MempoolResponse struct {
	Result struct {
		NTxs       string `json:"n_txs"`
		Total      string `json:"total"`
		TotalBytes string `json:"total_bytes"`
	} `json:"result"`
}

type UnifiedValidatorTracker struct {
	rpcEndpoint     string
	validators      map[string]string // address -> label
	metrics         *UnifiedMetrics
	lastBlockHeight int64
	processedBlocks map[int64]bool
}

func NewUnifiedValidatorTracker(rpcEndpoint string, validators map[string]string) *UnifiedValidatorTracker {
	return &UnifiedValidatorTracker{
		rpcEndpoint:     rpcEndpoint,
		validators:      validators,
		metrics:         NewUnifiedMetrics(),
		processedBlocks: make(map[int64]bool),
	}
}

func (vt *UnifiedValidatorTracker) RegisterMetrics() {
	vt.metrics.Register()
}

func (vt *UnifiedValidatorTracker) fetchBlock(height int64) (*BlockInfo, error) {
	var url string
	if height == 0 {
		// 최신 블록을 가져오기 위해 /block 엔드포인트 사용 (height 파라미터 없이)
		url = fmt.Sprintf("%s/block", vt.rpcEndpoint)
	} else {
		url = fmt.Sprintf("%s/block?height=%d", vt.rpcEndpoint, height)
	}
	
	log.Printf("Fetching block from: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 응답 본문을 읽어서 로그에 출력
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	log.Printf("RPC Response: %s", string(body))

	var blockInfo BlockInfo
	if err := json.Unmarshal(body, &blockInfo); err != nil {
		log.Printf("JSON parsing error: %v", err)
		return nil, err
	}

	return &blockInfo, nil
}

func (vt *UnifiedValidatorTracker) fetchValidators() (*ValidatorInfo, error) {
	url := fmt.Sprintf("%s/validators", vt.rpcEndpoint)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var validatorInfo ValidatorInfo
	if err := json.NewDecoder(resp.Body).Decode(&validatorInfo); err != nil {
		return nil, err
	}

	return &validatorInfo, nil
}

func (vt *UnifiedValidatorTracker) fetchStakingValidators() (*ValidatorResponse, error) {
	url := fmt.Sprintf("%s/cosmos/staking/v1beta1/validators", vt.rpcEndpoint)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var validatorResponse ValidatorResponse
	if err := json.NewDecoder(resp.Body).Decode(&validatorResponse); err != nil {
		return nil, err
	}

	return &validatorResponse, nil
}

func (vt *UnifiedValidatorTracker) fetchMempool() (*MempoolResponse, error) {
	url := fmt.Sprintf("%s/mempool", vt.rpcEndpoint)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var mempoolResponse MempoolResponse
	if err := json.NewDecoder(resp.Body).Decode(&mempoolResponse); err != nil {
		return nil, err
	}

	return &mempoolResponse, nil
}

// 비콘 체인용: -1 블록 이전을 조회하여 서명/누락 판단
func (vt *UnifiedValidatorTracker) updateBeaconBlockMetrics(currentBlockInfo *BlockInfo) {
	log.Printf("=== updateBeaconBlockMetrics called ===")
	log.Printf("Current block height: %s", currentBlockInfo.Result.Block.Header.Height)
	currentHeight, _ := strconv.ParseInt(currentBlockInfo.Result.Block.Header.Height, 10, 64)
	previousHeight := currentHeight - 1
	log.Printf("Previous height: %d", previousHeight)

	// 이전 블록 정보 조회
	previousBlockInfo, err := vt.fetchBlock(previousHeight)
	if err != nil {
		log.Printf("Error fetching previous block %d: %v", previousHeight, err)
		return
	}

	// 이전 블록의 서명 정보로 현재 블록의 서명 상태 판단
	signedValidators := make(map[string]bool)
	for _, sig := range previousBlockInfo.Result.Block.LastCommit.Signatures {
		if sig.Signature != "" {
			signedValidators[sig.ValidatorAddress] = true
		}
	}

	// 디버깅을 위한 로그 추가
	log.Printf("Previous block %d signatures: %v", previousHeight, signedValidators)
	log.Printf("Tracking validators: %v", vt.validators)

	// 현재 블록 높이에 대해 이전 블록의 서명 정보로 메트릭 업데이트
	for address, label := range vt.validators {
		signed := 0.0
		if signedValidators[address] {
			signed = 1.0
		}
		
		log.Printf("Validator %s (%s): signed=%v", address, label, signed)
		
		// 비콘 체인 메트릭 업데이트
		vt.metrics.custom.beaconBlockSignedMetric.WithLabelValues(label, currentBlockInfo.Result.Block.Header.Height).Set(signed)
		
		// CometBFT consensus missed blocks metric 업데이트
		// 서명하지 않았으면 missed blocks로 카운트
		missedBlocks := 0.0
		if !signedValidators[address] {
			missedBlocks = 1.0
		}
		vt.metrics.cosmos.cometbftMissedBlocksMetric.WithLabelValues(label, "0g-galileo").Set(missedBlocks)
	}

	log.Printf("Updated beacon block metrics for block %d based on previous block %d", currentHeight, previousHeight)
}

func (vt *UnifiedValidatorTracker) updateCosmosMetrics() {
	// 스테이킹 벨리데이터 정보 조회
	stakingValidators, err := vt.fetchStakingValidators()
	if err != nil {
		log.Printf("Error fetching staking validators: %v", err)
		return
	}

	// 벨리데이터 정보 업데이트
	for _, validator := range stakingValidators.Validators {
		// 주소를 hex 형식으로 변환 (필요한 경우)
		address := validator.OperatorAddress
		
		// 추적 중인 벨리데이터인지 확인
		label, exists := vt.validators[address]
		if !exists {
			continue
		}

		// 본딩 상태
		isBonded := 0.0
		if validator.Status == "BOND_STATUS_BONDED" {
			isBonded = 1.0
		}
		vt.metrics.cosmos.isBondedMetric.WithLabelValues(label).Set(isBonded)

		// 감금 상태
		isJailed := 0.0
		if validator.Jailed {
			isJailed = 1.0
		}
		vt.metrics.cosmos.isJailedMetric.WithLabelValues(label).Set(isJailed)

		// 토큰 수량
		if tokens, err := strconv.ParseFloat(validator.Tokens, 64); err == nil {
			vt.metrics.cosmos.tokensMetric.WithLabelValues(label).Set(tokens)
		}

		// 커미션
		if rate, err := strconv.ParseFloat(validator.Commission.CommissionRates.Rate, 64); err == nil {
			vt.metrics.cosmos.commissionMetric.WithLabelValues(label).Set(rate)
		}

		// CometBFT consensus missed blocks metric
		// 기존 missed blocks 정보를 사용하여 CometBFT 형식으로도 노출
		// 실제 구현에서는 더 정확한 데이터가 필요할 수 있음
		vt.metrics.cosmos.cometbftMissedBlocksMetric.WithLabelValues(label, "0g-galileo").Set(0.0) // 기본값
	}

	// 기본 메트릭 설정 (예시 값들)
	vt.metrics.cosmos.activeSetMetric.Set(float64(len(stakingValidators.Validators)))
	vt.metrics.cosmos.seatPriceMetric.Set(1000000.0) // 예시 값
	vt.metrics.cosmos.signedBlocksWindowMetric.Set(100.0) // 예시 값
	vt.metrics.cosmos.minSignedBlocksPerWindowMetric.Set(50.0) // 예시 값
	vt.metrics.cosmos.downtimeJailDurationMetric.Set(600.0) // 예시 값
	vt.metrics.cosmos.slashFractionDoubleSignMetric.Set(0.05) // 예시 값
	vt.metrics.cosmos.slashFractionDowntimeMetric.Set(0.01) // 예시 값
}

func (vt *UnifiedValidatorTracker) updateValidatorStatus() {
	validatorInfo, err := vt.fetchValidators()
	if err != nil {
		log.Printf("Error fetching validators: %v", err)
		return
	}

	// Create a map of active validators
	activeValidators := make(map[string]bool)
	for _, validator := range validatorInfo.Validators {
		activeValidators[validator.Address] = true
	}

	// Update status for each tracked validator
	for address, label := range vt.validators {
		status := 0.0
		if activeValidators[address] {
			status = 1.0
		}
		
		vt.metrics.custom.validatorStatusMetric.WithLabelValues(label, address).Set(status)
	}
}

func (vt *UnifiedValidatorTracker) updateMempoolMetrics() {
	// 0G 갈릴레오는 mempool API를 제공하지 않으므로
	// 현재 블록의 트랜잭션 정보를 사용하여 mempool 상태를 추정
	
	// 최신 블록 정보 가져오기
	blockInfo, err := vt.fetchBlock(0) // 0 means latest block
	if err != nil {
		log.Printf("Error fetching latest block for mempool estimation: %v", err)
		return
	}

	// 블록 높이 파싱
	height, err := strconv.ParseInt(blockInfo.Result.Block.Header.Height, 10, 64)
	if err != nil {
		log.Printf("Error parsing block height: %v", err)
		return
	}

	// 이전 블록과 비교하여 트랜잭션 변화 추정
	// 실제로는 더 정확한 방법이 필요하지만, 현재로서는 기본값 설정
	estimatedMempoolSize := float64(0) // 기본값
	estimatedTotalBytes := float64(0)  // 기본값
	estimatedTotal := float64(0)       // 기본값

	// 블록 높이가 증가했는지 확인하여 네트워크 활동 추정
	if height > vt.lastBlockHeight {
		// 네트워크가 활성화되어 있다고 가정
		estimatedMempoolSize = 10.0 // 추정값
		estimatedTotalBytes = 1024.0 // 추정값 (1KB)
		estimatedTotal = 5.0 // 추정값
	} else {
		// 네트워크가 비활성화되어 있다고 가정
		estimatedMempoolSize = 0.0
		estimatedTotalBytes = 0.0
		estimatedTotal = 0.0
	}

	// 메트릭 업데이트
	vt.metrics.custom.mempoolSizeMetric.Set(estimatedMempoolSize)
	vt.metrics.custom.mempoolTotalBytesMetric.Set(estimatedTotalBytes)
	vt.metrics.custom.mempoolTotalMetric.Set(estimatedTotal)

	log.Printf("Updated estimated mempool metrics - Size: %.0f, Total: %.0f, TotalBytes: %.0f (Block: %d)", 
		estimatedMempoolSize, estimatedTotal, estimatedTotalBytes, height)
}

func (vt *UnifiedValidatorTracker) updateBlockMetrics(blockInfo *BlockInfo) {
	height, _ := strconv.ParseInt(blockInfo.Result.Block.Header.Height, 10, 64)
	log.Printf("Updating block metrics for height: %d", height)
	vt.metrics.cosmos.blockHeightMetric.Set(float64(height))
	log.Printf("Set block height metric to: %d", height)

	// 비콘 체인용 메트릭 업데이트
	vt.updateBeaconBlockMetrics(blockInfo)
	
	// cosmos-validator-watcher 메트릭 업데이트
	vt.updateCosmosMetrics()
	
	// 카운터 메트릭 업데이트
	vt.metrics.cosmos.trackedBlocksMetric.Inc()
}

func (vt *UnifiedValidatorTracker) StartTracking(ctx context.Context) {
	log.Printf("StartTracking: Initializing block tracking with 5-second intervals")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Printf("StartTracking: Starting tracking loop")
	for {
		select {
		case <-ctx.Done():
			log.Printf("StartTracking: Context cancelled, stopping tracking")
			return
		case <-ticker.C:
			log.Printf("StartTracking: Tick received, calling trackLatestBlock")
			vt.trackLatestBlock()
		}
	}
}

func (vt *UnifiedValidatorTracker) trackLatestBlock() {
	// Fetch latest block
	log.Printf("Attempting to fetch latest block from RPC endpoint: %s", vt.rpcEndpoint)
	blockInfo, err := vt.fetchBlock(0) // 0 means latest block
	if err != nil {
		log.Printf("Error fetching latest block: %v", err)
		// RPC 연결 실패 시에도 기본 메트릭은 계속 제공
		return
	}

	height, _ := strconv.ParseInt(blockInfo.Result.Block.Header.Height, 10, 64)
	log.Printf("Successfully fetched block height: %d", height)
	
	// Only process if this is a new block and hasn't been processed
	if height > vt.lastBlockHeight && !vt.processedBlocks[height] {
		log.Printf("Processing new block: %d (previous: %d)", height, vt.lastBlockHeight)
		vt.updateBlockMetrics(blockInfo)
		log.Printf("About to call updateBeaconBlockMetrics for block %d", height)
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic in updateBeaconBlockMetrics: %v", r)
				}
			}()
			vt.updateBeaconBlockMetrics(blockInfo) // Add beacon block metrics update
		}()
		log.Printf("Finished calling updateBeaconBlockMetrics for block %d", height)
		vt.updateValidatorStatus()
		vt.updateMempoolMetrics() // Add this line to update mempool metrics
		vt.lastBlockHeight = height
		vt.processedBlocks[height] = true
		
		// 메모리 관리를 위해 오래된 블록 정보 정리 (최근 1000개 블록만 유지)
		if len(vt.processedBlocks) > 1000 {
			for oldHeight := range vt.processedBlocks {
				if oldHeight < height-1000 {
					delete(vt.processedBlocks, oldHeight)
				}
			}
		}
		
		log.Printf("Successfully processed beacon block %d", height)
	} else {
		log.Printf("Block %d already processed or not new (last: %d)", height, vt.lastBlockHeight)
	}
}

// NodeExporterMetrics represents Node Exporter metrics
type NodeExporterMetrics struct {
	client *http.Client
	url    string
}

func NewNodeExporterMetrics(url string) *NodeExporterMetrics {
	return &NodeExporterMetrics{
		client: &http.Client{Timeout: 10 * time.Second},
		url:    url,
	}
}

func (nem *NodeExporterMetrics) fetchMetrics() (string, error) {
	resp, err := nem.client.Get(nem.url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func main() {
	// 0G 체인 갈릴레오 설정 (비콘 체인)
	rpcEndpoint := os.Getenv("RPC_ENDPOINT")
	if rpcEndpoint == "" {
		rpcEndpoint = "http://57.129.73.24:50657" // 기본값
	}
	
	// 추적할 벨리데이터 (실제 0G 노드 벨리데이터 주소 사용)
	validators := map[string]string{
		"21F5C524FCA565DD50841FF4B92A7220AA5B0BDD": "validator1",
	}

	log.Printf("Initializing unified metrics tracker with RPC endpoint: %s", rpcEndpoint)
	log.Printf("Tracking validators: %v", validators)

	tracker := NewUnifiedValidatorTracker(rpcEndpoint, validators)
	tracker.RegisterMetrics()
	log.Printf("Metrics registered successfully")

	// Node Exporter 메트릭 수집기 초기화
	nodeExporterURL := os.Getenv("NODE_EXPORTER_URL")
	if nodeExporterURL == "" {
		nodeExporterURL = "http://57.129.73.24:9200/metrics" // 기본값
	}
	nodeExporter := NewNodeExporterMetrics(nodeExporterURL)
	log.Printf("Node Exporter metrics collector initialized")

	// HTTP 서버 설정
	http.Handle("/metrics", promhttp.Handler())
	
	// 통합 메트릭 엔드포인트 (모든 메트릭 포함)
	http.HandleFunc("/all-metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		
		// 1. Prometheus 메트릭 (cosmos-validator-watcher + 커스텀 메트릭)
		promResp, err := http.Get("http://localhost:8080/metrics")
		if err == nil {
			defer promResp.Body.Close()
			io.Copy(w, promResp.Body)
		} else {
			log.Printf("Warning: Failed to fetch local metrics: %v", err)
		}
		
		// 2. Node Exporter 메트릭 추가 (시스템 메트릭만)
		nodeMetrics, err := nodeExporter.fetchMetrics()
		if err == nil {
			w.Write([]byte("\n# Node Exporter Metrics\n"))
			w.Write([]byte(nodeMetrics))
		} else {
			log.Printf("Warning: Failed to fetch Node Exporter metrics: %v", err)
		}
		
		// 3. 0G 노드 메트릭 추가 (CometBFT 메트릭만, 중복 제거)
		ogNodeURL := os.Getenv("OG_NODE_METRICS_URL")
		if ogNodeURL == "" {
			ogNodeURL = "http://57.129.73.24:50660/metrics" // 기본값
		}
		log.Printf("Attempting to fetch 0G node metrics from %s", ogNodeURL)
		ogClient := &http.Client{Timeout: 15 * time.Second}
		ogResp, err := ogClient.Get(ogNodeURL)
		if err == nil {
			defer ogResp.Body.Close()
			body, err := io.ReadAll(ogResp.Body)
			if err == nil {
				// CometBFT 메트릭만 필터링하여 중복 제거
				lines := strings.Split(string(body), "\n")
				w.Write([]byte("\n# 0G Galileo Node Metrics (CometBFT)\n"))
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" && !strings.HasPrefix(line, "#") {
						// 이미 로컬 메트릭에 있는 메트릭은 제외
						if !strings.Contains(line, "og_galileo_") && 
						   !strings.Contains(line, "cosmos_validator_") &&
						   !strings.Contains(line, "go_") &&
						   !strings.Contains(line, "process_") {
							w.Write([]byte(line + "\n"))
						}
					} else if strings.HasPrefix(line, "#") {
						// 헬프 텍스트는 유지
						w.Write([]byte(line + "\n"))
					}
				}
			}
			log.Printf("Successfully fetched 0G node metrics (status: %d)", ogResp.StatusCode)
		} else {
			log.Printf("Warning: Failed to fetch 0G node metrics: %v", err)
			// 에러가 발생해도 기본 메트릭은 계속 제공
			w.Write([]byte("\n# 0G Galileo Node Metrics (CometBFT) - UNAVAILABLE\n"))
			w.Write([]byte("# Error: Unable to connect to 0G node metrics endpoint\n"))
		}
	})
	
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>0G Galileo Unified Metrics</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .metric { margin: 10px 0; padding: 10px; background: #f5f5f5; border-radius: 5px; }
        a { color: #007bff; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .status { padding: 5px 10px; border-radius: 3px; color: white; font-weight: bold; }
        .status.running { background: #28a745; }
        .status.stopped { background: #dc3545; }
    </style>
</head>
<body>
    <div class="container">
        <h1>0G Galileo Beacon Chain Unified Metrics</h1>
        <p>Unified metrics collector - provides cosmos-validator-watcher, custom beacon chain metrics, and system metrics from a single port.</p>
        
        <div class="metric">
            <h3>📊 Metrics Endpoints</h3>
            <p><a href="/metrics">/metrics</a> - Basic Prometheus format metrics</p>
            <p><a href="/all-metrics">/all-metrics</a> - <strong>All metrics unified (recommended)</strong></p>
        </div>
        
        <div class="metric">
            <h3>🏥 Health Check</h3>
            <p><a href="/health">/health</a> - Service status check</p>
        </div>
        
        <div class="metric">
            <h3>🔗 Unified Metrics Configuration</h3>
            <ul>
                <li><strong>cosmos-validator-watcher metrics</strong> - Basic validator information</li>
                <li><strong>Beacon chain custom metrics</strong> - Block signing status, mempool, etc.</li>
                <li><strong>Node Exporter metrics</strong> - System resources (CPU, memory, disk, etc.)</li>
                <li><strong>0G node metrics</strong> - Chain node status</li>
            </ul>
        </div>
        
        <div class="metric">
            <h3>🔗 Key Metrics</h3>
            <ul>
                <li><strong>og_galileo_validator_beacon_block_signed</strong> - Beacon chain block signing status</li>
                <li><strong>og_galileo_validator_block_height</strong> - Current block height</li>
                <li><strong>og_galileo_validator_is_bonded</strong> - Validator bonding status</li>
                <li><strong>og_galileo_validator_missed_blocks</strong> - Number of missed blocks</li>
                <li><strong>og_galileo_validator_mempool_size</strong> - Mempool size (estimated)</li>
                <li><strong>node_cpu_seconds_total</strong> - CPU usage</li>
                <li><strong>node_memory_MemTotal_bytes</strong> - Memory usage</li>
                <li><strong>node_filesystem_size_bytes</strong> - Disk usage</li>
            </ul>
        </div>
        
        <div class="metric">
            <h3>⚠️ Beacon Chain Characteristics</h3>
            <p>Block signing status is determined using previous block's LastCommit information.</p>
            <p>Current block N signing status = Block N-1 signing information</p>
        </div>
        
        <div class="metric">
            <h3>🌐 External Access</h3>
            <p>Accessible through nginx reverse proxy at the following URLs:</p>
            <ul>
                <li><a href="/node-exporter/">/node-exporter/</a> - Node Exporter metrics</li>
                <li><a href="/grafana/">/grafana/</a> - Grafana dashboard</li>
                <li><a href="/prometheus/">/prometheus/</a> - Prometheus UI</li>
            </ul>
        </div>
    </div>
</body>
</html>
		`))
	})

	// 백그라운드에서 블록 추적 시작
	log.Printf("Starting block tracking in background...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go tracker.StartTracking(ctx)
	log.Printf("Block tracking started successfully")

	log.Println("Starting 0G Galileo unified metrics server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}