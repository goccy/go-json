package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/goccy/go-json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/tools/benchmark/parse"
	"golang.org/x/xerrors"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/api/slides/v1"
)

const (
	benchgraphToken = "benchgraph-token.json"
)

type BenchmarkCodec string

const (
	UnknownCodec BenchmarkCodec = "Unknown"
	Encoder      BenchmarkCodec = "Encode"
	Decoder      BenchmarkCodec = "Decode"
)

func (c BenchmarkCodec) String() string {
	return string(c)
}

type BenchmarkKind string

const (
	UnknownKind        BenchmarkKind = "Unknown"
	SmallStruct        BenchmarkKind = "SmallStruct"
	SmallStructCached  BenchmarkKind = "SmallStructCached"
	MediumStruct       BenchmarkKind = "MediumStruct"
	MediumStructCached BenchmarkKind = "MediumStructCached"
	LargeStruct        BenchmarkKind = "LargeStruct"
	LargeStructCached  BenchmarkKind = "LargeStructCached"
)

func (k BenchmarkKind) String() string {
	return string(k)
}

type BenchmarkTarget string

const (
	UnknownTarget  BenchmarkTarget = "Unknown"
	EncodingJson   BenchmarkTarget = "EncodingJson"
	GoJson         BenchmarkTarget = "GoJson"
	GoJsonNoEscape BenchmarkTarget = "GoJsonNoEscape"
	GoJsonColored  BenchmarkTarget = "GoJsonColored"
	FFJson         BenchmarkTarget = "FFJson"
	JsonIter       BenchmarkTarget = "JsonIter"
	EasyJson       BenchmarkTarget = "EasyJson"
	Jettison       BenchmarkTarget = "Jettison"
	GoJay          BenchmarkTarget = "GoJay"
	SegmentioJson  BenchmarkTarget = "SegmentioJson"
)

func (t BenchmarkTarget) String() string {
	return string(t)
}

func (t BenchmarkTarget) DisplayName() string {
	switch t {
	case EncodingJson:
		return "encoding/json"
	case GoJson:
		return "goccy/go-json"
	case GoJsonNoEscape:
		return "goccy/go-json (noescape)"
	case GoJsonColored:
		return "goccy/go-json (colored)"
	case FFJson:
		return "pquerna/ffjson"
	case JsonIter:
		return "json-iterator/go"
	case EasyJson:
		return "mailru/easyjson"
	case Jettison:
		return "wl2L/jettison"
	case GoJay:
		return "francoispqt/gojay"
	case SegmentioJson:
		return "segmentio/encoding/json"
	default:
		return ""
	}
}

var credFile string

func init() {
	// How to create credentials.json: https://developers.google.com/workspace/guides/create-credentials#desktop
	flag.StringVar(&credFile, "cred", "credentials.json", "specify to path to credential file to use google spreadsheets and slides APIs")
}

func stringptr(s string) *string {
	return &s
}

func floatptr(v float64) *float64 {
	return &v
}

