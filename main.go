package main

import (
	"fmt"
	"github.com/kreuzwerker/gva"
	"gopkg.in/alecthomas/kingpin.v2"
	"net"
	"os"
	"strings"
)

type subArg []string

var (
	prepared_command []string
	extra_flag       string

	app = kingpin.New("varnishadm-ninja", "Varnish CLI client")

	varnish_terminal_address = app.Flag("terminal", "Varnish management terminal address").Short('T').Default("127.0.0.1:6082").String()
	varnish_terminal_secret  = app.Flag("secret", "Secret authentication code").Short('S').String()

	command_vcl_load            = app.Command("vcl.load", "Compile and load the VCL file under the name provided.")
	command_vcl_load_configname = command_vcl_load.Arg("configname", "Alias for configuration").Required().String()
	command_vcl_load_filename   = command_vcl_load.Arg("filename", "Configuration filename").Required().String()
	command_vcl_load_temp       = command_vcl_load.Arg("temperature", "VCL temperature").String()

	command_vcl_inline                   = app.Command("vcl.inline", "Compile and load the VCL data under the name provided")
	command_vcl_inline_configname        = command_vcl_inline.Arg("configname", "Alias for configuration").Required().String()
	command_vcl_inline_quoted_vcl_string = command_vcl_inline.Arg("quoted_vcl_string", "VCL data").Required().String()
	command_vcl_inline_temp              = command_vcl_inline.Arg("temperature", "VCL temperature").String()

	command_vcl_use            = app.Command("vcl.use", "Switch to the named configuration immediately")
	command_vcl_use_configname = command_vcl_use.Arg("configname", "Alias for configuration").Required().String()

	command_vcl_discard            = app.Command("vcl.discard", "Unload the named configuration (when possible)")
	command_vcl_discard_configname = command_vcl_discard.Arg("configname", "Alias for configuration").Required().String()

	command_vcl_list = app.Command("vcl.list", "List all loaded configuration")

	command_vcl_show            = app.Command("vcl.show", "Display the source code for the specified configuration.")
	command_vcl_show_configname = command_vcl_show.Arg("configname", "Alias for configuration").Required().String()

	command_vcl_state            = app.Command("vcl.state", "Force the state of the specified configuration. State is any of auto, warm or cold values.")
	command_vcl_state_configname = command_vcl_state.Arg("configname", "Alias for configuration").Required().String()
	command_vcl_state_temp       = command_vcl_state.Arg("temperature", "VCL temperature").Required().String()

	command_param_show             = app.Command("param.show", "Show parameters and their values.")
	command_param_show_description = command_param_show.Flag("description", "Display description on each parameter").Short('l').Bool()
	command_param_show_parameter   = command_param_show.Arg("parameter", "Show description and value for specified parameter").String()

	command_param_set           = app.Command("param.set", "Set parameter value")
	command_param_set_parameter = command_param_set.Arg("parameter", "Parameter name").Required().String()
	command_param_set_value     = command_param_set.Arg("value", "Value for parameter").Required().String()

	command_panic_show = app.Command("panic.show", "Return the last panic, if any.")

	command_panic_clear          = app.Command("param.clear", "Clear the last panic, if any.")
	command_panic_clear_counters = command_panic_clear.Flag("counters", "Clear related varnishstat counters(s)").Short('z').Bool()

	command_storage_list = app.Command("storage.list", "Lists the defined storage backends")

	command_backend_list               = app.Command("backend.list", "List backends")
	command_backend_list_probe_details = command_backend_list.Flag("probe", "Iclude probe details").Short('p').Bool()
	command_backend_list_expression    = command_backend_list.Arg("expression", "Expression to filter backends").String()

	command_backend_set_health            = app.Command("backend.set_health", "Set health status on the backends. ")
	command_backend_set_health_expression = command_backend_set_health.Arg("expression", "Expression to filter backends").Required().String()
	command_backend_set_health_state      = command_backend_set_health.Arg("state", "State is any of auto, healthy or sick values.").Required().String()

	command_ban      = app.Command("ban", "Mark obsolete all objects where all the conditions match")
	command_ban_args = SubArg(command_ban.Arg("arguments", "Arguments for ban command"))

	command_ban_list = app.Command("ban.list", "List the active bans")
)

func (i *subArg) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *subArg) String() string {
	return ""
}

func (i *subArg) IsCumulative() bool {
	return true
}

func SubArg(s kingpin.Settings) (target *[]string) {
	target = new([]string)
	s.SetValue((*subArg)(target))
	return
}

func SetVCLtemperature(state string, quit int) string {
	switch state {
	case "auto":
		return "auto"
	case "cold":
		return"cold"
	case "warm":
		return "warm"
	default:
		if quit == 1 {
			println("Wrong VCL temperature specified")
			os.Exit(1)
		}
		return ""
	}
}

