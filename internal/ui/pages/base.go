package pages
import (
	"lazytea-mobile/internal/data"
	"lazytea-mobile/internal/network"
	"lazytea-mobile/internal/utils"
	"fyne.io/fyne/v2"
)
type PageBase struct {
	client  *network.Client
	storage *data.Storage
	logger  *utils.Logger
	content fyne.CanvasObject
}
func NewPageBase(client *network.Client, storage *data.Storage, logger *utils.Logger) *PageBase {
	return &PageBase{
		client:  client,
		storage: storage,
		logger:  logger,
	}
}
func (p *PageBase) GetClient() *network.Client {
	return p.client
}
func (p *PageBase) GetStorage() *data.Storage {
	return p.storage
}
func (p *PageBase) GetLogger() *utils.Logger {
	return p.logger
}
func (p *PageBase) GetContent() fyne.CanvasObject {
	return p.content
}
func (p *PageBase) SetContent(content fyne.CanvasObject) {
	p.content = content
}