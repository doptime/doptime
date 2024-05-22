package dynamicdev

import (
	"os"
	"os/exec"

	"github.com/doptime/doptime/api"
)

type CompileAndRunGoCodeIn struct {
	SourceCode string
	FileName   string
	RunOnce    bool
}

var ApiCompileAndRunGoCode = api.Api(func(req *CompileAndRunGoCodeIn) (ok bool, err error) {
	// Step 1: Write the source code to a temporary file
	tmpFile, err := os.CreateTemp("", "*.go")
	if err != nil {
		return false, err
	}
	defer os.Remove(tmpFile.Name()) // Clean up file afterwards

	if _, err := tmpFile.Write([]byte(req.FileName)); err != nil {
		return false, err
	}
	if err := tmpFile.Close(); err != nil {
		return false, err
	}

	// Step 2: Compile the Go program
	outBinary := tmpFile.Name() + ".exe"
	cmdCompile := exec.Command("go", "build", "-o", outBinary, tmpFile.Name())
	if err := cmdCompile.Run(); err != nil {
		return false, err
	}
	defer os.Remove(outBinary) // Clean up the binary afterwards

	// Step 3: Run the compiled program
	cmdRun := exec.Command(outBinary)
	cmdRun.Stdout = os.Stdout
	cmdRun.Stderr = os.Stderr
	if err := cmdRun.Run(); err != nil {
		return false, err
	}
	return true, nil
}).Func
