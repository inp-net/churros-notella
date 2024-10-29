package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"git.inpt.fr/churros/notella"
	"github.com/invopop/jsonschema"
)

func main() {
	reflector := new(jsonschema.Reflector)
	if err := reflector.AddGoComments("git.inpt.fr/churros/notella", "./"); err != nil {
		fmt.Printf("Error adding Go comments: %v\n", err)
	}
	schema := reflector.Reflect(&notella.Message{})
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		fmt.Printf("Error generating schema: %v\n", err)
		return
	}

	// Set up quicktype command to read from stdin
	cmd := exec.Command("npm", "exec", "quicktype", "--", "--lang=ts", "--src-lang=schema", "--just-types", "--top-level=Message")

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
	os.WriteFile("types.d.ts", output, 0644)

	// Also save useful constants
	os.WriteFile("constants.ts", []byte(fmt.Sprintf("export const STREAM_NAME = %q; export const SUBJECT_NAME = %q;", notella.StreamName, notella.SubjectName)), 0644)
}
