package conventions

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	//SysHeaderPrefix common of system GRPC metadata and HTTP headers
	SysHeaderPrefix = "x-sbr-"
)

const (
	//UserAgentHeader web requests
	UserAgentHeader = "user-agent"

	//RemoteHostnameHeader -
	RemoteHostnameHeader = "hostname"

	//UserAgentHeaderNoVer - because [UserAgentHeader] contains grpc version as "foo grpc-go/1.53.0"
	UserAgentHeaderNoVer = "user-agent-no-ver"
)

const (
	//LoggerLevelHeader notes to change log level in current context of operation
	LoggerLevelHeader = SysHeaderPrefix + "log-lvl"

	//AppNameHeader holds application name for incoming outgoing requests
	AppNameHeader = SysHeaderPrefix + "app-name"

	//AppVersionHeader holds application version for incoming outgoing requests
	AppVersionHeader = SysHeaderPrefix + "app-ver"
)

//ClientName user agent extractor
var ClientName clientNameExtractor

//Incoming extracts user agent from incoming context
func (a clientNameExtractor) Incoming(ctx context.Context, defVal string) string {
	if ret, ok := a.extractClientName(ctx, a.mdIncoming); ok {
		return ret
	}
	return defVal
}

//Outgoing extracts user agent from outgoing context
func (a clientNameExtractor) Outgoing(ctx context.Context, defVal string) string {
	if ret, ok := a.extractClientName(ctx, a.mdOutgoing); ok {
		return ret
	}
	return defVal
}

type clientNameExtractor struct{}

func (clientNameExtractor) mdIncoming(ctx context.Context) metadata.MD {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		return md
	}
	return nil
}

func (clientNameExtractor) mdOutgoing(ctx context.Context) metadata.MD {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		return md
	}
	return nil
}

func (clientNameExtractor) extractClientName(ctx context.Context, mdExtractor func(ctx context.Context) metadata.MD) (string, bool) {
	var (
		res string
		ok  bool
	)
	if md := mdExtractor(ctx); md != nil {
		for _, k := range []string{AppNameHeader, UserAgentHeader} {
			if v := md[k]; len(v) > 0 {
				res, ok = v[0], true
				break
			}
		}
	}
	const suffix = "grpc-go/"
	if ok && len(res) > 0 {
		n := strings.Index(res, suffix)
		if n > 0 {
			res = strings.TrimRight(res[:n], " ")
		}
	}
	return res, ok
}
