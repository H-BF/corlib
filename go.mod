module github.com/H-BF/corlib

go 1.21

require (
	github.com/ahmetb/go-linq/v3 v3.2.0
	github.com/c-robinson/iplib v1.0.8
	github.com/cenkalti/backoff/v4 v4.1.3
	github.com/emirpasic/gods v1.18.1
	github.com/go-chi/chi v1.5.4
	github.com/go-openapi/spec v0.20.6
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/google/nftables v0.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.2
	github.com/hashicorp/go-retryablehttp v0.7.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/miekg/dns v1.1.61
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common v0.39.0
	github.com/rakyll/statik v0.1.7
	github.com/satori/go.uuid v1.2.0
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cast v1.6.0
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.9.0
	github.com/vishvananda/netlink v1.3.0
	github.com/vishvananda/netns v0.0.4
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/exporters/jaeger v1.9.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.9.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.9.0
	go.opentelemetry.io/otel/sdk v1.9.0
	go.opentelemetry.io/otel/trace v1.24.0
	go.uber.org/multierr v1.9.0
	go.uber.org/zap v1.22.0
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842
	golang.org/x/sys v0.21.0
	google.golang.org/genproto/googleapis/api v0.0.0-20240311132316-a219d84964c2
	google.golang.org/grpc v1.62.1
	google.golang.org/protobuf v1.33.0
)

replace github.com/google/nftables v0.3.0 => github.com/H-BF/nftables v0.3.0-dev

replace github.com/vishvananda/netlink v1.3.0 => github.com/H-BF/netlink v1.3.0-dev

//TODO: need upgrade to actual version
replace go.opentelemetry.io/otel v1.24.0 => go.opentelemetry.io/otel v1.9.0

//TODO: need upgrade to actual version
replace go.opentelemetry.io/otel/trace v1.24.0 => go.opentelemetry.io/otel/trace v1.9.0

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mdlayher/netlink v1.7.2 // indirect
	github.com/mdlayher/socket v0.5.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.opentelemetry.io/proto/otlp v0.18.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	google.golang.org/genproto v0.0.0-20240213162025-012b6fc9bca9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240314234333-6e1732d8331c // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
