# 🚀 0G Galileo Prometheus 모니터링 시스템 배포 가이드

## 📦 배포 패키지 정보
- **파일명**: `0g_prometheus_monitoring_system_20250804_224608.tar.gz`
- **크기**: 32KB
- **생성일**: 2025-08-04 22:46:08

## 📋 포함된 구성 요소

### 🔧 핵심 서비스
- **Prometheus**: 메트릭 수집 및 저장
- **Grafana**: 대시보드 및 시각화
- **Unified Metrics**: 0G Galileo 체인 메트릭 수집기
- **Nginx**: 리버스 프록시 및 웹 서버
- **Alertmanager**: 알림 관리 (설정 파일 포함)

### 📁 디렉토리 구조
```
0g_prometheus_deploy/
├── docker-compose.yml          # Docker 서비스 구성
├── install.sh                  # 자동 설치 스크립트
├── start.sh                    # 서비스 시작 스크립트
├── README.md                   # 프로젝트 문서
├── DEPLOYMENT_GUIDE.md         # 상세 배포 가이드
├── CURRENT_STATUS.md           # 현재 시스템 상태
├── LICENSE                     # MIT 라이선스
├── .gitignore                  # Git 무시 파일
├── prometheus-unified.yml      # Prometheus 설정
├── alertmanager/
│   └── alertmanager.yml        # 알림 관리자 설정
├── grafana/
│   ├── provisioning/           # Grafana 자동 설정
│   │   ├── dashboards/        # 대시보드 설정
│   │   └── datasources/       # 데이터 소스 설정
│   └── dashboards/            # 대시보드 JSON 파일
├── nginx/
│   └── nginx.conf             # Nginx 설정
└── unified-metrics/
    ├── main.go                # 메트릭 수집기 소스 코드
    └── go.mod                 # Go 모듈 설정
```

## 🚀 빠른 배포 방법

### 1. 압축 파일 다운로드
```bash
# 압축 파일을 대상 서버에 업로드
scp 0g_prometheus_monitoring_system_20250804_224608.tar.gz user@your-server:/tmp/
```

### 2. 압축 해제
```bash
cd /opt
tar -xzf /tmp/0g_prometheus_monitoring_system_20250804_224608.tar.gz
cd 0g_prometheus_deploy
```

### 3. 환경 설정
```bash
# RPC 엔드포인트 설정
export RPC_ENDPOINT="http://your-node-ip:50657"
export NODE_EXPORTER_URL="http://your-node-ip:9200/metrics"
export OG_NODE_METRICS_URL="http://your-node-ip:50660/metrics"

# 또는 docker-compose.yml에서 직접 수정
sed -i 's/your-node-ip/YOUR_ACTUAL_IP/g' docker-compose.yml
```

### 4. 자동 설치 및 시작
```bash
# 실행 권한 부여
chmod +x install.sh start.sh

# 자동 설치 실행
./install.sh

# 서비스 시작
./start.sh
```

## 🔧 수동 배포 방법

### 1. Docker 설치 확인
```bash
docker --version
docker-compose --version
```

### 2. 서비스 시작
```bash
# 모든 서비스 시작
docker-compose up -d

# 로그 확인
docker-compose logs -f
```

### 3. 상태 확인
```bash
# 컨테이너 상태 확인
docker-compose ps

# 메트릭 엔드포인트 확인
curl http://localhost:8080/health
curl http://localhost:8080/all-metrics
```

## 🌐 접속 URL

배포 완료 후 다음 URL로 접속 가능:

- **Grafana 대시보드**: `http://your-server-ip/grafana/`
- **Prometheus UI**: `http://your-server-ip/prometheus/`
- **통합 메트릭**: `http://your-server-ip/all-metrics/`
- **Node Exporter**: `http://your-server-ip/node-exporter/`

## 🔐 기본 로그인 정보

- **Grafana**: admin / admin
- **Prometheus**: 기본 인증 없음

## 📊 모니터링 메트릭

### 주요 모니터링 포인트
- **Validator Performance**: 벨리데이터 성능 및 블록 서명 상태
- **System Health**: CPU, 메모리, 디스크, 네트워크 사용률
- **Chain Status**: 블록 높이, 메모풀 크기, 컨센서스 상태
- **Application Performance**: Go 애플리케이션 성능 지표
- **Network & Connectivity**: P2P 네트워크 연결 상태

### 핵심 메트릭
- `og_galileo_validator_missed_blocks` - 놓친 블록 수
- `og_galileo_beacon_block_signed` - 블록 서명 상태
- `cometbft_consensus_validator_missed_blocks` - CometBFT 컨센서스 오류
- `node_cpu_seconds_total` - CPU 사용률
- `node_memory_MemTotal_bytes` - 메모리 사용량

## 🛠️ 문제 해결

### 일반적인 문제
1. **포트 충돌**: `netstat -tlnp | grep 9090`로 확인 후 프로세스 종료
2. **RPC 연결 실패**: RPC_ENDPOINT 환경변수 확인
3. **메트릭 중복**: unified-metrics 필터링 로직 확인
4. **컨테이너 시작 실패**: `docker-compose logs`로 로그 확인

### 로그 확인
```bash
# 특정 서비스 로그 확인
docker-compose logs unified-metrics
docker-compose logs prometheus
docker-compose logs grafana

# 실시간 로그 모니터링
docker-compose logs -f
```

## 📞 지원

문제가 발생하면 다음을 확인하세요:
1. `CURRENT_STATUS.md` - 현재 시스템 상태
2. `DEPLOYMENT_GUIDE.md` - 상세 배포 가이드
3. `README.md` - 프로젝트 문서

## 📝 라이선스

이 프로젝트는 MIT 라이선스 하에 배포됩니다. 자세한 내용은 `LICENSE` 파일을 참조하세요.

---

**배포 완료 후**: 시스템이 정상적으로 작동하는지 확인하고, 필요에 따라 알림 설정을 구성하세요. 