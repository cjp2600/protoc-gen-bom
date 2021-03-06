package plugin

import (
	"fmt"
	"path"
	"strings"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"

	bom "github.com/cjp2600/protoc-gen-bom/plugin/options"
)

type MongoPlugin struct {
	*generator.Generator
	generator.PluginImports
	EmptyFiles     []string
	currentPackage string
	currentFile    *generator.FileDescriptor
	generateCrud   bool

	PrivateEntities map[string]PrivateEntity
	ConvertEntities map[string]ConvertEntity
	Fields          map[string][]*descriptor.FieldDescriptorProto

	usePrimitive bool
	useTime      bool
	useStrconv   bool
	useMongoDr   bool
	useCrud      bool
	useUnsafe    bool
	localName    string
}

type ConvertEntity struct {
	nameFrom string
	nameTo   string
	message  *generator.Descriptor
}

type PrivateEntity struct {
	name    string
	items   []*descriptor.FieldDescriptorProto
	message *generator.Descriptor
}

var ServiceName string

func NewMongoPlugin(generator *generator.Generator) *MongoPlugin {
	return &MongoPlugin{Generator: generator}
}

func (p *MongoPlugin) GenerateImports(file *generator.FileDescriptor) {
	if p.usePrimitive {
		p.Generator.PrintImport("primitive", "go.mongodb.org/mongo-driver/bson/primitive")
	}
	p.Generator.PrintImport("valid", "github.com/asaskevich/govalidator")
	p.Generator.PrintImport("context", "context")
	p.Generator.PrintImport("os", "os")
	p.Generator.PrintImport("time", "time")
	p.Generator.PrintImport("mongo", "go.mongodb.org/mongo-driver/mongo")
	p.Generator.PrintImport("options", "go.mongodb.org/mongo-driver/mongo/options")
	p.Generator.PrintImport("readpref", "go.mongodb.org/mongo-driver/mongo/readpref")
	p.Generator.PrintImport("bom", "github.com/cjp2600/bom")

	//p.Generator.PrintImport("context", "context")
	if p.useTime {
		p.Generator.PrintImport("time", "time")
		p.Generator.PrintImport("ptypes", "github.com/golang/protobuf/ptypes")
	}
	if p.useStrconv {
		p.Generator.PrintImport("strconv", "strconv")
	}

	if p.useUnsafe {
		p.Generator.PrintImport("unsafe", "unsafe")
	}
}

func (p *MongoPlugin) Init(g *generator.Generator) {
	generator.RegisterPlugin(NewMongoPlugin(g))
	p.Generator = g
}

func (p *MongoPlugin) GenerateName(name string) string {
	return name + "Mongo"
}

func (p *MongoPlugin) Generate(file *generator.FileDescriptor) {
	p.PrivateEntities = make(map[string]PrivateEntity)
	p.ConvertEntities = make(map[string]ConvertEntity)
	p.Fields = make(map[string][]*descriptor.FieldDescriptorProto)

	p.PluginImports = generator.NewPluginImports(p.Generator)
	p.localName = generator.FileName(file)
	ServiceName = p.GetServiceName(file)

	p.usePrimitive = false
	p.GenerateBomObject()

	for _, msg := range file.Messages() {
		name := strings.Trim(p.GenerateName(msg.GetName()), " ")
		p.Fields[name] = msg.GetField()
		if val, ok := p.PrivateEntities[name]; ok {
			val.items = msg.GetField()
			p.PrivateEntities[name] = val
		}
	}

	for _, msg := range file.Messages() {
		if bomMessage, ok := p.getMessageOptions(msg); ok {
			if bomMessage.GetModel() {
				name := p.GenerateName(msg.GetName())

				p.setCovertEntities(msg, name)

				p.GenerateToPB(msg)
				p.GenerateToObject(msg)
				p.GenerateObjectId(msg)
				// todo: доделать генерацию конвертации в связанную модель
				//p.GenerateBoundMessage(msg)

				// генерируем основные методы модели
				p.generateModelsStructures(msg)
				p.generateValidationMethods(msg)

				// добавляем круд
				if bomMessage.GetCrud() {
					p.useCrud = true
					p.GenerateBomConnection(msg)
					p.GenerateGetBom(msg)
					p.GenerateContructor(msg)
					p.GenerateInsertMethod(msg)
					p.GenerateFindOneMethod(msg)
					p.GenerateFieldNameMethod(msg)
					p.GenerateWhereFieldMethod(msg)
					p.GenerateUpdateAllMethod(msg)
					p.GenerateUpdateWithoutConditionAllMethod(msg)
					p.GerateWhereId(msg)
					p.GenerateFindMethod(msg)
					p.GenerateEachMethod(msg)
					p.GenerateWhereMethod(msg)
					p.GenerateWhereInMethod(msg)
					p.GenerateOrWhereMethod(msg)
					p.GenerateUpdateOneMethod(msg)
				}

			}
		}
	}
	// generate merge and covert methods
	p.generateEntitiesMethods()
}

func (w *MongoPlugin) setCovertEntities(message *generator.Descriptor, name string) {
	opt, ok := w.getMessageOptions(message)
	if ok {
		if entity := opt.GetConvertTo(); len(entity) > 0 {
			st := strings.Split(entity, ",")
			if len(st) > 0 {
				for _, str := range st {
					nameTo := strings.Trim(w.GenerateName(str), " ")
					w.ConvertEntities[name+":"+nameTo] = ConvertEntity{
						nameFrom: name,
						nameTo:   nameTo,
						message:  message,
					}
				}
			}
		}
	}
}

func (p *MongoPlugin) getMessageOptions(message *generator.Descriptor) (*bom.BomMessageOptions, bool) {
	opt := message.GetOptions()
	if opt != nil {
		v, err := proto.GetExtension(opt, bom.E_Opts)
		if err != nil {
			return nil, false
		}
		bomMessage, ok := v.(*bom.BomMessageOptions)
		if !ok {
			return nil, false
		}
		return bomMessage, true
	}
	return nil, false
}

func (p *MongoPlugin) getFieldOptions(field *descriptor.FieldDescriptorProto) *bom.BomFieldOptions {
	if field.Options == nil {
		return nil
	}
	v, err := proto.GetExtension(field.Options, bom.E_Field)
	if err != nil {
		return nil
	}
	opts, ok := v.(*bom.BomFieldOptions)
	if !ok {
		return nil
	}
	return opts
}

//GenerateBoundMessage
func (p *MongoPlugin) GenerateBoundMessage(message *generator.Descriptor) {
	if opt, ok := p.getMessageOptions(message); ok {
		if len(opt.GetBoundMessage()) > 0 {
			bm := opt.GetBoundMessage()
			//mName := p.GenerateName(message.GetName())
			mName := message.GetName()
			bm = generator.CamelCase(bm)

			p.P(`func (e *`, mName, `) ToBound() (*`, bm, `) {`)
			for _, field := range message.GetField() {
				fieldName := field.GetName()
				fieldName = generator.CamelCase(fieldName)
			}
			p.P(`}`)
		}
	}
}

// GenerateObjectId
func (p *MongoPlugin) GenerateObjectId(message *generator.Descriptor) {
	mName := p.GenerateName(message.GetName())
	for _, field := range message.GetField() {
		fieldName := field.GetName()
		fieldName = generator.CamelCase(fieldName)
		bomField := p.getFieldOptions(field)
		if bomField != nil && bomField.Tag.GetMongoObjectId() {
			if bomField.GetTag().GetIsID() {
				p.usePrimitive = true
				p.P(`func (e *`, mName, `) WithObject`, fieldName, `() *`, mName, ` {`)
				p.P(`e.`, fieldName, ` = primitive.NewObjectID() // create object id`)
				p.P(`return e`)
				p.P(`}`)
			}
		}
	}
}

//p.GenerateContructor(msg)
func (p *MongoPlugin) getCollection(message *generator.Descriptor) string {
	collection := strings.ToLower(message.GetName())
	bomMessage, ok := p.getMessageOptions(message)
	if ok {
		if clt := bomMessage.GetCollection(); len(clt) > 0 {
			collection = clt
		}
	}
	return collection
}

func (p *MongoPlugin) GetServiceName(file *generator.FileDescriptor) string {
	var name string
	for _, svc := range file.Service {
		if svc != nil && svc.Name != nil {
			return *svc.Name
		}
	}
	name = *file.Name
	if ext := path.Ext(name); ext == ".proto" || ext == ".protodevel" {
		name = name[:len(name)-len(ext)]
	}
	return name
}

func (p *MongoPlugin) GenerateContructor(message *generator.Descriptor) {
	gName := p.GenerateName(message.GetName())
	bomMessage, ok := p.getMessageOptions(message)
	if ok {
		collection := strings.ToLower(message.GetName())
		if clt := bomMessage.GetCollection(); len(clt) > 0 {
			collection = clt
		}
		p.P(`//`)
		p.P(`// create `, gName, ` mongo model of protobuf `, message.GetName())
		p.P(`//`)
		p.P(`func New`, gName, `() *`, gName, ` {`)
		p.P(`return &`, gName, `{bom:  `, ServiceName, `BomWrapper().WithColl("`, collection, `")}`)
		p.P(`}`)
	}
}

