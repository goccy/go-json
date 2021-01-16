.PHONY: cover
cover:
	@ go test -coverprofile=cover.tmp.out . ; \
	cat cover.tmp.out | grep -v "encode_optype.go" > cover.out; \
	rm cover.tmp.out

cover-html: cover
	go tool cover -html=cover.out
