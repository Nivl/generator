package generate

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/serenize/snaker"
	"github.com/urfave/cli"
)

// Regex to parse db struct tag like `db:"field_name"`
var dbFieldRegex, _ = regexp.Compile("`db:\"([a-zA-Z0-9_-]+),?.*\"`")

// ModelField represents a field from the struct we are parsing
type ModelField struct {
	Name   string
	DbName string
}

// Model represent a model to generate
type Model struct {
	// Name contains the name of the Model
	Name string

	// Table contains the name of the SQL table
	Table string

	// FileName cotains the name of the file containing the model
	FileName string

	// PackageName cotains the name of the package containing the model
	PackageName string

	// FullPath cotains the path to the file containing the model
	FullPath string

	// Fields cotains all the fields of the model
	Fields []*ModelField

	// Excluded contains all the method not to generate
	Excluded []string

	// IsSingle is used to name the CRUD methods accordingly to the package
	// Example: here's the name of the finder by id depending on IsSingle
	// false, GetUserByID
	// true: GetByID
	IsSingle bool

	// UseUUID is used to auto-populate the ID fields
	UseUUID bool
}

// ModelTemplateVars contains all the variable needed to render the new file
type ModelTemplateVars struct {
	ModelName      string
	ModelNameLC    string
	TableName      string
	ModelVar       string
	PackageName    string
	CreateStmt     string
	CreateStmtArgs string
	UpdateStmt     string
	UpdateStmtArgs string
	FieldsAsArray  string
	Excluded       []string
	IsSingle       bool
	UseUUID        bool
}

// Generate returns true if the element has not been excluded
func (mtv *ModelTemplateVars) Generate(wanted string) bool {
	for _, name := range mtv.Excluded {
		if name == wanted {
			return false
		}
	}
	return true
}

// OptionalName returns the model name if single is set to false. returns
// an empty string otherwise
func (mtv *ModelTemplateVars) OptionalName() string {
	name := ""
	if !mtv.IsSingle {
		name = mtv.ModelName
	}
	return name
}

// setDefault control what has been set in the model, and set default values where needed
func (m *Model) setDefault() error {
	if m.Name == "" {
		return errors.New("model name missing")
	}

	if m.FileName == "" {
		return errors.New("filename missing. use -f to specify one")
	}

	if m.PackageName == "" {
		return errors.New("package name missing. use -p to specify one")
	}

	if m.Table == "" {
		m.Table = snaker.CamelToSnake(m.Name)
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	m.FullPath = path.Join(pwd, m.FileName)
	return nil
}

// Parse parses and render a model
func (m *Model) Parse() error {
	if err := m.setDefault(); err != nil {
		return err
	}

	// Open the file
	file, err := os.Open(m.FullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Put the content of the file in a string
	fileStr, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// Parse the file
	astFile, err := parser.ParseFile(token.NewFileSet(), "", fileStr, parser.AllErrors)
	if err != nil {
		return err
	}
	if err := m.parseTarget(astFile); err != nil {
		return err
	}

	return m.generateAll()
}

// generateModelFile generates the model file
func (m *Model) generateModelFile(vars *ModelTemplateVars) error {
	// Array of Fields
	fieldsAsArray := make([]string, len(m.Fields))
	for i, field := range m.Fields {
		fieldsAsArray[i] = fmt.Sprintf(`"%s"`, field.DbName)
	}
	vars.FieldsAsArray = strings.Join(fieldsAsArray, ", ")

	// Create Statement
	createFields := make([]string, len(m.Fields))
	createValues := make([]string, len(m.Fields))
	for i, field := range m.Fields {
		createFields[i] = field.DbName
		createValues[i] = ":" + field.DbName
	}
	vars.CreateStmt = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		vars.TableName,
		strings.Join(createFields, ", "),
		strings.Join(createValues, ", "),
	)

	// Update Statement
	updateFields := make([]string, len(m.Fields))

	for i, field := range m.Fields {
		updateFields[i] = fmt.Sprintf("%s=:%s", field.DbName, field.DbName)
	}
	vars.UpdateStmt = fmt.Sprintf(
		"UPDATE %s SET %s WHERE id=:id",
		vars.TableName,
		strings.Join(updateFields, ", "),
	)

	// Get the template and parse it with the variables we have
	t, err := template.New("model").Parse(modelTpl)
	if err != nil {
		fmt.Println(err)
		return err
	}
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, vars); err != nil {
		fmt.Println(err)
		return err
	}
	output := strings.TrimSpace(buf.String())

	// Write the new file to the disk
	newFile, err := os.Create(m.generatedFileNameModel())
	if err != nil {
		return err
	}
	defer newFile.Close()
	if _, err := newFile.WriteString(output); err != nil {
		return err
	}

	return nil
}

