module stefansaftic/docker-gelf-multi-log-driver

go 1.13

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Graylog2/go-gelf v0.0.0-20170811154226-7ebf4f536d8f // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/beeker1121/goque v2.0.1+incompatible
	github.com/containerd/fifo v0.0.0-20190816180239-bda0ff6ed73c // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/cortexproject/cortex v0.2.0
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-plugins-helpers v0.0.0-20181025120712-1e6269c305b8
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-kit/kit v0.8.0
	github.com/gogo/protobuf v1.3.0
	github.com/jinzhu/copier v0.0.0-20190625015134-976e0346caa8
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.6.0
	github.com/sirupsen/logrus v1.4.2
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/tonistiigi/fifo v0.0.0-20190816180239-bda0ff6ed73c
	github.com/weaveworks/common v0.0.0-20190917143411-a2b2a6303c33
	golang.org/x/net v0.0.0-20190916140828-c8589233b77d // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	gotest.tools v2.2.0+incompatible // indirect
)

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20190327083406-200b524eff60

replace github.com/Graylog2/go-gelf => gopkg.in/Graylog2/go-gelf.v2 v2.0.0-20180326133423-4dbb9d721348
