package util

import (
	"ZhiShanYunXue/setting"
	"database/sql"
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// QAAnswer 任务答案 结构体
type QAAnswer struct {
	QaTitle  string `json:"qa_title" binding:"required"`
	QaNumber int    `json:"qa_number" binding:"required"`
	QaAnswer string `json:"qa_answer" binding:"required"`
}

var (
	db *sql.DB
)

// 初始化
func init() {
	// 打开数据库
	logger, _ := NewLogger()
	var err error
	db, err = sql.Open(setting.DbDriverName, setting.DbName)
	if err != nil {
		// 处理错误
		logger.Error("打开数据库失败: %v", err)
		return
	}

}

// InitSqlite 初始化sqlite数据库
func InitSqlite() {
	logger, _ := NewLogger()

	// 创建任务表
	if err := CreateTable(db, "tasks"); err != nil {
		// 处理错误
		logger.Error("创建任务表错误: %v", err)
		return
	}
	// 创建任务数据表
	if err := CreateTable(db, "task_data"); err != nil {
		// 处理错误
		logger.Error("创建任务数据表错误: %v", err)
		return
	}
	// 创建任务和题目关联表
	if err := CreateTable(db, "task_qa_relations"); err != nil {
		// 处理错误
		logger.Error("创建任务和题目关联表错误: %v", err)
		return
	}
	// 创建学生和任务关联表
	if err := CreateTable(db, "student_task_answers"); err != nil {
		// 处理错误
		logger.Error("创建学生和任务关联表错误: %v", err)
		return
	}
	// 创建任务时间表
	if err := CreateTable(db, "task_time"); err != nil {
		// 处理错误
		logger.Error("创建任务时间表错误: %v", err)
	}
}

// CreateTable 创建表
func CreateTable(db *sql.DB, table string) error {
	// 声明变量
	var s string
	switch table {
	case "tasks":
		// 任务
		s = `create table if not exists tasks
		(
			task_id          TEXT not null,
			task_title       TEXT not null,
			task_description TEXT not null,
			publish_time     TEXT not null,
			Deadline         TEXT not null
		)`

	case "task_data":
		// 任务数据(题目)
		s = `create table if not exists task_data
		(
			qa_id    TEXT not null,
			q_title  TEXT,
			qa_number INT not null,
			q_choice TEXT not null
		)`

	case "task_qa_relations":
		// 任务和题目关联表
		s = `create table if not exists task_qa_relations
		(
			task_id TEXT not null
				references tasks (task_id),
			qa_id   TEXT not null
				references task_data (qa_id),
			primary key (task_id, qa_id)
		)`
	case "student_task_answers":
		s = `create table if not exists student_task_answers (
		student_id TEXT not null, -- 学生id 暂时使用5位学号代替
		task_id TEXT not null,   -- 任务ID
		qa_id TEXT not null,     -- 问题ID
		answer TEXT not null,    -- 学生的答案
		UNIQUE (student_id, task_id, qa_id) -- 确保组合唯一，避免同一学生在同一任务中对同一问题重复作答
		)`
	case "task_time":
		s = `create table if not exists task_time
		(
			student_id       text not null,
			task_id          text not null,
			get_task_time    text not null,
			push_answer_time text not null,
			unique (student_id, task_id)
		)`
	}
	// 写入数据库
	_, err := db.Exec(s)
	return err
}

// CheckFieldValueExist 检查字段值是否存在
func CheckFieldValueExist(table string, field string, fieldValue string) bool {
	// 定义日志
	logger, _ := NewLogger()
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", table, field)
	var count int
	err := db.QueryRow(query, fieldValue).Scan(&count)
	if err != nil {
		logger.Error("查询字段失败 ", err)
		return false
	}

	if count != 0 {
		return true
	}

	return false

}

// GenerateTaskId 生成任务id
func GenerateTaskId(maxTries int) string {
	logger, _ := NewLogger()

	for maxTries > 0 {
		taskId := uuid.NewV4().String()

		if !(CheckFieldValueExist("tasks", "task_id", taskId)) {
			return taskId
		}
		logger.Error(fmt.Sprintf("分配任务id失败 尝试次数: %d 当前task_id: %s", setting.MaxTries-maxTries+1, taskId))

		maxTries--
	}

	return ""
}

// GenerateQaId 生成题目id
func GenerateQaId(maxTries int) string {
	logger, _ := NewLogger()

	for maxTries > 0 {
		QaId := uuid.NewV4().String()

		if !(CheckFieldValueExist("task_data", "qa_id", QaId)) {
			return QaId
		}
		logger.Error(fmt.Sprintf("分配任务id失败 尝试次数: %d 当前task_id: %s", setting.MaxTries-maxTries+1, QaId))

		maxTries--
	}

	return ""
}

// AddTask 添加任务
func AddTask(taskId, taskTitle, taskDescription, deadline string, answers []QAAnswer) (success bool, err error) {
	logger, _ := NewLogger()

	// 插入tasks数据库
	taskStmt, err := db.Prepare(`INSERT INTO tasks (task_id, task_title, task_description, publish_time, Deadline) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return false, err
	}
	defer func() {
		closeErr := taskStmt.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}()
	_, err = taskStmt.Exec(taskId, taskTitle, taskDescription, time.Now().Format("2006-01-02 15:04:05.000"), deadline)
	if err != nil {
		return false, err
	}
	if !CheckFieldValueExist("tasks", "task_id", taskId) {
		return false, fmt.Errorf("添加任务失败")
	}
	logger.Info("添加任务成功")

	dataStmt, err := db.Prepare(`INSERT INTO task_data (qa_id, q_title, qa_number, q_choice) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return false, err
	}
	defer func() {
		closeErr := dataStmt.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}()

	relationStmt, err := db.Prepare(`INSERT INTO task_qa_relations (task_id, qa_id) VALUES (?, ?)`)
	if err != nil {
		return false, err
	}
	defer func() {
		closeErr := relationStmt.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}()

	for _, answer := range answers {
		// 在服务端为每个题目生成QaId(题目id)
		QaId := GenerateQaId(setting.MaxTries)

		// 插入task_data数据库
		_, err = dataStmt.Exec(QaId, fmt.Sprintf(answer.QaTitle), answer.QaNumber, answer.QaAnswer)
		if err != nil {
			return false, err
		}
		if !CheckFieldValueExist("task_data", "qa_id", QaId) {
			return false, fmt.Errorf("添加题目失败")
		}
		logger.Infof("添加题目数据成功")

		// 插入task_qa_relations关联表
		_, err = relationStmt.Exec(taskId, QaId)
		if err != nil {
			return false, err
		}
		if !CheckFieldValueExist("task_qa_relations", "qa_id", QaId) {
			return false, fmt.Errorf("添加题目关联失败")
		}
		logger.Infof("添加题目与任务关联成功")
	}

	return true, nil
}

// TaskInfo 任务信息 结构体
type TaskInfo struct {
	TaskTitle       string
	TaskDescription string
	PublishTime     string
	Deadline        string
}

// GetInfo 获取任务信息
func GetInfo(taskId string) (taskInfo *TaskInfo, err error) {
	logger, _ := NewLogger()

	// 初始化 TaskInfo 结构体实例
	taskInfo = &TaskInfo{}

	// 获取tasks中的数据
	stmt, err := db.Prepare(`SELECT task_title, task_description, publish_time, Deadline FROM tasks WHERE task_id = ?`)
	if err != nil {
		return nil, err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			logger.Error(err)
		}
	}(stmt)

	rows, err := stmt.Query(taskId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Error(err)
		}
	}(rows)

	if !rows.Next() {
		return nil, errors.New("找不到任务")
	}

	err = rows.Scan(&taskInfo.TaskTitle, &taskInfo.TaskDescription, &taskInfo.PublishTime, &taskInfo.Deadline)
	if err != nil {
		return nil, err
	}

	logger.Info("获取任务成功")
	return taskInfo, nil
}

