package v1

import (
	"ZhiShanYunXue/setting"
	"ZhiShanYunXue/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// NewTaskRequest 新建任务 请求结构体
type NewTaskRequest struct {
	TaskTitle       string          `json:"task_title" binding:"required"`
	TaskDescription string          `json:"task_description" binding:"required"`
	Deadline        string          `json:"deadline" binding:"required"`
	Answers         []util.QAAnswer `json:"answers"`
}

// NewTask 新建任务
func NewTask(c *gin.Context) {
	// 日志记录
	logger, _ := util.NewLogger()
	// 绑定请求参数
	var req NewTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, Data{
			Code: http.StatusUnprocessableEntity,
			Msg:  "请求格式错误或缺少必要参数",
		})
		return
	}
	logger.Info("验证数据成功")
	// 在服务端生成任务Id
	taskId := util.GenerateTaskId(setting.MaxTries)

	logger.Info(req.Answers)
	// 操作数据库 - 添加任务
	_, err := util.AddTask(taskId, req.TaskTitle, req.TaskDescription, req.Deadline, req.Answers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Data{
			Code: http.StatusInternalServerError,
			Msg:  "生成任务失败",
		})
		return
	}
	c.JSON(http.StatusCreated, Data{
		Code: http.StatusCreated,
		Msg:  "生成任务成功！",
		Data: gin.H{"task_id": taskId},
	})
	return
}

// GetInfoRequest 获取任务信息 请求结构体
type GetInfoRequest struct {
	TaskID string `form:"task_id" binding:"required"`
}

// GetInfo 获取任务信息
func GetInfo(c *gin.Context) {
	// 日志记录
	logger, _ := util.NewLogger()
	// 绑定请求参数
	var req GetInfoRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, Data{
			Code: http.StatusUnprocessableEntity,
			Msg:  "请求格式错误或缺少必要参数",
		})
		return
	}
	logger.Info("验证数据成功")
	// 操作数据库
	answersInfo, err := util.GetInfo(req.TaskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Data{
			Code: http.StatusInternalServerError,
			Msg:  "获取任务失败",
		})
		return
	}
	c.JSON(http.StatusOK, Data{
		Code: http.StatusOK,
		Data: answersInfo,
	})
}

// GetTaskDataRequest 获取任务列表 请求结构体
type GetTaskDataRequest struct {
	StudentId string `form:"student_id" binding:"required"`
	TaskId    string `form:"task_id" binding:"required"`
}

// GetTaskData 获取任务数据
func GetTaskData(c *gin.Context) {
	// 日志记录
	logger, _ := util.NewLogger()
	// 绑定请求参数
	var req GetTaskDataRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, Data{
			Code: http.StatusUnprocessableEntity,
			Msg:  "请求格式错误或缺少必要参数",
		})
		return
	}
	logger.Info("验证数据成功")
	// 操作数据库
	taskData, err := util.GetTaskData(req.TaskId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, Data{
			Code: http.StatusInternalServerError,
			Msg:  "获取任务失败",
		})
		return
	}

	// 写入获取任务的时间
	_, err = util.MarkGetTaskTime(req.StudentId, req.TaskId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Data{
			Code: http.StatusInternalServerError,
			Msg:  "写入开始时间失败",
		})
		return
	}

	c.JSON(http.StatusOK, Data{
		Code: http.StatusOK,
		Data: taskData,
	})
}

// PushAnswerRequest 提交答案 请求结构体
type PushAnswerRequest struct {
	StudentId string              `json:"student_id" binding:"required"`
	TaskId    string              `json:"task_id" binding:"required"`
	TaskData  *[]util.StuTaskData `json:"task_data" binding:"required"`
}

// PushAnswer 提交答案
func PushAnswer(c *gin.Context) {
	// 日志记录
	logger, _ := util.NewLogger()
	// 绑定请求参数
	var req PushAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, Data{
			Code: http.StatusUnprocessableEntity,
			Msg:  "请求格式错误或缺少必要参数",
		})
		return
	}
	logger.Info("验证数据成功")

	// 写入数据库
	_, err := util.PushTaskData(req.StudentId, req.TaskId, req.TaskData)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: student_task_answers.student_id, student_task_answers.task_id, student_task_answers.qa_id" {
			c.JSON(http.StatusInternalServerError, Data{
				Code: http.StatusInternalServerError,
				Msg:  "禁止重复提交",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, Data{
			Code: http.StatusInternalServerError,
			Msg:  "提交答案失败",
		})
		return
	}

	// 写入答题时间
	_, err = util.PushAnswerTime(req.StudentId, req.TaskId, time.Now().Format("2006-01-02 15:04:05.000"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, Data{
			Code: http.StatusInternalServerError,
			Msg:  "写入答题时间失败",
		})
		return
	}

	c.JSON(http.StatusOK, Data{
		Code: http.StatusOK,
		Msg:  "提交答案成功",
	})
	return
}

// GetReportRequest 获取报告 请求结构体
type GetReportRequest struct {
	StudentId string `form:"student_id" binding:"required"`
	TaskId    string `form:"task_id" binding:"required"`
}

// GetReport 获取报告
func GetReport(c *gin.Context) {
	// 日志记录
	logger, _ := util.NewLogger()

	// 绑定请求参数
	var req GetReportRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, Data{
			Code: http.StatusUnprocessableEntity,
			Msg:  "请求格式错误或缺少必要参数",
		})
		return
	}
	logger.Info("验证数据成功")

	// 从数据获取数据
	reportData, err := util.GetReportData(req.StudentId, req.TaskId)
	if err != nil || reportData.TaskData == nil {
		c.JSON(http.StatusInternalServerError, Data{
			Code: http.StatusInternalServerError,
			Msg:  "获取报告失败",
		})
		return
	}
	c.JSON(http.StatusOK, Data{
		Code: http.StatusOK,
		Data: reportData,
	})
}

// GetStatusReportDataRequest 获取学生任务状态报告数据 请求结构体
type GetStatusReportDataRequest struct {
	TaskId string `form:"task_id" binding:"required"`
}

// GetStatusReportData 获取学生任务状态报告数据
func GetStatusReportData(c *gin.Context) {
	// 日志记录
	logger, _ := util.NewLogger()
	// 绑定请求参数
	var req GetStatusReportDataRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, Data{
			Code: http.StatusUnprocessableEntity,
			Msg:  "请求格式错误或缺少必要参数",
		})
		return
	}
	logger.Info("验证数据成功")
	reportData, err := util.GetStatusReportData(req.TaskId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Data{
			Code: http.StatusInternalServerError,
			Msg:  "获取报告失败",
		})
		return
	}
	c.JSON(http.StatusOK, Data{
		Code: http.StatusOK,
		Data: reportData,
	})
}
