package app

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/samaydhawan/exa-cli/internal/config"
)

func (a *App) writeEnvelope(data any, meta Meta, commandDefaultFormat string) error {
	meta.Format = a.effectiveFormat(commandDefaultFormat)
	meta.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	envelope := Envelope{
		Meta: meta,
		Data: data,
	}
	return a.renderEnvelope(envelope)
}

func (a *App) renderEnvelope(envelope Envelope) error {
	switch envelope.Meta.Format {
	case "json":
		data, err := json.MarshalIndent(envelope, "", "  ")
		if err != nil {
			return wrap(1, err)
		}
		_, err = fmt.Fprintln(a.out, string(data))
		return err
	case "jsonl":
		encoder := json.NewEncoder(a.out)
		if err := encoder.Encode(map[string]any{"type": "meta", "data": envelope.Meta}); err != nil {
			return wrap(1, err)
		}
		for _, record := range jsonlRecords(envelope.Data) {
			if err := encoder.Encode(map[string]any{"type": "record", "data": record}); err != nil {
				return wrap(1, err)
			}
		}
		return nil
	case "llm":
		_, err := fmt.Fprintln(a.out, renderLLM(envelope, a.options.Verbose))
		return err
	case "markdown":
		_, err := fmt.Fprintln(a.out, renderMarkdown(envelope, a.options.Verbose))
		return err
	default:
		_, err := fmt.Fprintln(a.out, renderTable(envelope))
		return err
	}
}

func (a *App) effectiveFormat(commandDefault string) string {
	if a.options.FormatSource == "default" && commandDefault != "" {
		return commandDefault
	}
	if a.options.Format == "" {
		if commandDefault != "" {
			return commandDefault
		}
		return config.DefaultFormat
	}
	return a.options.Format
}
