package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

var (
	successColor = color.New(color.FgGreen, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	warningColor = color.New(color.FgYellow, color.Bold)
	infoColor    = color.New(color.FgCyan)
)

func Success(message string) {
	successColor.Println(message)
}

func Error(message string) {
	errorColor.Fprintln(os.Stderr, message)
}

func Warning(message string) {
	warningColor.Println(message)
}

func Info(message string) {
	infoColor.Println(message)
}

type Spinner struct {
	spinner *spinner.Spinner
}

func NewSpinner(message string) *Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Color("cyan")
	return &Spinner{spinner: s}
}

func (s *Spinner) Start() {
	s.spinner.Start()
}

func (s *Spinner) Stop() {
	s.spinner.Stop()
}

type Table struct {
	headers []string
	rows    [][]string
}

func NewTable(headers []string) *Table {
	return &Table{headers: headers, rows: [][]string{}}
}

func (t *Table) AddRow(row []string) {
	t.rows = append(t.rows, row)
}

func (t *Table) Render() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(t.headers)
	table.SetBorder(true)
	table.SetRowLine(false)
	table.AppendBulk(t.rows)
	table.Render()
}

func PrintJSON(data interface{}) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonBytes))
	return nil
}

func PrintYAML(data interface{}) error {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Println(string(yamlBytes))
	return nil
}