// GenerateFindMethod
func (p *MongoPlugin) GenerateWhereMethod(message *generator.Descriptor) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`//Deprecated: should use WhereConditions or WhereEq`)
	p.P(`func (e *`, gName, `) Where(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.WhereEq(field, value)`)
	p.P(` return e`)
	p.P(`}`)

	p.P()
	p.P(`// WhereEq method`)
	p.P(`func (e *`, gName, `) WhereEq(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.WhereEq(field, value)`)
	p.P(` return e`)
	p.P(`}`)
	p.P()
	p.P(`// WhereGt method`)
	p.P(`func (e *`, gName, `) WhereGt(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.WhereGt(field, value)`)
	p.P(` return e`)
	p.P(`}`)
	p.P(`// WhereGte method`)
	p.P(`func (e *`, gName, `) WhereGte(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.WhereGte(field, value)`)
	p.P(` return e`)
	p.P(`}`)
	p.P(`// WhereLt method`)
	p.P(`func (e *`, gName, `) WhereLt(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.WhereLt(field, value)`)
	p.P(` return e`)
	p.P(`}`)
	p.P(`// WhereLte method`)
	p.P(`func (e *`, gName, `) WhereLte(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.WhereLte(field, value)`)
	p.P(` return e`)
	p.P(`}`)
	////	bmQuery := mongoModel.GetBom().
	////		WithLimit(&bom.Limit{Page: q.Page, Size: q.Size})

	p.P()
	p.P(`// Limit method`)
	p.P(`func (e *`, gName, `) Limit(page int32, size int32) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.WithLimit(&bom.Limit{Page: page, Size: size})`)
	p.P(` return e`)
	p.P(`}`)

	p.P()
	p.P(`// Size method`)
	p.P(`func (e *`, gName, `) Size(size int32) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.WithSize(size)`)
	p.P(` return e`)
	p.P(`}`)

	p.P()
	p.P(`// LastId method`)
	p.P(`func (e *`, gName, `) LastId(lastId string) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.WithLastID(lastId)`)
	p.P(` return e`)
	p.P(`}`)

	p.P()
	p.P(`// Sort method`)
	p.P(`func (e *`, gName, `) Sort(sortField string, sortType string) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(`if sortField == "id" {`)
	p.P(`sortField = "_id"`)
	p.P(`}`)
	p.P(` e.bom.WithSort(&bom.Sort{Field: sortField, Type: sortType})`)
	p.P(` return e`)
	p.P(`}`)

	p.P()
	p.P(`// Find with pagination`)
	p.P(`func (e *`, gName, `) ListWithPagination() ([]*`, gName, `, *bom.Pagination, error) {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(`var items []*`, gName)
	p.P(`paginator, err := e.bom.ListWithPagination(func(cur *mongo.Cursor) error {`)
	p.P(`var result `, gName, ``)
	p.P(`err := cur.Decode(&result)`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.P(`items = append(items, &result)`)
	p.P(`return nil`)
	p.P(`})`)
	p.P(`return items, paginator, err`)
	p.P(`}`)

	p.P()
	p.P(`// List with last id for fast pagination`)
	p.P(`func (e *`, gName, `) ListWithLastID() ([]*`, gName, `, string, error) {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(`var items []*`, gName)
	p.P(`lastId, err := e.bom.ListWithLastID(func(cur *mongo.Cursor) error {`)
	p.P(`var result `, gName, ``)
	p.P(`err := cur.Decode(&result)`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.P(`items = append(items, &result)`)
	p.P(`return nil`)
	p.P(`})`)
	p.P(`return items, lastId, err`)
	p.P(`}`)

	p.P()
	p.P(`// Find list`)
	p.P(`func (e *`, gName, `) List() ([]*`, gName, `, error) {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(`var items []*`, gName)

	p.P(`err := e.bom.List(func(cur *mongo.Cursor) error {`)
	p.P(`var result `, gName, ``)
	p.P(`err := cur.Decode(&result)`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.P(`items = append(items, &result)`)
	p.P(`return nil`)
	p.P(`})`)

	p.P(`return items, err`)
	p.P(`}`)

	for _, field := range message.GetField() {
		bomField := p.getFieldOptions(field)
		fieldName := field.GetName()
		if bomField != nil && bomField.Tag.GetIsID() {

			fn := strings.ToLower(fieldName)
			if fn == "id" {
				fn = "_id"
			}

			p.P()
			p.P(`// Get bulk map`)
			p.P(`func (e *`, gName, `) GetBulkMap(ids []string) (map[string]*`, gName, `, error) {`)
			p.P(`result := make(map[string]*`, gName, `)`)
			p.P(`items, err := e.WhereIn("`, fn, `", bom.ToObjects(ids)).List()`)
			p.P(`if err != nil {`)
			p.P(`return result, err`)
			p.P(`}`)

			p.P(`for _, v := range items {`)
			p.P(`result[v.`, generator.CamelCase(fieldName), `.Hex()] = v`)
			p.P(`}`)

			p.P(`return result, nil`)
			p.P(`}`)

			p.P()
			p.P(`// Get bulk map`)
			p.P(`func (e *`, gName, `) GetBulk(ids []string) ([]*`, gName, `, error) {`)
			p.P(`return e.WhereIn("`, fn, `", bom.ToObjects(ids)).List()`)
			p.P(`}`)
		}
	}
}

// GenerateEachMethod
func (p *MongoPlugin) GenerateEachMethod(message *generator.Descriptor) {
	source := message.GetName()
	gName := p.GenerateName(message.GetName())

	var useWhereId = false
	for _, field := range message.GetField() {
		fieldName := field.GetName()
		if strings.ToLower(fieldName) == "id" {
			useWhereId = true
		}
	}

	if useWhereId {
		p.P()
		p.P(`// Iteration - full iteration method (note that an anonymous function return false to continue) `)
		p.P(`// the method is based on the last element pagination mechanism `)
		p.P(`func (e *`, gName, `) Iteration(fn func (`, strings.ToLower(source), ` []*`, gName, `) bool , size int32) error { `)
		p.P(`// check if bom object is nil`)
		p.P(`if e.bom == nil {`)
		p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
		p.P(`}`)
		p.P(`// set size`)
		p.P(`e.Size(size)`)
		p.P(``, strings.ToLower(source), `, lastId, err := e.ListWithLastID()`)

		p.P(`// first start`)
		p.P(`if !fn(`, strings.ToLower(source), `) {`)
		p.P(`return nil`)
		p.P(`}`)

		p.P(`for len(lastId) > 0 {`)
		p.P(``, strings.ToLower(source), `, lastId, err = e.LastId(lastId).ListWithLastID()`)
		p.P(`if !fn(`, strings.ToLower(source), `) {`)
		p.P(`continue`)
		p.P(`}`)
		p.P(`}`)

		p.P(`return err`)
		p.P(`}`)
	}
}

func (p *MongoPlugin) GenerateMongoConnection() {
	p.P(`// MongoClient - create mongo connection`)
	p.P(`var MongoClient *mongo.Client`)
	p.P()
	p.P(`// MongoConnection - connection`)
	p.P(`func MongoConnection() (*mongo.Client, error) {`)

	p.P(`ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)`)

	p.P(`dbUrl := os.Getenv("MONGO_URL")`)
	p.P(`if len(dbUrl) == 0 {`)
	p.P(`return nil, fmt.Errorf("MONGO_URL env is empty")`)
	p.P(`}`)

	p.P(`client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbUrl))`)

	p.P(`if err != nil {`)
	p.P(`return client, err`)
	p.P(`}`)

	p.P(`ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)`)
	p.P(`err = client.Ping(ctx, readpref.Primary())`)

	p.P(`if err != nil {`)
	p.P(`return client, err`)
	p.P(`}`)

	p.P(`MongoClient = client`)

	p.P(`dbName := os.Getenv("MONGO_DB_NAME")`)
	p.P(`if len(dbName) == 0 {`)
	p.P(`dbName = "`, strings.ToLower(ServiceName), `"`)
	p.P(`}`)
	p.P(`client.Database(dbName)`)

	p.P(`return client, nil`)
	p.P(`}`)
	p.P()

}

func (p *MongoPlugin) GenerateBomObject() {
	p.useMongoDr = true
	p.P()
	p.GenerateMongoConnection()

	p.P(`// create Bom wrapper (`, ServiceName, `)`)
	p.P(`func `, ServiceName, `BomWrapper() *bom.Bom {`)
	p.P(`dbName := os.Getenv("MONGO_DB_NAME")`)
	p.P(`if len(dbName) == 0 {`)
	p.P(`dbName = "`, strings.ToLower(ServiceName), `"`)
	p.P(`}`)
	p.P(`bomObject, _ := bom.New(`)
	p.P(`bom.SetMongoClient(MongoClient),`)
	p.P(`bom.SetDatabaseName(dbName),`)
	p.P(`)`)
	p.P(`// set global var`)
	p.P(`return bomObject`)
	p.P(`}`)
	p.P()
}