// TeaTaskData 教师的任务数据
type TeaTaskData struct {
	QaId     string            `json:"qa_id"`
	QaTitle  string            `json:"q_title"`
	QaNumber int               `json:"qa_number"`
	QaChoice map[string]string `json:"q_choice"` // 存储问题的选项
}

// GetTaskData 获取学生任务数据
func GetTaskData(taskId string) (taskData *[]TeaTaskData, err error) {
	logger, _ := NewLogger()
	taskData = &[]TeaTaskData{}

	stmt, err := db.Prepare(`SELECT qa_id, q_title, q_choice, qa_number FROM task_data WHERE qa_id IN (SELECT qa_id FROM task_qa_relations WHERE task_id = ?)`)
	if err != nil {
		return nil, err
	}
	defer func(stmt *sql.Stmt) {
		closeErr := stmt.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}(stmt)

	rows, err := stmt.Query(taskId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		closeErr := rows.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}(rows)

	for rows.Next() {
		var qa TeaTaskData
		var qAnswer string

		err = rows.Scan(&qa.QaId, &qa.QaTitle, &qAnswer, &qa.QaNumber)
		if err != nil {
			return nil, err
		}

		// 构造默认的QaChoice
		qa.QaChoice = map[string]string{
			"A": "这是答题卡，填写答案即可。",
			"B": "这是答题卡，填写答案即可。",
			"C": "这是答题卡，填写答案即可。",
			"D": "这是答题卡，填写答案即可。",
		}

		*taskData = append(*taskData, qa)
		logger.Info("获取任务数据成功")
	}

	if len(*taskData) == 0 {
		return nil, errors.New("获取任务数据失败")
	}

	return taskData, nil
}

