package handlers

import (
	"database/sql"

	"golang.org/x/oauth2"
)

type App struct {
	DB           *sql.DB
	FileUpload   *string
	InstanceType *string
	OutputFormat *string
	Output       *string
	Concurrency  *int
	Oauth2       *oauth2.Config
}

type SalesforceOauth struct {
	Instance string
	Token    oauth2.Token
}

type wsMessageReceived struct {
	Cmd      string                 `json:"cmd"`
	Query    string                 `json:"query"`
	Data     map[string]interface{} `json:"data"`
	Filename string                 `json:"filename"`
}

type wsMessageSend struct {
	TotalSize int  `json:"totalSize"`
	Records   int  `json:"records"`
	Done      bool `json:"done"`
}

type Records struct {
	TotalSize      int                      `json:"totalSize"`
	Done           bool                     `json:"done"`
	NextRecordsURL string                   `json:"nextRecordsUrl"`
	Records        []map[string]interface{} `json:"records"`
}

type Describe struct {
	Name   string           `json:"name"`
	Fields []DescribeFields `json:"fields"`
}

type DescribeFields struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Upload struct {
	Title        string `json:"Title"`
	Description  string `json:"Description"`
	PathOnClient string `json:"PathOnClient"`
	VersionData  string `json:"VersionData"`
}

type UploadResponse struct {
	Id      string   `json:"id"`
	Success bool     `json:"success"`
	Error   []string `json:"error"`
}

type ContentVersionResponse struct {
	Id                string `json:"id"`
	ContentDocumentId string `json:"ContentDocumentId"`
}

type ContentDocumentLink struct {
	LinkedEntityId    string `json:"LinkedEntityId"`
	ContentDocumentId string `json:"ContentDocumentId"`
}

type Result struct {
	OldContentDocumentId string
	ContentDocumentId    string
	LinkedEntityId       string
	FilePath             string
	Error                string `json:"Error,omitempty"`
}

type Download struct {
	FilePath          string `json:"FilePath"`
	Title             string `json:"Title"`
	FileExtension     string `json:"FileExtension"`
	PathOnClient      string `json:"PathOnClient"`
	ContentDocumentId string `json:"ContentDocumentId"`
	LinkedEntityId    string `json:"LinkedEntityId"`
	Description       string `json:"Description"`
}
