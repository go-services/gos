templates:
	resources -declare -var=FS -package=template -output=template/assets.go $$(find assets)
	gofmt -s -w template/assets.go
