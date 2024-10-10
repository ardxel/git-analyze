#!/bin/bash

MAX_DISK_SIZE=5000 \
	MAX_REPO_SIZE=100 \
	SYNC_EVERY=300 \
	USE_FILE_WORKERS=1 \
	APP_TAG="0.8.0" \
	docker-compose -f ./docker-compose.prod.yml up -d