// StuTaskData 学生的任务数据
type StuTaskData struct {
	QaId      string `json:"qa_id"`
	QAnswer   string `json:"q_answer"`
	SpendTime string `json:"spend_time"`
}

// PushTaskData 更新任务数据
func PushTaskData(StudentId string, taskId string, taskData *[]StuTaskData) (success bool, err error) {
	logger, _ := NewLogger()
	// 写入student_task_answers数据库
	stmt, err := db.Prepare(`INSERT INTO student_task_answers (student_id, task_id, qa_id, answer) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return false, err
	}
	defer func(stmt *sql.Stmt) {
		closeErr := stmt.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}(stmt)
	for _, data := range *taskData {
		_, err = stmt.Exec(StudentId, taskId, data.QaId, data.QAnswer)
		if err != nil {
			return false, err
		}
	}
	logger.Info("更新任务数据成功")
	return true, nil

}

// MarkGetTaskTime 写入获取任务的时间
func MarkGetTaskTime(StudentId string, taskId string) (success bool, err error) {
	logger, _ := NewLogger()

	// 检查是否已存在对应的记录
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM task_time WHERE student_id = ? AND task_id = ?`, StudentId, taskId).Scan(&count)
	if err != nil {
		return false, err
	}

	// 如果记录已存在，则返回成功并忽略插入操作
	if count > 0 {
		logger.Info("任务时间已存在，无需更新")
		return true, nil
	}

	// 准备插入语句
	timeStmt, err := db.Prepare(`INSERT INTO task_time (student_id, task_id, get_task_time ,push_answer_time) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return false, err
	}
	defer func() {
		closeErr := timeStmt.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}()

	// 执行插入操作
	_, err = timeStmt.Exec(StudentId, taskId, time.Now().Format("2006-01-02 15:04:05.000"), "")
	if err != nil {
		return false, err
	}

	logger.Info("更新任务时间成功")
	return true, nil
}

// PushAnswerTime 学生答题时间
func PushAnswerTime(StudentId string, taskId string, finishedTime string) (success bool, err error) {
	logger, _ := NewLogger()

	// 更新学生答题时间
	timeStmt, err := db.Prepare(`UPDATE task_time SET push_answer_time = ? WHERE student_id = ? AND task_id = ?`)
	if err != nil {
		return false, err
	}
	defer func() {
		closeErr := timeStmt.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}()
	_, err = timeStmt.Exec(finishedTime, StudentId, taskId)
	if err != nil {
		return false, err
	}
	logger.Info("更新学生答题时间成功")
	return true, nil
}

// TaskData 报告数据结构体
type TaskData struct {
	QaID      string `json:"qa_id"`
	QaNumber  int    `json:"qa_number"`
	TeaAnswer string `json:"tea_answer"`
	StuAnswer string `json:"stu_answer"`
}

// StuTaskReport 学生任务报告数据请求结构体
type StuTaskReport struct {
	TaskTitle  string
	FinishTime string
	SpendTime  string
	TaskData   []TaskData
}

func GetReportData(StudentId string, taskId string) (*StuTaskReport, error) {
	logger, _ := NewLogger()
	// 获取任务信息的任务标题
	info, err := GetInfo(taskId)
	if err != nil {
		return nil, err
	}

	// 初始化报告数据，并填充 TaskTitle
	report := &StuTaskReport{
		TaskTitle: info.TaskTitle,
	}

	spendTime := ""
	finishTime := ""

	// 从task_time获取学生答题时间
	reportStmt, err := db.Prepare(`SELECT push_answer_time, get_task_time FROM task_time WHERE student_id = ? AND task_id = ?`)
	defer func() {
		closeErr := reportStmt.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}()
	defer func() {
		closeErr := reportStmt.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}()
	rows, err := reportStmt.Query(StudentId, taskId)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}()
	for rows.Next() {
		var pushAnswerTime string
		var getTaskTime string
		err = rows.Scan(&pushAnswerTime, &getTaskTime)
		if err != nil {
			return nil, err
		}
		// 获取学生答题时间
		if pushAnswerTime != "" {
			finishTime = pushAnswerTime
		}
		// 获取学生获取任务时间
		if getTaskTime != "" {
			spendTime = GetSpendTimeInSeconds(getTaskTime, pushAnswerTime)
		}
	}

	// 根据task_id和qa_id关联，从tasks、task_qa_relations和task_data表中获取正确答案q_choice
	qaRelationStmt, err := db.Prepare(`SELECT td.qa_id, td.q_choice, td.qa_number FROM task_data td INNER JOIN task_qa_relations tqr ON td.qa_id = tqr.qa_id WHERE tqr.task_id = ?`)
	if err != nil {
		return nil, err
	}
	defer func(qaRelationStmt *sql.Stmt) {
		err := qaRelationStmt.Close()
		if err != nil {

		}
	}(qaRelationStmt)

	rowsQARelation, err := qaRelationStmt.Query(taskId)
	if err != nil {
		return nil, err
	}
	defer func(rowsQARelation *sql.Rows) {
		err := rowsQARelation.Close()
		if err != nil {

		}
	}(rowsQARelation)

	// 根据student_id和task_id从student_task_answers取出学生答题内容
	stuAnswerStmt, err := db.Prepare(`SELECT qa_id, answer FROM student_task_answers WHERE student_id = ? AND task_id = ?`)
	if err != nil {
		return nil, err
	}
	defer func(stuAnswerStmt *sql.Stmt) {
		err := stuAnswerStmt.Close()
		if err != nil {

		}
	}(stuAnswerStmt)

	rowsStuAnswer, err := stuAnswerStmt.Query(StudentId, taskId)
	if err != nil {
		return nil, err
	}
	defer func(rowsStuAnswer *sql.Rows) {
		err := rowsStuAnswer.Close()
		if err != nil {

		}
	}(rowsStuAnswer)

	var taskDataList []TaskData
	qaMap := make(map[string]QAChoice)

	for rowsQARelation.Next() {
		var qaID string
		var qChoice string
		var qNumber int
		err = rowsQARelation.Scan(&qaID, &qChoice, &qNumber)
		if err != nil {
			return nil, err
		}
		qaChoice := QAChoice{qChoice, qNumber}
		qaMap[qaID] = qaChoice
	}

	for rowsStuAnswer.Next() {
		var qaID string
		var stuAnswer string
		err = rowsStuAnswer.Scan(&qaID, &stuAnswer)
		if err != nil {
			return nil, err
		}
		teaAnswer, ok := qaMap[qaID]
		if !ok {
			continue // 如果找不到对应的题目答案，则跳过
		}
		taskData := TaskData{
			QaID:      qaID,
			QaNumber:  teaAnswer.QNumber,
			TeaAnswer: teaAnswer.Answer,
			StuAnswer: stuAnswer,
		}
		taskDataList = append(taskDataList, taskData)
	}

	report.SpendTime = spendTime
	report.FinishTime = finishTime
	report.TaskData = taskDataList

	reports := *report
	return &reports, nil
}

// GetSpendTimeInSeconds 获得时间差
func GetSpendTimeInSeconds(getTaskTime, pushAnswerTime string) string {
	// 获取时间差
	t1, _ := time.Parse("2006-01-02 15:04:05.000", getTaskTime)
	t2, _ := time.Parse("2006-01-02 15:04:05.000", pushAnswerTime)

	diff := t2.Sub(t1)
	// 将时间差转换为秒并格式化为整数字符串
	secondsDiff := int(diff.Seconds())
	return strconv.Itoa(secondsDiff)
}

// QAChoice 单个题目及其答案选项结构体
type QAChoice struct {
	Answer  string `json:"answer"`
	QNumber int    `json:"question_number"`
}

// AnswerItem 定义AnswerItem结构体
type AnswerItem struct {
	QaID     string `json:"qa_id"`
	QaNumber int    `json:"qa_number"`
	Answer   string `json:"answer"`
}

// StudentAnswer 定义StudentAnswer结构体
type StudentAnswer struct {
	UserID  string       `json:"user_id"`
	Answers []AnswerItem `json:"answers"`
}

// StatusTaskData 定义TaskData结构体
type StatusTaskData struct {
	TaskTitle     string          `json:"task_title"`
	CorrectAnswer []AnswerItem    `json:"correctAnswer"`
	StudentAnswer []StudentAnswer `json:"studentAnswer"`
}

// GetStatusReportData 获取学生任务状态报告数据
func GetStatusReportData(taskId string) (*StatusTaskData, error) {
	logger, _ := NewLogger()

	// 获取任务信息的任务标题
	info, err := GetInfo(taskId)
	if err != nil {
		return nil, err
	}
	// 初始化报告数据，并填充 TaskTitle
	data := &StatusTaskData{
		TaskTitle: info.TaskTitle,
	}

	// 获取正确答案
	qaRelationStmt, err := db.Prepare(`SELECT td.qa_id, td.q_choice, td.qa_number FROM task_data td INNER JOIN task_qa_relations tqr ON td.qa_id = tqr.qa_id WHERE tqr.task_id = ?`)
	defer func(qaRelationStmt *sql.Stmt) {
		err := qaRelationStmt.Close()
		if err != nil {

		}
	}(qaRelationStmt)
	defer func() {
		closeErr := qaRelationStmt.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}()
	rowsQARelation, err := qaRelationStmt.Query(taskId)
	defer func(rowsQARelation *sql.Rows) {
		err := rowsQARelation.Close()
		if err != nil {

		}
	}(rowsQARelation)
	defer func() {
		closeErr := rowsQARelation.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}()
	for rowsQARelation.Next() {
		var qaID string
		var qChoice string
		var qNumber int
		err = rowsQARelation.Scan(&qaID, &qChoice, &qNumber)
		if err != nil {
			return nil, err
		}
		data.CorrectAnswer = append(data.CorrectAnswer, AnswerItem{qaID, qNumber, qChoice})
	}

	// 在student_task_answers通过task_id获取所有学生针对此任务的答题内容，并整合到StatusTaskData结构体中
	stuAnswerStmt, err := db.Prepare(`SELECT student_id, qa_id, answer FROM student_task_answers WHERE task_id = ?`)
	if err != nil {
		return nil, fmt.Errorf("准备查询学生答题记录SQL语句时出错: %v", err)
	}
	defer func() {
		closeErr := stuAnswerStmt.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}()
	rowsStuAnswer, err := stuAnswerStmt.Query(taskId)
	if err != nil {
		return nil, fmt.Errorf("查询学生答题记录时出错: %v", err)
	}
	defer func(rowsStuAnswer *sql.Rows) {
		err := rowsStuAnswer.Close()
		if err != nil {

		}
	}(rowsStuAnswer)

	var studentAnswers []StudentAnswer
	for rowsStuAnswer.Next() {
		var studentID string
		var qaID string
		var stuAnswer string
		err = rowsStuAnswer.Scan(&studentID, &qaID, &stuAnswer)
		if err != nil {
			return nil, fmt.Errorf("扫描学生答题记录结果时出错: %v", err)
		}

		// 通过qa_id在task_data获取题目序号，这里假设每道题目对应的结果唯一
		qaRelationStmt, err := db.Prepare(`SELECT qa_number FROM task_data WHERE qa_id = ?`)
		if err != nil {
			return nil, fmt.Errorf("准备查询题目序号SQL语句时出错: %v", err)
		}
		defer func() {
			closeErr := qaRelationStmt.Close()
			if closeErr != nil {
				logger.Error(closeErr)
			}
		}()
		var qNumber int
		err = qaRelationStmt.QueryRow(qaID).Scan(&qNumber)
		if err != nil {
			return nil, fmt.Errorf("查询题目序号时出错: %v", err)
		}

		// 构建StudentAnswer结构体并添加到列表中
		studentAnswer := StudentAnswer{UserID: studentID, Answers: []AnswerItem{{QaID: qaID, QaNumber: qNumber, Answer: stuAnswer}}}
		studentAnswers = append(studentAnswers, studentAnswer)

		// 更新或创建StatusTaskData中的StudentAnswer字段（这里假设每个学生的答案会分组在一起）
		found := false
		for i, sa := range data.StudentAnswer {
			if sa.UserID == studentID {
				found = true
				data.StudentAnswer[i].Answers = append(data.StudentAnswer[i].Answers, AnswerItem{qaID, qNumber, stuAnswer})
				break
			}
		}
		if !found {
			data.StudentAnswer = append(data.StudentAnswer, studentAnswer)
		}
	}

	// 确保在处理完所有学生答案后返回数据
	return data, nil

}
