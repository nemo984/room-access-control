#!/bin/bash

while IFS='=' read -r key value; do
    export "$key"="$value"
done < .env

air