func (p *MongoPlugin) GenerateBomConnection(message *generator.Descriptor) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`// set custom bom wrapper`)
	p.P(`func (e *`, gName, `) SetBom(bom *bom.Bom) *`, gName, ` {`)
	p.P(` e.bom = bom.WithColl("`, p.getCollection(message), `") `)
	p.P(` return e`)
	p.P(`}`)
}

func (p *MongoPlugin) GenerateGetBom(message *generator.Descriptor) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`// GetSourceBom - Get the source object`)
	p.P(`func (e *`, gName, `) GetBom() *bom.Bom {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` return e.bom`)
	p.P(`}`)
}

func (p *MongoPlugin) GenerateWhereInMethod(message *generator.Descriptor) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`// WhereIn method`)
	p.P(`func (e *`, gName, `) WhereIn(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.WhereIn(field, value)`)
	p.P(` return e`)
	p.P(`}`)

	p.P()
	p.P(`// WhereNotIn method`)
	p.P(`func (e *`, gName, `) WhereNotIn(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(`// exist in bom version >= 1.0.11`)
	p.P(` e.bom.NotWhereIn(field, value)`)
	p.P(` return e`)
	p.P(`}`)
}

func (p *MongoPlugin) GenerateOrWhereMethod(message *generator.Descriptor) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`//Deprecated: should use OrWhereEq`)
	p.P(`func (e *`, gName, `) OrWhere(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.OrWhereEq(field, value)`)
	p.P(` return e`)
	p.P(`}`)
	p.P(`// WhereNotEq`)
	p.P(`func (e *`, gName, `) WhereNotEq(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.WhereNotEq(field, value)`)
	p.P(` return e`)
	p.P(`}`)
	p.P(`// OrWhereEq method`)
	p.P(`func (e *`, gName, `) OrWhereEq(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.OrWhereEq(field, value)`)
	p.P(` return e`)
	p.P(`}`)
	p.P(`// OrWhereGt method`)
	p.P(`func (e *`, gName, `) OrWhereGt(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.OrWhereGt(field, value)`)
	p.P(` return e`)
	p.P(`}`)
	p.P(`// OrWhereGte method`)
	p.P(`func (e *`, gName, `) OrWhereGte(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.OrWhereGte(field, value)`)
	p.P(` return e`)
	p.P(`}`)
	p.P(`// OrWhereLt method`)
	p.P(`func (e *`, gName, `) OrWhereLt(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.OrWhereLt(field, value)`)
	p.P(` return e`)
	p.P(`}`)
	p.P(`// OrWhereLte method`)
	p.P(`func (e *`, gName, `) OrWhereLte(field string, value interface{}) *`, gName, ` {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(` e.bom.OrWhereLte(field, value)`)
	p.P(` return e`)
	p.P(`}`)
}

// GenerateFindMethod
func (p *MongoPlugin) GenerateFindMethod(message *generator.Descriptor) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`// Find  find method`)
	p.P(`func (e *`, gName, `) FindOne() (*`, gName, `, error) {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(`mongoModel := `, gName, `{}`)
	p.P(` err := e.bom.`)
	p.P(` FindOne(func(s *mongo.SingleResult) error {`)
	p.P(` err := s.Decode(&mongoModel)`)
	p.P(` if err != nil {`)
	p.P(` return err`)
	p.P(` }`)
	p.P(` return nil`)
	p.P(` })`)
	p.P(` return &mongoModel, err`)
	p.P(`}`)
}

//GenerateFieldName method
func (p *MongoPlugin) GenerateFieldNameMethod(message *generator.Descriptor) {

	fields := message.GetField()
	opt, ok := p.getMessageOptions(message)
	if ok {
		if table := opt.GetMerge(); len(table) > 0 {
			st := strings.Split(table, ",")
			if len(st) > 0 {
				for _, str := range st {
					if val, ok := p.Fields[p.GenerateName(str)]; ok {
						for _, f1 := range val {
							fields = append(fields, f1)
						}
					}
				}
			}
		}
	}

	for _, field := range fields {

		//nullable := gogoproto.IsNullable(field)
		repeated := field.IsRepeated()
		fieldName := field.GetName()
		//oneOf := field.OneofIndex != nil
		//goTyp, _ := p.GoType(message, field)
		fieldName = generator.CamelCase(fieldName)
		mName := p.GenerateName(message.GetName())

		if !field.IsMessage() && !repeated {
			p.useMongoDr = true
			p.P(`// Get`, fieldName, `FieldName - mongo field name`)
			p.P(`func (e *`, mName, `) Get`, fieldName, `FieldName() string {`)
			fn := strings.ToLower(fieldName)
			if fn == "id" {
				fn = "_id"
			}
			p.P(`return "`, fn, `"`)
			p.P(`}`)
			p.P()
		}

	}
}

//GenerateWhereFieldMethod
func (p *MongoPlugin) GenerateWhereFieldMethod(message *generator.Descriptor) {

	fields := message.GetField()
	opt, ok := p.getMessageOptions(message)
	if ok {
		if table := opt.GetMerge(); len(table) > 0 {
			st := strings.Split(table, ",")
			if len(st) > 0 {
				for _, str := range st {
					if val, ok := p.Fields[p.GenerateName(str)]; ok {
						for _, f1 := range val {
							fields = append(fields, f1)
						}
					}
				}
			}
		}
	}

	for _, field := range fields {
		repeated := field.IsRepeated()
		fieldName := field.GetName()
		goTyp, _ := p.GoType(message, field)
		fieldName = generator.CamelCase(fieldName)
		mName := p.GenerateName(message.GetName())

		if !field.IsMessage() && !repeated {
			p.useMongoDr = true
			p.P(`// Where`, fieldName, `Eq - filter method `)
			p.P(`func (e *`, mName, `) Where`, fieldName, `Eq(`, fieldName, ` `, goTyp, `) *`, mName, ` {`)
			fn := `e.Get` + fieldName + `FieldName()`
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				p.P(`return e.WhereEq(`, fn, `, bom.ToObj(`, fieldName, `))`)
			} else {
				p.P(`return e.WhereEq(`, fn, `, `, fieldName, ` )`)
			}
			p.P(`}`)
			p.P()
		}
		if !field.IsMessage() && !repeated {
			p.useMongoDr = true
			p.P(`// OrWhere`, fieldName, `Eq - filter method `)
			p.P(`func (e *`, mName, `) OrWhere`, fieldName, `Eq(`, fieldName, ` `, goTyp, `) *`, mName, ` {`)
			fn := `e.Get` + fieldName + `FieldName()`
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				p.P(`return e.OrWhereEq(`, fn, `, bom.ToObj(`, fieldName, `))`)
			} else {
				p.P(`return e.OrWhereEq(`, fn, `, `, fieldName, ` )`)
			}
			p.P(`}`)
			p.P()
		}
		if !field.IsMessage() && !repeated && !field.IsBool() {
			p.useMongoDr = true
			p.P(`// Where`, fieldName, `In - filter method `)
			p.P(`func (e *`, mName, `) Where`, fieldName, `In(`, fieldName, ` []`, goTyp, `) *`, mName, ` {`)
			fn := `e.Get` + fieldName + `FieldName()`
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				p.P(`return e.WhereIn(`, fn, `, bom.ToObjects(`, fieldName, `))`)
			} else {
				p.P(`return e.WhereIn(`, fn, `, `, fieldName, ` )`)
			}
			p.P(`}`)
			p.P()
		}
		if !field.IsMessage() && !repeated {
			p.useMongoDr = true
			p.P(`// Where`, fieldName, `NotEq - filter method `)
			p.P(`func (e *`, mName, `) Where`, fieldName, `NotEq(`, fieldName, ` `, goTyp, `) *`, mName, ` {`)
			fn := `e.Get` + fieldName + `FieldName()`
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				p.P(`return e.WhereNotEq(`, fn, `, bom.ToObj(`, fieldName, `))`)
			} else {
				p.P(`return e.WhereNotEq(`, fn, `, `, fieldName, ` )`)
			}
			p.P(`}`)
			p.P()
		}
		if !field.IsMessage() && !repeated && field.IsScalar() && !field.IsBool() {
			p.useMongoDr = true
			p.P(`// Where`, fieldName, `Gt - filter method `)
			p.P(`func (e *`, mName, `) Where`, fieldName, `Gt(`, fieldName, ` `, goTyp, `) *`, mName, ` {`)
			fn := `e.Get` + fieldName + `FieldName()`
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				p.P(`return e.WhereGt(`, fn, `, bom.ToObj(`, fieldName, `))`)
			} else {
				p.P(`return e.WhereGt(`, fn, `, `, fieldName, ` )`)
			}
			p.P(`}`)
			p.P()
		}
		if !field.IsMessage() && !repeated && field.IsScalar() && !field.IsBool() {
			p.useMongoDr = true
			p.P(`// OrWhere`, fieldName, `Gt - filter method `)
			p.P(`func (e *`, mName, `) OrWhere`, fieldName, `Gt(`, fieldName, ` `, goTyp, `) *`, mName, ` {`)
			fn := `e.Get` + fieldName + `FieldName()`
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				p.P(`return e.OrWhereGt(`, fn, `, bom.ToObj(`, fieldName, `))`)
			} else {
				p.P(`return e.OrWhereGt(`, fn, `, `, fieldName, ` )`)
			}
			p.P(`}`)
			p.P()
		}
		if !field.IsMessage() && !repeated && field.IsScalar() && !field.IsBool() {
			p.useMongoDr = true
			p.P(`// Where`, fieldName, `Gte - filter method `)
			p.P(`func (e *`, mName, `) Where`, fieldName, `Gte(`, fieldName, ` `, goTyp, `) *`, mName, ` {`)
			fn := `e.Get` + fieldName + `FieldName()`
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				p.P(`return e.WhereGte(`, fn, `, bom.ToObj(`, fieldName, `))`)
			} else {
				p.P(`return e.WhereGte(`, fn, `, `, fieldName, ` )`)
			}
			p.P(`}`)
			p.P()
		}
		if !field.IsMessage() && !repeated && field.IsScalar() && !field.IsBool() {
			p.useMongoDr = true
			p.P(`// OrWhere`, fieldName, `Gte - filter method `)
			p.P(`func (e *`, mName, `) OrWhere`, fieldName, `Gte(`, fieldName, ` `, goTyp, `) *`, mName, ` {`)
			fn := `e.Get` + fieldName + `FieldName()`
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				p.P(`return e.OrWhereGte(`, fn, `, bom.ToObj(`, fieldName, `))`)
			} else {
				p.P(`return e.OrWhereGte(`, fn, `, `, fieldName, ` )`)
			}
			p.P(`}`)
			p.P()
		}
		if !field.IsMessage() && !repeated && field.IsScalar() && !field.IsBool() {
			p.useMongoDr = true
			p.P(`// Where`, fieldName, `Lt - filter method `)
			p.P(`func (e *`, mName, `) Where`, fieldName, `Lt(`, fieldName, ` `, goTyp, `) *`, mName, ` {`)
			fn := `e.Get` + fieldName + `FieldName()`
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				p.P(`return e.WhereLt(`, fn, `, bom.ToObj(`, fieldName, `))`)
			} else {
				p.P(`return e.WhereLt(`, fn, `, `, fieldName, ` )`)
			}
			p.P(`}`)
			p.P()
		}
		if !field.IsMessage() && !repeated && field.IsScalar() && !field.IsBool() {
			p.useMongoDr = true
			p.P(`// OrWhere`, fieldName, `Lt - filter method `)
			p.P(`func (e *`, mName, `) OrWhere`, fieldName, `Lt(`, fieldName, ` `, goTyp, `) *`, mName, ` {`)
			fn := `e.Get` + fieldName + `FieldName()`
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				p.P(`return e.OrWhereLt(`, fn, `, bom.ToObj(`, fieldName, `))`)
			} else {
				p.P(`return e.OrWhereLt(`, fn, `, `, fieldName, ` )`)
			}
			p.P(`}`)
			p.P()
		}
		if !field.IsMessage() && !repeated && field.IsScalar() && !field.IsBool() {
			p.useMongoDr = true
			p.P(`// Where`, fieldName, `Lte - filter method `)
			p.P(`func (e *`, mName, `) Where`, fieldName, `Lte(`, fieldName, ` `, goTyp, `) *`, mName, ` {`)
			fn := `e.Get` + fieldName + `FieldName()`
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				p.P(`return e.WhereLte(`, fn, `, bom.ToObj(`, fieldName, `))`)
			} else {
				p.P(`return e.WhereLte(`, fn, `, `, fieldName, ` )`)
			}
			p.P(`}`)
			p.P()
		}
		if !field.IsMessage() && !repeated && field.IsScalar() && !field.IsBool() {
			p.useMongoDr = true
			p.P(`// OrWhere`, fieldName, `Lte - filter method `)
			p.P(`func (e *`, mName, `) OrWhere`, fieldName, `Lte(`, fieldName, ` `, goTyp, `) *`, mName, ` {`)
			fn := `e.Get` + fieldName + `FieldName()`
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				p.P(`return e.OrWhereLte(`, fn, `, bom.ToObj(`, fieldName, `))`)
			} else {
				p.P(`return e.OrWhereLte(`, fn, `, `, fieldName, ` )`)
			}
			p.P(`}`)
			p.P()
		}

	}
}

