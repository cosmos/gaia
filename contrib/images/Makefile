all: gaiad-env

gaiad-env: gaiad-rmi
	docker build --tag cosmos/gaiad -f gaiad-env/Dockerfile \
		$(shell git rev-parse --show-toplevel)

gaiad-dlv: gaiad-rmi
	docker build --tag cosmos/gaiad -f gaiad-dlv/Dockerfile \
		$(shell git rev-parse --show-toplevel)

gaiad-rmi:
	docker rmi cosmos/gaiad 2>/dev/null; true

.PHONY: all gaiad-env gaiad-dlv gaiad-rmi
