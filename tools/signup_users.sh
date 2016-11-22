#!/bin/bash

if [ $# -ne 3 ]; then
  echo "Usage: $0 user_prefix password count"
  exit 1
fi

server=${DUET_SERVER:-"https://api.helloduet.com"}
prefix=$1
password=$2
count=$3

echo "Signing up $count users with prefix '$1'"

for ((i=0; i < $count; i++)); do
  user=$prefix$i
  curl "$server/rest/signup" --fail -H "Content-Type: application/json" -X POST \
    -d '{ "username": "'$user'", "password": "'$password'" }'
  if [ $? -ne 0 ]; then
    echo "Request failed for user '$user'"
    exit 1
  fi
done