//GenerateFindOneMethod
func (p *MongoPlugin) GenerateFindOneMethod(message *generator.Descriptor) {

	fields := message.GetField()
	opt, ok := p.getMessageOptions(message)
	if ok {
		if table := opt.GetMerge(); len(table) > 0 {
			st := strings.Split(table, ",")
			if len(st) > 0 {
				for _, str := range st {
					if val, ok := p.Fields[p.GenerateName(str)]; ok {
						for _, f1 := range val {
							fields = append(fields, f1)
						}
					}
				}
			}
		}
	}

	for _, field := range fields {

		//nullable := gogoproto.IsNullable(field)
		repeated := field.IsRepeated()
		fieldName := field.GetName()
		//oneOf := field.OneofIndex != nil
		goTyp, _ := p.GoType(message, field)
		fieldName = generator.CamelCase(fieldName)
		mName := p.GenerateName(message.GetName())

		if !field.IsMessage() && !repeated {
			p.useMongoDr = true
			p.P(`// FindOneBy`, fieldName, ` - find method`)
			p.P(`func (e *`, mName, `) FindOneBy`, fieldName, `(`, fieldName, ` `, goTyp, `) (*`, mName, `, error) {`)
			p.P(`// check if bom object is nil`)
			p.P(`if e.bom == nil {`)
			p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
			p.P(`}`)
			p.P(`mongoModel := `, mName, `{}`)
			p.P(` err := e.bom.`)
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				fn := `e.Get` + fieldName + `FieldName()`
				p.P(` WhereEq(`, fn, `, bom.ToObj(`, fieldName, `)).`)
			} else {
				fn := `e.Get` + fieldName + `FieldName()`
				p.P(` WhereEq(`, fn, `, `, fieldName, ` ).`)
			}

			p.P(` FindOne(func(s *mongo.SingleResult) error {`)
			p.P(` err := s.Decode(&mongoModel)`)
			p.P(` if err != nil {`)
			p.P(` return err`)
			p.P(` }`)
			p.P(` return nil`)
			p.P(` })`)
			p.P(` return &mongoModel, err`)
			p.P(`}`)
			p.P()
		}

	}
}

//GenerateUpdateOneMethod
func (p *MongoPlugin) GenerateUpdateOneMethod(message *generator.Descriptor) {
	var useWhereId = false
	for _, field := range message.GetField() {
		fieldName := field.GetName()
		if strings.ToLower(fieldName) == "id" {
			useWhereId = true
		}
	}
	mName := p.GenerateName(message.GetName())

	fields := message.GetField()
	opt, ok := p.getMessageOptions(message)
	if ok {
		if table := opt.GetMerge(); len(table) > 0 {
			st := strings.Split(table, ",")
			if len(st) > 0 {
				for _, str := range st {
					if val, ok := p.Fields[p.GenerateName(str)]; ok {
						for _, f1 := range val {
							fields = append(fields, f1)
						}
					}
				}
			}
		}
	}

	for _, field := range fields {
		fieldName := field.GetName()

		if strings.EqualFold(fieldName, "id") {
			continue
		}

		goTyp, _ := p.GoType(message, field)
		fieldName = generator.CamelCase(fieldName)

		if strings.EqualFold(goTyp, "*timestamppb.Timestamp") {
			goTyp = "time.Time"
			p.useTime = true
		} else if p.IsMap(field) {
			m, _ := p.GoMapTypeCustomMongo(nil, field)
			goTyp = m.GoType
		} else {
			if field.IsMessage() {
				goTyp = p.GenerateName(goTyp)
			}
		}

		p.useMongoDr = true
		p.P(`// Update`, fieldName, ` - update field`)
		p.P(`func (e *`, mName, `) Update`, fieldName, `(`, fieldName, ` `, goTyp, `, updateAt bool) (*`, mName, `, error) {`)

		p.P(`// check if bom object is nil`)
		p.P(`if e.bom == nil {`)
		p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
		p.P(`}`)

		if useWhereId {
			p.P(`// check if fil _id field`)
			p.P(`if !e.Id.IsZero() {`)
			p.P(`e.WhereId(e.Id.Hex())`)
			p.P(`}`)
		}

		p.P(`// mongoModel := `, mName, `{}`)
		p.P(` if !updateAt {`)
		p.P(` _, err := e.bom.UpdateRaw(primitive.D{`)
		p.P(` {Key: "$set", Value: primitive.D{{"`, strings.ToLower(fieldName), `", `, fieldName, `}}},`)
		p.P(` })`)
		p.P(` if err != nil {`)
		p.P(` return e, err`)
		p.P(` }`)

		p.P(` } else {`)
		p.P(` _, err := e.bom.UpdateRaw(primitive.D{`)
		p.P(` {Key: "$set", Value: primitive.D{{"`, strings.ToLower(fieldName), `", `, fieldName, `}}},`)
		p.P(` {Key: "$currentDate", Value: primitive.D{{"updatedat", true}}},`)
		p.P(` })`)
		p.P(` if err != nil {`)
		p.P(` return e, err`)
		p.P(` }`)
		p.P(` }`)

		p.P(` return e, nil`)
		p.P(` }`)
		p.P()

	}
}

