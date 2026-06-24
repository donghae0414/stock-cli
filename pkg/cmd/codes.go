package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"

	"stock-cli/pkg/config"
	"stock-cli/pkg/kiwoom"
	"stock-cli/pkg/stocklookup"
)

const defaultLookupLimit = 10

var newStockInfoClient = func(creds config.Credentials) stockInfoClient {
	return kiwoom.NewClient(creds.AppKey, creds.SecretKey)
}

var codesCmd = cli.Command{
	Name:     "codes",
	Usage:    "Resolve Kiwoom stock codes",
	Category: "API RESOURCE",
	Suggest:  true,
	Commands: []*cli.Command{
		&codesLookupCmd,
	},
}

var codesLookupCmd = cli.Command{
	Name:            "lookup",
	Usage:           "Resolve stock names to six-digit stock codes",
	Suggest:         true,
	SkipFlagParsing: true,
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "name",
			Usage: "Stock name query; repeat for multiple names",
		},
		&cli.StringFlag{
			Name:  "limit",
			Usage: "Maximum candidates per query",
			Value: strconv.Itoa(defaultLookupLimit),
		},
	},
	Action:          handleCodesLookup,
	HideHelpCommand: true,
}

type stockInfoClient interface {
	StockInfoRows(context.Context, []string) ([]kiwoom.StockInfoRow, error)
}

type codesLookupOptions struct {
	Names []string
	Limit string
}

type lookupValidationError struct {
	message string
}

func (e lookupValidationError) Error() string {
	return e.message
}

type lookupConfigError struct {
	err error
}

func (e lookupConfigError) Error() string {
	return e.err.Error()
}

func (e lookupConfigError) Unwrap() error {
	return e.err
}

type lookupUpstreamError struct {
	err error
}

func (e lookupUpstreamError) Error() string {
	return e.err.Error()
}

func (e lookupUpstreamError) Unwrap() error {
	return e.err
}

func handleCodesLookup(ctx context.Context, cmd *cli.Command) error {
	opts, unusedArgs, showHelp := parseCodesLookupArgs(cmd.Args().Slice())
	if showHelp {
		showCodesLookupHelp(cmd)
		return nil
	}
	return runCodesLookup(ctx, opts, unusedArgs)
}

func showCodesLookupHelp(cmd *cli.Command) {
	tmpl := cmd.CustomHelpTemplate
	if tmpl == "" {
		tmpl = cli.CommandHelpTemplate
	}
	cli.HelpPrinter(cmd.Root().Writer, tmpl, cmd)
}

func parseCodesLookupArgs(args []string) (codesLookupOptions, []string, bool) {
	opts := codesLookupOptions{Limit: strconv.Itoa(defaultLookupLimit)}
	unusedArgs := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h":
			return opts, nil, true
		case arg == "--name":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
				opts.Names = append(opts.Names, "")
				continue
			}
			opts.Names = append(opts.Names, args[i+1])
			i++
		case strings.HasPrefix(arg, "--name="):
			opts.Names = append(opts.Names, strings.TrimPrefix(arg, "--name="))
		case arg == "--limit":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
				opts.Limit = ""
				continue
			}
			opts.Limit = args[i+1]
			i++
		case strings.HasPrefix(arg, "--limit="):
			opts.Limit = strings.TrimPrefix(arg, "--limit=")
		default:
			unusedArgs = append(unusedArgs, arg)
		}
	}
	return opts, unusedArgs, false
}

func runCodesLookup(ctx context.Context, opts codesLookupOptions, unusedArgs []string) error {
	names, limit, err := parseCodesLookupOptions(opts, unusedArgs)
	if err != nil {
		return encodeLookupFailure(err)
	}

	creds, err := config.Load()
	if err != nil {
		return encodeLookupFailure(lookupConfigError{err: err})
	}
	if creds.AppKey == "" || creds.SecretKey == "" {
		return encodeLookupFailure(lookupConfigError{err: fmt.Errorf(config.MissingCredentialsMessage)})
	}

	rows, err := newStockInfoClient(creds).StockInfoRows(ctx, nil)
	if err != nil {
		return encodeLookupFailure(lookupUpstreamError{err: err})
	}

	output := stocklookup.Lookup(names, rows, limit)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func parseCodesLookupOptions(opts codesLookupOptions, unusedArgs []string) ([]string, int, error) {
	if len(unusedArgs) > 0 {
		return nil, 0, lookupValidationError{message: fmt.Sprintf("unexpected extra arguments: %v", unusedArgs)}
	}
	names, err := validateLookupNames(opts.Names)
	if err != nil {
		return nil, 0, err
	}
	limit, err := validateLookupLimit(opts.Limit)
	if err != nil {
		return nil, 0, err
	}
	return names, limit, nil
}

func validateLookupNames(names []string) ([]string, error) {
	cleaned := make([]string, 0, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return nil, lookupValidationError{message: "stock name must not be blank"}
		}
		cleaned = append(cleaned, trimmed)
	}
	if len(cleaned) == 0 {
		return nil, lookupValidationError{message: "at least one --name is required"}
	}
	return cleaned, nil
}

func validateLookupLimit(limit string) (int, error) {
	trimmed := strings.TrimSpace(limit)
	if trimmed == "" {
		return 0, lookupValidationError{message: "limit must be an integer"}
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, lookupValidationError{message: "limit must be an integer"}
	}
	if parsed < 1 {
		return 0, lookupValidationError{message: "limit must be at least 1"}
	}
	return parsed, nil
}

func encodeLookupFailure(err error) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(stocklookup.Envelope{
		OK:      false,
		Queries: []stocklookup.Query{},
		Errors: []stocklookup.ErrorEntry{{
			Type:    lookupErrorType(err),
			Message: err.Error(),
		}},
	})
	return cli.Exit("", 1)
}

func lookupErrorType(err error) string {
	var validationErr lookupValidationError
	if errors.As(err, &validationErr) {
		return "ValidationError"
	}
	var configErr lookupConfigError
	if errors.As(err, &configErr) {
		return "ConfigError"
	}
	return "KiwoomClientError"
}
