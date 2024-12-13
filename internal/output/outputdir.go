package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/softwaresale/lambdagen/internal/codegen"
	"github.com/softwaresale/lambdagen/internal/model"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

type outputNode struct {
	serviceDef *model.ServiceDefinition
	method     *model.HandlerDefinition
}

func (node outputNode) Metadata() model.LambdaMetadata {

	handlerPath := node.method.Path
	basePath, ok := node.serviceDef.Config["base_path"]
	if ok {
		handlerPath = filepath.Join(basePath, handlerPath)
	}

	return model.LambdaMetadata{
		Path:   handlerPath,
		Method: node.method.Method,
	}
}

// Manager is responsible for verifying that all rendered handlers form a valid API
type Manager struct {
	baseOutputDir string
	outputs       map[string]outputNode
	uniquePaths   map[string]string
}

func NewManager(rootModDir, lambdaDir string) *Manager {

	outputDir := filepath.Join(rootModDir, lambdaDir)

	return &Manager{
		baseOutputDir: outputDir,
		outputs:       make(map[string]outputNode),
	}
}

func (output *Manager) CreateOutputDir() error {
	return os.MkdirAll(output.baseOutputDir, 0755)
}

func (output *Manager) Render() error {
	var err error
	for outputPath, node := range output.outputs {
		// make the output path
		err = os.MkdirAll(outputPath, os.ModePerm)
		if err != nil {
			return err
		}

		outputFile, err := os.Create(filepath.Join(outputPath, "main.go"))
		if err != nil {
			return fmt.Errorf("error while making main file: %w", err)
		}

		err = codegen.TranslateHandler(outputFile, *node.serviceDef, *node.method)
		if err != nil {
			return fmt.Errorf("error while translating handler: %w", err)
		}

		err = output.outputMetadata(node, outputPath)
		if err != nil {
			return fmt.Errorf("error while writing metadata: %w", err)
		}
	}

	return nil
}

func (output *Manager) outputMetadata(node outputNode, lambdaDir string) error {

	// get an optional base path
	basePath, ok := node.serviceDef.Config["base_path"]
	if ok {
		node.method.Path = filepath.Join(basePath, node.method.Path)
	}

	outputPath := filepath.Join(lambdaDir, "spec.json")
	metadata := model.LambdaMetadata{
		Method: node.method.Method,
		Path:   node.method.Path,
	}

	bytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("error while marshalling data: %w", err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error while creating output file: %w", err)
	}

	defer func() {
		err := outputFile.Close()
		if err != nil {
			panic(err)
		}
	}()

	_, err = outputFile.Write(bytes)
	if err != nil {
		return fmt.Errorf("error while writing output file: %w", err)
	}

	_, err = outputFile.Write([]byte{'\n'})
	if err != nil {
		return fmt.Errorf("error while writing output file: %w", err)
	}

	return nil
}

// Register registers a service handler to be outputted
func (output *Manager) Register(serviceDef *model.ServiceDefinition, handler *model.HandlerDefinition) error {

	handlerName, err := extractHandlerName(serviceDef)
	if err != nil {
		panic(err)
	}

	// make the output dir
	lambdaDirectoryName := strings.Join([]string{handlerName, handler.HandlerMethodName}, "_")
	lambdaDirectoryPath := filepath.Join(output.baseOutputDir, lambdaDirectoryName)

	if _, existing := output.uniquePaths[lambdaDirectoryPath]; existing {
		return errors.New("lambda directory already exists")
	}

	outputNode := outputNode{
		serviceDef: serviceDef,
		method:     handler,
	}

	// ensure that path is not mapped with the given method
	method, ok := output.uniquePaths[outputNode.Metadata().Path]
	if ok {
		// this is mapped, if the methods are the same, then that's a problem
		if method == outputNode.method.Method {
			// problem
			return fmt.Errorf("path %s %s is already mapped", method, outputNode.Metadata().Path)
		}
	}

	// work out the output
	output.outputs[lambdaDirectoryPath] = outputNode

	return nil
}

func extractHandlerName(serviceDef *model.ServiceDefinition) (string, error) {
	var handlerName string
	switch handlerType := serviceDef.Type.(type) {
	case *types.Named:
		handlerName = handlerType.Obj().Name()

	default:
		return "", errors.New("handler was not a named type")
	}

	return handlerName, nil
}