//GerateWhereId
func (p *MongoPlugin) GerateWhereId(message *generator.Descriptor) {
	mName := p.GenerateName(message.GetName())
	for _, field := range message.GetField() {
		fieldName := field.GetName()
		fieldName = generator.CamelCase(fieldName)
		bomField := p.getFieldOptions(field)
		if bomField != nil && bomField.Tag.GetMongoObjectId() {

			if bomField.GetTag().GetIsID() {
				p.usePrimitive = true
				f := strings.ToLower(fieldName)
				if f == "id" {
					f = "_id"
				}
				p.P(`func (e *`, mName, `) Where`, fieldName, `(id string) *`, mName, ` {`)
				p.P(`// check if bom object is nil`)
				p.P(`if e.bom == nil {`)
				p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
				p.P(`}`)
				p.P(`e.bom.WhereEq("`, f, `", bom.ToObj(id))`)
				p.P(`return e`)
				p.P(`}`)

				p.P()
				p.P(`func (e *`, mName, `) Where`, fieldName, `s (ids []string) *`, mName, ` {`)
				p.P(`// check if bom object is nil`)
				p.P(`if e.bom == nil {`)
				p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
				p.P(`}`)
				p.P(`e.bom.WhereIn("`, f, `", bom.ToObjects(ids))`)
				p.P(`return e`)
				p.P(`}`)
				p.P()
			}

		}
	}
}

//GenerateInsertMethod
func (p *MongoPlugin) GenerateInsertMethod(message *generator.Descriptor) {
	//typeName := p.GenerateName(message.GetName())
	mName := p.GenerateName(message.GetName())
	p.usePrimitive = true
	useId := false
	p.P()
	p.P(`// InsertOne method`)
	p.P(`func (e *`, mName, `) InsertOne() (*`, mName, `, error) {`)

	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)

	for _, field := range message.GetField() {
		fieldName := field.GetName()
		fieldName = generator.CamelCase(fieldName)
		bomField := p.getFieldOptions(field)
		if bomField != nil && bomField.Tag.GetMongoObjectId() {
			if bomField.GetTag().GetIsID() {
				useId = true
				p.P(`e.`, fieldName, ` = primitive.NewObjectID() // create object id`)
			}
		}
	}
	if useId {
		p.P(`res, err := e.bom.InsertOne(e)`)
	} else {
		p.P(`_, err := e.bom.InsertOne(e)`)
	}
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if useId {
		p.P(`if insertId, ok := res.InsertedID.(primitive.ObjectID); ok {`)
		p.P(`e.Id = insertId`)
		p.P(`}`)
	}
	p.P(`return e, nil`)
	p.P(`}`)
}

//GenerateBehaviorInterface
func (p *MongoPlugin) GenerateBehaviorInterface(message *generator.Descriptor) {
	p.In()
	typeName := p.GenerateName(message.GetName())
	p.P(`// The following are interfaces you can implement for special behavior during Mongo/PB conversions`)
	p.P(`// of type `, typeName, ` the arg will be the target, the caller the one being converted from`)
	p.P()
	p.P()
	for _, desc := range [][]string{
		{"BeforeToMongo", typeName, " called before default ToMongo code"},
		{"AfterToMongo", typeName, " called after default ToMongo code"},
		{"BeforeToPB", message.GetName(), " called before default ToPB code"},
		{"AfterToPB", message.GetName(), " called after default ToPB code"},
	} {
		p.P(`// `, typeName, desc[0], desc[2])
		p.P(`type `, typeName, `With`, desc[0], ` interface {`)
		p.P(desc[0], `(*`, desc[1], `) error`)
		p.P(`}`)
		p.P()
	}
}

//GenerateUpdateAllMethod
func (p *MongoPlugin) GenerateUpdateAllMethod(message *generator.Descriptor) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`//Update - update model method, a check is made on existing fields.`)
	p.P(`func (e *`, gName, `) Update (updateAt bool) (*`, gName, `, error) {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(`var flatFields []primitive.E`)
	p.P(`var upResult primitive.D`)

	fields := message.GetField()
	opt, ok := p.getMessageOptions(message)
	if ok {
		if table := opt.GetMerge(); len(table) > 0 {
			st := strings.Split(table, ",")
			if len(st) > 0 {
				for _, str := range st {
					if val, ok := p.Fields[p.GenerateName(str)]; ok {
						for _, f1 := range val {
							fields = append(fields, f1)
						}
					}
				}
			}
		}
	}

	for _, field := range fields {
		fieldName := field.GetName()

		if strings.ToLower(fieldName) == "id" {
			p.P(`// check if fil _id field`)
			p.P(`if !e.Id.IsZero() {`)
			p.P(`e.WhereId(e.Id.Hex())`)
			p.P(`}`)
		}

		// skip _id field UpdatedAt
		if strings.ToLower(fieldName) == "id" || strings.ToLower(fieldName) == "createdat" || strings.ToLower(fieldName) == "updatedat" {
			continue
		}
		bomField := p.getFieldOptions(field)
		if bomField != nil && bomField.Tag.GetSkip() {
			// skip field
			continue
		}

		//e.ManufactureId.Hex() != "000000000000000000000000"

		// find goType
		goTyp, _ := p.GoType(message, field)
		fieldName = generator.CamelCase(fieldName)
		mapName := strings.ToLower(fieldName)
		oneOf := field.OneofIndex != nil

		if oneOf {

			p.P(`// set `, fieldName)
			p.P(`if e.`, fieldName, ` != nil {`)
			p.P(`flatFields = append(flatFields, primitive.E{Key: "`, mapName, `", Value: e.Get`, fieldName, `()})`)
			p.P(`}`)

		} else if field.IsScalar() {

			if strings.ToLower(goTyp) == "bool" {
				p.P(`// set `, fieldName)
				p.P(`if e.`, fieldName, ` {`)
				p.P(`flatFields = append(flatFields, primitive.E{Key: "`, mapName, `", Value: e.`, fieldName, `})`)
				p.P(`}`)
			} else {
				p.P(`// set `, fieldName)
				p.P(`if e.`, fieldName, ` > 0 {`)
				p.P(`flatFields = append(flatFields, primitive.E{Key: "`, mapName, `", Value: e.`, fieldName, `})`)
				p.P(`}`)
			}

		} else if bomField != nil && bomField.Tag.GetMongoObjectId() {

			if !field.IsRepeated() {
				p.P(`// set `, fieldName)
				p.P(`if e.`, fieldName, `.Hex() != "000000000000000000000000" {`)
				p.P(`flatFields = append(flatFields, primitive.E{Key: "`, mapName, `", Value: e.`, fieldName, `})`)
				p.P(`}`)
			} else {

				p.P(`// set `, fieldName)
				p.P(`if len(e.`, fieldName, `) > 0 {`)
				p.P(`flatFields = append(flatFields, primitive.E{Key: "`, mapName, `", Value: e.`, fieldName, `})`)
				p.P(`}`)
			}

		} else if strings.EqualFold(goTyp, "*timestamppb.Timestamp") {
			goTyp = "time.Time"
			p.useTime = true

			p.P(`// set `, fieldName)
			p.P(`if !e.`, fieldName, `.IsZero() {`)
			p.P(`flatFields = append(flatFields, primitive.E{Key: "`, mapName, `", Value: e.`, fieldName, `})`)
			p.P(`}`)

		} else if p.IsMap(field) {
			m, _ := p.GoMapTypeCustomMongo(nil, field)
			goTyp = m.GoType

			p.P(`// set `, fieldName)
			p.P(`if len(e.`, fieldName, `) > 0 {`)
			p.P(`flatFields = append(flatFields, primitive.E{Key: "`, mapName, `", Value: e.`, fieldName, `})`)
			p.P(`}`)

		} else {

			if field.IsMessage() {
				goTyp = p.GenerateName(goTyp)

				p.P(`// set `, fieldName)
				p.P(`if e.`, fieldName, ` != nil {`)
				p.P(`flatFields = append(flatFields, primitive.E{Key: "`, mapName, `", Value: e.`, fieldName, `})`)
				p.P(`}`)

			} else {

				if field.IsEnum() {
					p.P(`// set `, fieldName)
					p.P(`if len(e.`, fieldName, `.String()) > 0 {`)
					p.P(`flatFields = append(flatFields, primitive.E{Key: "`, mapName, `", Value: e.`, fieldName, `})`)
					p.P(`}`)
				} else {
					p.P(`// set `, fieldName)
					p.P(`if len(e.`, fieldName, `) > 0 {`)
					p.P(`flatFields = append(flatFields, primitive.E{Key: "`, mapName, `", Value: e.`, fieldName, `})`)
					p.P(`}`)
				}
			}

		}

	}
	p.P(` if updateAt {`)

	p.P(`upResult = primitive.D{`)
	p.P(`{"$set", flatFields},`)
	p.P(`{"$currentDate", primitive.D{{"updatedat", true}}},`)
	p.P(`}`)

	p.P(` } else {`)

	p.P(`upResult = primitive.D{`)
	p.P(`{"$set", flatFields},`)
	p.P(`}`)

	p.P(` }`)

	p.P(`_, err := e.bom.UpdateRaw(upResult)`)
	p.P(` if err != nil {`)
	p.P(` return e, err`)
	p.P(` }`)

	p.P(` return e, nil`)
	p.P(`}`)
	p.P()
}

