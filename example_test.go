package httpcompression_test

import (
	"log"
	"net/http"

	"github.com/nako-ruru/httpcompression"
	"github.com/nako-ruru/httpcompression/contrib/andybalholm/brotli"
)

func Example() {
	// Create a compression adapter with default configuration
	compress, err := httpcompression.DefaultAdapter()
	if err != nil {
		log.Fatal(err)
	}
	// Define your handler, and apply the compression adapter.
	http.Handle("/", compress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	})))
	// ...
}

func ExampleCustom() {
	brEnc, err := brotli.New(brotli.Options{})
	if err != nil {
		log.Fatal(err)
	}
	_, _ = httpcompression.Adapter(
		httpcompression.Compressor(brotli.Encoding, 1, brEnc),
		httpcompression.Prefer(httpcompression.PreferServer),
		httpcompression.MinSize(100),
		httpcompression.ContentTypes([]string{
			"image/jpeg",
			"image/gif",
			"image/png",
		}, true),
	)
}

func ExampleWithDictionary() {
	_, _ = httpcompression.DefaultAdapter(
		// We need to pick a custom content-encoding name. It is recommended to:
		// - avoid names that contain standard names (e.g. "gzip", "deflate" or "br")
		// - include the dictionary ID, so that multiple dictionaries can be used (including
		//   e.g. multiple versions of the same dictionary)
		httpcompression.Prefer(httpcompression.PreferServer),
		httpcompression.MinSize(0),
		httpcompression.ContentTypes([]string{
			"image/jpeg",
			"image/gif",
			"image/png",
		}, true),
	)
}
