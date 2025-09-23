.PHONY: build-all build-gateway build-social build-user build-leaderboard build-log-service push-all push-gateway push-social push-user push-leaderboard push-log-service 

build-all: build-gateway build-social build-user build-leaderboard


build-gateway:
	docker build -f dockerfile/gateway.Dockerfile -t 192.168.101.2:5000/gateway:v1.0.0 .

build-social:
	docker build -f dockerfile/social-service.Dockerfile -t 192.168.101.2:5000/social:v1.0.0 .

build-user:
	docker build -f dockerfile/user-service.Dockerfile -t 192.168.101.2:5000/user:v1.0.0 .

build-leaderboard:
	docker build -f dockerfile/leaderboard-service.Dockerfile -t 192.168.101.2:5000/leaderboard:v1.0.0 .

build-log-service:
	docker build -f dockerfile/log-service.Dockerfile -t 192.168.101.2:5000/log-service:v1.0.0 .

push-all: push-gateway push-social push-user push-leaderboard push-log-service

push-gateway:
	docker push 192.168.101.2:5000/gateway:v1.0.0

push-social:
	docker push 192.168.101.2:5000/social:v1.0.0

push-user:
	docker push 192.168.101.2:5000/user:v1.0.0

push-leaderboard:
	docker push 192.168.101.2:5000/leaderboard:v1.0.0

push-log-service:
	docker push 192.168.101.2:5000/log-service:v1.0.0
