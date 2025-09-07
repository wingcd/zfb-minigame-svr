package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web"
)

type UploadController struct {
	web.Controller
}

// UploadFile 上传文件
func (c *UploadController) UploadFile() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取上传的文件
	file, header, err := c.GetFile("file")
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1002, "获取上传文件失败: "+err.Error(), nil)
		return
	}
	defer file.Close()

	// 获取参数
	fileType := c.GetString("fileType", "image") // image, document, archive, other
	isPublic := c.GetString("isPublic", "false") == "true"

	// 检查文件大小
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if header.Size > maxSize {
		utils.ErrorResponse(&c.Controller, 1002, "文件大小超过限制(10MB)", nil)
		return
	}

	// 检查文件类型
	allowedTypes := map[string][]string{
		"image":    {".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"},
		"document": {".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt"},
		"archive":  {".zip", ".rar", ".7z", ".tar", ".gz"},
		"other":    {}, // 其他类型不限制
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if types, exists := allowedTypes[fileType]; exists && len(types) > 0 {
		allowed := false
		for _, allowedExt := range types {
			if ext == allowedExt {
				allowed = true
				break
			}
		}
		if !allowed {
			utils.ErrorResponse(&c.Controller, 1002, "不支持的文件类型", nil)
			return
		}
	}

	// 生成文件路径
	uploadDir := "uploads"
	if !isPublic {
		uploadDir = "uploads/private"
	}

	// 按日期创建子目录
	dateDir := time.Now().Format("2006/01/02")
	fullDir := filepath.Join(uploadDir, dateDir)

	// 创建目录
	err = os.MkdirAll(fullDir, 0755)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "创建上传目录失败: "+err.Error(), nil)
		return
	}

	// 生成唯一文件名
	hash := md5.New()
	hash.Write([]byte(header.Filename + time.Now().String()))
	fileName := fmt.Sprintf("%x%s", hash.Sum(nil), ext)
	filePath := filepath.Join(fullDir, fileName)

	// 保存文件
	err = c.SaveToFile("file", filePath)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "保存文件失败: "+err.Error(), nil)
		return
	}

	// 计算文件MD5
	fileMD5, err := calculateFileMD5(filePath)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "计算文件MD5失败: "+err.Error(), nil)
		return
	}

	// 保存文件信息到数据库
	fileInfo := models.FileInfo{
		OriginalName: header.Filename,
		FileName:     fileName,
		FilePath:     filePath,
		FileSize:     header.Size,
		FileType:     fileType,
		FileExt:      ext,
		FileMD5:      fileMD5,
		IsPublic:     isPublic,
		UploadBy:     claims.UserID,
		UploadTime:   time.Now().Unix(),
	}

	err = models.SaveFileInfo(&fileInfo)
	if err != nil {
		// 删除已上传的文件
		os.Remove(filePath)
		utils.ErrorResponse(&c.Controller, 1003, "保存文件信息失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "上传文件", "上传文件: "+header.Filename)

	result := map[string]interface{}{
		"fileId":       fileInfo.Id,
		"fileName":     fileName,
		"originalName": header.Filename,
		"fileSize":     header.Size,
		"fileType":     fileType,
		"filePath":     filePath,
		"isPublic":     isPublic,
	}

	utils.SuccessResponse(&c.Controller, "上传成功", result)
}

// GetFileList 获取文件列表
func (c *UploadController) GetFileList() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取分页参数
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")
	fileType := c.GetString("fileType", "")
	_ = c.GetString("keyword", "") // 暂时不使用关键词搜索
	uploadBy := c.GetString("uploadBy", "")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 创建过滤条件
	filters := make(map[string]interface{})
	if fileType != "" {
		filters["fileType"] = fileType
	}
	if uploadBy != "" {
		if uploadByInt, err := strconv.ParseInt(uploadBy, 10, 64); err == nil {
			filters["uploadBy"] = uploadByInt
		}
	}

	// 获取文件列表
	files, total, err := models.GetFileList(page, pageSize, filters)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取文件列表失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"files":    files,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// GetFileInfo 获取文件信息
func (c *UploadController) GetFileInfo() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取文件ID
	fileIdStr := c.Ctx.Input.Param(":id")
	if fileIdStr == "" {
		utils.ErrorResponse(&c.Controller, 1002, "文件ID不能为空", nil)
		return
	}

	fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1002, "文件ID格式错误", nil)
		return
	}

	// 获取文件信息
	fileInfo, err := models.GetFileInfo(fileId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取文件信息失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", fileInfo)
}

