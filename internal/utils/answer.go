package utils

import (
	"encoding/json"
	"strings"
	"testogo/internal/model/entity"
)

// CheckAnswer checks if the user's answer is correct
func CheckAnswer(questionType entity.QuestionType, options, correctAnswer, userAnswer string) bool {
	// 去除前后空格
	correct := strings.TrimSpace(correctAnswer)
	user := strings.TrimSpace(userAnswer)

	// 对于选择题（choice），需要处理字母答案到选项内容的映射
	if questionType == entity.TypeChoice || questionType == entity.TypeMultiChoice || questionType == entity.TypeJudge {
		// 尝试解析选项JSON
		if options != "" {
			var optionsList []string
			if err := json.Unmarshal([]byte(options), &optionsList); err == nil && len(optionsList) > 0 {
				// 检查用户答案是否为字母形式（A、B、C、D等）
				if len(user) == 1 && user >= "A" && user <= "Z" {
					optionIndex := int(user[0] - 'A')
					if optionIndex >= 0 && optionIndex < len(optionsList) {
						// 将字母答案转换为对应的选项内容
						user = strings.TrimSpace(optionsList[optionIndex])
					}
				} else if len(user) == 1 && user >= "a" && user <= "z" {
					optionIndex := int(user[0] - 'a')
					if optionIndex >= 0 && optionIndex < len(optionsList) {
						// 将字母答案转换为对应的选项内容
						user = strings.TrimSpace(optionsList[optionIndex])
					}
				}
			}
		}
	}

	// 直接比较（保持原始大小写）
	if correct == user {
		return true
	}

	// 转为小写进行比较（兼容性）
	correctLower := strings.ToLower(correct)
	userLower := strings.ToLower(user)
	return correctLower == userLower
}