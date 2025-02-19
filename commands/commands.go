package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/Qitmeer/qitmeer-cli/rpc/client"
)

const (
	// DefaultServiceNameSpace default api
	DefaultServiceNameSpace = "qitmeer"
	// MinerNameSpace  miner api
	MinerNameSpace = "miner"
)

// RootCmd command obj
var RootCmd = &cobra.Command{
	Use:  "qitmeer-cli",
	Long: `qitmeer cli is a RPC tool for the qitmeer and qitmeer-wallet`,
	//PersistentPreRunE: rootCmdPreRun,

}

var RPCCfg = &client.Config{}
var RPCVersion = "1.0"
var Format = false

var RootSubCmdGroups = make(map[string][]*cobra.Command)

func init() {
	//RootCmd.SetHelpCommand(helpCmd)

	RootCmd.SetUsageFunc(UseageFunc)

}

func getResString(method string, args []interface{}) (rs string, err error) {
	reqData, err := client.MakeRequestData(RPCVersion, 1, method, args)
	if err != nil {
		return
	}

	resResult, err := client.SendPostRequest(reqData, RPCCfg)
	if err != nil {
		return
	}

	rs = string(resResult)
	return

}

//
func output(dataStr string) {
	if Format {
		var str bytes.Buffer
		_ = json.Indent(&str, []byte(dataStr), "", "    ")
		fmt.Println(str.String())
	} else {
		fmt.Println(dataStr)
	}
}

var cmdHelpTpl = `Usage:{{if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
{{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
{{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}
{{MakeHelpTpl}} {{end}}{{if .HasAvailableLocalFlags}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// MakeHelpTpl make root sub group
func MakeHelpTpl() string {
	cmdTpl := `{{range $group,$Commands :=.}}
{{$group}} Commands: {{range $Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
	{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}
{{end}}`

	buf := bytes.NewBuffer([]byte{})

	t := template.New("cc")
	t.Funcs(template.FuncMap{
		"rpad": rpad,
	})
	t.Parse(cmdTpl)
	t.Execute(buf, RootSubCmdGroups)

	return string(buf.Bytes())
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func UseageFunc(c *cobra.Command) error {

	groupTpl := MakeHelpTpl()
	tpl := strings.Replace(cmdHelpTpl, "{{MakeHelpTpl}}", groupTpl, -1)

	err := tmpl(c.OutOrStdout(), tpl, c)
	if err != nil {
		c.Println(err)
	}

	return nil
}

// tmpl executes the given template text on data, writing the result to w.
func tmpl(w io.Writer, text string, data interface{}) error {
	t := template.New("top")
	t.Funcs(templateFuncs)
	template.Must(t.Parse(text))
	return t.Execute(w, data)
}

var templateFuncs = template.FuncMap{
	"trim":                    strings.TrimSpace,
	"trimRightSpace":          trimRightSpace,
	"trimTrailingWhitespaces": trimRightSpace,
	"appendIfNotPresent":      appendIfNotPresent,
	"rpad":                    rpad,
	"gt":                      Gt,
	"eq":                      Eq,
}

func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// appendIfNotPresent will append stringToAppend to the end of s, but only if it's not yet present in s.
func appendIfNotPresent(s, stringToAppend string) string {
	if strings.Contains(s, stringToAppend) {
		return s
	}
	return s + " " + stringToAppend
}

// FIXME Gt is unused by cobra and should be removed in a version 2. It exists only for compatibility with users of cobra.

// Gt takes two types and checks whether the first type is greater than the second. In case of types Arrays, Chans,
// Maps and Slices, Gt will compare their lengths. Ints are compared directly while strings are first parsed as
// ints and then compared.
func Gt(a interface{}, b interface{}) bool {
	var left, right int64
	av := reflect.ValueOf(a)

	switch av.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		left = int64(av.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		left = av.Int()
	case reflect.String:
		left, _ = strconv.ParseInt(av.String(), 10, 64)
	}

	bv := reflect.ValueOf(b)

	switch bv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		right = int64(bv.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		right = bv.Int()
	case reflect.String:
		right, _ = strconv.ParseInt(bv.String(), 10, 64)
	}

	return left > right
}

// Eq takes two types and checks whether they are equal. Supported types are int and string. Unsupported types will panic.
func Eq(a interface{}, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		panic("Eq called on unsupported type")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return av.Int() == bv.Int()
	case reflect.String:
		return av.String() == bv.String()
	}
	return false
}
