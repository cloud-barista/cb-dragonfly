
# cb-dragonfly
Cloud-Barista Integrated Monitoring Framework

***

## [목차]

1. [설치 개요](#설치-개요)
2. [설치 절차](#설치-절차)
3. [설치 & 실행 상세 정보](#설치--실행-상세-정보)

***


## [설치 개요]
- 설치 환경: 리눅스(검증시험:Ubuntu 18.04)

## [설치 절차]

- Git 설치
- Go 설치
- 실시간 모니터링 데이터 저장소 설치
- 시계열 모니터링 데이터 저장소 설치
- 멀티 클라우드 모니터링 프레임워크 (cb-dragonfly) 설치
- 멀티 클라우드 모니터링 프레임워크 (cb-dragonfly) 실행

## [설치 & 실행 상세 정보]

- Git 설치
  - `$ sudo apt update`
  - `$ sudo apt install git`
  - `$ sudo apt-get install git-core`

- Go 설치
  - https://golang.org/doc/install 
  (2019년 11월 현재 `$ sudo apt install golang` 으로 설치하면 1.10 설치됨. 이 링크에서 1.12 이상 버전으로 설치할 것(Go mod 호환성 문제))
  - `$ wget https://dl.google.com/go/go1.13.4.linux-amd64.tar.gz` (설치 파일 다운로드)
  - `$ sudo tar -C /usr/local -xzf go1.13.4.linux-amd64.tar.gz` (압축해제)
  - `$ sudo vim ~/.bashrc 파일 맨 아래에 export GOROOT=$PATH:/usr/local/go/bin` (GOPATH 환경변수 추가)
  - `$ source ~/.bashrc` (수정한 bashrc 파일 반영)
  - `$ go version` (버전 확인)

- 실시간 모니터링 데이터 저장소 설치
  - etcd 설치(3.3.11) 및 실행
  
        - `$ wget https://github.com/coreos/etcd/releases/download/v3.3.11/etcd-v3.3.11-linux-amd64.tar.gz` (설치 파일 다운로드)
        - `$ sudo tar -xvf etcd-v3.3.11-linux-amd64.tar.gz` (압축해제)
        - `$ sudo mv etcd-v3.3.11-linux-amd64/etcd* /usr/local/bin/` (추출된 실행파일을 로컬 저장소로 이동)
        - `$ etcd --version` (버전 확인)
    
        - `$ sudo mkdir -p /var/lib/etcd/` (Etcd 구성 파일 폴더 생성)
        - `$ sudo mkdir /etc/etcd` (데이터 폴더 생성)
    
        - `$ sudo groupadd --system etcd` (etcd 시스템 그룹 생성)
        - `$ sudo useradd -s /sbin/nologin --system -g etcd etcd` (etcd 시스템 사용자 생성)
        - `$ sudo chown -R etcd:etcd /var/lib/etcd/` (/var/lib/etcd/ 폴더 소유권을 etcd사용자로 설정)
    
        - `$ sudo vim /etc/systemd/system/etcd.service` (etcd에 대한 새로운 시스템 서비스 파일 작성)
        (바로 밑에 코드 붙여넣기 후)
        - `$ sudo systemctl  daemon-reload` (데몬 재시작)
        - `$ sudo systemctl  start etcd.service` (etcd 서비스 시작)
        
  - etcd.service 등록
          
          [Unit]
          Description=etcd key-value store
          Documentation=https://github.com/etcd-io/etcd
          After=network.target

          [Service]
          User=etcd
          Type=notify
          Environment=ETCD_DATA_DIR=/var/lib/etcd
          Environment=ETCD_NAME=%m
          ExecStart=/usr/local/bin/etcd
          Restart=always
          RestartSec=10s
          LimitNOFILE=40000

          [Install]
          WantedBy=multi-user.target


- 모니터링 시계열 데이터 저장소 설치

  - influxdb 설치(1.7.8) 및 실행
  
        - `$ wget https://dl.influxdata.com/influxdb/releases/influxdb_1.7.8_amd64.deb` (패키지 파일 다운로드)
        - `$ sudo dpkg -i influxdb_1.7.8_amd64.deb` (패키지 설치)
        - `$ sudo systemctl start influxdb` (influxDB 서비스 시작)
        - `$ influx --version` (버전 확인)
    
        - `$ influx` (influxDB 사용하기)
            - CREATE DATABASE cbmon

- 멀티 클라우드 모니터링 프레임워크 (cb-dragonfly) 설치

    - Git Project Clone

          - `$ sudo git config --global color.ui "auto"` (Git 소스에 색 구분)
          - `$ sudo git clone https://github.com/cloud-barista/cb-dragonfly.git` (Git 프로젝트 CLone)
          - `username = {{GitUserEmail}}` (Clone시 자격여부 확인 : 자신의 Git Email 입력)
          - `Password = {{GitUserPW}}`    (Clone시 자격여부 확인 : 자신의 Git PW 입력)
    
    - Go mod 기반 의존성 라이브러리 로드
          
          - `$ cd ~/cb-mon` (clone한 프로젝트 파일로 들어가기)
          - `$ go mod download` (.mod 파일에 등록된 라이브러리 다운로드 실행)
    
    - Go mod 기반 의존성 라이브러리 다운로드 확인
    
          - `$ go mod verify` (다운로드 확인)
    
    - 환경변수 설정
          
          - `$ sudo vim conf/setup.env` (실행에 필요한 PATH를 처리할 파일 생성  (현 위치: ~/cb-mon))
               setup.env에 추가
                
                export CBSTORE_ROOT=~/cb-mon
                export CBLOG_ROOT=~/cb-mon
                export CBMON_PATH=~/cb-mon
                export SPIDER_URL=http://localhost:1024
                
          - `$ source conf/setup.env` (수정한 setup.env 반영)         
          - `$ go run pkg/manager/main/main.go` (실행)
    
    - config 파일 설정 (config 파일에 influxdb IP 및 사용자 정보, etcd IP 정보, cb-dragonfly 호스트 IP 정보를 순차적으로 입력)
          
          -  # influxdb connection info
             influxdb:
              endpoint_url: http://{{influxdb_ip}}:8086
              database: cbmon
              user_name: {{user_name}}
              password: {{password}}

            # etcd connection info
            etcd:
              endpoint_url: http://{{etcd_ip}}:2379
              ttl: 60                 # time to live (s)

            # collect manager configuration info
            collect_manager:
              collector_ip: {{collector_ip}}
              collector_port: 8094    # udp port
              collector_count: 1      # default collector count
    

- 멀티 클라우드 모니터링 프레임워크 실행

     - 프로젝트 빌드 및 실행

              - `$ cd ~/cb-mon`
              - `$ go run pkg/manager/main/main.go` (cb-dragonfly 프로젝트 빌드 및 실행)
              

