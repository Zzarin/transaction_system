#!/usr/bin/env bash
set -a allexport # export all variables created next
source .env
set +a allexport # stop exporting
$CMD "./app"