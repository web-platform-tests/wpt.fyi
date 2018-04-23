#!/bin/bash

GRAY='\033[0;90m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

function verbose() {
  printf "${GRAY}[  $(date +'%Y-%m-%d %H:%M:%S')  VERB  ]  $1${NC}\n"
}

function debug() {
  printf "${GRAY}[  $(date +'%Y-%m-%d %H:%M:%S')  DEBUG ]  $1${NC}\n"
}

function info() {
  printf "${GREEN}[  $(date +'%Y-%m-%d %H:%M:%S')  INFO  ]  $1${NC}\n"
}

function warn() {
  printf "${YELLOW}[  $(date +'%Y-%m-%d %H:%M:%S')  WARN  ]  $1${NC}\n"
}

function error() {
  printf "${RED}[  $(date +'%Y-%m-%d %H:%M:%S')  ERROR ]  $1${NC}\n"
}

# Usage: fatal "Message" [exitCode]
function fatal() {
  printf "${RED}[  $(date +'%Y-%m-%d %H:%M:%S')  FATAL ]  $1${NC}\n"
  exit ${2:-1}
}

function confirm() {
  warn "${1} (Y/n)"
  exec < /dev/tty
  read -n 1 CH

  if [ "${CH}" == "y" ] || [ "${CH}" == "Y" ]; then
    return 0
  else
    return 1
  fi
}
