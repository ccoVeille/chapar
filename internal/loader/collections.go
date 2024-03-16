package loader

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/mirzakhany/chapar/internal/domain"
)

// Directory structure of collections
// /collections
//   /collection1
//  	 _collection.yaml  # Metadata file
//		 request1.yaml
//       request2.yaml
//   /collection2
//  	 _collection.yaml  # Metadata file
//		 request1.yaml
//       request2.yaml

func LoadCollections() ([]*domain.Collection, error) {
	dir, err := GetCollectionsDir()
	if err != nil {
		return nil, err
	}

	out := make([]*domain.Collection, 0)

	// Walk through the collections directory
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory
		if path == dir {
			return nil
		}

		// If it's a directory, it's a collection
		if info.IsDir() {
			col, err := loadCollection(path)
			if err != nil {
				fmt.Println("failed to load collection", path, err)
				return err
			}
			out = append(out, col)
		}

		// Skip further processing since we're only interested in directories here
		return filepath.SkipDir
	})

	return out, err
}

func loadCollection(collectionPath string) (*domain.Collection, error) {
	// Read the collection metadata
	collectionMetadataPath := filepath.Join(collectionPath, "_collection.yaml")
	collectionMetadata, err := os.ReadFile(collectionMetadataPath)
	if err != nil {
		return nil, err
	}

	collection := &domain.Collection{}
	if err = yaml.Unmarshal(collectionMetadata, collection); err != nil {
		fmt.Println(collectionMetadataPath, err)
		return nil, err
	}

	collection.FilePath = collectionMetadataPath
	collection.Spec.Requests = make([]*domain.Request, 0)

	// Load requests in the collection
	files, err := os.ReadDir(collectionPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || file.Name() == "_collection.yaml" {
			continue // Skip directories and the collection metadata file
		}

		requestPath := filepath.Join(collectionPath, file.Name())
		req, err := LoadFromYaml[domain.Request](requestPath)
		if err != nil {
			return nil, err
		}

		setRequestDefaultValues(req)

		req.FilePath = requestPath
		req.CollectionName = collection.MetaData.Name
		collection.Spec.Requests = append(collection.Spec.Requests, req)
	}
	return collection, nil
}

func GetCollectionsDir() (string, error) {
	dir, err := CreateConfigDir()
	if err != nil {
		return "", err
	}

	cdir := path.Join(dir, collectionsDir)
	if _, err := os.Stat(cdir); os.IsNotExist(err) {
		if err := os.Mkdir(cdir, 0755); err != nil {
			return "", err
		}
	}

	return cdir, nil
}

func UpdateCollection(collection *domain.Collection) error {
	if collection.FilePath == "" {
		dirName, err := getNewCollectionDirName(collection.MetaData.Name)
		if err != nil {
			return err
		}
		collection.FilePath = filepath.Join(dirName, "_collection.yaml")
	}

	if err := SaveToYaml(collection.FilePath, collection); err != nil {
		return err
	}

	// Get the directory name
	dirName := path.Dir(collection.FilePath)
	// Change the directory name to the collection name
	if collection.MetaData.Name != path.Base(dirName) {
		// replace last part of the path with the new name
		newDirName := path.Join(path.Dir(dirName), collection.MetaData.Name)
		if err := os.Rename(dirName, newDirName); err != nil {
			return err
		}
		collection.FilePath = filepath.Join(newDirName, "_collection.yaml")
	}

	return nil
}

func DeleteCollection(collection *domain.Collection) error {
	return os.RemoveAll(path.Dir(collection.FilePath))
}
