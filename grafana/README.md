# Grafana 데이터소스 설정 가이드

## 자동 설정된 데이터소스

이 프로젝트는 자동으로 다음 데이터소스를 설정합니다:

### 1. Prometheus (기본 데이터소스)
- **URL**: `http://prometheus:9090`
- **타입**: Prometheus
- **접근 방식**: Proxy
- **기본 데이터소스**: 예

## 수동으로 추가할 수 있는 데이터소스

### 1. 통합 메트릭 수집기 (Prometheus 형식)
- **URL**: `http://unified-metrics:8080/metrics`
- **타입**: Prometheus
- **접근 방식**: Proxy
- **설명**: cosmos-validator-watcher + 커스텀 비콘 체인 메트릭

### 2. Node Exporter (시스템 메트릭)
- **URL**: `http://172.17.0.1:9200/metrics`
- **타입**: Prometheus
- **접근 방식**: Proxy
- **설명**: 시스템 리소스 메트릭

## 주의사항

### ❌ 잘못된 설정
- **URL**: `http://localhost:8080` (HTML 페이지 반환)
- **문제**: Grafana가 Prometheus API를 기대하지만 HTML 페이지를 받음
- **오류**: "response from prometheus couldn't be parsed. it is non-json"

### ✅ 올바른 설정
- **URL**: `http://localhost:8080/metrics` (Prometheus 형식 메트릭)
- **설명**: `/metrics` 엔드포인트는 Prometheus 형식의 메트릭을 반환

## 데이터소스 추가 방법

1. Grafana에 로그인 (http://localhost:3000, admin/admin123)
2. 설정 → 데이터소스 → 새 데이터소스 추가
3. Prometheus 선택
4. URL 입력 (위의 올바른 URL 중 하나)
5. 저장 및 테스트

## 사용 가능한 메트릭

### Prometheus 데이터소스
- 모든 수집된 메트릭 (통합 메트릭 + Node Exporter + 0G 노드)

### 통합 메트릭 데이터소스
- `og_galileo_validator_*` - 벨리데이터 메트릭
- `og_galileo_validator_beacon_block_signed` - 비콘 체인 서명 상태

### Node Exporter 데이터소스
- `node_cpu_*` - CPU 메트릭
- `node_memory_*` - 메모리 메트릭
- `node_filesystem_*` - 디스크 메트릭
- `node_network_*` - 네트워크 메트릭 