//go:build go1.25

package benchmark

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	gojson "github.com/goccy/go-json"
	gojsonv2 "github.com/goccy/go-json/v2"
)

// The responses of an HTTP server: MarshalWrite of github.com/goccy/go-json/v2 writes a value to the
// http.ResponseWriter from the buffer of the encoder, without the copy of the result which Marshal returns, and
// MarshalWriteOf takes the value by its type, without the copy of it to an interface value: a value is written
// without an allocation, also by the MarshalJSONTo methods of its types.

// discardResponse is an http.ResponseWriter which keeps nothing, as the writer of a connection which is flushed.
type discardResponse struct{ header http.Header }

func (r *discardResponse) Header() http.Header         { return r.header }
func (r *discardResponse) Write(b []byte) (int, error) { return len(b), nil }
func (r *discardResponse) WriteHeader(int)             {}

func BenchmarkHTTPResponseWriter(b *testing.B) {
	var w http.ResponseWriter = &discardResponse{header: http.Header{"Content-Type": {"application/json"}}}
	small := *NewSmallPayload()
	orders := newOrders[orderID, point]()
	write := func(b *testing.B, fn func() error) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if err := fn(); err != nil {
				b.Fatal(err)
			}
		}
	}
	b.Run("Small/GoJson/Marshal", func(b *testing.B) {
		write(b, func() error {
			out, err := gojson.Marshal(small)
			if err == nil {
				_, err = w.Write(out)
			}
			return err
		})
	})
	b.Run("Small/GoJson/MarshalOf", func(b *testing.B) {
		write(b, func() error {
			out, err := gojson.MarshalOf(small)
			if err == nil {
				_, err = w.Write(out)
			}
			return err
		})
	})
	b.Run("Small/GoJsonV2/MarshalWrite", func(b *testing.B) {
		write(b, func() error { return gojsonv2.MarshalWrite(w, small) })
	})
	b.Run("Small/GoJsonV2/MarshalWriteOf", func(b *testing.B) {
		write(b, func() error { return gojsonv2.MarshalWriteOf(w, small) })
	})
	b.Run("Orders/GoJson/MarshalOf", func(b *testing.B) {
		write(b, func() error {
			out, err := gojson.MarshalOf(orders)
			if err == nil {
				_, err = w.Write(out)
			}
			return err
		})
	})
	b.Run("Orders/GoJsonV2/MarshalWriteOf", func(b *testing.B) {
		write(b, func() error { return gojsonv2.MarshalWriteOf(w, orders) })
	})
}

// BenchmarkHTTPServer serves the values through net/http: the allocations of the server and of the client are the
// ones of the handlers which write bytes encoded once, which the others are compared with. mallocs/op counts the
// allocations which happen once in some requests, which allocs/op rounds: the contexts of the encoder, which a GC
// drops from their pool, and their buffers, which grow again.
func BenchmarkHTTPServer(b *testing.B) {
	small := *NewSmallPayload()
	encodedSmall, err := gojsonv2.Marshal(small)
	if err != nil {
		b.Fatal(err)
	}
	orders := newOrders[orderID, point]()
	encodedOrders, err := gojsonv2.Marshal(orders)
	if err != nil {
		b.Fatal(err)
	}
	for _, h := range []struct {
		name string
		fn   http.HandlerFunc
	}{
		{"Small/Bytes", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(encodedSmall) }},
		{"Small/GoJsonV2/MarshalWriteOf", func(w http.ResponseWriter, _ *http.Request) { _ = gojsonv2.MarshalWriteOf(w, small) }},
		{"Orders/Bytes", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(encodedOrders) }},
		{"Orders/GoJsonV2/MarshalWriteOf", func(w http.ResponseWriter, _ *http.Request) { _ = gojsonv2.MarshalWriteOf(w, orders) }},
	} {
		b.Run(h.name, func(b *testing.B) {
			srv := httptest.NewServer(h.fn)
			defer srv.Close()
			client := srv.Client()
			req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			if err != nil {
				b.Fatal(err)
			}
			buf := make([]byte, 64<<10)
			b.ReportAllocs()
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			for i := 0; i < b.N; i++ {
				resp, err := client.Do(req)
				if err != nil {
					b.Fatal(err)
				}
				for {
					if _, err := resp.Body.Read(buf); err != nil {
						break
					}
				}
				_ = resp.Body.Close()
			}
			runtime.ReadMemStats(&after)
			b.ReportMetric(float64(after.Mallocs-before.Mallocs)/float64(b.N), "mallocs/op")
		})
	}
}
