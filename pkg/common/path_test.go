// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"path/filepath"
	"strings"
	"testing"

	"gotest.tools/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/cli-utils/pkg/testutil"
)

const (
	packageDir        = "test-pkg-dir"
	inventoryFilename = "inventory.yaml"
	podAFilename      = "pod-a.yaml"
	podBFilename      = "pod-b.yaml"
)

var (
	inventoryFilePath = filepath.Join(packageDir, inventoryFilename)
	podAFilePath      = filepath.Join(packageDir, podAFilename)
	podBFilePath      = filepath.Join(packageDir, podBFilename)
)

func setupTestFilesystem(t *testing.T) testutil.TestFilesystem {
	// Create the test filesystem, and add package config files
	// to it.
	t.Log("Creating test filesystem")
	tf := testutil.Setup(t, packageDir)
	t.Logf("Adding File: %s", inventoryFilePath)
	tf.WriteFile(t, inventoryFilePath, inventoryConfigMap)
	t.Logf("Adding File: %s", podAFilePath)
	tf.WriteFile(t, podAFilePath, podA)
	t.Logf("Adding File: %s", podBFilePath)
	tf.WriteFile(t, podBFilePath, podB)
	return tf
}

var inventoryConfigMap = []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: test-namespace
  name: inventory
  labels:
    cli-utils.sigs.k8s.io/inventory-id: test-inventory
`)

var podA = []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: pod-a
  namespace: test-namespace
  labels:
    name: test-pod-label
spec:
  containers:
  - name: kubernetes-pause
    image: k8s.gcr.io/pause:2.0
`)

var podB = []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: pod-a
  namespace: test-namespace
  labels:
    name: test-pod-label
spec:
  containers:
  - name: kubernetes-pause
    image: k8s.gcr.io/pause:2.0
`)

func TestProcessPaths(t *testing.T) {
	tf := setupTestFilesystem(t)
	defer tf.Clean()

	trueVal := true
	testCases := map[string]struct {
		paths                     []string
		expectedFileNameFlags     genericclioptions.FileNameFlags
		errFromDemandOneDirectory string
	}{
		"empty slice means reading from StdIn": {
			paths: []string{},
			expectedFileNameFlags: genericclioptions.FileNameFlags{
				Filenames: &[]string{"-"},
			},
		},
		"single file in slice is error; must be directory": {
			paths: []string{podAFilePath},
			expectedFileNameFlags: genericclioptions.FileNameFlags{
				Filenames: nil,
				Recursive: nil,
			},
			errFromDemandOneDirectory: "argument 'test-pkg-dir/pod-a.yaml' is not but must be a directory",
		},
		"single dir in slice": {
			paths: []string{tf.GetRootDir()},
			expectedFileNameFlags: genericclioptions.FileNameFlags{
				Filenames: &[]string{tf.GetRootDir()},
				Recursive: &trueVal,
			},
		},
		"multiple arguments is an error": {
			paths: []string{podAFilePath, podBFilePath},
			expectedFileNameFlags: genericclioptions.FileNameFlags{
				Filenames: nil,
				Recursive: nil,
			},
			errFromDemandOneDirectory: "specify exactly one directory path argument; rejecting [test-pkg-dir/pod-a.yaml test-pkg-dir/pod-b.yaml]",
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			fileNameFlags, err := DemandOneDirectory(tc.paths)
			assert.DeepEqual(t, tc.expectedFileNameFlags, fileNameFlags)
			if err != nil && err.Error() != tc.errFromDemandOneDirectory {
				assert.Equal(t, err.Error(), tc.errFromDemandOneDirectory)
			}
		})
	}
}

func TestExpandDirErrors(t *testing.T) {
	tf := setupTestFilesystem(t)
	defer tf.Clean()

	testCases := map[string]struct {
		packageDirPath []string
		expandedPaths  []string
		isError        bool
	}{
		"empty path is error": {
			packageDirPath: []string{},
			isError:        true,
		},
		"more than one path is error": {
			packageDirPath: []string{"fakedir1", "fakedir2"},
			isError:        true,
		},
		"path that is not dir is error": {
			packageDirPath: []string{"fakedir1"},
			isError:        true,
		},
		"root package dir excludes inventory object": {
			packageDirPath: []string{tf.GetRootDir()},
			expandedPaths: []string{
				filepath.Join(packageDir, "pod-a.yaml"),
				filepath.Join(packageDir, "pod-b.yaml"),
			},
			isError: false,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			trueVal := true
			filenameFlags := genericclioptions.FileNameFlags{
				Filenames: &tc.packageDirPath,
				Recursive: &trueVal,
			}
			actualFlags, err := ExpandPackageDir(filenameFlags)
			if tc.isError && err == nil {
				t.Fatalf("expected error but received none")
			}
			if !tc.isError {
				if err != nil {
					t.Fatalf("unexpected error received: %v", err)
				}
				actualPaths := *actualFlags.Filenames
				if len(tc.expandedPaths) != len(actualPaths) {
					t.Errorf("expected config filepaths (%s), got (%s)",
						tc.expandedPaths, actualPaths)
				}
				for _, expected := range tc.expandedPaths {
					if !filepathExists(expected, actualPaths) {
						t.Errorf("expected config filepath (%s) in actual filepaths (%s)",
							expected, actualPaths)
					}
				}
				// Check the inventory object is not in the filename flags.
				for _, actualPath := range actualPaths {
					if strings.Contains(actualPath, "inventory.yaml") {
						t.Errorf("inventory object should be excluded")
					}
				}
			}
		})
	}
}

// filepathExists returns true if the passed "filepath" is a substring
// of any of the passed full "filepaths"; false otherwise. For example:
// if filepath = "test/a.yaml", and filepaths includes "/tmp/test/a.yaml",
// this function returns true.
func filepathExists(filepath string, filepaths []string) bool {
	for _, fp := range filepaths {
		if strings.Contains(fp, filepath) {
			return true
		}
	}
	return false
}