// DownloadFile 下载文件
func (c *UploadController) DownloadFile() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取文件ID
	fileIdStr := c.Ctx.Input.Param(":id")
	if fileIdStr == "" {
		utils.ErrorResponse(&c.Controller, 1002, "文件ID不能为空", nil)
		return
	}

	fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1002, "文件ID格式错误", nil)
		return
	}

	// 获取文件信息
	fileInfo, err := models.GetFileInfo(fileId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取文件信息失败: "+err.Error(), nil)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(fileInfo.FilePath); os.IsNotExist(err) {
		utils.ErrorResponse(&c.Controller, 1003, "文件不存在", nil)
		return
	}

	// 更新下载次数
	models.UpdateFileDownloadCount(fileId)

	// 设置下载头
	c.Ctx.Output.Header("Content-Disposition", "attachment; filename="+fileInfo.OriginalName)
	c.Ctx.Output.Header("Content-Type", "application/octet-stream")

	// 输出文件
	http.ServeFile(c.Ctx.ResponseWriter, c.Ctx.Request, fileInfo.FilePath)
}

// DeleteFile 删除文件
func (c *UploadController) DeleteFile() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取文件ID
	fileIdStr := c.Ctx.Input.Param(":id")
	if fileIdStr == "" {
		utils.ErrorResponse(&c.Controller, 1002, "文件ID不能为空", nil)
		return
	}

	fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1002, "文件ID格式错误", nil)
		return
	}

	// 获取文件信息
	fileInfo, err := models.GetFileInfo(fileId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取文件信息失败: "+err.Error(), nil)
		return
	}

	// 删除物理文件
	if _, err := os.Stat(fileInfo.FilePath); err == nil {
		os.Remove(fileInfo.FilePath)
	}

	// 删除数据库记录
	err = models.DeleteFileInfo(fileId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "删除文件记录失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "删除文件", "删除文件: "+fileInfo.OriginalName)

	utils.SuccessResponse(&c.Controller, "删除成功", nil)
}

// BatchDeleteFiles 批量删除文件
func (c *UploadController) BatchDeleteFiles() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取文件ID列表
	fileIdsStr := c.GetStrings("fileIds")
	if len(fileIdsStr) == 0 {
		utils.ErrorResponse(&c.Controller, 1002, "文件ID列表不能为空", nil)
		return
	}

	var fileIds []int
	for _, idStr := range fileIdsStr {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			fileIds = append(fileIds, id)
		}
	}

	if len(fileIds) == 0 {
		utils.ErrorResponse(&c.Controller, 1002, "有效的文件ID列表不能为空", nil)
		return
	}

	// 转换fileIds为int64类型
	var fileIds64 []int64
	for _, id := range fileIds {
		fileIds64 = append(fileIds64, int64(id))
	}

	// 批量删除文件
	err := models.BatchDeleteFiles(fileIds64)
	deletedCount := len(fileIds64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "批量删除文件失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "批量删除文件", fmt.Sprintf("批量删除%d个文件", deletedCount))

	result := map[string]interface{}{
		"deletedCount": deletedCount,
	}

	utils.SuccessResponse(&c.Controller, "删除成功", result)
}

// GetUploadStats 获取上传统计
func (c *UploadController) GetUploadStats() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取参数
	days := c.GetString("days", "7")
	dayCount, err := strconv.Atoi(days)
	if err != nil || dayCount <= 0 || dayCount > 30 {
		dayCount = 7
	}

	// 获取上传统计
	stats, err := models.GetUploadStats()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取上传统计失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", stats)
}

// CleanupFiles 清理无效文件
func (c *UploadController) CleanupFiles() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 清理无效文件
	cleanedCount, err := models.CleanupInvalidFiles()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "清理无效文件失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "清理无效文件", fmt.Sprintf("清理了%d个无效文件", cleanedCount))

	result := map[string]interface{}{
		"cleanedCount": cleanedCount,
	}

	utils.SuccessResponse(&c.Controller, "清理成功", result)
}

// calculateFileMD5 计算文件MD5值
func calculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
