package openapi

import (
	"github.com/pb33f/libopenapi/datamodel"
	v3 "github.com/pb33f/libopenapi/datamodel/low/3.0"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

var doc *v3.Document

func init() {
	data, _ := ioutil.ReadFile("../test_specs/burgershop.openapi.yaml")
	info, _ := datamodel.ExtractSpecInfo(data)
	doc, _ = CreateDocument(info)
}

func BenchmarkCreateDocument(b *testing.B) {
	data, _ := ioutil.ReadFile("../test_specs/burgershop.openapi.yaml")
	info, _ := datamodel.ExtractSpecInfo(data)
	for i := 0; i < b.N; i++ {
		doc, _ = CreateDocument(info)
	}
}

func TestCreateDocument(t *testing.T) {
	assert.Equal(t, "3.0.1", doc.Version.Value)
	assert.Equal(t, "Burger Shop", doc.Info.Value.Title.Value)
	assert.NotEmpty(t, doc.Info.Value.Title.Value)
}

func TestCreateDocument_Info(t *testing.T) {
	assert.Equal(t, "https://pb33f.io", doc.Info.Value.TermsOfService.Value)
	assert.Equal(t, "pb33f", doc.Info.Value.Contact.Value.Name.Value)
	assert.Equal(t, "buckaroo@pb33f.io", doc.Info.Value.Contact.Value.Email.Value)
	assert.Equal(t, "https://pb33f.io", doc.Info.Value.Contact.Value.URL.Value)
	assert.Equal(t, "pb33f", doc.Info.Value.License.Value.Name.Value)
	assert.Equal(t, "https://pb33f.io/made-up", doc.Info.Value.License.Value.URL.Value)
}

func TestCreateDocument_Servers(t *testing.T) {
	assert.Len(t, doc.Servers, 2)
	server1 := doc.Servers[0]
	server2 := doc.Servers[1]

	// server 1
	assert.Equal(t, "{scheme}://api.pb33f.io", server1.Value.URL.Value)
	assert.NotEmpty(t, server1.Value.Description.Value)
	assert.Len(t, server1.Value.Variables.Value, 1)
	assert.Len(t, server1.Value.Variables.Value["scheme"].Value.Enum, 2)
	assert.Equal(t, server1.Value.Variables.Value["scheme"].Value.Default.Value, "https")
	assert.NotEmpty(t, server1.Value.Variables.Value["scheme"].Value.Description.Value)

	// server 2
	assert.Equal(t, "https://{domain}.{host}.com", server2.Value.URL.Value)
	assert.NotEmpty(t, server2.Value.Description.Value)
	assert.Len(t, server2.Value.Variables.Value, 2)
	assert.Equal(t, server2.Value.Variables.Value["domain"].Value.Default.Value, "api")
	assert.NotEmpty(t, server2.Value.Variables.Value["domain"].Value.Description.Value)
	assert.NotEmpty(t, server2.Value.Variables.Value["host"].Value.Description.Value)
	assert.Equal(t, server2.Value.Variables.Value["host"].Value.Default.Value, "pb33f.io")
	assert.Equal(t, "1.2", doc.Info.Value.Version.Value)
}

func TestCreateDocument_Tags(t *testing.T) {
	assert.Len(t, doc.Tags, 2)

	// tag1
	assert.Equal(t, "Burgers", doc.Tags[0].Value.Name.Value)
	assert.NotEmpty(t, doc.Tags[0].Value.Description.Value)
	assert.NotNil(t, doc.Tags[0].Value.ExternalDocs.Value)
	assert.Equal(t, "https://pb33f.io", doc.Tags[0].Value.ExternalDocs.Value.URL.Value)
	assert.NotEmpty(t, doc.Tags[0].Value.ExternalDocs.Value.URL.Value)
	assert.Len(t, doc.Tags[0].Value.Extensions, 7)

	for key, extension := range doc.Tags[0].Value.Extensions {
		switch key.Value {
		case "x-internal-ting":
			assert.Equal(t, "somethingSpecial", extension.Value)
		case "x-internal-tong":
			assert.Equal(t, int64(1), extension.Value)
		case "x-internal-tang":
			assert.Equal(t, 1.2, extension.Value)
		case "x-internal-tung":
			assert.Equal(t, true, extension.Value)
		case "x-internal-arr":
			assert.Len(t, extension.Value, 2)
			assert.Equal(t, "one", extension.Value.([]interface{})[0].(string))
		case "x-internal-arrmap":
			assert.Len(t, extension.Value, 2)
			assert.Equal(t, "now", extension.Value.([]interface{})[0].(map[string]interface{})["what"])
		case "x-something-else":
			// crazy times in the upside down. this API should be avoided for the higher up use cases.
			// this is why we will need a higher level API to this model, this looks cool and all, but dude.
			assert.Equal(t, "now?", extension.Value.(map[string]interface{})["ok"].([]interface{})[0].(map[string]interface{})["what"])
		}

	}

	/// tag2
	assert.Equal(t, "Dressing", doc.Tags[1].Value.Name.Value)
	assert.NotEmpty(t, doc.Tags[1].Value.Description.Value)
	assert.NotNil(t, doc.Tags[1].Value.ExternalDocs.Value)
	assert.Equal(t, "https://pb33f.io", doc.Tags[1].Value.ExternalDocs.Value.URL.Value)
	assert.NotEmpty(t, doc.Tags[1].Value.ExternalDocs.Value.URL.Value)
	assert.Len(t, doc.Tags[1].Value.Extensions, 0)

}

func TestCreateDocument_Paths(t *testing.T) {
	assert.Len(t, doc.Paths.Value.PathItems, 6)
	burgerId := doc.Paths.Value.FindPath("/burgers/{burgerId}")
	assert.NotNil(t, burgerId)
	assert.Len(t, burgerId.Value.Get.Value.Parameters.Value, 2)
	param := burgerId.Value.Get.Value.Parameters.Value[1]
	assert.Equal(t, "burgerHeader", param.Value.Name.Value)
	prop := param.Value.Schema.Value.FindProperty("burgerTheme")
	assert.Equal(t, "something about a theme?", prop.Value.Description.Value)
	assert.Equal(t, "big-mac", param.Value.Example.Value)

	// check content
	pContent := param.Value.FindContent("application/json")
	assert.Equal(t, "somethingNice", pContent.Value.Example.Value)

	encoding := pContent.Value.FindPropertyEncoding("burgerTheme")
	assert.NotNil(t, encoding.Value)
	assert.Len(t, encoding.Value.Headers.Value, 1)

	header := encoding.Value.FindHeader("someHeader")
	assert.NotNil(t, header.Value)
	assert.Equal(t, "this is a header", header.Value.Description.Value)
	assert.Equal(t, "string", header.Value.Schema.Value.Type.Value)

	// check request body on operation
	burgers := doc.Paths.Value.FindPath("/burgers")
	assert.NotNil(t, burgers.Value.Post.Value)

	burgersPost := burgers.Value.Post.Value
	assert.Equal(t, "createBurger", burgersPost.OperationId.Value)
	assert.Equal(t, "Create a new burger", burgersPost.Summary.Value)
	assert.NotEmpty(t, burgersPost.Description.Value)

	requestBody := burgersPost.RequestBody.Value

	assert.NotEmpty(t, requestBody.Description.Value)
	content := requestBody.FindContent("application/json").Value

	assert.NotNil(t, content)
	assert.Len(t, content.Schema.Value.Properties.Value, 4)
	assert.Len(t, content.GetAllExamples(), 2)

	ex := content.FindExample("pbjBurger")
	assert.NotNil(t, ex.Value)
	assert.NotEmpty(t, ex.Value.Summary.Value)
	assert.NotNil(t, ex.Value.Value.Value)

	if n, ok := ex.Value.Value.Value.(map[string]interface{}); ok {
		assert.Len(t, n, 2)
		assert.Equal(t, 3, n["numPatties"])
	} else {
		assert.Fail(t, "should easily be convertable. something changed!")
	}

	cb := content.FindExample("cakeBurger")
	assert.NotNil(t, cb.Value)
	assert.NotEmpty(t, cb.Value.Summary.Value)
	assert.NotNil(t, cb.Value.Value.Value)

	if n, ok := cb.Value.Value.Value.(map[string]interface{}); ok {
		assert.Len(t, n, 2)
		assert.Equal(t, "Chocolate Cake Burger", n["name"])
		assert.Equal(t, 5, n["numPatties"])
	} else {
		assert.Fail(t, "should easily be convertable. something changed!")
	}

	// check responses
	responses := burgersPost.Responses.Value
	assert.NotNil(t, responses)
	assert.Len(t, responses.Codes, 3)

	okCode := responses.FindResponseByCode("200")
	assert.NotNil(t, okCode.Value)
	assert.Equal(t, "A tasty burger for you to eat.", okCode.Value.Description.Value)

	// check headers are populated
	assert.Len(t, okCode.Value.Headers.Value, 1)
	okheader := okCode.Value.FindHeader("UseOil")
	assert.NotNil(t, okheader.Value)
	assert.Equal(t, "this is a header", okheader.Value.Description.Value)

	respContent := okCode.Value.FindContent("application/json").Value
	assert.NotNil(t, respContent)

	assert.NotNil(t, respContent.Schema.Value)
	assert.Len(t, respContent.Schema.Value.Required.Value, 2)

	respExample := respContent.FindExample("quarterPounder")
	assert.NotNil(t, respExample.Value)
	assert.NotNil(t, respExample.Value.Value.Value)

	if n, ok := respExample.Value.Value.Value.(map[string]interface{}); ok {
		assert.Len(t, n, 2)
		assert.Equal(t, "Quarter Pounder with Cheese", n["name"])
		assert.Equal(t, 1, n["numPatties"])
	} else {
		assert.Fail(t, "should easily be convertable. something changed!")
	}

	// check links
	links := okCode.Value.Links
	assert.NotNil(t, links.Value)
	assert.Len(t, links.Value, 2)
	assert.Equal(t, "locateBurger", okCode.Value.FindLink("LocateBurger").Value.OperationId.Value)

}