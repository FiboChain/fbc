#!/usr/bin/env bash

NUM_NODE=4

# tackle size chronic goose deny inquiry gesture fog front sea twin raise
# acid pulse trial pill stumble toilet annual upgrade gold zone void civil
# antique onion adult slot sad dizzy sure among cement demise submit scare
# lazy cause kite fence gravity regret visa fuel tone clerk motor rent
HARDCODED_MNEMONIC=true

set -e
set -o errexit
set -a
set -m

set -x # activate debugging

source fbc.profile
WRAPPEDTX=false
PRERUN=false
NUM_RPC=0
WHITE_LIST=0x9119cE8bEA989Ee6D19514f4a8390F3503689D79,\
0x3934c56b25c36a7f28DE3522b8Fd0c55B7D54722,\
0x7E4A75569041996E869a0749Ed342fB13bD7d249,\
0x631f017259c07E2A57fbda3f269DbdCF00c7979e,\
testnet-node-ids


while getopts "r:isn:b:p:c:Sxwk:" opt; do
  case $opt in
  i)
    echo "FBCHAIN_INIT"
    FBCHAIN_INIT=1
    ;;
  r)
    echo "NUM_RPC=$OPTARG"
    NUM_RPC=$OPTARG
    ;;
  w)
    echo "WRAPPEDTX=$OPTARG"
    WRAPPEDTX=true
    ;;
  x)
    echo "PRERUN=$OPTARG"
    PRERUN=true
    ;;
  s)
    echo "FBCHAIN_START"
    FBCHAIN_START=1
    ;;
  k)
    echo "LOG_SERVER"
    LOG_SERVER="--log-server $OPTARG"
    ;;
  c)
    echo "Test_CASE"
    Test_CASE="--consensus-testcase $OPTARG"
    ;;
  n)
    echo "NUM_NODE=$OPTARG"
    NUM_NODE=$OPTARG
    ;;
  b)
    echo "BIN_NAME=$OPTARG"
    BIN_NAME=$OPTARG
    ;;
  S)
    STREAM_ENGINE="analysis&mysql&localhost:3306,notify&redis&localhost:6379,kline&pulsar&localhost:6650"
    echo "$STREAM_ENGINE"
    ;;
  p)
    echo "IP=$OPTARG"
    IP=$OPTARG
    ;;
  \?)
    echo "Invalid option: -$OPTARG"
    ;;
  esac
done

echorun() {
  echo "------------------------------------------------------------------------------------------------"
  echo "["$@"]"
  $@
  echo "------------------------------------------------------------------------------------------------"
}

killbyname() {
  NAME=$1
  ps -ef|grep "$NAME"|grep -v grep |awk '{print "kill -9 "$2", "$8}'
  ps -ef|grep "$NAME"|grep -v grep |awk '{print "kill -9 "$2}' | sh
  echo "All <$NAME> killed!"
}

init() {
  killbyname ${BIN_NAME}

  (cd ${FBCHAIN_TOP} && make install VenusHeight=1)

  rm -rf cache

  echo "=================================================="
  echo "===== Generate testnet configurations files...===="
  echorun fbchaind testnet --v $1 --r $2 -o cache -l \
    --chain-id ${CHAIN_ID} \
    --starting-ip-address ${IP} \
    --base-port ${BASE_PORT} \
    --keyring-backend test
}
recover() {
  killbyname ${BIN_NAME}
  (cd ${FBCHAIN_TOP} && make install VenusHeight=1)
  rm -rf cache
  cp -rf nodecache cache
}

