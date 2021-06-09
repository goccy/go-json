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

	"github.com/goccy/go-json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/xerrors"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/api/slides/v1"
)

const (
	benchgraphToken = "benchgraph-token.json"
)

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
	config, err := google.ConfigFromJSON(b, sheets.SpreadsheetsScope, slides.PresentationsScope) //"https://spreadsheets.google.com/feeds"
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

func generateChart(svc *sheets.Service, spreadSheet *sheets.Spreadsheet, title string) (int64, error) {
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
								LegendPosition: "RIGHT_LEGEND",
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
		return 0, xerrors.Errorf("failed to generate chart: %w", err)
	}

	for _, rep := range res.Replies {
		if rep.AddChart == nil {
			continue
		}
		return rep.AddChart.Chart.ChartId, nil
	}
	return 0, xerrors.Errorf("failed to find chartID")
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

func addChartToPresentation(presentationService *slides.PresentationsService, presentation *slides.Presentation, slideID string, spreadSheetID string, chartID int64) error {
	if _, err := presentationService.BatchUpdate(presentation.PresentationId, &slides.BatchUpdatePresentationRequest{
		Requests: []*slides.Request{
			{
				CreateSheetsChart: &slides.CreateSheetsChartRequest{
					LinkingMode:   "LINKED",
					SpreadsheetId: spreadSheetID,
					ChartId:       chartID,
					ElementProperties: &slides.PageElementProperties{
						PageObjectId: slideID,
						Size: &slides.Size{
							Width: &slides.Dimension{
								Magnitude: 1024,
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

func _main(args []string) error {
	ctx := context.Background()
	client, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("failed to create client: %w", err)
	}

	sheetService, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return xerrors.Errorf("failed to create service for sheet: %w", err)
	}
	headers := []string{"TypeA", "TypeB", "TypeC"}
	targets := []string{"targetA", "targetB"}
	data := map[string][]float64{
		targets[0]: []float64{10, 100, 1000},
		targets[1]: []float64{20, 200, 2000},
	}
	spreadSheet, err := createSpreadsheet(sheetService, "go-json benchmark results", headers, targets, data)
	if err != nil {
		return xerrors.Errorf("failed to create spreadsheet: %w", err)
	}
	chartID, err := generateChart(sheetService, spreadSheet, "Benchmark Result")
	if err != nil {
		return xerrors.Errorf("failed to generate chart: %w", err)
	}
	log.Println("spreadSheetID = ", spreadSheet.SpreadsheetId, "chartID = ", chartID)

	slideService, err := slides.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return xerrors.Errorf("failed to create service for slide: %w", err)
	}
	presentationService := slides.NewPresentationsService(slideService)
	presentation, slideID, err := createPresentationWithEmptySlide(presentationService)
	if err != nil {
		return xerrors.Errorf("failed to create presentation with slide: %w", err)
	}
	if err := addChartToPresentation(presentationService, presentation, slideID, spreadSheet.SpreadsheetId, chartID); err != nil {
		return xerrors.Errorf("failed to add chart to presentation: %w", err)
	}
	if err := downloadChartImage(presentationService, presentation, "bench.png"); err != nil {
		return xerrors.Errorf("failed to download chart image: %w", err)
	}
	return nil
}

func main() {
	if err := _main(os.Args); err != nil {
		log.Fatalf("%+v", err)
	}
}
