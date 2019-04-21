package plugin

import (
	bom "github.com/cjp2600/protoc-gen-bom/plugin/options"
	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"path"
	"strings"
)

type MongoPlugin struct {
	*generator.Generator
	generator.PluginImports
	EmptyFiles     []string
	currentPackage string
	currentFile    *generator.FileDescriptor
	generateCrud   bool

	usePrimitive bool
	useTime      bool
	useStrconv   bool
	useMongoDr   bool
	useCrud      bool
	localName    string
}

var ServiceName string

func NewMongoPlugin(generator *generator.Generator) *MongoPlugin {
	return &MongoPlugin{Generator: generator}
}

func (p *MongoPlugin) GenerateImports(file *generator.FileDescriptor) {
	if p.usePrimitive {
		p.Generator.PrintImport("primitive", "go.mongodb.org/mongo-driver/bson/primitive")
	}
	p.Generator.PrintImport("os", "os")
	p.Generator.PrintImport("bom", "github.com/cjp2600/bom")
	//p.Generator.PrintImport("context", "context")
	if p.useTime {
		p.Generator.PrintImport("time", "time")
		p.Generator.PrintImport("ptypes", "github.com/golang/protobuf/ptypes")
	}
	if p.useStrconv {
		p.Generator.PrintImport("strconv", "strconv")
	}
	if p.useMongoDr {
		//"go.mongodb.org/mongo-driver/mongo"
		p.Generator.PrintImport("mongo", "go.mongodb.org/mongo-driver/mongo")
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
	p.PluginImports = generator.NewPluginImports(p.Generator)
	p.localName = generator.FileName(file)
	ServiceName = p.GetServiceName(file)

	p.usePrimitive = false
	p.GenerateBomObject()
	for _, msg := range file.GetMessageType() {
		if bomMessage, ok := p.getMessageOptions(msg); ok {
			if bomMessage.GetModel() {

				p.GenerateToPB(msg)
				p.GenerateToObject(msg)
				p.GenerateObjectId(msg)
				// todo: доделать генерацию конвертации в связанную модель
				//p.GenerateBoundMessage(msg)

				// добавляем круд
				if bomMessage.GetCrud() {
					p.useCrud = true
					p.GenerateBomConnection(msg)
					p.GenerateGetBom(msg)
					p.GenerateContructor(msg)
					p.GenerateInsertMethod(msg)
					p.GenerateFindOneMethod(msg)
					p.GenerateFindMethod(msg)
					p.GenerateWhereMethod(msg)
					p.GenerateWhereInMethod(msg)
					p.GenerateOrWhereMethod(msg)
				}

				// генерируем основные методы модели
				p.generateModelsStructures(msg)

			}
		}
	}
}

func (p *MongoPlugin) getMessageOptions(msg *descriptor.DescriptorProto) (*bom.BomMessageOptions, bool) {
	opt := msg.GetOptions()
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
func (p *MongoPlugin) GenerateBoundMessage(message *descriptor.DescriptorProto) {
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
func (p *MongoPlugin) GenerateObjectId(message *descriptor.DescriptorProto) {
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
func (p *MongoPlugin) getCollection(message *descriptor.DescriptorProto) string {
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

func (p *MongoPlugin) GenerateContructor(message *descriptor.DescriptorProto) {
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
		p.P(`if GlobalBom == nil {`)
		p.P(`panic("bom object not found")`)
		p.P(`}`)
		p.P(`return &`, gName, `{bom:  GlobalBom.WithColl("`, collection, `")}`)
		p.P(`}`)
	}
}

// GenerateFindMethod
func (p *MongoPlugin) GenerateWhereMethod(message *descriptor.DescriptorProto) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`// Where method`)
	p.P(`func (e *`, gName, `) Where(field string, value interface{}) *`, gName, ` {`)
	p.P(` e.bom.Where(field, value)`)
	p.P(` return e`)
	p.P(`}`)

	////	bmQuery := mongoModel.GetBom().
	////		WithLimit(&bom.Limit{Page: q.Page, Size: q.Size})

	p.P()
	p.P(`// Limit method`)
	p.P(`func (e *`, gName, `) Limit(page int32, size int32) *`, gName, ` {`)
	p.P(` e.bom.WithLimit(&bom.Limit{Page: page, Size: size})`)
	p.P(` return e`)
	p.P(`}`)

	p.P()
	p.P(`// Sort method`)
	p.P(`func (e *`, gName, `) Sort(sortField string, sortType string) *`, gName, ` {`)
	p.P(`if sortField == "id" {`)
	p.P(`sortField = "_id"`)
	p.P(`}`)
	p.P(` e.bom.WithSort(&bom.Sort{Field: sortField, Type: sortType})`)
	p.P(` return e`)
	p.P(`}`)

	p.P()
	p.P(`// Find with pagination`)
	p.P(`func (e *`, gName, `) ListWithPagination() ([]*`, gName, `, *bom.Pagination, error) {`)
	p.P(`var items []*`, gName)
	p.P(`paginator, err := e.bom.ListWithPagination(func(cur *mongo.Cursor) error {`)
	p.P(`var result `, gName, ``)
	p.P(`err := cur.Decode(&result)`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.P(`items = append(items, result)`)
	p.P(`return nil`)
	p.P(`})`)
	p.P(`return items, paginator, err`)
	p.P(`}`)

	p.P()
	p.P(`// Find list`)
	p.P(`func (e *`, gName, `) List() ([]*`, gName, `, error) {`)
	p.P(`var items []*`, gName)

	p.P(`err := e.bom.List(func(cur *mongo.Cursor) error {`)
	p.P(`var result `, gName, ``)
	p.P(`err := cur.Decode(&result)`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.P(`items = append(items, result)`)
	p.P(`return nil`)
	p.P(`})`)

	p.P(`return items, err`)
	p.P(`}`)

	p.P()
	p.P(`// Get bulk map`)
	p.P(`func (e *`, gName, `) GetBulkMap(ids []string) (map[string]*`, gName, `, error) {`)
	p.P(`result = make(map[string]*`, gName, `)`)
	p.P(`items, err := e.WhereIn("_id", bom.ToObjects(ids)).List()`)
	p.P(`if err != nil {`)
	p.P(`return result, err`)
	p.P(`}`)

	p.P(`for k, v := range items {`)
	p.P(`result[k] = v`)
	p.P(`}`)

	p.P(`return result, nil`)
	p.P(`}`)

	p.P()
	p.P(`// Get bulk map`)
	p.P(`func (e *`, gName, `) GetBulk(ids []string) ([]*`, gName, `, error) {`)
	p.P(`return e.WhereIn("_id", bom.ToObjects(ids)).List()`)
	p.P(`}`)

}

func (p *MongoPlugin) GenerateBomObject() {
	p.useMongoDr = true
	p.P()
	p.P(`// global bom Object`)
	p.P(`var GlobalBom *bom.Bom`)
	p.P()
	p.P(`// create Bom wrapper (`, ServiceName, `)`)
	p.P(`func `, ServiceName, `BomWrapper(client *mongo.Client) error {`)
	p.P(`dbName := os.Getenv("MONGO_DB_NAME")`)
	p.P(`if len(dbName) == 0 {`)
	p.P(`dbName = "`, strings.ToLower(ServiceName), `"`)
	p.P(`}`)
	p.P(`bomObject, err := bom.New(`)
	p.P(`bom.SetMongoClient(client),`)
	p.P(`bom.SetDatabaseName(dbName),`)
	p.P(`)`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.P(`// set global var`)
	p.P(`GlobalBom = bomObject`)
	p.P(`return nil`)
	p.P(`}`)
	p.P()
}

func (p *MongoPlugin) GenerateBomConnection(message *descriptor.DescriptorProto) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`// set custom bom wrapper`)
	p.P(`func (e *`, gName, `) SetBom(bom *bom.Bom) *`, gName, ` {`)
	p.P(` e.bom = bom.WithColl("`, p.getCollection(message), `") `)
	p.P(` return e`)
	p.P(`}`)
}

func (p *MongoPlugin) GenerateGetBom(message *descriptor.DescriptorProto) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`// GetSourceBom - Get the source object`)
	p.P(`func (e *`, gName, `) GetBom() *bom.Bom {`)
	p.P(` return e.bom`)
	p.P(`}`)
}

func (p *MongoPlugin) GenerateWhereInMethod(message *descriptor.DescriptorProto) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`// WhereIn method`)
	p.P(`func (e *`, gName, `) WhereIn(field string, value interface{}) *`, gName, ` {`)
	p.P(` e.bom.InWhere(field, value)`)
	p.P(` return e`)
	p.P(`}`)
}

