//+build dev

package ui

import (
	"sync"
	"github.com/aprice/embed/loader"
)

var _embeddedContentLoader loader.Loader
var _initOnce sync.Once

// GetEmbeddedContent returns the Loader for embedded content files.
func GetEmbeddedContent() loader.Loader {
	_initOnce.Do(func() {
		_embeddedContentLoader = loader.NewOnDisk("ui")
	})
	return _embeddedContentLoader
}
