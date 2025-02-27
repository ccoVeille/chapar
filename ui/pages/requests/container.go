package requests

import (
	"github.com/chapar-rest/chapar/ui/chapartheme"
	"github.com/chapar-rest/chapar/ui/widgets"

	"gioui.org/layout"
	"github.com/chapar-rest/chapar/internal/domain"
)

const (
	TypeRequest    = "request"
	TypeCollection = "collection"

	TypeMeta = "Type"
)

type Container interface {
	Layout(gtx layout.Context, theme *chapartheme.Theme) layout.Dimensions
	SetOnDataChanged(f func(id string, data any))
	SetOnTitleChanged(f func(title string))
	SetDataChanged(changed bool)
	SetOnSave(f func(id string))
	ShowPrompt(title, content, modalType string, onSubmit func(selectedOption string, remember bool), options ...widgets.Option)
	HidePrompt()
}

type RestContainer interface {
	SetHTTPResponse(response domain.HTTPResponseDetail)
	GetHTTPResponse() *domain.HTTPResponseDetail
	SetPostRequestSetPreview(preview string)
	ShowSendingRequestLoading()
	HideSendingRequestLoading()
	SetQueryParams(params []domain.KeyValue)
	SetPathParams(params []domain.KeyValue)
	SetURL(url string)
	SetPostRequestSetValues(set domain.PostRequestSet)
	SetOnPostRequestSetChanged(f func(id string, statusCode int, item, from, fromKey string))
	SetOnBinaryFileSelect(f func(id string))
	SetBinaryBodyFilePath(filePath string)
	SetOnFormDataFileSelect(f func(requestId, fieldId string))
	AddFileToFormData(fieldId, filePath string)
}
