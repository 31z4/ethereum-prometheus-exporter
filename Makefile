tokenInterface:
	@solc --abi token/erc20.sol -o build --overwrite
	@abigen --abi=build/ERC20.abi --pkg=token --out=token/erc20.go