func SetBackendState(state string, quit int) string {
	switch state {
	case "auto":
		return "auto"
	case "healthy":
		return"healthy"
	case "sick":
		return "sick"
	default:
		if quit == 1 {
			println("Wrong backend state")
			os.Exit(1)
		}
		return ""
	}
}

func DescribeReturnCode(code int) string {
	switch code {
	case 100:
		return "Syntax Error"
	case 101:
		return "Unknown request"
	case 102:
		return "Unimplemented"
	case 104:
		return "Too few arguments"
	case 105:
		return "Too many arguments"
	case 106:
		return "Command failed"
	case 107:
		return "Authentication required"
	case 200:
		return "OK"
	case 400:
		return "CLI communication error"
	default:
		return "Command failed"
	}
}

func init() {

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case command_vcl_load.FullCommand():
		prepared_command = append(prepared_command,
			command_vcl_load.FullCommand(),
			*command_vcl_load_configname,
			*command_vcl_load_filename,
			SetVCLtemperature(*command_vcl_load_temp, 0))

	case command_vcl_inline.FullCommand():
		prepared_command = append(prepared_command,
			command_vcl_load.FullCommand(),
			*command_vcl_inline_configname,
			fmt.Sprintf("%q", *command_vcl_inline_quoted_vcl_string),
			SetVCLtemperature(*command_vcl_inline_temp, 0))

	case command_vcl_use.FullCommand():
		prepared_command = append(prepared_command, command_vcl_use.FullCommand(), *command_vcl_use_configname)

	case command_vcl_discard.FullCommand():
		prepared_command = append(prepared_command, command_vcl_discard.FullCommand(), *command_vcl_discard_configname)

	case command_vcl_list.FullCommand():
		prepared_command = append(prepared_command, command_vcl_list.FullCommand())

	case command_vcl_show.FullCommand():
		prepared_command = append(prepared_command, command_vcl_show.FullCommand(), *command_vcl_show_configname)

	case command_vcl_state.FullCommand():
		prepared_command = append(prepared_command,
			command_vcl_state.FullCommand(),
			*command_vcl_state_configname,
			SetVCLtemperature(*command_vcl_state_temp, 1))

	case command_param_show.FullCommand():
		if *command_param_show_description == true {
			extra_flag = "-l"
		}
		prepared_command = append(prepared_command,
			command_param_show.FullCommand(),
			extra_flag,
			*command_param_show_parameter)

	case command_param_set.FullCommand():
		prepared_command = append(prepared_command,
			command_param_set.FullCommand(),
			*command_param_set_parameter,
			*command_param_set_value)

	case command_panic_show.FullCommand():
		prepared_command = append(prepared_command, command_panic_show.FullCommand())

	case command_panic_clear.FullCommand():
		if *command_panic_clear_counters == true {
			extra_flag = "-z"
		}
		prepared_command = append(prepared_command,
			command_panic_clear.FullCommand(),
			extra_flag)

	case command_storage_list.FullCommand():
		prepared_command = append(prepared_command, command_storage_list.FullCommand())

	case command_backend_list.FullCommand():
		if *command_backend_list_probe_details == true {
			extra_flag = "-p"
		}
		prepared_command = append(prepared_command,
			command_backend_list.FullCommand(),
			extra_flag,
			*command_backend_list_expression)

	case command_backend_set_health.FullCommand():
		prepared_command = append(prepared_command,
			command_backend_set_health.FullCommand(),
			*command_backend_set_health_expression,
			SetBackendState(*command_backend_set_health_state, 1))

	case command_ban.FullCommand():
		if len(*command_ban_args) < 3 {
			println("Too few arguments for ban command")
			os.Exit(1)
		}
		prepared_command = append(prepared_command, command_ban.FullCommand(), strings.Join(*command_ban_args, " "))

	case command_ban_list.FullCommand():
		prepared_command = append(prepared_command, command_ban_list.FullCommand())

	}
}

func main() {

	varnish_secret := fmt.Sprintf("%s\n", *varnish_terminal_secret)

	varnish_terminal_interface, err := net.ResolveTCPAddr("tcp", *varnish_terminal_address)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := gva.NewConnection(varnish_terminal_interface.IP.String(), uint16(varnish_terminal_interface.Port), &varnish_secret)
	if err != nil {
		println("Connection failed:", err.Error())
		os.Exit(1)
	}

	resp, err := conn.Cmd(prepared_command[0], strings.Join(prepared_command[1:], " "))
	if err != nil {
		println("Command failed:", err.Error())
		os.Exit(1)
	}

	fmt.Printf("%v (%v)\n", resp.Status, DescribeReturnCode(resp.Status))
	if !resp.IsSuccess() {
		os.Exit(1)
	}	else {
		println(resp.Body)
	}
}
