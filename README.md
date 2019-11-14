# cb-dragonfly
Cloud-Barista Integrated Monitoring Framework

***

## [목차]

0. [VM 접속](#vm-접속)
1. [설치 개요](#설치-개요)
2. [설치 절차](#설치-절차)
3. [설치 & 실행 상세 정보](#설치--실행-상세-정보)

***


## [VM 접속]

- PrivateKey.pem 사용 SSH 접속
  - `$ ssh cb-user@52.141.22.160 -i CB-Test.pem` 으로 vm에 접속

## [설치 개요]
- 설치 환경: 리눅스(검증시험:Ubuntu 18.04)

## [설치 절차]

- Go 설치 & Git 설치
- 환경 변수 설정

## [설치 & 실행 상세 정보]

- Git 설치
  - `$ sudo apt update`
  - `$ sudo apt install git`

- Go 설치
  - https://golang.org/doc/install 
  (2019년 11월 현재 `$ sudo apt install golang` 으로 설치하면 1.10 설치됨. 이 링크에서 1.12 이상 버전으로 설치할 것)
  - `$ wget https://dl.google.com/go/go1.13.4.linux-amd64.tar.gz` (파일 다운로드)
  - `$ sudo tar -C /usr/local -xzf go1.13.4.linux-amd64.tar.gz` (압축해제)
  - `$ export PATH=$PATH:/usr/local/go/bin` (PATH 환경변수에 추가)
  - `$ go version` (버전 확인)
  
  - `$ source $HOME/.profile` (추가한 환경변수 즉시 반영)

- 모니터링 데이터베이스 저장소 (의존 라이브러리 다운로드)
  - etcd 설치(3.3.11) 및 실행
    --- `$ wget https://github.com/coreos/etcd/releases/download/v3.3.11/etcd-v3.3.11-linux-amd64.tar.gz` (파일 다운로드)
    --- `$ sudo tar -xvf etcd-v3.3.11-linux-amd64.tar.gz` (압축해제)
    --- `$ sudo mv etcd-v3.3.11-linux-amd64/etcd* /usr/local/bin/` (추출된 실행파일을 로컬 저장소로 이동)
    --- `$ etcd --version` (버전 확인)
    
    --- 
   
  - influxdb (1.7.8) 및 실행
    -- `$ wget https://dl.influxdata.com/influxdb/releases/influxdb_1.7.8_amd64.deb` (다운로드)
    -- `$ sudo dpkg -i influxdb_1.7.8_amd64.deb` (압축해제)
    -- `$ sudo apt-get update && sudo apt-get install influxdb` (InfluxDB 서비스 설치)
    -- `$ sudo systemctl start influxdb` (influxDB 서비스 시작)
    -- `$ influx --version` (버전 확인)
    
    -- `$ influx` (influxDB 사용하기)
    --- CREATE DATABASE cbmon
    --- USE cbmon

- 멀티 클라우드 모니터링 설치

    Git Project Clone
    
    Go mod 의존성 라이브러리 로드
    
    라이브러리 실행 (go run ....)
    
    config 파일 설정

  
### 수정중
