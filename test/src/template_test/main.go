package main

import (
	"os"
	"text/template"
)

func main() {
	//const_str_output := `{{"\"output\""}}`
	//tmpl, _ := template.New("test1").Parse(const_str_output)

	//	const_rawstr_output := "{{`\"output\"`}}"
	//	tmpl, _ := template.New("test1").Parse(const_rawstr_output)

	//func_output := `{{ printf "%q" "output"}}`
	//tmpl, _ := template.New("test1").Parse(func_output)

	//pipeline_output := `{{"output" | printf "%q"}}`
	//tmpl, _ := template.New("test1").Parse(pipeline_output)

	//parenth_args := `{{ printf "%q" (print "out" "put")}}`
	//tmpl, _ := template.New("test1").Parse(parenth_args)

	//elabrate_call := `{{ "put" | printf "%s%s" "out" | printf "%q"}}`
	//tmpl, _ := template.New("test1").Parse(elabrate_call)

	//long_chain := `{{ "output" | printf "%s" | printf "%q" }}`
	//tmpl, _ := template.New("test1").Parse(long_chain)

	//with := `{{with "output"}} {{ printf "%q" .}}{{end}}`
	//tmpl, _ := template.New("test1").Parse(with)

	//with_var := `{{ with $x := "output" | printf "%q"}} {{$x}} {{end}}`
	//tmpl, _ := template.New("test1").Parse(with_var)

	//with_var2 := `{{ with $x := "output" }}{{ printf "%q" $x }} {{end}}`
	//tmpl, _ := template.New("test1").Parse(with_var2)
	with_var_pipeline := `{{ with $x := "output"}} {{$x | printf "%q" }}{{end}}`
	tmpl, _ := template.New("test1").Parse(with_var_pipeline)
	tmpl.Execute(os.Stdout, nil)
}