func createClient(ctx context.Context) (*http.Client, error) {
	b, err := ioutil.ReadFile(credFile)
	if err != nil {
		return nil, xerrors.Errorf("failed to read credential file: %w", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, sheets.SpreadsheetsScope, slides.PresentationsScope)
	if err != nil {
		return nil, xerrors.Errorf("failed to create config from scope: %w", err)
	}

	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first time.
	tok, err := tokenFromFile()
	if err != nil {
		tok = getTokenFromWeb(ctx, config)
		if err := saveToken(tok); err != nil {
			return nil, xerrors.Errorf("failed to save token: %w", err)
		}
	}
	return config.Client(ctx, tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFromFile() (*oauth2.Token, error) {
	f, err := os.Open(benchgraphToken)
	if err != nil {
		return nil, xerrors.Errorf("failed to open %s: %w", benchgraphToken, err)
	}
	defer f.Close()
	tok := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(tok); err != nil {
		return nil, xerrors.Errorf("failed to decode token: %w", err)
	}
	return tok, nil
}

func saveToken(token *oauth2.Token) error {
	log.Printf("Saving credential file to: %s\n", benchgraphToken)
	f, err := os.OpenFile(benchgraphToken, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return xerrors.Errorf("failed to create file for saving token: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(token); err != nil {
		return xerrors.Errorf("failed to write token: %w", err)
	}
	return nil
}

func createSpreadsheet(svc *sheets.Service, title string, headers []string, targets []string, data map[string][]float64) (*sheets.Spreadsheet, error) {
	rows := []*sheets.RowData{}
	headerRow := &sheets.RowData{Values: []*sheets.CellData{{UserEnteredValue: &sheets.ExtendedValue{}}}}
	for _, header := range headers {
		headerRow.Values = append(headerRow.Values, &sheets.CellData{
			UserEnteredValue: &sheets.ExtendedValue{
				StringValue: stringptr(header),
			},
		})
	}
	rows = append(rows, headerRow)
	for _, target := range targets {
		target := target
		targetRow := &sheets.RowData{
			Values: []*sheets.CellData{
				{
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: stringptr(target),
					},
				},
			},
		}
		for _, v := range data[target] {
			v := v
			targetRow.Values = append(targetRow.Values, &sheets.CellData{
				UserEnteredValue: &sheets.ExtendedValue{
					NumberValue: floatptr(v),
				},
			})
		}
		rows = append(rows, targetRow)
	}
	sheetsData := []*sheets.Sheet{
		{
			Properties: &sheets.SheetProperties{
				Title:     "benchmark",
				SheetType: "GRID",
			},
			Data: []*sheets.GridData{{RowData: rows}},
		},
	}
	spreadSheet, err := svc.Spreadsheets.Create(&sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{Title: title},
		Sheets:     sheetsData,
	}).Do()
	if err != nil {
		return nil, xerrors.Errorf("failed to create spreadsheet: %w", err)
	}
	spreadSheet.Sheets[0].Data = sheetsData[0].Data
	return spreadSheet, nil
}

func generateChart(svc *sheets.Service, spreadSheet *sheets.Spreadsheet, title string) (*sheets.EmbeddedChart, error) {
	sheet := spreadSheet.Sheets[0]
	rows := sheet.Data[0].RowData
	rowSize := int64(len(rows))
	colSize := int64(len(rows[0].Values))
	series := []*sheets.BasicChartSeries{}
	for i := int64(1); i < colSize; i++ {
		series = append(series, &sheets.BasicChartSeries{
			Series: &sheets.ChartData{
				SourceRange: &sheets.ChartSourceRange{
					Sources: []*sheets.GridRange{
						{
							SheetId:          sheet.Properties.SheetId,
							StartColumnIndex: i,
							EndColumnIndex:   i + 1,
							StartRowIndex:    0,
							EndRowIndex:      rowSize,
						},
					},
				},
			},
			TargetAxis: "BOTTOM_AXIS",
		})
	}
	res, err := svc.Spreadsheets.BatchUpdate(spreadSheet.SpreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddChart: &sheets.AddChartRequest{
					Chart: &sheets.EmbeddedChart{
						Position: &sheets.EmbeddedObjectPosition{
							NewSheet: true,
						},
						Spec: &sheets.ChartSpec{
							Title: title,
							BasicChart: &sheets.BasicChartSpec{
								ChartType:      "BAR",
								LegendPosition: "TOP_LEGEND",
								Domains: []*sheets.BasicChartDomain{
									{
										Domain: &sheets.ChartData{
											SourceRange: &sheets.ChartSourceRange{
												Sources: []*sheets.GridRange{
													{
														SheetId:          sheet.Properties.SheetId,
														StartColumnIndex: 0,
														EndColumnIndex:   1,
														StartRowIndex:    0,
														EndRowIndex:      rowSize,
													},
												},
											},
										},
									},
								},
								Series:      series,
								HeaderCount: 1,
							},
						},
					},
				},
			},
		},
	}).Do()
	if err != nil {
		return nil, xerrors.Errorf("failed to generate chart: %w", err)
	}

	for _, rep := range res.Replies {
		if rep.AddChart == nil {
			continue
		}
		return rep.AddChart.Chart, nil
	}
	return nil, xerrors.Errorf("failed to find chartID")
}