// generateTests generates the test file
func (m *Model) generateTestsFile(vars *ModelTemplateVars) error {
	// Get the template and parse it with the variables we have
	t, err := template.New("tests").Parse(modelTestTpl)
	if err != nil {
		fmt.Println(err)
		return err
	}
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, vars); err != nil {
		fmt.Println(err)
		return err
	}
	output := strings.TrimSpace(buf.String())

	// Write the new file to the disk
	newFile, err := os.Create(m.generatedFileNameTests())
	if err != nil {
		return err
	}
	defer newFile.Close()
	if _, err := newFile.WriteString(output); err != nil {
		return err
	}

	return nil
}

// generate generates the model file
func (m *Model) generateAll() error {
	vars := &ModelTemplateVars{
		ModelName:   m.Name,
		ModelNameLC: strings.ToLower(m.Name),
		TableName:   m.Table,
		ModelVar:    string(strings.ToLower(m.Name)[0]),
		PackageName: m.PackageName,
		IsSingle:    m.IsSingle,
		Excluded:    m.Excluded,
		UseUUID:     m.UseUUID,
	}

	// if the ModelVar is `t`, we need to add an extra letter to avoid conflict
	// with the `t` variable from `t *testing.T`
	if vars.ModelVar == "t" {
		vars.ModelVar += string(strings.ToLower(m.Name)[1])
	}

	if err := m.generateModelFile(vars); err != nil {
		return nil
	}
	if err := m.generateTestsFile(vars); err != nil {
		return nil
	}
	return nil
}

// generatedFileNameModel returns the file name of the new file
func (m *Model) generatedFileNameModel() string {
	return strings.TrimSuffix(m.FullPath, ".go") + "_generated.go"
}

// generatedFileNameTests returns the file name of the new test file
func (m *Model) generatedFileNameTests() string {
	return strings.TrimSuffix(m.generatedFileNameModel(), ".go") + "_test.go"
}

// parseTarget parses the source file to get the Model fields
func (m *Model) parseTarget(f *ast.File) error {
	obj, ok := f.Scope.Objects[m.Name]
	if !ok {
		return fmt.Errorf("could not find type %s in %s", m.Name, m.FullPath)
	}
	typeSpec, ok := obj.Decl.(*ast.TypeSpec)
	if !ok {
		return fmt.Errorf("%s is not a type", m.Name)
	}
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return fmt.Errorf("%s is not a struct", m.Name)
	}

	for _, field := range structType.Fields.List {
		// We do not handle structs with no name, and un-exported fields
		// Also, I'm not sure in what case we can have more than one name?
		if len(field.Names) > 0 && field.Names[0].IsExported() {
			// Lets be sure the field has a Tag
			if field.Tag == nil {
				continue
			}
			dbName := dbFieldRegex.FindStringSubmatch(field.Tag.Value)
			// for `db:"name"` the func returns [`db:"name"` name], and we want "name"
			if len(dbName) != 2 {
				continue
			}

			newField := &ModelField{
				Name:   field.Names[0].Name,
				DbName: dbName[1],
			}

			m.Fields = append(m.Fields, newField)
		}
	}

	return nil
}

// GenModel is used to generate a new model
func GenModel(c *cli.Context) error {
	// Parse the excluded data
	excluded := strings.Split(c.String("exclude"), ",")
	for i, name := range excluded {
		excluded[i] = strings.TrimSpace(name)
	}

	model := &Model{
		Name:        c.Args().First(),
		Table:       c.String("table"),
		FileName:    c.String("file"),
		PackageName: c.String("package"),
		IsSingle:    c.BoolT("single"),
		Excluded:    excluded,
		UseUUID:     c.BoolT("use-uuid"),
	}

	return model.Parse()
}
