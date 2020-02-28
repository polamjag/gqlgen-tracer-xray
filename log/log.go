package log

import (
	stdlog "log"
	"os"
	"strings"

	"github.com/aws/aws-xray-sdk-go/xray"
)

var Logger = stdlog.New(os.Stderr, "", stdlog.LstdFlags|stdlog.Lmicroseconds)

func Start(seg *xray.Segment) {
	Logger.Printf("%s > Start %s %s", prefix(depthOf(seg)), seg.Name, seg.ID)
}

func Close(seg *xray.Segment, atDefer bool) {
	Logger.Printf("%s < Close (at defer?=%v) %s %s", prefix(depthOf(seg)), atDefer, seg.Name, seg.ID)
}

func depthOf(seg *xray.Segment) int {
	meta, ok := seg.Metadata["default"]
	if !ok {
		return 0
	}
	// Logger.Printf("segment name=%s metadata=%#v", seg.Name, meta)
	if _, ok := meta["gql.variables"]; ok {
		// maybe operation
		return 0
	}
	obj := meta["gql.object"]
	if obj.(string) == "Query" {
		return 1
	}
	return 2
}

func prefix(depth int) string {
	return strings.Repeat("  ", depth)
}
