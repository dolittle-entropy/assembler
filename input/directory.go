package input

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
	"os"
	"path"
)

type DirectoryInput struct {
	path    string
	watcher *fsnotify.Watcher
	logger  *zerolog.Logger
}

func NewDirectoryInput(config *koanf.Koanf, logger *zerolog.Logger) (*DirectoryInput, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	path := config.String("input.directory")
	if err := watcher.Add(path); err != nil {
		return nil, err
	}

	loggerWithPath := logger.With().Str("path", path).Logger()

	input := &DirectoryInput{
		path:    path,
		watcher: watcher,
		logger:  &loggerWithPath,
	}

	go input.listenForChanges()

	if err := input.createEventsForExistingFiles(); err != nil {
		return nil, err
	}

	return input, nil
}

func (di *DirectoryInput) onFileUpdated(name string) {
	logger := di.logger.With().Str("file", name).Logger()

	contents, err := os.ReadFile(name)
	if err != nil {
		logger.Error().Err(err).Msg("Could not read input file")
		return
	}

	data := unstructured.Unstructured{}
	if err := json.Unmarshal(contents, &data.Object); err != nil {
		logger.Error().Err(err).Msg("Could not parse input file as Unstructured")
		return
	}

	fmt.Println(data.GetName(), data.GetNamespace())
}

func (di *DirectoryInput) onFileRemoved(name string) {

}

func (di *DirectoryInput) listenForChanges() {
	defer di.logger.Warn().Msg("Watcher loop finished")
	defer di.watcher.Close()

	di.logger.Info().Msg("Watching directory for changes...")

	for {
		select {
		case event, ok := <-di.watcher.Events:
			if !ok {
				return
			}
			if event.Op == fsnotify.Create || event.Op == fsnotify.Write {
				di.onFileUpdated(event.Name)
			} else if event.Op == fsnotify.Remove || event.Op == fsnotify.Rename {
				di.onFileRemoved(event.Name)
			}
		case err, ok := <-di.watcher.Errors:
			if !ok {
				return
			}
			di.logger.Error().Err(err).Msg("Error received while watching directory")
		}
	}
}

func (di *DirectoryInput) createEventsForExistingFiles() error {
	files, err := ioutil.ReadDir(di.path)
	if err != nil {
		return err
	}

	for _, file := range files {
		name := path.Join(di.path, file.Name())
		di.watcher.Events <- fsnotify.Event{
			Name: name,
			Op:   fsnotify.Create,
		}
	}

	return nil
}
