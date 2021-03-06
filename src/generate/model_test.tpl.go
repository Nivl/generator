package generate

var modelTestTpl = `package {{.PackageName}}

// Code generated; DO NOT EDIT.

import (
	{{ if .Generate "JoinSQL" -}}"strings"{{- end }}

	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/satori/go.uuid"

	"github.com/Nivl/go-sqldb/implementations/mocksqldb"

	gomock "github.com/golang/mock/gomock"

	{{ if .Generate "doCreate" }}"github.com/Nivl/go-types/datetime"{{ end }}
)

{{ if .Generate "JoinSQL" -}}
func TestJoin{{.OptionalName}}SQL(t *testing.T) {
	fields := []string{ {{.FieldsAsArray}} }
	totalFields := len(fields)
	output := Join{{.OptionalName}}SQL("tofind")

	assert.Equal(t, totalFields*2, strings.Count(output, "tofind."), "wrong number of fields returned")
	assert.True(t, strings.HasSuffix(output, "\""), "JoinSQL() output should end with a \"")
}
{{- end }}

{{ if .Generate "Get" -}}
func TestGet{{.OptionalName}}ByID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expectedID := "4408d5e1-b510-42cb-8ff8-788948a246dd"
	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().GetID(&{{.ModelName}}{}, expectedID, nil)

	_, err := Get{{.OptionalName}}ByID(mockDB, expectedID)
	assert.NoError(t, err, "Get{{.OptionalName}}ByID() should not have failed")
}
{{- end }}

{{ if .Generate "GetAny" -}}
func TestGetAny{{.OptionalName}}ByID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expectedID := "4408d5e1-b510-42cb-8ff8-788948a246dd"
	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().GetID(&{{.ModelName}}{}, expectedID, nil)

	_, err := GetAny{{.OptionalName}}ByID(mockDB, expectedID)
	assert.NoError(t, err, "Get{{.OptionalName}}ByID() should not have failed")
}
{{- end }}

{{ if and (.Generate "Save") (.UseUUID) -}}
func Test{{.ModelName}}SaveNew(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().InsertSuccess(&{{.ModelName}}{})

	{{.ModelVar}} := &{{.ModelName}}{}
	err := {{.ModelVar}}.Save(mockDB)

	assert.NoError(t, err, "Save() should not have fail")
	assert.NotEmpty(t, {{.ModelVar}}.ID, "ID should have been set")
}

func Test{{.ModelName}}SaveExisting(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().UpdateSuccess(&{{.ModelName}}{})

	{{.ModelVar}} := &{{.ModelName}}{}
	id := uuid.NewV4().String()
	{{.ModelVar}}.ID = id
	err := {{.ModelVar}}.Save(mockDB)

	assert.NoError(t, err, "Save() should not have fail")
	assert.Equal(t, id, {{.ModelVar}}.ID, "ID should not have changed")
}
{{- end }}

{{ if .Generate "Create" -}}
func Test{{.ModelName}}Create(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().InsertSuccess(&{{.ModelName}}{})

	{{.ModelVar}} := &{{.ModelName}}{}
	{{ if not (.UseUUID) }}{{.ModelVar}}.ID = uuid.NewV4().String(){{ end}}
	err := {{.ModelVar}}.Create(mockDB)

	assert.NoError(t, err, "Create() should not have fail")
	assert.NotEmpty(t, {{.ModelVar}}.ID, "ID should have been set")
	assert.NotNil(t, {{.ModelVar}}.CreatedAt, "CreatedAt should have been set")
	assert.NotNil(t, {{.ModelVar}}.UpdatedAt, "UpdatedAt should have been set")
}


{{ if .UseUUID }}
func Test{{.ModelName}}CreateWithID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)

	{{.ModelVar}} := &{{.ModelName}}{}
	{{.ModelVar}}.ID = uuid.NewV4().String()

	err := {{.ModelVar}}.Create(mockDB)
	assert.Error(t, err, "Create() should have fail")
}
{{- end }}
{{- end }}

{{ if .Generate "doCreate" -}}
func Test{{.ModelName}}DoCreate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().InsertSuccess(&{{.ModelName}}{})

	{{.ModelVar}} := &{{.ModelName}}{}
	{{ if not (.UseUUID) }}{{.ModelVar}}.ID = uuid.NewV4().String(){{ end}}
	err := {{.ModelVar}}.doCreate(mockDB)

	assert.NoError(t, err, "doCreate() should not have fail")
	assert.NotEmpty(t, {{.ModelVar}}.ID, "ID should have been set")
	assert.NotNil(t, {{.ModelVar}}.CreatedAt, "CreatedAt should have been set")
	assert.NotNil(t, {{.ModelVar}}.UpdatedAt, "UpdatedAt should have been set")
}

func Test{{.ModelName}}DoCreateWithDate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().InsertSuccess(&{{.ModelName}}{})

	createdAt := datetime.Now().AddDate(0, 0, 1)
	{{.ModelVar}} := &{{.ModelName}}{CreatedAt: createdAt}
	{{ if not (.UseUUID) }}{{.ModelVar}}.ID = uuid.NewV4().String(){{ end }}
	err := {{.ModelVar}}.doCreate(mockDB)

	assert.NoError(t, err, "doCreate() should not have fail")
	assert.NotEmpty(t, {{.ModelVar}}.ID, "ID should have been set")
	assert.True(t, {{.ModelVar}}.CreatedAt.Equal(createdAt), "CreatedAt should not have been updated")
	assert.NotNil(t, {{.ModelVar}}.UpdatedAt, "UpdatedAt should have been set")
}

func Test{{.ModelName}}DoCreateFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().InsertError(&{{.ModelName}}{}, errors.New("sql error"))

	{{.ModelVar}} := &{{.ModelName}}{}
	err := {{.ModelVar}}.doCreate(mockDB)

	assert.Error(t, err, "doCreate() should have fail")
}
{{- end }}


{{ if .Generate "Update" -}}
func Test{{.ModelName}}Update(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().UpdateSuccess(&{{.ModelName}}{})

	{{.ModelVar}} := &{{.ModelName}}{}
	{{.ModelVar}}.ID = uuid.NewV4().String()
	err := {{.ModelVar}}.Update(mockDB)

	assert.NoError(t, err, "Update() should not have fail")
	assert.NotNil(t, {{.ModelVar}}.UpdatedAt, "UpdatedAt should have been set")
}

func Test{{.ModelName}}UpdateWithoutID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	{{.ModelVar}} := &{{.ModelName}}{}
	err := {{.ModelVar}}.Update(mockDB)

	assert.Error(t, err, "Update() should not have fail")
}
{{- end }}


{{ if .Generate "doUpdate" -}}
func Test{{.ModelName}}DoUpdate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().UpdateSuccess(&{{.ModelName}}{})

	{{.ModelVar}} := &{{.ModelName}}{}
	{{.ModelVar}}.ID = uuid.NewV4().String()
	err := {{.ModelVar}}.doUpdate(mockDB)

	assert.NoError(t, err, "doUpdate() should not have fail")
	assert.NotEmpty(t, {{.ModelVar}}.ID, "ID should have been set")
	assert.NotNil(t, {{.ModelVar}}.UpdatedAt, "UpdatedAt should have been set")
}

func Test{{.ModelName}}DoUpdateWithoutID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	{{.ModelVar}} := &{{.ModelName}}{}
	err := {{.ModelVar}}.doUpdate(mockDB)

	assert.Error(t, err, "doUpdate() should not have fail")
}

func Test{{.ModelName}}DoUpdateFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().UpdateError(&{{.ModelName}}{}, errors.New("sql error"))

	{{.ModelVar}} := &{{.ModelName}}{}
	{{.ModelVar}}.ID = uuid.NewV4().String()
	err := {{.ModelVar}}.doUpdate(mockDB)

	assert.Error(t, err, "doUpdate() should have fail")
}
{{- end }}

{{ if .Generate "Delete" -}}
func Test{{.ModelName}}Delete(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().DeletionSuccess()

	{{.ModelVar}} := &{{.ModelName}}{}
	{{.ModelVar}}.ID = uuid.NewV4().String()
	err := {{.ModelVar}}.Delete(mockDB)

	assert.NoError(t, err, "Delete() should not have fail")
}

func Test{{.ModelName}}DeleteWithoutID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	{{.ModelVar}} := &{{.ModelName}}{}
	err := {{.ModelVar}}.Delete(mockDB)

	assert.Error(t, err, "Delete() should have fail")
}

func Test{{.ModelName}}DeleteError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocksqldb.NewMockQueryable(mockCtrl)
	mockDB.EXPECT().DeletionError(errors.New("sql error"))

	{{.ModelVar}} := &{{.ModelName}}{}
	{{.ModelVar}}.ID = uuid.NewV4().String()
	err := {{.ModelVar}}.Delete(mockDB)

	assert.Error(t, err, "Delete() should have fail")
}
{{- end }}

{{ if .Generate "IsZero" -}}
func Test{{.ModelName}}IsZero(t *testing.T) {
	empty := &{{.ModelName}}{}
	assert.True(t, empty.IsZero(), "IsZero() should return true for empty struct")

	var nilStruct *{{.ModelName}}
	assert.True(t, nilStruct.IsZero(), "IsZero() should return true for nil struct")

	valid := &{{.ModelName}}{ID: uuid.NewV4().String()}
	assert.False(t, valid.IsZero(), "IsZero() should return false for valid struct")
}
{{- end }}`
