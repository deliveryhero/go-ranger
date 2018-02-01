package ranger_os_test

import (
	"io/ioutil"
	"os"
	"testing"

	goranger "github.com/foodora/go-ranger/ranger_os"
	"github.com/stretchr/testify/assert"
)

func TestExportEnvVars_InvalidFile(t *testing.T) {
	err := goranger.ExportEnvVars("./invalid-file.sh")
	assert.True(t, os.IsNotExist(err))
}

func TestExportEnvVars_IgnoreEmptyLines(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "go-ranger-utils")
	defer os.Remove(tmpfile.Name())

	content := []byte("\n\n")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	err := goranger.ExportEnvVars(tmpfile.Name())
	assert.Nil(t, err)
}

func TestExportEnvVars_IgnoreComments(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "go-ranger-utils")
	defer os.Remove(tmpfile.Name())

	content := []byte(`# ignore empty line`)
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	err := goranger.ExportEnvVars(tmpfile.Name())
	assert.Nil(t, err)
}

func TestExportEnvVars_WithSpaceInKey(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "go-ranger-utils")
	defer os.Remove(tmpfile.Name())

	content := []byte(`export COMPANY =pandora`)
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	err := goranger.ExportEnvVars(tmpfile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "pandora", os.Getenv("COMPANY"))
}

func TestExportEnvVars_WithSpaceInValue(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "go-ranger-utils")
	defer os.Remove(tmpfile.Name())

	content := []byte(`export COMPANY= pandora`)
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	err := goranger.ExportEnvVars(tmpfile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "pandora", os.Getenv("COMPANY"))
}

func TestExportEnvVars_WithoutExport(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "go-ranger-utils")
	defer os.Remove(tmpfile.Name())

	content := []byte(`COMPANY=pandora`)
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	err := goranger.ExportEnvVars(tmpfile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "pandora", os.Getenv("COMPANY"))
}

func TestExportEnvVars_WithSingleQuote(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "go-ranger-utils")
	defer os.Remove(tmpfile.Name())

	content := []byte(`COMPANY='pandora'`)
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	err := goranger.ExportEnvVars(tmpfile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "pandora", os.Getenv("COMPANY"))
}

func TestExportEnvVars_WithDoubleQuote(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "go-ranger-utils")
	defer os.Remove(tmpfile.Name())

	content := []byte(`COMPANY="pandora"`)
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	err := goranger.ExportEnvVars(tmpfile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "pandora", os.Getenv("COMPANY"))
}

func TestExportEnvVars_WithoutQuotes(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "go-ranger-utils")
	defer os.Remove(tmpfile.Name())

	content := []byte(`COMPANY=pandora`)
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	err := goranger.ExportEnvVars(tmpfile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "pandora", os.Getenv("COMPANY"))
}

func TestExportEnvVars_WithEmptyValue(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "go-ranger-utils")
	defer os.Remove(tmpfile.Name())

	content := []byte(`COMPANY=`)
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	err := goranger.ExportEnvVars(tmpfile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "", os.Getenv("COMPANY"))
}

func TestExportEnvVars_AllTogether(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "go-ranger-utils")
	defer os.Remove(tmpfile.Name())

	content := []byte(`
# ignore empty line

# comments too
WITH_SPACE_KEY =pandora
WITH_SPACE_VALUE= pandora
WITHOUT_EXPORT=1
export SINGLE_QUOTES=' pandora'
export DOUBLE_QUOTES=" pandora"
export WITHOUT_QUOTES=pandora
export EMPTY_VALUE=
`)
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	err := goranger.ExportEnvVars(tmpfile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "pandora", os.Getenv("WITH_SPACE_KEY"))
	assert.Equal(t, "pandora", os.Getenv("WITH_SPACE_VALUE"))
	assert.Equal(t, "1", os.Getenv("WITHOUT_EXPORT"))
	assert.Equal(t, " pandora", os.Getenv("SINGLE_QUOTES"))
	assert.Equal(t, " pandora", os.Getenv("DOUBLE_QUOTES"))
	assert.Equal(t, "pandora", os.Getenv("WITHOUT_QUOTES"))
	assert.Equal(t, "", os.Getenv("EMPTY_VALUE"))
}
