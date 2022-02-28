package ortfodb

// configurationJSONSchema is the entire json string from ortfodb.yaml.json.schema
const configurationJSONSchema = `{
	"$schema": "http://json-schema.org/schema",
	"$id": "ortfodb.yaml",
	"definitions": {
		"validate_check": {
			"type": [
				"string",
				"boolean"
			],
			"enum": [
				"warn",
				"error",
				"fatal",
				"info",
				"off",
				false,
				true
			]
		}
	},
	"properties": {
		"build steps": {
			"type": "object",
			"properties": {
				"extract colors": {
					"type": "object",
					"properties": {
						"extract": {
							"enum": [
								"primary",
								"secondary",
								"tertiary"
							]
						},
						"default file name": {
							"type": "array",
							"items": {
								"type": "string"
							}
						}
					}
				},
				"make gifs": {
					"type": "object",
					"properties": {
						"file name template": {
							"type": "string"
						}
					}
				},
				"make thumbnails": {
					"type": "object",
					"properties": {
						"widths": {
							"type": "array",
							"items": {
								"type": "integer"
							}
						},
						"input file": {
							"type": "string"
						},
						"file name template": {
							"type": "string"
						}
					}
				}
			}
		},
		"features": {
			"type": "object",
			"properties": {
				"made with": {
					"type": "boolean"
				},
				"media hoisting": {
					"type": "boolean"
				}
			}
		},
		"validate": {
			"type": "object",
			"properties": {
				"checks": {
					"type": "object",
					"properties": {
						"schema compliance": {
							"$ref": "#/definitions/validate_check",
							"default": "fatal"
						},
						"work folder uniqueness": {
							"$ref": "#/definitions/validate_check",
							"default": "fatal"
						},
						"work folder safeness": {
							"$ref": "#/definitions/validate_check",
							"default": "error"
						},
						"yaml header": {
							"$ref": "#/definitions/validate_check",
							"default": "error"
						},
						"title presence": {
							"$ref": "#/definitions/validate_check",
							"default": "error"
						},
						"title uniqueness": {
							"$ref": "#/definitions/validate_check",
							"default": "error"
						},
						"tags presence": {
							"$ref": "#/definitions/validate_check",
							"default": "warn"
						},
						"tags knowledge": {
							"$ref": "#/definitions/validate_check",
							"default": "error"
						},
						"working media": {
							"$ref": "#/definitions/validate_check",
							"default": "warn"
						},
						"working urls": {
							"$ref": "#/definitions/validate_check",
							"default": false
						}
					}
				}
			}
		},
		"markdown": {
			"type": "object",
			"properties": {
				"abbreviations": {
					"type": "boolean"
				},
				"definition lists": {
					"type": "boolean"
				},
				"admonitions": {
					"type": "boolean"
				},
				"markdown in HTML": {
					"type": "boolean"
				},
				"new-line-to-line-break": {
					"type": "boolean"
				},
				"smarty pants": {
					"type": "boolean"
				},
				"anchored headings": {
					"type": "boolean"
				},
				"custom syntaxes": {
					"type": "array",
					"items": {
						"type": "object",
						"properties": {
							"from": {
								"type": "string"
							},
							"to": {
								"type": "string"
							}
						}
					}
				}
			}
		}
	}
}
`

// databaseJSONSchema is the entire json string from database.json.schema
const databaseJSONSchema = `{
	"$schema": "http://json-schema.org/schema",
	"$id": "portfoliodb-database",
	"type": "array",
	"items": {
		"type": "object",
		"required": [
			"id",
			"metadata",
			"paragraphs",
			"title",
			"media",
			"links",
			"footnotes"
		],
		"properties": {
			"id": {
				"type": "string"
			},
			"metadata": {
				"type": "object"
			},
			"paragraphs": {
				"type": "object",
				"additionalProperties": {
					"type": "array",
					"items": {
						"type": "object",
						"required": [
							"id",
							"content"
						],
						"properties": {
							"id": {
								"type": "string"
							},
							"content": {
								"type": "string"
							}
						}
					}
				}
			},
			"title": {
				"type": "object",
				"additionalProperties": {
					"type": "string"
				}
			},
			"media": {
				"type": "object",
				"additionalProperties": {
					"type": "array",
					"items": {
						"type": "object",
						"required": [
							"id",
							"alt",
							"title",
							"source",
							"content_type",
							"size",
							"dimensions",
							"duration",
							"online"
						],
						"properties": {
							"id": {
								"type": "string"
							},
							"alt": {
								"type": "string"
							},
							"title": {
								"type": "string"
							},
							"source": {
								"type": "string"
							},
							"content_type": {
								"type": "string"
							},
							"size": {
								"type": "number"
							},
							"dimensions": {
								"type": "object",
								"required": [
									"width",
									"height",
									"aspect_ratio"
								],
								"properties": {
									"width": {
										"type": "number"
									},
									"height": {
										"type": "number"
									},
									"aspect_ratio": {
										"type": "number"
									}
								}
							},
							"duration": {
								"type": "number"
							},
							"online": {
								"type": "boolean"
							},
							"attributes": {
								"type": "object",
								"properties": {
									"looped": {
										"type": "boolean",
										"description": "Whether to add looped to the potential HTML element's attributes (<video>, <audio>)."
									},
									"autoplay": {
										"type": "boolean",
										"description": "Whether to add autoplay to the potential HTML element's attributes (<video>, <audio>)."
									},
									"muted": {
										"type": "boolean",
										"description": "Whether to add muted to the potential HTML element's attributes (<video>, <audio>)."
									},
									"playsinline": {
										"type": "boolean",
										"description": "Whether to add playsinline to the potential HTML element's attributes (<video>, <audio>)."
									},
									"controls": {
										"type": "boolean",
										"description": "Whether to add controls to the potential HTML element's attributes (<video>, <audio>)."
									}
								}
							}
						}
					}
				}
			},
			"links": {
				"type": "object",
				"additionalProperties": {
					"type": "array",
					"items": {
						"type": "object",
						"required": [
							"id",
							"name",
							"title",
							"url"
						],
						"properties": {
							"id": {
								"type": "string"
							},
							"name": {
								"type": "string"
							},
							"title": {
								"type": "string"
							},
							"url": {
								"type": "string"
							}
						}
					}
				}
			},
			"footnotes": {
				"type": "object",
				"additionalProperties": {
					"type": "array",
					"items": {
						"type": "object",
						"required": [
							"name",
							"content"
						],
						"properties": {
							"name": {
								"type": "string"
							},
							"content": {
								"type": "string"
							}
						}
					}
				}
			}
		}
	}
}
`
