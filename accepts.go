package httpcompression

import (
	"math"
	"strconv"
	"strings"
)

const (
	// defaultQValue is the default qvalue to assign to an encoding if no explicit qvalue is set.
	// This is actually kind of ambiguous in RFC 2616, so hopefully it's correct.
	// The examples seem to indicate that it is.
	defaultQValue = 1.0
)

// encoding picks the content-encoding to use for the response
func encoding(acceptEncoding string, compressors comps, prefer PreferType) string {
	var bestEncoding string
	var bestQvalue float64
	var bestPriority int

	for _, ss := range strings.Split(acceptEncoding, ",") {
		encoding, qvalue := parseCoding(ss)
		if encoding == "" {
			continue
		}
		if qvalue == 0 {
			continue
		}
		compressor, ok := compressors[encoding]
		if !ok {
			continue
		}

		if bestEncoding == "" {
			bestEncoding, bestQvalue, bestPriority = encoding, qvalue, compressor.priority
		} else {
			if prefer == PreferServer {
				if bestPriority < compressor.priority ||
					(bestPriority == compressor.priority && bestQvalue < qvalue) ||
					(bestPriority == compressor.priority && bestQvalue == qvalue && strings.Compare(bestEncoding, encoding) < 0) {
					bestEncoding, bestQvalue, bestPriority = encoding, qvalue, compressor.priority
				}
			} else {
				if bestQvalue < qvalue ||
					(bestQvalue == qvalue && bestPriority < compressor.priority) ||
					(bestQvalue == qvalue && bestPriority == compressor.priority && strings.Compare(bestEncoding, encoding) < 0) {
					bestEncoding, bestQvalue, bestPriority = encoding, qvalue, compressor.priority
				}
			}
		}
	}

	return bestEncoding
}

// parseCoding parses a single conding (content-coding with an optional qvalue),
// as might appear in an Accept-Encoding header. It attempts to forgive minor
// formatting errors.
func parseCoding(s string) (coding string, qvalue float64) {
	qvalue = defaultQValue

	p := strings.IndexRune(s, ';')
	if p != -1 {
		if part := strings.Replace(s[p+1:], " ", "", -1); strings.HasPrefix(part, "q=") {
			qvalue, _ = strconv.ParseFloat(part[2:], 64)
			if qvalue < 0.0 || math.IsNaN(qvalue) {
				qvalue = 0.0
			} else if qvalue > 1.0 {
				qvalue = 1.0
			}
		}
	} else {
		p = len(s)
	}
	coding = strings.ToLower(strings.TrimSpace(s[:p]))
	return
}
