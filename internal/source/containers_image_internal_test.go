package source

import (
	"archive/tar"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/rand"
)

func TestContainersImage_applyLayerFilter(t *testing.T) {
	type testCase struct {
		name       string
		srcPaths   []string
		tarHeaders []tar.Header
		assertion  func(*tar.Header, bool, error)
	}
	for _, tc := range []testCase{
		{
			name:     "everything found when srcPaths represent root",
			srcPaths: []string{"", "/"},
			tarHeaders: []tar.Header{
				{
					Name: "file",
				},
				{
					Name: "/file",
				},
				{
					Name: "/nested/file",
				},
				{
					Name: "/deeply/nested/file",
				},
			},
			assertion: func(tarHeader *tar.Header, keep bool, err error) {
				assert.True(t, keep)
				assert.NoError(t, err)
			},
		},
		{
			name:     "nothing found outside of srcPath",
			srcPaths: []string{"source"},
			tarHeaders: []tar.Header{
				{
					Name: "elsewhere",
				},
				{
					Name: "/elsewhere",
				},
				{
					Name: "/nested/elsewhere",
				},
				{
					Name: "/deeply/nested/elsewhere",
				},
			},
			assertion: func(tarHeader *tar.Header, keep bool, err error) {
				assert.False(t, keep)
				assert.NoError(t, err)
			},
		},
		{
			name:     "absolute paths are trimmed",
			srcPaths: []string{"source", "/source"},
			tarHeaders: []tar.Header{
				{
					Name: "source",
				},
				{
					Name: "/source",
				},
				{
					Name: "source/nested/elsewhere",
				},
				{
					Name: "/source/nested/elsewhere",
				},
				{
					Name: "source/deeply/nested/elsewhere",
				},
				{
					Name: "/source/deeply/nested/elsewhere",
				},
			},
			assertion: func(tarHeader *tar.Header, keep bool, err error) {
				assert.True(t, keep)
				assert.NoError(t, err)
			},
		},
		{
			name:     "up level source paths are not supported",
			srcPaths: []string{"../not-supported"},
			tarHeaders: []tar.Header{
				{
					Name: "anything",
				},
			},
			assertion: func(tarHeader *tar.Header, keep bool, err error) {
				assert.False(t, keep)
				assert.ErrorContains(t, err, "error getting relative path")
			},
		},
		{
			name:     "up level tar headers are not supported",
			srcPaths: []string{"fine"},
			tarHeaders: []tar.Header{
				{
					Name: "../not-supported",
				},
				{
					Name: "../fine",
				},
			},
			assertion: func(tarHeader *tar.Header, keep bool, err error) {
				assert.False(t, keep)
				assert.NoError(t, err)
			},
		},
		{
			name:     "correctly sets permissions, uid/gid, and file attributes",
			srcPaths: []string{"foo"},
			tarHeaders: []tar.Header{
				{
					Name: "foo/bar",
					Mode: 0000,
					Uid:  rand.Int(),
					Gid:  rand.Int(),
					Xattrs: map[string]string{ //nolint:staticcheck
						"foo": "bar",
					},
					PAXRecords: map[string]string{
						"fizz": "buzz",
					},
				},
			},
			assertion: func(tarHeader *tar.Header, keep bool, err error) {
				assert.True(t, keep)
				assert.NoError(t, err)
				assert.Equal(t, "foo/bar", tarHeader.Name)
				assert.Equal(t, int64(0700), tarHeader.Mode)
				assert.Equal(t, os.Getuid(), tarHeader.Uid)
				assert.Equal(t, os.Getgid(), tarHeader.Gid)
				assert.Nil(t, tarHeader.PAXRecords)
				assert.Nil(t, tarHeader.Xattrs) //nolint:staticcheck
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, srcPath := range tc.srcPaths {
				f := applyLayerFilter(srcPath)
				for _, tarHeader := range tc.tarHeaders {
					keep, err := f(&tarHeader)
					tc.assertion(&tarHeader, keep, err)
				}
			}
		})
	}
}