//GenerateUpdateAllMethod
func (p *MongoPlugin) GenerateUpdateWithoutConditionAllMethod(message *generator.Descriptor) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`//Update - update model method, a check is made on existing fields.`)
	p.P(`func (e *`, gName, `) UpdateWithoutCondition (updateAt bool) (*`, gName, `, error) {`)
	p.P(`// check if bom object is nil`)
	p.P(`if e.bom == nil {`)
	p.P(`e.SetBom(`, ServiceName, `BomWrapper())`)
	p.P(`}`)
	p.P(`var flatFields []primitive.E`)
	p.P(`var upResult primitive.D`)

	fields := message.GetField()
	opt, ok := p.getMessageOptions(message)
	if ok {
		if table := opt.GetMerge(); len(table) > 0 {
			st := strings.Split(table, ",")
			if len(st) > 0 {
				for _, str := range st {
					if val, ok := p.Fields[p.GenerateName(str)]; ok {
						for _, f1 := range val {
							fields = append(fields, f1)
						}
					}
				}
			}
		}
	}

	for _, field := range fields {
		fieldName := field.GetName()

		if strings.ToLower(fieldName) == "id" {
			p.P(`// check if fil _id field`)
			p.P(`if !e.Id.IsZero() {`)
			p.P(`e.WhereId(e.Id.Hex())`)
			p.P(`}`)
		}

		// skip _id field UpdatedAt
		if strings.ToLower(fieldName) == "id" || strings.ToLower(fieldName) == "createdat" || strings.ToLower(fieldName) == "updatedat" {
			continue
		}
		bomField := p.getFieldOptions(field)
		if bomField != nil && bomField.Tag.GetSkip() {
			continue
		}

		// find goType
		fieldName = generator.CamelCase(fieldName)
		mapName := strings.ToLower(fieldName)
		oneOf := field.OneofIndex != nil

		if oneOf {
			p.P(`// set `, fieldName)
			p.P(`flatFields = append(flatFields, primitive.E{Key: "`, mapName, `", Value: e.Get`, fieldName, `()})`)
		} else {
			p.P(`flatFields = append(flatFields, primitive.E{Key: "`, mapName, `", Value: e.`, fieldName, `})`)
		}
	}
	p.P(` if updateAt {`)

	p.P(`upResult = primitive.D{`)
	p.P(`{"$set", flatFields},`)
	p.P(`{"$currentDate", primitive.D{{"updatedat", true}}},`)
	p.P(`}`)

	p.P(` } else {`)

	p.P(`upResult = primitive.D{`)
	p.P(`{"$set", flatFields},`)
	p.P(`}`)

	p.P(` }`)

	p.P(`_, err := e.bom.UpdateRaw(upResult)`)
	p.P(` if err != nil {`)
	p.P(` return e, err`)
	p.P(` }`)

	p.P(` return e, nil`)
	p.P(`}`)
	p.P()
}

func (g *MongoPlugin) GoMapTypeCustomPB(d *generator.Descriptor, field *descriptor.FieldDescriptorProto) (*generator.GoMapDescriptor, bool) {
	var isMessage = false
	if d == nil {
		byName := g.ObjectNamed(field.GetTypeName())
		desc, ok := byName.(*generator.Descriptor)
		if byName == nil || !ok || !desc.GetOptions().GetMapEntry() {
			g.Fail(fmt.Sprintf("field %s is not a map", field.GetTypeName()))
			return nil, false
		}
		d = desc
	}

	m := &generator.GoMapDescriptor{
		KeyField:   d.Field[0],
		ValueField: d.Field[1],
	}

	// Figure out the Go types and tags for the key and value types.
	m.KeyAliasField, m.ValueAliasField = g.GetMapKeyField(field, m.KeyField), g.GetMapValueField(field, m.ValueField)
	keyType, _ := g.GoType(d, m.KeyAliasField)
	valType, _ := g.GoType(d, m.ValueAliasField)

	// We don't use stars, except for message-typed values.
	// Message and enum types are the only two possibly foreign types used in maps,
	// so record their use. They are not permitted as map keys.
	keyType = strings.TrimPrefix(keyType, "*")
	switch *m.ValueAliasField.Type {
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		valType = strings.TrimPrefix(valType, "*")
		g.RecordTypeUse(m.ValueAliasField.GetTypeName())
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if !gogoproto.IsNullable(m.ValueAliasField) {
			valType = strings.TrimPrefix(valType, "*")
		}
		if !gogoproto.IsStdType(m.ValueAliasField) && !gogoproto.IsCustomType(field) && !gogoproto.IsCastType(field) {
			isMessage = true
			g.RecordTypeUse(m.ValueAliasField.GetTypeName())
		}
	default:
		if gogoproto.IsCustomType(m.ValueAliasField) {
			if !gogoproto.IsNullable(m.ValueAliasField) {

				valType = strings.TrimPrefix(valType, "*")
			}
			if !gogoproto.IsStdType(field) {
				g.RecordTypeUse(m.ValueAliasField.GetTypeName())
			}
		} else {

			valType = strings.TrimPrefix(valType, "*")
		}
	}

	m.GoType = fmt.Sprintf("map[%s]%s", keyType, valType)
	return m, isMessage
}

func (g *MongoPlugin) GoMapTypeCustomMongo(d *generator.Descriptor, field *descriptor.FieldDescriptorProto) (*generator.GoMapDescriptor, bool) {
	var isMessage = false
	if d == nil {
		byName := g.ObjectNamed(field.GetTypeName())
		desc, ok := byName.(*generator.Descriptor)
		if byName == nil || !ok || !desc.GetOptions().GetMapEntry() {
			g.Fail(fmt.Sprintf("field %s is not a map", field.GetTypeName()))
			return nil, false
		}
		d = desc
	}

	m := &generator.GoMapDescriptor{
		KeyField:   d.Field[0],
		ValueField: d.Field[1],
	}

	// Figure out the Go types and tags for the key and value types.
	m.KeyAliasField, m.ValueAliasField = g.GetMapKeyField(field, m.KeyField), g.GetMapValueField(field, m.ValueField)
	keyType, _ := g.GoType(d, m.KeyAliasField)
	valType, _ := g.GoType(d, m.ValueAliasField)

	// We don't use stars, except for message-typed values.
	// Message and enum types are the only two possibly foreign types used in maps,
	// so record their use. They are not permitted as map keys.
	keyType = strings.TrimPrefix(keyType, "*")
	switch *m.ValueAliasField.Type {
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		valType = strings.TrimPrefix(valType, "*")
		g.RecordTypeUse(m.ValueAliasField.GetTypeName())
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if !gogoproto.IsNullable(m.ValueAliasField) {
			valType = strings.TrimPrefix(valType, "*")
		}
		if !gogoproto.IsStdType(m.ValueAliasField) && !gogoproto.IsCustomType(field) && !gogoproto.IsCastType(field) {
			valType = g.GenerateName(valType)
			isMessage = true
			g.RecordTypeUse(m.ValueAliasField.GetTypeName())
		}
	default:
		if gogoproto.IsCustomType(m.ValueAliasField) {
			if !gogoproto.IsNullable(m.ValueAliasField) {

				valType = strings.TrimPrefix(valType, "*")
			}
			if !gogoproto.IsStdType(field) {
				g.RecordTypeUse(m.ValueAliasField.GetTypeName())
			}
		} else {

			valType = strings.TrimPrefix(valType, "*")
		}
	}

	m.GoType = fmt.Sprintf("map[%s]%s", keyType, valType)
	return m, isMessage
}

