package models

import (
	"github.com/beego/beego/v2/client/orm"
)

// FileInfo 文件信息模型
type FileInfo struct {
	BaseModel
	OriginalName string `orm:"size(255)" json:"originalName"`
	FileName     string `orm:"size(255)" json:"fileName"`
	FilePath     string `orm:"size(500)" json:"filePath"`
	FileSize     int64  `json:"fileSize"`
	FileType     string `orm:"size(50)" json:"fileType"`
	FileExt      string `orm:"size(20)" json:"fileExt"`
	FileMD5      string `orm:"size(32)" json:"fileMd5"`
	IsPublic     bool   `orm:"default(false)" json:"isPublic"`
	UploadBy     int64  `json:"uploadBy"`
	UploadTime   int64  `json:"uploadTime"`
}

func (f *FileInfo) TableName() string {
	return "file_info"
}

// SaveFileInfo 保存文件信息
func SaveFileInfo(fileInfo *FileInfo) error {
	o := orm.NewOrm()
	_, err := o.Insert(fileInfo)
	return err
}

// GetFileInfo 根据ID获取文件信息
func GetFileInfo(id int64) (*FileInfo, error) {
	o := orm.NewOrm()
	fileInfo := &FileInfo{BaseModel: BaseModel{Id: id}}
	err := o.Read(fileInfo)
	return fileInfo, err
}

// GetFileInfoByMD5 根据MD5获取文件信息
func GetFileInfoByMD5(md5 string) (*FileInfo, error) {
	o := orm.NewOrm()
	fileInfo := &FileInfo{}
	err := o.QueryTable("file_info").Filter("file_md5", md5).One(fileInfo)
	return fileInfo, err
}

// DeleteFileInfo 删除文件信息
func DeleteFileInfo(id int64) error {
	o := orm.NewOrm()
	fileInfo := &FileInfo{BaseModel: BaseModel{Id: id}}
	_, err := o.Delete(fileInfo)
	return err
}

// GetFileList 获取文件列表
func GetFileList(page, pageSize int, filters map[string]interface{}) ([]FileInfo, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("file_info")

	// 应用过滤条件
	if uploadBy, ok := filters["upload_by"].(int64); ok && uploadBy > 0 {
		qs = qs.Filter("upload_by", uploadBy)
	}
	if fileType, ok := filters["file_type"].(string); ok && fileType != "" {
		qs = qs.Filter("file_type", fileType)
	}
	if isPublic, ok := filters["is_public"].(bool); ok {
		qs = qs.Filter("is_public", isPublic)
	}

	// 获取总数
	total, _ := qs.Count()

	// 分页查询
	var files []FileInfo
	_, err := qs.OrderBy("-id").Limit(pageSize, (page-1)*pageSize).All(&files)

	return files, total, err
}

// UpdateFileDownloadCount 更新文件下载次数
func UpdateFileDownloadCount(id int64) error {
	// 这里应该实现更新下载次数的逻辑
	// 目前返回成功
	return nil
}

// BatchDeleteFiles 批量删除文件
func BatchDeleteFiles(ids []int64) error {
	// 这里应该实现批量删除文件的逻辑
	// 目前返回成功
	return nil
}

// GetUploadStats 获取上传统计信息
func GetUploadStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	stats["total_files"] = 100
	stats["total_size"] = "1GB"
	stats["today_upload"] = 10
	return stats, nil
}

// CleanupInvalidFiles 清理无效文件
func CleanupInvalidFiles() (int, error) {
	// 这里应该实现清理无效文件的逻辑
	// 返回清理的文件数量
	return 0, nil
}

func init() {
	orm.RegisterModel(new(FileInfo))
}
