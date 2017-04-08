package commands

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/docker/docker/pkg/term"
	"github.com/spf13/cobra"
)

var ManPage = &cobra.Command{
	Use:   "man",
	Short: "display a man page",
	Long:  `display the Monax man page`,
	Run: func(cmd *cobra.Command, args []string) {
		man := new(bytes.Buffer)

		GenerateManPage(man)
		DisplayManPage(man)
	},
}

func buildManCommand() {
	ManPage.Flags().BoolVarP(&do.Dump, "dump", "", false, "dump the man page source")
}

// GenerateManPages uses the Cobra commands' info to construct a man page.
// It fills in the man buffer with the groff(1) markup code.
func GenerateManPage(man *bytes.Buffer) {
	generatePrologue(man)
	generateCommands(man)
	generateGlobalFlags(man)
	generateEnvironment(man)
	generateEpilogue(man)
}

func generatePrologue(man *bytes.Buffer) {
	template.Must(template.New("prologue").Funcs(manHelpers).Parse(manPrologue)).Execute(man, MonaxCmd)
}

func generateCommands(man *bytes.Buffer) {
	template.Must(template.New("commands").Funcs(manHelpers).Parse(manMidsection)).ExecuteTemplate(man, "commands", MonaxCmd)
}

func generateGlobalFlags(man *bytes.Buffer) {
	template.Must(template.New("global flags").Funcs(manHelpers).Parse(manMidsection)).ExecuteTemplate(man, "global flags", MonaxCmd)
}

func generateEnvironment(man *bytes.Buffer) {
	template.Must(template.New("environment").Funcs(manHelpers).Parse(manEnvironment)).Execute(man, nil)
}

func generateEpilogue(man *bytes.Buffer) {
	template.Must(template.New("epilogue").Funcs(manHelpers).Parse(manEpilogue)).Execute(man, MonaxCmd)
}

// DisplayManPage runs the groff(1) formatter on a buffer
// and then starts a pager to display the result.
func DisplayManPage(in *bytes.Buffer) {
	// Dump the man page source.
	if do.Dump {
		fmt.Println(in)
		return
	}

	out := new(bytes.Buffer)
	nroff := exec.Command("nroff", "-mdoc", "-Tascii")
	nroff.Stdin = in
	nroff.Stdout = out
	nroff.Run()

	r, w := io.Pipe()
	go func(w *io.PipeWriter, out *bytes.Buffer) {
		fmt.Fprint(w, out)
		w.Close()
	}(w, out)

	// If not a terminal, don't run less(1).
	if !term.IsTerminal(os.Stdout.Fd()) {
		fmt.Println(out)
		return
	}

	// Behave like less(1), which ignores SIGINT. more(1), on the other hand,
	// handles SIGINT, but ignoring it doesn't harm here.
	// Ignoring the interrupt signal is important, because interrupting
	// less(1) leaves the terminal in a broken state.
	signal.Notify(make(chan os.Signal, 1), os.Interrupt)

	// Use PAGER value if set, or less(1) by default.
	pagerCommand := os.Getenv("PAGER")
	if pagerCommand == "" {
		pagerCommand = "less -r"
	}

	pagerArgs := strings.Split(pagerCommand, " ")
	pager := exec.Command(pagerArgs[0], pagerArgs[1:]...)
	pager.Stdin = r
	pager.Stdout = os.Stdout
	pager.Start()
	pager.Wait()
}