func createPresentationWithEmptySlide(presentationService *slides.PresentationsService) (*slides.Presentation, string, error) {
	presentation, err := presentationService.Create(&slides.Presentation{}).Do()
	if err != nil {
		return nil, "", xerrors.Errorf("failed to create presentation: %w", err)
	}
	res, err := presentationService.BatchUpdate(presentation.PresentationId, &slides.BatchUpdatePresentationRequest{
		Requests: []*slides.Request{
			{
				CreateSlide: &slides.CreateSlideRequest{
					InsertionIndex: 0,
				},
			},
		},
	}).Do()
	if err != nil {
		return nil, "", xerrors.Errorf("failed to update presentation: %w", err)
	}
	for _, rep := range res.Replies {
		if rep.CreateSlide == nil {
			continue
		}
		return presentation, rep.CreateSlide.ObjectId, nil
	}
	return nil, "", xerrors.Errorf("failed to find slide objectID")
}

func addChartToPresentation(presentationService *slides.PresentationsService, presentation *slides.Presentation, slideID string, spreadSheetID string, chart *sheets.EmbeddedChart) error {
	if _, err := presentationService.BatchUpdate(presentation.PresentationId, &slides.BatchUpdatePresentationRequest{
		Requests: []*slides.Request{
			{
				CreateSheetsChart: &slides.CreateSheetsChartRequest{
					LinkingMode:   "LINKED",
					SpreadsheetId: spreadSheetID,
					ChartId:       chart.ChartId,
					ElementProperties: &slides.PageElementProperties{
						PageObjectId: slideID,
						Size: &slides.Size{
							Width: &slides.Dimension{
								Magnitude: 1200,
								Unit:      "PT",
							},
							Height: &slides.Dimension{
								Magnitude: 800,
								Unit:      "PT",
							},
						},
					},
				},
			},
		},
	}).Do(); err != nil {
		return xerrors.Errorf("failed to add chart: %w", err)
	}
	return nil
}

func downloadChartImage(presentationService *slides.PresentationsService, presentation *slides.Presentation, path string) error {
	gotPresentation, err := presentationService.Get(presentation.PresentationId).Do()
	if err != nil {
		return xerrors.Errorf("failed to get presentation: %w", err)
	}
	for _, slide := range gotPresentation.Slides {
		for _, pe := range slide.PageElements {
			if pe.SheetsChart == nil {
				continue
			}
			if err := downloadImage(pe.SheetsChart.ContentUrl, path); err != nil {
				return xerrors.Errorf("failed to download image: %w", err)
			}
		}
	}
	return nil
}

func downloadImage(url, path string) error {
	res, err := http.Get(url)
	if err != nil {
		return xerrors.Errorf("failed to get content from %s: %w", url, err)
	}
	defer res.Body.Close()
	file, err := os.Create(path)
	if err != nil {
		return xerrors.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, res.Body); err != nil {
		return xerrors.Errorf("failed to copy file: %w", err)
	}
	return nil
}

type BenchmarkData struct {
	*parse.Benchmark
	Codec  BenchmarkCodec
	Kind   BenchmarkKind
	Target BenchmarkTarget
}

func toBenchmarkCodec(name string) (BenchmarkCodec, error) {
	switch {
	case strings.Contains(name, "_Encode_"):
		return Encoder, nil
	case strings.Contains(name, "_Decoder_"):
		return Decoder, nil
	default:
		return UnknownCodec, xerrors.Errorf("cannot find codec from %s", name)
	}
}

func toBenchmarkKind(name string) (BenchmarkKind, error) {
	switch {
	case strings.Contains(name, "_SmallStruct_"):
		return SmallStruct, nil
	case strings.Contains(name, "_MediumStruct_"):
		return MediumStruct, nil
	case strings.Contains(name, "_LargeStruct_"):
		return LargeStruct, nil
	case strings.Contains(name, "_SmallStructCached_"):
		return SmallStructCached, nil
	case strings.Contains(name, "_MediumStructCached_"):
		return MediumStructCached, nil
	case strings.Contains(name, "_LargeStructCached_"):
		return LargeStructCached, nil
	default:
		return UnknownKind, xerrors.Errorf("cannot find kind from %s", name)
	}
}