func (p *MongoPlugin) generateModelsStructures(message *generator.Descriptor) {
	p.In()
	mName := p.GenerateName(message.GetName())
	p.P(`// create MongoDB Model from protobuf (`, p.GenerateName(message.GetName()), `)`)
	p.P(`type `, p.GenerateName(message.GetName()), ` struct {`)

	opt, ok := p.getMessageOptions(message)
	if ok {
		if table := opt.GetMerge(); len(table) > 0 {
			st := strings.Split(table, ",")
			if len(st) > 0 {
				for _, str := range st {
					p.P(p.GenerateName(str))
					p.PrivateEntities[strings.Trim(p.GenerateName(str), " ")] = PrivateEntity{name: p.GenerateName(message.GetName()), message: message}
				}
			}
		}
	}

	type useUnsafeMethod struct {
		fieldName string
		goTyp     string
		mName     string
	}

	var nsafeScope []useUnsafeMethod

	for _, field := range message.GetField() {
		var tagString string
		fieldName := field.GetName()
		oneOf := field.OneofIndex != nil

		goTyp, _ := p.GoType(message, field)
		fieldName = generator.CamelCase(fieldName)

		bomField := p.getFieldOptions(field)
		if bomField != nil && bomField.Tag.GetSkip() {
			// skip field
			continue
		}

		if oneOf && bomField != nil && bomField.Tag.GetMongoObjectId() {
			goTyp = "primitive.ObjectID"
		}
		if oneOf && strings.EqualFold(goTyp, "*timestamppb.Timestamp") {
			goTyp = "time.Time"
		}

		if oneOf {
			p.useUnsafe = true

			nsafeScope = append(nsafeScope, useUnsafeMethod{
				fieldName: fieldName,
				goTyp:     goTyp,
				mName:     mName,
			})
		}

		tagString = "`"
		tagString = tagString + `json:"` + strings.ToLower(fieldName) + `"`

		if bomField != nil && bomField.Tag != nil {
			validTag := bomField.Tag.GetValidator()
			if len(validTag) > 0 {
				tagString = tagString + " "
				tagString = tagString + `valid:"` + validTag + `"`
			}
		}
		tagString = tagString + "`"

		if oneOf {
			if strings.EqualFold(goTyp, "*timestamppb.Timestamp") {
				p.P(fieldName, ` *time.Time`, tagString)
				p.useTime = true
			} else {
				p.P(fieldName, ` *`, goTyp, tagString)
			}
		} else if bomField != nil && bomField.Tag.GetMongoObjectId() {

			repeated := field.IsRepeated()
			if repeated {
				p.P(fieldName, ` `, `[]primitive.ObjectID`, tagString)
			} else {
				idName := tagString
				if bomField.Tag.GetIsID() {
					idName = "`_id, omitempty`"
				}
				p.P(fieldName, ` `, `primitive.ObjectID`, idName)
			}
			p.usePrimitive = true

		} else if p.IsMap(field) {
			m, _ := p.GoMapTypeCustomMongo(nil, field)
			//_, keyField, keyAliasField := m.GoType, m.KeyField, m.KeyAliasField

			p.P(fieldName, ` `, m.GoType, tagString)

		} else if (field.IsMessage() && !gogoproto.IsCustomType(field) && !gogoproto.IsStdType(field)) || p.IsGroup(field) {
			if strings.EqualFold(goTyp, "*timestamppb.Timestamp") {
				p.P(fieldName, ` time.Time`, tagString)
				p.useTime = true
			} else {
				p.P(fieldName, ` `, p.GenerateName(goTyp), tagString)
			}
		} else {
			p.P(fieldName, ` `, goTyp, tagString)
		}
	}
	p.P(`bom  *bom.Bom`)
	p.P(`}`)
	p.Out()
	p.P(``)

	for _, s := range nsafeScope {
		p.P(``)
		p.P(`//Check method `, s.fieldName, ` - update field`)
		p.P(`func (e *`, mName, `) Get`, s.fieldName, `() `, s.goTyp, ` {`)
		p.P(`var resp `, s.goTyp)
		p.P(`if e.`, s.fieldName, ` != nil {`)
		p.P(`resp = *((*`, s.goTyp, `)(unsafe.Pointer(e.`, s.fieldName, `)))`)
		p.P(`}`)
		p.P(`return resp`)
		p.P(`}`)
		p.P(``)
	}

}

func (p *MongoPlugin) GenerateToPB(message *generator.Descriptor) {
	p.In()
	mName := p.GenerateName(message.GetName())
	p.P(`func (e *`, mName, `) ToPB() *`, message.GetName(), ` {`)
	p.P(`var resp `, message.GetName())

	for _, field := range message.GetField() {
		bomField := p.getFieldOptions(field)
		if bomField != nil && bomField.Tag.GetSkip() {
			// skip field
			continue
		}
		p.fieldsToPBStructure(field, message, bomField)
	}

	p.P(`return &resp`)
	p.P(`}`)
	p.Out()
	p.P(``)
}

func (w *MongoPlugin) generateValidationMethods(message *generator.Descriptor) {
	w.P(`// isValid - validation method of the described protobuf structure `)
	mName := w.GenerateName(message.GetName())
	w.P(`func (e *`, mName, `) IsValid() error {`)
	w.P(`if _, err := valid.ValidateStruct(e); err != nil {`)
	w.P(`return err`)
	w.P(`}`)
	w.P(`return nil`)
	w.P(`}`)
	w.Out()
	w.P(``)
}

func (p *MongoPlugin) fieldsToPBStructure(field *descriptor.FieldDescriptorProto, message *generator.Descriptor, bomField *bom.BomFieldOptions) {
	fieldName := field.GetName()
	fieldName = generator.CamelCase(fieldName)
	oneof := field.OneofIndex != nil

	goTyp, _ := p.GoType(message, field)
	p.In()

	if p.IsMap(field) {
		m, ism := p.GoMapTypeCustomPB(nil, field)
		_, keyField, keyAliasField := m.GoType, m.KeyField, m.KeyAliasField
		keygoTyp, _ := p.GoType(nil, keyField)
		keygoTyp = strings.Replace(keygoTyp, "*", "", 1)
		keygoAliasTyp, _ := p.GoType(nil, keyAliasField)
		keygoAliasTyp = strings.Replace(keygoAliasTyp, "*", "", 1)
		//keyCapTyp := generator.CamelCase(keygoTyp)

		p.P(`tt`, fieldName, ` := make(`, m.GoType, `)`)
		p.P(`for k, v := range e.`, fieldName, ` {`)
		p.In()
		if ism {
			p.P(`tt`, fieldName, `[k] = v.ToPB()`)
		} else {
			p.P(`tt`, fieldName, `[k] = v`)
		}
		p.Out()
		p.P(`}`)
		p.P(`resp.`, fieldName, ` = tt`, fieldName)

	} else if (field.IsMessage() && !gogoproto.IsCustomType(field) && !gogoproto.IsStdType(field)) || p.IsGroup(field) {

		if strings.EqualFold(goTyp, "*timestamppb.Timestamp") {

			if oneof {

				p.P(`if e.`, fieldName, ` != nil {`)
				p.P(`ptap`, fieldName, `, _ := ptypes.TimestampProto(e.Get`, fieldName, `())`)
				sourceName := p.GetFieldName(message, field)
				interfaceName := p.Generator.OneOfTypeName(message, field)
				p.P(`resp.`, sourceName, ` = &`, interfaceName, `{ptap`, fieldName, `}`)
				p.P(`}`)

			} else {
				p.P(`ptap`, fieldName, `, _ := ptypes.TimestampProto(e.`, fieldName, `)`)
				p.useTime = true
				p.P(`resp.`, fieldName, ` = ptap`, fieldName)
			}

		} else if field.IsMessage() {

			repeated := field.IsRepeated()
			if repeated {
				p.P(`// create nested pb`)
				p.P(`var sub`, fieldName, goTyp)
				p.P(`if e.`, fieldName, ` != nil {`)
				p.P(`if len(e.`, fieldName, `) > 0 {`)

				p.P(`for _, b := range `, `e.`, fieldName, `{`)
				p.P(`sub`, fieldName, ` = append(sub`, fieldName, `, b.ToPB())`)
				p.P(`}`)

				p.P(`}`)
				p.P(`}`)

				p.P(`resp.`, fieldName, ` = sub`, fieldName)
			} else {

				p.P(`// create single pb`)
				p.P(`if e.`, fieldName, ` != nil {`)
				p.P(`resp.`, fieldName, ` = e.`, fieldName, `.ToPB()`)
				p.P(`}`)

			}

		} else {
			p.P(`resp.`, fieldName, ` = e.`, fieldName)
		}

	} else {

		if oneof {

			if bomField != nil && bomField.Tag.GetMongoObjectId() {

				sourceName := p.GetFieldName(message, field)
				interfaceName := p.Generator.OneOfTypeName(message, field)

				p.P(`objectValue := e.Get`, fieldName, `()`)
				p.P(`if !objectValue.IsZero() {`)
				p.P(`resp.`, sourceName, ` = &`, interfaceName, `{objectValue.Hex()}`)
				p.P(`}`)

			} else {
				// if one of click
				sourceName := p.GetFieldName(message, field)
				interfaceName := p.Generator.OneOfTypeName(message, field)
				p.P(`resp.`, sourceName, ` = &`, interfaceName, `{e.Get`, fieldName, `()}`)
			}

		} else if bomField != nil && bomField.Tag.GetMongoObjectId() {
			repeated := field.IsRepeated()
			if repeated {
				p.P(`if len(e.`, fieldName, `) > 0 {`)
				p.P(`var sub`, fieldName, goTyp)
				p.P(`for _, b := range `, `e.`, fieldName, `{`)
				p.P(`if !b.IsZero() {`)
				p.P(`sub`, fieldName, ` = append(sub`, fieldName, `, b.Hex())`)
				p.P(`}`)
				p.P(`}`)
				p.P(`resp.`, fieldName, ` = sub`, fieldName)
				p.P(`}`)

			} else {
				p.P(`if !e.`, fieldName, `.IsZero() {`)
				p.P(`resp.`, fieldName, ` = e.`, fieldName, `.Hex()`)
				p.P(`}`)

			}
		} else {
			p.P(`resp.`, fieldName, ` = e.`, fieldName)
		}
	}
	p.Out()
}

