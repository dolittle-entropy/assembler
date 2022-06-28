package input

import (
	"dolittle.io/kokk/resources"
	"github.com/fsnotify/fsnotify"
	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path"
)

type TypeConverter interface {
	Convert(object *unstructured.Unstructured) (*resources.Resource, error)
}

type DirectoryInput struct {
	path       string
	watcher    *fsnotify.Watcher
	converter  TypeConverter
	repository map[string]resources.Resource
	fileIDs    map[string]string
	logger     *zerolog.Logger
}

func NewDirectoryInput(config *koanf.Koanf, converter TypeConverter, logger *zerolog.Logger) (*DirectoryInput, error) {
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
		path:       path,
		watcher:    watcher,
		converter:  converter,
		repository: make(map[string]resources.Resource),
		fileIDs:    make(map[string]string),
		logger:     &loggerWithPath,
	}

	go input.listenForChanges()

	if err := input.createEventsForExistingFiles(); err != nil {
		return nil, err
	}

	return input, nil
}

func (di *DirectoryInput) Get(id string) (*resources.Resource, error) {
	if resource, found := di.repository[id]; found {
		return &resource, nil
	}

	return nil, ResourceNotFound
}

func (di *DirectoryInput) List() []resources.Resource {
	list := make([]resources.Resource, 0, len(di.repository))
	for _, resource := range di.repository {
		list = append(list, resource)
	}
	return list
}

func (di *DirectoryInput) onFileUpdated(name string) {
	logger := di.logger.With().Str("method", "onFileUpdated").Str("file", name).Logger()

	contents, err := os.ReadFile(name)
	if err != nil {
		logger.Error().Err(err).Msg("Could not read input file")
		return
	}

	resource := unstructured.Unstructured{}
	if err := yaml.Unmarshal(contents, &resource.Object); err != nil {
		logger.Error().Err(err).Msg("Could not parse input file as Unstructured")
		return
	}

	gvk := resource.GroupVersionKind()
	logger = logger.With().Str("group", gvk.Group).Str("version", gvk.Version).Str("kind", gvk.Kind).Logger()

	converted, err := di.converter.Convert(&resource)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to convert resource")
		return
	}

	if _, exists := di.repository[converted.Id]; exists {
		if di.fileIDs[name] != converted.Id {
			logger.Warn().Str("id", converted.Id).Msg("Resource already described in another file, skipping")
			return
		}
	}

	di.repository[converted.Id] = *converted
	di.fileIDs[name] = converted.Id
	logger.Trace().Str("id", converted.Id).Msg("Added resource to repository")
}

func (di *DirectoryInput) onFileRemoved(name string) {
	logger := di.logger.With().Str("method", "onFileUpdated").Str("file", name).Logger()

	id, found := di.fileIDs[name]
	if !found {
		logger.Warn().Msg("File was not already loaded, ignoring")
		return
	}

	delete(di.repository, id)
	delete(di.fileIDs, id)
	logger.Trace().Str("id", id).Msg("Removed resource from repository")
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