func toBenchmarkTarget(name string) (BenchmarkTarget, error) {
	v := strings.ToLower(name)
	switch {
	case strings.Contains(v, strings.ToLower("EncodingJson")):
		return EncodingJson, nil
	case strings.Contains(v, strings.ToLower("GoJson")):
		switch {
		case strings.Contains(v, strings.ToLower("NoEscape")):
			return GoJsonNoEscape, nil
		case strings.Contains(v, strings.ToLower("Colored")):
			return GoJsonColored, nil
		}
		return GoJson, nil
	case strings.Contains(v, strings.ToLower("FFJson")):
		return FFJson, nil
	case strings.Contains(v, strings.ToLower("JsonIter")):
		return JsonIter, nil
	case strings.Contains(v, strings.ToLower("EasyJson")):
		return EasyJson, nil
	case strings.Contains(v, strings.ToLower("Jettison")):
		return Jettison, nil
	case strings.Contains(v, strings.ToLower("GoJay")):
		return GoJay, nil
	case strings.Contains(v, strings.ToLower("SegmentioJson")):
		return SegmentioJson, nil
	default:
		return UnknownTarget, xerrors.Errorf("cannot find target from %s", name)
	}
}

func createBenchmarkData(bench *parse.Benchmark) (*BenchmarkData, error) {
	codec, err := toBenchmarkCodec(bench.Name)
	if err != nil {
		return nil, xerrors.Errorf("failed to convert benchmark codec: %w", err)
	}
	kind, err := toBenchmarkKind(bench.Name)
	if err != nil {
		return nil, xerrors.Errorf("failed to convert benchmark kind: %w", err)
	}
	target, err := toBenchmarkTarget(bench.Name)
	if err != nil {
		return nil, xerrors.Errorf("failed to convert benchmark target: %w", err)
	}
	return &BenchmarkData{
		Benchmark: bench,
		Codec:     codec,
		Kind:      kind,
		Target:    target,
	}, nil
}

func createAllBenchmarkData(data string) ([]*BenchmarkData, error) {
	allBenchData := []*BenchmarkData{}
	for _, line := range strings.Split(data, "\n") {
		if !strings.HasPrefix(line, "Benchmark") {
			continue
		}
		bench, err := parse.ParseLine(line)
		if err != nil {
			return nil, xerrors.Errorf("failed to parse line: %w", err)
		}
		benchData, err := createBenchmarkData(bench)
		if err != nil {
			return nil, xerrors.Errorf("failed to create benchmark data: %w", err)
		}
		allBenchData = append(allBenchData, benchData)
	}
	return allBenchData, nil
}

type Graph struct {
	Title string
	Codec BenchmarkCodec
	Kinds []BenchmarkKind
}

func (g *Graph) existsKind(target BenchmarkKind) bool {
	for _, kind := range g.Kinds {
		if kind == target {
			return true
		}
	}
	return false
}

func generateBenchmarkGraph(ctx context.Context, client *http.Client, g *Graph, data []*BenchmarkData) error {
	headers := []string{}
	for _, kind := range g.Kinds {
		headers = append(headers, kind.String())
	}
	targetMap := map[string][]*BenchmarkData{}
	targetToData := map[string][]float64{}
	for _, v := range data {
		if g.Codec != v.Codec {
			continue
		}
		if !g.existsKind(v.Kind) {
			continue
		}
		name := v.Target.DisplayName()
		targetMap[name] = append(targetMap[name], v)
		targetToData[name] = append(targetToData[name], v.NsPerOp)
	}
	targets := []string{}
	for k := range targetMap {
		targets = append(targets, k)
	}

	sheetService, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return xerrors.Errorf("failed to create service for sheet: %w", err)
	}
	spreadSheet, err := createSpreadsheet(sheetService, g.Title, headers, targets, targetToData)
	if err != nil {
		return xerrors.Errorf("failed to create spreadsheet: %w", err)
	}
	chart, err := generateChart(sheetService, spreadSheet, g.Title)
	if err != nil {
		return xerrors.Errorf("failed to generate chart: %w", err)
	}
	log.Println("spreadSheetID = ", spreadSheet.SpreadsheetId, "chartID = ", chart.ChartId)

	slideService, err := slides.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return xerrors.Errorf("failed to create service for slide: %w", err)
	}
	presentationService := slides.NewPresentationsService(slideService)
	presentation, slideID, err := createPresentationWithEmptySlide(presentationService)
	if err != nil {
		return xerrors.Errorf("failed to create presentation with slide: %w", err)
	}
	if err := addChartToPresentation(presentationService, presentation, slideID, spreadSheet.SpreadsheetId, chart); err != nil {
		return xerrors.Errorf("failed to add chart to presentation: %w", err)
	}
	if err := downloadChartImage(presentationService, presentation, "bench.png"); err != nil {
		return xerrors.Errorf("failed to download chart image: %w", err)
	}
	return nil
}