run() {

  index=$1
  seed_mode=$2
  p2pport=$3
  rpcport=$4
  restport=$5
  p2p_seed_opt=$6
  p2p_seed_arg=$7

  if [ "$(uname -s)" == "Darwin" ]; then
      sed -i "" 's/"enable_call": false/"enable_call": true/' cache/node${index}/fbchaind/config/genesis.json
      sed -i "" 's/"enable_create": false/"enable_create": true/' cache/node${index}/fbchaind/config/genesis.json
      sed -i "" 's/"enable_contract_blocked_list": false/"enable_contract_blocked_list": true/' cache/node${index}/fbchaind/config/genesis.json
  else
      sed -i 's/"enable_call": false/"enable_call": true/' cache/node${index}/fbchaind/config/genesis.json
      sed -i 's/"enable_create": false/"enable_create": true/' cache/node${index}/fbchaind/config/genesis.json
      sed -i 's/"enable_contract_blocked_list": false/"enable_contract_blocked_list": true/' cache/node${index}/fbchaind/config/genesis.json
  fi

  fbchaind add-genesis-account 0x9119cE8bEA989Ee6D19514f4a8390F3503689D79 900000000fibo --home cache/node${index}/fbchaind
  fbchaind add-genesis-account 0x3934c56b25c36a7f28DE3522b8Fd0c55B7D54722 900000000fibo --home cache/node${index}/fbchaind
  fbchaind add-genesis-account 0x7E4A75569041996E869a0749Ed342fB13bD7d249 900000000fibo --home cache/node${index}/fbchaind
  fbchaind add-genesis-account 0x631f017259c07E2A57fbda3f269DbdCF00c7979e 900000000fibo --home cache/node${index}/fbchaind

  LOG_LEVEL=main:info,*:error,consensus:error,state:info

  echorun nohup fbchaind start \
    --home cache/node${index}/fbchaind \
    --p2p.seed_mode=$seed_mode \
    --p2p.allow_duplicate_ip \
    --dynamic-gp-mode=2 \
    --enable-wtx=${WRAPPEDTX} \
    --mempool.node_key_whitelist ${WHITE_LIST} \
    --p2p.pex=false \
    --p2p.addr_book_strict=false \
    $p2p_seed_opt $p2p_seed_arg \
    --p2p.laddr tcp://${IP}:${p2pport} \
    --rpc.laddr tcp://${IP}:${rpcport} \
    --log_level ${LOG_LEVEL} \
    --chain-id ${CHAIN_ID} \
    --upload-delta=false \
    --enable-gid \
    --consensus.timeout_commit 3800ms \
    --enable-blockpart-ack=false \
    --append-pid=true \
    ${LOG_SERVER} \
    --elapsed DeliverTxs=0,Round=1,CommitRound=1,Produce=1 \
    --rest.laddr tcp://localhost:$restport \
    --enable-preruntx=$PRERUN \
    --consensus-role=v$index \
    --active-view-change=true \
    ${Test_CASE} \
    --keyring-backend test >cache/val${index}.log 2>&1 &

#     --iavl-enable-async-commit \    --consensus-testcase case12.json \
#     --upload-delta \
#     --enable-preruntx \
#     --mempool.node_key_whitelist="nodeKey1,nodeKey2" \
#    --mempool.node_key_whitelist ${WHITE_LIST} \
}

#prometheus_listen_addr
#prof_laddr
function start() {
  killbyname ${BIN_NAME}
  index=0

  echo "============================================"
  echo "=========== Startup seed node...============"
  ((restport = REST_PORT)) # for evm tx
  run $index true ${seedp2pport} ${seedrpcport} $restport
  seed=$(fbchaind tendermint show-node-id --home cache/node${index}/fbchaind)

  echo "============================================"
  echo "======== Startup validator nodes...========="
  for ((index = 1; index < ${1}; index++)); do
    ((p2pport = BASE_PORT_PREFIX + index * 100 + P2P_PORT_SUFFIX))
    ((rpcport = BASE_PORT_PREFIX + index * 100 + RPC_PORT_SUFFIX))  # for fbchaincli
    ((restport = index * 100 + REST_PORT)) # for evm tx
    run $index false ${p2pport} ${rpcport} $restport --p2p.seeds ${seed}@${IP}:${seedp2pport}
  done
  echo "start node done"
}

if [ -z ${IP} ]; then
  IP="127.0.0.1"
fi

if [ ! -z "${FBCHAIN_INIT}" ]; then
	((NUM_VAL=NUM_NODE-NUM_RPC))
  init ${NUM_VAL} ${NUM_RPC}
fi

if [ ! -z "${FBCHAIN_RECOVER}" ]; then
  recover ${NUM_NODE}
fi

if [ ! -z "${FBCHAIN_START}" ]; then
  start ${NUM_NODE}
fi