func (p *MongoPlugin) GenerateOrWhereMethod(message *descriptor.DescriptorProto) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`// OrWhere method`)
	p.P(`func (e *`, gName, `) OrWhere(field string, value interface{}) *`, gName, ` {`)
	p.P(` e.bom.OrWhere(field, value)`)
	p.P(` return e`)
	p.P(`}`)
}

// GenerateFindMethod
func (p *MongoPlugin) GenerateFindMethod(message *descriptor.DescriptorProto) {
	gName := p.GenerateName(message.GetName())
	p.P()
	p.P(`// Find  find method`)
	p.P(`func (e *`, gName, `) FindOne() (*`, gName, `, error) {`)
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

//GenerateFindOneMethod
func (p *MongoPlugin) GenerateFindOneMethod(message *descriptor.DescriptorProto) {

	for _, field := range message.GetField() {
		des := &generator.Descriptor{
			DescriptorProto: message,
		}

		//nullable := gogoproto.IsNullable(field)
		repeated := field.IsRepeated()
		fieldName := field.GetName()
		//oneOf := field.OneofIndex != nil
		goTyp, _ := p.GoType(des, field)
		fieldName = generator.CamelCase(fieldName)
		mName := p.GenerateName(message.GetName())

		if !field.IsMessage() && !repeated {
			p.useMongoDr = true
			p.P(`// FindOneBy`, fieldName, ` - find method`)
			p.P(`func (e *`, mName, `) FindOneBy`, fieldName, `(`, fieldName, ` `, goTyp, `) (*`, mName, `, error) {`)
			p.P(`mongoModel := `, mName, `{}`)
			p.P(` err := e.bom.`)
			bomField := p.getFieldOptions(field)
			if bomField != nil && bomField.Tag.GetMongoObjectId() {
				fn := strings.ToLower(fieldName)
				if fn == "id" {
					fn = "_id"
				}
				p.P(` Where("`, fn, `", bom.ToObj(`, fieldName, `)).`)
			} else {
				p.P(` Where("`, strings.ToLower(fieldName), `", `, fieldName, ` ).`)
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

//GenerateInsertMethod
func (p *MongoPlugin) GenerateInsertMethod(message *descriptor.DescriptorProto) {
	//typeName := p.GenerateName(message.GetName())
	mName := p.GenerateName(message.GetName())
	p.usePrimitive = true
	useId := false
	p.P()
	p.P(`// InsertOne method`)
	p.P(`func (e *`, mName, `) InsertOne() (*`, mName, `, error) {`)
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
func (p *MongoPlugin) GenerateBehaviorInterface(message *descriptor.DescriptorProto) {
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

func (p *MongoPlugin) generateModelsStructures(message *descriptor.DescriptorProto) {
	p.In()
	p.P(`// create MongoDB Model from protobuf (`, p.GenerateName(message.GetName()), `)`)
	p.P(`type `, p.GenerateName(message.GetName()), ` struct {`)
	oneofs := make(map[string]struct{})
	for _, field := range message.GetField() {
		des := &generator.Descriptor{
			DescriptorProto: message,
		}
		//nullable := gogoproto.IsNullable(field)
		//repeated := field.IsRepeated()
		fieldName := field.GetName()
		oneOf := field.OneofIndex != nil
		goTyp, _ := p.GoType(des, field)
		fieldName = generator.CamelCase(fieldName)

		bomField := p.getFieldOptions(field)
		if bomField != nil && bomField.Tag.GetSkip() {
			// skip field
			continue
		}

		if oneOf {
			if _, ok := oneofs[fieldName]; ok {
				continue
			} else {
				oneofs[fieldName] = struct{}{}
			}
		}

		if bomField != nil && bomField.Tag.GetMongoObjectId() {

			repeated := field.IsRepeated()
			if repeated {
				p.P(fieldName, ` `, `[]primitive.ObjectID`)
			} else {
				idName := ""
				if bomField.Tag.GetIsID() {
					idName = "`_id, omitempty`"
				}
				p.P(fieldName, ` `, `primitive.ObjectID`, idName)
			}
			p.usePrimitive = true

		} else if p.IsMap(field) {
			m := p.GoMapType(nil, field)
			//_, keyField, keyAliasField := m.GoType, m.KeyField, m.KeyAliasField
			p.P(fieldName, ` `, m.GoType)
		} else if (field.IsMessage() && !gogoproto.IsCustomType(field) && !gogoproto.IsStdType(field)) || p.IsGroup(field) {
			if strings.ToLower(goTyp) == "*timestamp.timestamp" {
				p.P(fieldName, ` time.Time`)
				p.useTime = true
			} else {
				p.P(fieldName, ` `, p.GenerateName(goTyp))
			}
		} else {
			p.P(fieldName, ` `, goTyp)
		}
	}
	if p.useCrud {
		p.P(`bom  *bom.Bom`)
	}
	p.P(`}`)
	p.Out()
	p.P(``)
}

func (p *MongoPlugin) GenerateToPB(message *descriptor.DescriptorProto) {
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
		p.GenerateFieldConversion(field, message, bomField)
	}

	p.P(`return &resp`)
	p.P(`}`)
	p.Out()
	p.P(``)
}

func (p *MongoPlugin) GenerateFieldConversion(field *descriptor.FieldDescriptorProto, message *descriptor.DescriptorProto, bomField *bom.BomFieldOptions) {
	fieldName := field.GetName()
	fieldName = generator.CamelCase(fieldName)
	des := &generator.Descriptor{
		DescriptorProto: message,
	}
	goTyp, _ := p.GoType(des, field)
	p.In()
	if p.IsMap(field) {
		m := p.GoMapType(nil, field)
		_, keyField, keyAliasField := m.GoType, m.KeyField, m.KeyAliasField
		keygoTyp, _ := p.GoType(nil, keyField)
		keygoTyp = strings.Replace(keygoTyp, "*", "", 1)
		keygoAliasTyp, _ := p.GoType(nil, keyAliasField)
		keygoAliasTyp = strings.Replace(keygoAliasTyp, "*", "", 1)
		//keyCapTyp := generator.CamelCase(keygoTyp)
		p.P(`tt`, fieldName, ` := make(`, m.GoType, `)`)
		p.P(`for k, v := range e.`, fieldName, ` {`)
		p.In()
		p.P(`tt`, fieldName, `[k] = v`)
		p.Out()
		p.P(`}`)
		p.P(`resp.`, fieldName, ` = tt`, fieldName)

	} else if (field.IsMessage() && !gogoproto.IsCustomType(field) && !gogoproto.IsStdType(field)) || p.IsGroup(field) {

		if strings.ToLower(goTyp) == "*timestamp.timestamp" {
			p.P(`ptap`, fieldName, `, _ := ptypes.TimestampProto(e.`, fieldName, `)`)
			p.useTime = true
			p.P(`resp.`, fieldName, ` = ptap`, fieldName)
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
		if bomField != nil && bomField.Tag.GetMongoObjectId() {

			repeated := field.IsRepeated()
			if repeated {

				p.P(`if len(e.`, fieldName, `) > 0 {`)
				p.P(`var sub`, fieldName, goTyp)
				p.P(`for _, b := range `, `e.`, fieldName, `{`)
				p.P(`sub`, fieldName, ` = append(sub`, fieldName, `, b.Hex())`)
				p.P(`}`)
				p.P(`resp.`, fieldName, ` = sub`, fieldName)
				p.P(`}`)

			} else {
				p.P(`resp.`, fieldName, ` = e.`, fieldName, `.Hex()`)
			}

		} else {
			p.P(`resp.`, fieldName, ` = e.`, fieldName)
		}
	}
	p.Out()
}

func (p *MongoPlugin) ToMongoGenerateFieldConversion(field *descriptor.FieldDescriptorProto, message *descriptor.DescriptorProto, bomField *bom.BomFieldOptions) {
	fieldName := field.GetName()
	fieldName = generator.CamelCase(fieldName)
	des := &generator.Descriptor{
		DescriptorProto: message,
	}
	goTyp, _ := p.GoType(des, field)
	p.In()

	if p.IsMap(field) {

		m := p.GoMapType(nil, field)
		_, keyField, keyAliasField := m.GoType, m.KeyField, m.KeyAliasField
		keygoTyp, _ := p.GoType(nil, keyField)
		keygoTyp = strings.Replace(keygoTyp, "*", "", 1)
		keygoAliasTyp, _ := p.GoType(nil, keyAliasField)
		keygoAliasTyp = strings.Replace(keygoAliasTyp, "*", "", 1)
		//keyCapTyp := generator.CamelCase(keygoTyp)
		p.P(`tt`, fieldName, ` := make(`, m.GoType, `)`)

		p.P(`for k, v := range e.`, fieldName, ` {`)
		p.In()
		p.P(`tt`, fieldName, `[k] = v`)
		p.Out()
		p.P(`}`)
		p.P(`resp.`, fieldName, ` = tt`, fieldName)

	} else if (field.IsMessage() && !gogoproto.IsCustomType(field) && !gogoproto.IsStdType(field)) || p.IsGroup(field) {

		if strings.ToLower(goTyp) == "*timestamp.timestamp" {
			p.useTime = true
			p.P(`// create time object`)
			p.P(`ut`, fieldName, ` := time.Unix(e.`, fieldName, `.GetSeconds(), int64(e.`, fieldName, `.GetNanos()))`)
			p.P(`resp.`, fieldName, ` = ut`, fieldName)
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
			p.P(`if len(e.`, fieldName, `) > 0 {`)
			p.P(`resp.`, fieldName, ` = bom.ToObj(e.`, fieldName, `)`)
			p.P(`}`)
		}

	} else {
		p.P(`resp.`, fieldName, ` = e.`, fieldName)
	}
	p.Out()
}

func (p *MongoPlugin) GenerateToObject(message *descriptor.DescriptorProto) {
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
	if ok {
		collection := strings.ToLower(message.GetName())
		if clt := bomMessage.GetCollection(); len(clt) > 0 {
			collection = clt
		}
		p.P()
		p.P(`// bom connection`)
		p.P(`if GlobalBom == nil {`)
		p.P(`panic("bom object not found")`)
		p.P(`}`)
		p.P(`resp.bom = GlobalBom.WithColl("`, collection, `")`)
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
