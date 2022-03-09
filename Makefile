all: docker-compose

generateABI:
	@solc --abi contracts/erc20.sol -o build --overwrite

transpileABIToGo:
	@abigen --abi=build/ERC20.abi --pkg=token --out=token/erc20.go

tokenInterface: generateABI transpileABIToGo

docker-compose: tokenInterface
	docker-compose build

mixin: lint-mixin

lint-mixin:
	mixtool lint mixin/mixin.libsonnet
