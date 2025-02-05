package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"git.inpt.fr/churros/notella"
	"github.com/invopop/jsonschema"

	esbuild "github.com/evanw/esbuild/pkg/api"
	ll "github.com/gwennlbh/label-logger-go"
)

func main() {
	ll.Log("Reflecting", "cyan", "structs")
	reflector := new(jsonschema.Reflector)
	if err := reflector.AddGoComments("git.inpt.fr/churros/notella", "./"); err != nil {
		fmt.Printf("Error adding Go comments: %v\n", err)
	}

	ll.Log("Writing", "cyan", "typescript types")
	writeTypescriptDefinition(reflector, "Message", &notella.Message{}, "typescript/message.ts")
	writeTypescriptDefinition(reflector, "HealthResponse", &notella.HealthResponse{}, "typescript/health.ts")
	reflector.FieldNameTag = "env"
	writeTypescriptDefinition(reflector, "Configuration", &notella.Configuration{}, "typescript/configuration.ts")

	// Also save useful constants
	ll.Log("Writing", "cyan", "exported constants")
	os.WriteFile("typescript/constants.ts", []byte(fmt.Sprintf("export const STREAM_NAME = '%s';\nexport const SUBJECT_NAME = '%s';\n", notella.StreamName, notella.SubjectName)), 0644)

	// Write barrel
	ll.Log("Writing", "cyan", "barrel file")
	os.WriteFile("typescript/index.ts", []byte(strings.Join([]string{
		"export * from './message.js';",
		"export * from './configuration.js';",
		"export * from './health.js';",
		"export * from './constants.js';",
	}, "\n")), 0644)

	ll.Log("Transpiling", "cyan", "to JS using esbuild")
	result := esbuild.Build(esbuild.BuildOptions{
		EntryPoints: []string{"typescript/index.ts"},
		Outdir:      "typescript-dist/",
		Bundle:      true,
		Sourcemap:   esbuild.SourceMapLinked,
		Format:      esbuild.FormatESModule,
		Platform:    esbuild.PlatformNeutral,
	})

	for _, msg := range result.Warnings {
		ll.Warn(formatEsbuildMessage(msg))
	}

	if len(result.Errors) > 0 {
		for _, msg := range result.Errors {
			ll.Error(formatEsbuildMessage(msg))
			os.Exit(1)
		}
	} else {
		for _, file := range result.OutputFiles {
			err := os.WriteFile(file.Path, file.Contents, 0o677)
			if err != nil {
				ll.ErrorDisplay("could not write %s [%s]", err, file.Path, file.Hash)
				os.Exit(1)
			}
			ll.Log("Wrote", "blue", "%s [dim][%s][reset]", file.Path, file.Hash)
		}
		ll.Log("Built", "green", "typescript library to [bold]typescript-dist/[reset]")
	}
}

func formatEsbuildMessage(msg esbuild.Message) string {
	notes := ""
	for _, note := range msg.Notes {
		notes += fmt.Sprintf("\nat %s: %s", formatEsbuildLocation(note.Location), note.Text)
	}
	return fmt.Sprintf("at %s: %s%s", formatEsbuildLocation(msg.Location), msg.Text, notes)
}

func formatEsbuildLocation(loc *esbuild.Location) string {
	if loc == nil {
		return ""
	}
	return fmt.Sprintf("[blue]%s:%d:%d[reset]", loc.File, loc.Line, loc.Column)
}

func writeTypescriptDefinition(reflector *jsonschema.Reflector, typename string, typ interface{}, filename string) {
	schema := reflector.Reflect(typ)
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		fmt.Printf("Error generating schema: %v\n", err)
		return
	}

	// Set up quicktype command to read from stdin
	cmd := exec.Command("npm", "exec", "quicktype", "--", "--lang=ts", "--src-lang=schema", "--just-types", fmt.Sprintf("--top-level=%s", typename))

	// Create a pipe to stdin for the quicktype command
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("Error creating stdin pipe: %v\n", err)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating stdout pipe: %v\n", err)
		return
	}

	stderr, err := cmd.StderrPipe() // Create a pipe for stderr
	if err != nil {
		fmt.Printf("Error creating stderr pipe: %v\n", err)
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting quicktype: %v\n", err)
		return
	}

	// Write the JSON schema to quicktype's stdin
	_, err = stdin.Write(schemaJSON)
	if err != nil {
		fmt.Printf("Error writing to stdin: %v\n", err)
		return
	}
	stdin.Close() // Important to close the pipe to signal EOF to quicktype

	output, err := io.ReadAll(stdout)
	if err != nil {
		fmt.Printf("Error reading from stdout: %v\n", err)
		return
	}

	// Read any error messages from quicktype's stderr
	errorOutput, err := io.ReadAll(stderr)
	if err != nil {
		fmt.Printf("Error reading quicktype stderr: %v\n", err)
		return
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for quicktype command: %v\n", err)
		// Print the stderr output in case of an error
		fmt.Printf("Quicktype stderr: %s\n", errorOutput)
		return
	}

	// Print or save the TypeScript output
	os.WriteFile(filename, output, 0644)
}