func (p *MongoPlugin) ToMongoGenerateFieldConversion(field *descriptor.FieldDescriptorProto, message *generator.Descriptor, bomField *bom.BomFieldOptions) {
	fieldName := field.GetName()
	fieldName = generator.CamelCase(fieldName)
	goTyp, _ := p.GoType(message, field)

	oneof := field.OneofIndex != nil

	p.In()

	if p.IsMap(field) {

		m, ism := p.GoMapTypeCustomMongo(nil, field)
		_, keyField, keyAliasField := m.GoType, m.KeyField, m.KeyAliasField
		keygoTyp, _ := p.GoType(nil, keyField)
		keygoTyp = strings.Replace(keygoTyp, "*", "", 1)
		keygoAliasTyp, _ := p.GoType(nil, keyAliasField)
		keygoAliasTyp = strings.Replace(keygoAliasTyp, "*", "", 1)

		p.P(`tt`, fieldName, ` := make(`, m.GoType, `)`)
		p.P(`for k, v := range e.`, fieldName, ` {`)
		p.In()
		if ism {
			p.P(`tt`, fieldName, `[k] = v.ToMongo()`)
		} else {
			p.P(`tt`, fieldName, `[k] = v`)
		}
		p.Out()
		p.P(`}`)
		p.P(`resp.`, fieldName, ` = tt`, fieldName)

	} else if (field.IsMessage() && !gogoproto.IsCustomType(field) && !gogoproto.IsStdType(field)) || p.IsGroup(field) {

		if strings.EqualFold(goTyp, "*timestamppb.Timestamp") {
			p.useTime = true

			if oneof {

				sourceName := p.GetFieldName(message, field)
				p.P(`// oneof link`)
				p.P(`if e.Get`, sourceName, `() != nil {`)
				p.P(`link`, fieldName, ` := e.Get`, fieldName, `()`)
				p.P(`if link`, fieldName, ` != nil {`)
				p.P(`ut`, fieldName, ` := time.Unix(link`, fieldName, `.GetSeconds(), int64(link`, fieldName, `.GetNanos()))`)
				p.P(`resp.`, fieldName, ` = &ut`, fieldName)
				p.P(`}`)
				p.P(`}`)
				p.P(``)

			} else {

				p.P(`// check if it's not a zero date and create time object`)
				p.P(`if e.`, fieldName, `.GetSeconds() > 0 {`)
				p.P(`ut`, fieldName, ` := time.Unix(e.`, fieldName, `.GetSeconds(), int64(e.`, fieldName, `.GetNanos()))`)
				p.P(`resp.`, fieldName, ` = ut`, fieldName)
				p.P(`}`)

			}

		} else if field.IsMessage() {
			repeated := field.IsRepeated()
			if repeated {
				p.P(`// create nested mongo`)
				p.P(`var sub`, fieldName, p.GenerateName(goTyp))
				p.P(`if e.`, fieldName, ` != nil {`)
				p.P(`if len(e.`, fieldName, `) > 0 {`)

				p.P(`for _, b := range `, `e.`, fieldName, `{`)
				p.P(`if b != nil {`)
				p.P(`sub`, fieldName, ` = append(sub`, fieldName, `, b.ToMongo())`)
				p.P(`}`)
				p.P(`}`)

				p.P(`}`)
				p.P(`}`)

				p.P(`resp.`, fieldName, ` = sub`, fieldName)
			} else {
				p.P(`// create single mongo`)
				p.P(`if e.`, fieldName, ` != nil {`)
				p.P(`resp.`, fieldName, ` = e.`, fieldName, `.ToMongo()`)
				p.P(`}`)
			}

		} else {
			p.P(`resp.`, fieldName, ` = e.`, fieldName)
		}

	} else if bomField != nil && bomField.Tag.GetMongoObjectId() {

		repeated := field.IsRepeated()
		if repeated {
			p.P(`if len(e.`, fieldName, `) > 0 {`)
			p.P(`resp.`, fieldName, ` = bom.ToObjects(e.`, fieldName, `)`)
			p.P(`}`)
		} else {
			if oneof {
				// if one of click
				sourceName := p.GetFieldName(message, field)
				p.P(`// oneof link`)
				p.P(`if e.Get`, sourceName, `() != nil {`)
				p.P(`link`, fieldName, ` := bom.ToObj(e.Get`, fieldName, `())`)
				p.P(`resp.`, fieldName, ` = &link`, fieldName, ``)
				p.P(`}`)
				p.P(``)

			} else {
				p.P(`if len(e.`, fieldName, `) > 0 {`)
				p.P(`resp.`, fieldName, ` = bom.ToObj(e.`, fieldName, `)`)
				p.P(`}`)
			}

		}

	} else {
		if oneof {

			// if one of click
			sourceName := p.GetFieldName(message, field)
			p.P(`// oneof link`)
			p.P(`if e.Get`, sourceName, `() != nil {`)
			p.P(`link`, fieldName, ` := e.Get`, fieldName, `()`)
			p.P(`resp.`, fieldName, ` = &link`, fieldName, ``)
			p.P(`}`)
			p.P(``)

		} else {
			p.P(`resp.`, fieldName, ` = e.`, fieldName)
		}
	}
	p.Out()
}

func (p *MongoPlugin) GenerateToObject(message *generator.Descriptor) {
	p.In()
	mName := p.GenerateName(message.GetName())
	p.P(`// ToMongo runs the BeforeToMongo hook if present, converts the fields of this`)
	p.P(`// object to Mongo format, runs the AfterToMongo hook, then returns the Mongo object`)
	p.P(`func (e *`, message.GetName(), `) ToMongo() *`, mName, ` {`)
	p.P(`var resp `, mName)

	for _, field := range message.GetField() {
		bomField := p.getFieldOptions(field)
		if bomField != nil && bomField.Tag.GetSkip() {
			// skip field
			continue
		}
		p.ToMongoGenerateFieldConversion(field, message, bomField)
	}

	bomMessage, ok := p.getMessageOptions(message)
	if ok && bomMessage.GetCrud() {
		collection := strings.ToLower(message.GetName())
		if clt := bomMessage.GetCollection(); len(clt) > 0 {
			collection = clt
		}
		p.P()
		p.P(`// bom connection`)
		p.P(`resp.bom = `, ServiceName, `BomWrapper().WithColl("`, collection, `")`)
		p.P()
	}

	p.P(`return &resp`)
	p.P(`}`)
	p.Out()
	p.P(``)
}

func (p *MongoPlugin) renderType(message string) string {
	return message
}

// Name identifies the plugin
func (p *MongoPlugin) Name() string {
	return "bom"
}

func (w *MongoPlugin) generateEntitiesMethods() {
	if len(w.PrivateEntities) > 0 {
		for key, value := range w.PrivateEntities {
			w.P(``)
			w.P(`// Merge - merge private structure (`, value.name, `)`)
			w.P(`func (e *`, value.name, `) Merge`, strings.Trim(key, " "), ` (m *`, key, `) *`, value.name, ` {`)

			for _, field := range value.items {
				fieldName := field.GetName()
				fieldName = generator.CamelCase(fieldName)
				w.P(`e.`, fieldName, ` = m.`, fieldName)
			}

			w.P(`return e`)
			w.P(`}`)
			w.Out()
			w.P(``)
		}
	}
	if len(w.ConvertEntities) > 0 {
		for _, value := range w.ConvertEntities {
			w.P(``)
			w.P(`// To`, strings.Trim(value.nameTo, " "), ` - convert structure (`, value.nameFrom, ` -> `, value.nameTo, `)`)
			w.P(`func (e *`, value.nameFrom, `) To`, strings.Trim(value.nameTo, " "), ` () *`, value.nameTo, ` {`)
			w.P(`var entity `, value.nameTo)
			if fieldsFrom, ok := w.Fields[value.nameFrom]; ok {
				if fieldsTo, ok := w.Fields[value.nameTo]; ok {
					for _, field := range fieldsFrom {
						for _, f := range fieldsTo {
							if field.GetName() == f.GetName() {
								fieldName := field.GetName()
								fieldName = generator.CamelCase(fieldName)

								// check is one of
								oneOfField := field.OneofIndex != nil
								if oneOfField {
									w.P(`entity.`, fieldName, ` = e.Get`, fieldName, `()`)
								} else {
									w.P(`entity.`, fieldName, ` = e.`, fieldName)
								}

							}
						}
					}
				}
			}
			for private, pe := range w.PrivateEntities {
				if pe.name == value.nameTo {
					if fields, ok := w.Fields[private]; ok {
						for _, f2 := range fields {
							fieldName := f2.GetName()
							fieldName = generator.CamelCase(fieldName)

							// check is one of
							oneOfField := f2.OneofIndex != nil

							if oneOfField {
								w.P(`entity.`, fieldName, ` = e.Get`, fieldName, `()`)
							} else {
								w.P(`entity.`, fieldName, ` = e.`, fieldName)
							}

						}
					}
				}
			}
			w.P(`return &entity`)
			w.P(`}`)
			w.Out()
			w.P(``)
		}
	}
}