func run(args []string) error {
	benchData, err := createAllBenchmarkData(`
goos: darwin
goarch: amd64
pkg: benchmark
Benchmark_Encode_SmallStruct_EncodingJson-16                     2135164               555 ns/op             256 B/op          2 allocs/op
Benchmark_Encode_SmallStruct_FFJson-16                           1348426               935 ns/op             584 B/op          9 allocs/op
Benchmark_Encode_SmallStruct_JsonIter-16                         1970002               598 ns/op             264 B/op          3 allocs/op
Benchmark_Encode_SmallStruct_EasyJson-16                         2202872               547 ns/op             720 B/op          4 allocs/op
Benchmark_Encode_SmallStruct_Jettison-16                         2610375               461 ns/op             256 B/op          2 allocs/op
Benchmark_Encode_SmallStruct_GoJay-16                            2763138               428 ns/op             624 B/op          2 allocs/op
Benchmark_Encode_SmallStruct_SegmentioJson-16                    4124536               291 ns/op             256 B/op          2 allocs/op
Benchmark_Encode_SmallStruct_GoJsonColored-16                    2530636               475 ns/op             432 B/op          2 allocs/op
Benchmark_Encode_SmallStruct_GoJson-16                           4308301               282 ns/op             256 B/op          2 allocs/op
Benchmark_Encode_SmallStruct_GoJsonNoEscape-16                   5406490               215 ns/op             144 B/op          1 allocs/op
Benchmark_Encode_SmallStructCached_EncodingJson-16               2386401               510 ns/op             144 B/op          1 allocs/op
Benchmark_Encode_SmallStructCached_FFJson-16                     1450132               829 ns/op             472 B/op          8 allocs/op
Benchmark_Encode_SmallStructCached_JsonIter-16                   2279529               526 ns/op             152 B/op          2 allocs/op
Benchmark_Encode_SmallStructCached_EasyJson-16                   2225763               543 ns/op             720 B/op          4 allocs/op
Benchmark_Encode_SmallStructCached_Jettison-16                   3059923               387 ns/op             144 B/op          1 allocs/op
Benchmark_Encode_SmallStructCached_GoJay-16                      3187108               372 ns/op             512 B/op          1 allocs/op
Benchmark_Encode_SmallStructCached_SegmentioJson-16              5128329               229 ns/op             144 B/op          1 allocs/op
Benchmark_Encode_SmallStructCached_GoJsonColored-16              3028646               403 ns/op             320 B/op          1 allocs/op
Benchmark_Encode_SmallStructCached_GoJson-16                     5458942               215 ns/op             144 B/op          1 allocs/op
Benchmark_Encode_SmallStructCached_GoJsonNoEscape-16             5725311               210 ns/op             144 B/op          1 allocs/op
PASS
ok      benchmark       33.928s
`)
	if err != nil {
		return xerrors.Errorf("failed to parse benchmark data: %w", err)
	}
	if benchData == nil {
		return nil
	}
	ctx := context.Background()
	client, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("failed to create client: %w", err)
	}
	graphs := []*Graph{
		{
			Title: "Encoding Benchmark ( Small / Medium Struct )",
			Codec: Encoder,
			Kinds: []BenchmarkKind{SmallStruct, MediumStruct},
		},
	}
	for _, graph := range graphs {
		if err := generateBenchmarkGraph(ctx, client, graph, benchData); err != nil {
			return xerrors.Errorf("failed to generate benchmark graph: %w", err)
		}
	}
	return nil
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatalf("%+v", err)
	}
}
