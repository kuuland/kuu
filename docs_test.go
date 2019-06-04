package kuu

import (
	"testing"
)

func TestMarshal(t *testing.T) {
	doc := Doc{
		Openapi: "3.0.1",
		Info: DocInfo{
			Title:       "Humansa",
			Description: "Humansa 模型默认RESTful接口文档",
			Contact: DocInfoContact{
				Email: "yinfxs@dexdev.me",
			},
			Version: "1.0.0",
		},
		Servers: []DocServer{
			{
				Url: "https://humansa.hofo.co",
			},
		},
		Tags: []DocTag{
			{
				Name:        "Member",
				Description: "会员资讯",
			},
			{
				Name:        "Address",
				Description: "会员地址",
			},
		},
		Paths: map[string]DocPathItem{
			"/member": {
				"post": DocPathDesc{
					Tags:        []string{"Member"},
					Summary:     "新增会员资讯",
					OperationID: "createMember",
					RequestBody: DocPathRequestBody{
						Required:    true,
						Description: "新增会员资讯信息",
						Content: map[string]DocPathContentItem{
							"application/json": {
								Schema: DocPathSchema{
									Ref: "#/components/schemas/Member",
								},
							},
						},
					},
					Responses: map[int]DocPathResponse{
						200: {
							Description: "接口调用成功",
							Content: map[string]DocPathContentItem{
								"application/json": {
									Schema: DocPathSchema{
										Ref: "#/components/schemas/Member",
									},
								},
							},
						},
					},
				},
			},
		},
		Components: DocComponent{
			Schemas: map[string]DocComponentSchema{
				"Member": {
					Type: "object",
					Properties: map[string]DocSchemaProperty{
						"ID": {
							Type: "integer",
						},
						"Avatar": {
							Type: "string",
						},
						"IsPassChanged": {
							Type:    "boolean",
							Default: false,
						},
					},
				},
			},
			SecuritySchemes: map[string]DocSecurityScheme{
				"api_key": {
					Type: "apiKey",
					Name: "api_key",
					In:   "header",
				},
			},
		},
	}
	yml := doc.Marshal()
	if len(yml) < 0 {
		t.Error("文档生成失败")
	}
}
