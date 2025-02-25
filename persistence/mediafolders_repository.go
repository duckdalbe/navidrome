package persistence

import (
	"context"

	"github.com/astaxie/beego/orm"
	"github.com/deluan/navidrome/conf"
	"github.com/deluan/navidrome/model"
)

type mediaFolderRepository struct {
	ctx context.Context
}

func NewMediaFolderRepository(ctx context.Context, o orm.Ormer) model.MediaFolderRepository {
	return &mediaFolderRepository{ctx}
}

func (r *mediaFolderRepository) Get(id string) (*model.MediaFolder, error) {
	mediaFolder := hardCoded()
	return &mediaFolder, nil
}

func (*mediaFolderRepository) GetAll() (model.MediaFolders, error) {
	mediaFolder := hardCoded()
	result := make(model.MediaFolders, 1)
	result[0] = mediaFolder
	return result, nil
}

func hardCoded() model.MediaFolder {
	mediaFolder := model.MediaFolder{ID: 0, Path: conf.Server.MusicFolder}
	mediaFolder.Name = "Music Library"
	return mediaFolder
}

var _ model.MediaFolderRepository = (*mediaFolderRepository)(nil)
