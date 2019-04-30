PKG := github.com/openshift/osde2e

out/osde2e: out
	go build -v -o $@ $(PKG)/cmd/osde2e

out:
	mkdir -p $@
