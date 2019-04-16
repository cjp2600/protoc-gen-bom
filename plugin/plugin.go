package plugin

import (
	bom "github.com/cjp2600/protoc-gen-bom/plugin/options"
	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
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
	localName    string
}

func NewMongoPlugin(generator *generator.Generator) *MongoPlugin {
	return &MongoPlugin{Generator: generator}
}

func (p *MongoPlugin) GenerateImports(file *generator.FileDescriptor) {
	if p.usePrimitive {
		p.Generator.PrintImport("primitive", "go.mongodb.org/mongo-driver/bson/primitive")
	}
	p.Generator.PrintImport("bom", "github.com/cjp2600/bom")
	//p.Generator.PrintImport("context", "context")
	if p.useTime {
		p.Generator.PrintImport("time", "time")
		p.Generator.PrintImport("ptypes", "github.com/golang/protobuf/ptypes")
	}
	if p.useStrconv {
		p.Generator.PrintImport("strconv", "strconv")
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
	p.usePrimitive = false

	for _, msg := range file.GetMessageType() {
		if bomMessage, ok := p.getMessageOptions(msg); ok {
			if bomMessage.GetModel() {
				p.generateModelsStructures(msg)
				if bomMessage.GetCrud() {
					p.GenerateBehaviorInterface(msg)
					p.GenerateToPB(msg)
					p.GenerateToObject(msg)
					p.GenerateBomConnect(msg)
				}
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

		if oneOf {
			if _, ok := oneofs[fieldName]; ok {
				continue
			} else {
				oneofs[fieldName] = struct{}{}
			}
		}

		bomField := p.getFieldOptions(field)
		if bomField != nil && bomField.Tag.GetMongoObjectId() {

			idName := ""
			if bomField.Tag.GetIsID() {
				idName = "`_id, omitempty`"
			}

			p.P(fieldName, ` `, `primitive.ObjectID`, idName)
			p.usePrimitive = true

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
	p.P(`}`)
	p.Out()
	p.P(``)
}

//GenerateBomConnect
func (p *MongoPlugin) GenerateBomConnect(message *descriptor.DescriptorProto) {
	bomMessage, ok := p.getMessageOptions(message)
	if ok && bomMessage.GetCrud() {
		p.In()
		mName := p.GenerateName(message.GetName())
		collection := strings.ToLower(message.GetName())
		p.P(`func (e *`, mName, `) WithBom(b *bom.Bom) *bom.Bom {`)
		if clt := bomMessage.GetCollection(); len(clt) > 0 {
			collection = clt
		}
		p.P(`return b.WithColl("`, collection, `")`)
		p.P(`}`)
		p.Out()
	}
}

func (p *MongoPlugin) GenerateToPB(message *descriptor.DescriptorProto) {
	p.In()
	mName := p.GenerateName(message.GetName())
	p.P(`func (e *`, mName, `) ToPB() (*`, message.GetName(), `, error) {`)
	p.P(`var resp *`, message.GetName())
	p.P(`var err error`)
	p.P(`if prehook, ok := interface{}(e).(`, mName, `WithBeforeToPB); ok {`)
	p.P(`if err = prehook.BeforeToPB(resp); err != nil {`)
	p.P(`return resp, err`)
	p.P(`}`)
	p.P(`}`)
	for _, field := range message.GetField() {
		bomField := p.getFieldOptions(field)
		if bomField != nil && bomField.Tag.GetSkip() {
			// skip field
			continue
		}
		p.GenerateFieldConversion(field, message, bomField)
	}
	p.P(`if posthook, ok := interface{}(e).(`, mName, `WithAfterToPB); ok {`)
	p.P(`err = posthook.AfterToPB(resp)`)
	p.P(`}`)
	p.P(`return resp, err`)
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
	if (field.IsMessage() && !gogoproto.IsCustomType(field) && !gogoproto.IsStdType(field)) || p.IsGroup(field) {

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
				p.P(`pb, err := b.ToPB()`)
				p.P(`if err != nil {`)
				p.P(`continue`)
				p.P(`}`)
				p.P(`sub`, fieldName, ` = append(sub`, fieldName, `, pb)`)
				p.P(`}`)

				p.P(`}`)
				p.P(`}`)

				p.P(`resp.`, fieldName, ` = sub`, fieldName)
			} else {
				p.P(`// create single pb`)
				p.P(`pb`, fieldName, `, _ := e.`, fieldName, `.ToPB()`)
				p.P(`resp.`, fieldName, ` = pb`, fieldName)
			}

		} else {
			p.P(`resp.`, fieldName, ` = e.`, fieldName)
		}

	} else {
		if bomField != nil && bomField.Tag.GetMongoObjectId() {
			p.P(`resp.`, fieldName, ` = e.`, fieldName, `.Hex()`)
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
	if (field.IsMessage() && !gogoproto.IsCustomType(field) && !gogoproto.IsStdType(field)) || p.IsGroup(field) {

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
				p.P(`pb, err := b.ToMongo()`)
				p.P(`if err != nil {`)
				p.P(`continue`)
				p.P(`}`)
				p.P(`sub`, fieldName, ` = append(sub`, fieldName, `, pb)`)
				p.P(`}`)

				p.P(`}`)
				p.P(`}`)

				p.P(`resp.`, fieldName, ` = sub`, fieldName)
			} else {
				p.P(`// create single mongo`)
				p.P(`pb`, fieldName, `, _ := e.`, fieldName, `.ToMongo()`)
				p.P(`resp.`, fieldName, ` = pb`, fieldName)
			}

		} else {
			p.P(`resp.`, fieldName, ` = e.`, fieldName)
		}

	} else if bomField != nil && bomField.Tag.GetMongoObjectId() {
		p.P(`resp.`, fieldName, ` = bom.ToObj(e.`, fieldName, `)`)
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
	p.P(`func (e *`, message.GetName(), `) ToMongo() (*`, mName, `, error) {`)
	p.P(`var resp *`, mName)
	p.P(`if prehook, ok := interface{}(e).(`, mName, `WithBeforeToMongo); ok {`)
	p.P(`if err := prehook.BeforeToMongo(resp); err != nil {`)
	p.P(`return resp, err`)
	p.P(`}`)
	p.P(`}`)
	for _, field := range message.GetField() {
		bomField := p.getFieldOptions(field)
		if bomField != nil && bomField.Tag.GetSkip() {
			// skip field
			continue
		}
		p.ToMongoGenerateFieldConversion(field, message, bomField)
	}
	p.P(`if posthook, ok := interface{}(e).(`, mName, `WithAfterToMongo); ok {`)
	p.P(`if err := posthook.AfterToMongo(resp); err != nil {`)
	p.P(`return resp, err`)
	p.P(`}`)
	p.P(`}`)
	p.P(`return resp, nil`)
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
